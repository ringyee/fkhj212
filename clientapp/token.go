package clientapp

import (
	"fmt"
	"sync"
	"time"

	"github.com/yjiong/fkhj212/packets"
)

type packetAndToken struct {
	p packets.Pduer
	t tokenCompleter
}

// Token defines the interface for the tokens used to indicate when
// actions have completed.
type Token interface {
	Wait() bool
	WaitTimeout(time.Duration) bool
	Error() error
}

// TokenErrorSetter ...
type TokenErrorSetter interface {
	setError(error)
}

type tokenCompleter interface {
	Token
	TokenErrorSetter
	flowComplete()
	getReCounts() int64
}

type baseToken struct {
	m                   sync.RWMutex
	retransmissionTimes int64
	complete            chan struct{}
	err                 error
}

func (b *baseToken) Wait() bool {
	<-b.complete
	return true
}

func (b *baseToken) WaitTimeout(d time.Duration) bool {
	timer := time.NewTimer(d)
	select {
	case <-b.complete:
		if !timer.Stop() {
			<-timer.C
		}
		return true
	case <-timer.C:
	}
	b.m.Lock()
	b.retransmissionTimes++
	b.m.Unlock()
	return false
}

func (b *baseToken) flowComplete() {
	select {
	case <-b.complete:
	default:
		close(b.complete)
	}
}

func (b *baseToken) getReCounts() int64 {
	b.m.RLock()
	defer b.m.RUnlock()
	return b.retransmissionTimes
}

func (b *baseToken) Error() error {
	b.m.RLock()
	defer b.m.RUnlock()
	return b.err
}

func (b *baseToken) setError(e error) {
	b.m.Lock()
	b.err = e
	b.flowComplete()
	b.m.Unlock()
}

func newToken() tokenCompleter {
	return &UploadMsgToken{baseToken: baseToken{complete: make(chan struct{})}}
}

// ConnectToken is an extension of Token containing the extra fields
// required to provide information about calls to Connect()
type ConnectToken struct {
	baseToken
	returnCode     byte
	sessionPresent bool
}

// ReturnCode returns the acknowledgement code in the connack sent
// in response to a Connect()
func (c *ConnectToken) ReturnCode() byte {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.returnCode
}

// SessionPresent returns a bool representing the value of the
// session present field in the connack sent in response to a Connect()
func (c *ConnectToken) SessionPresent() bool {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.sessionPresent
}

// UploadMsgToken ......
type UploadMsgToken struct {
	baseToken
	messageID uint16
}

// MessageID returns the  message ID that was assigned to the
// packet when it was sent to the data server
func (umt *UploadMsgToken) MessageID() uint16 {
	return umt.messageID
}

type pduIDs struct {
	sync.RWMutex
	index map[string]tokenCompleter
}

func (ps *pduIDs) cleanUp() {
	ps.Lock()
	for _, token := range ps.index {
		token.setError(fmt.Errorf("Connection lost before send Upload pdu completed"))
		token.flowComplete()
	}
	ps.index = make(map[string]tokenCompleter)
	ps.Unlock()
}

func (ps *pduIDs) free(id string) {
	ps.Lock()
	if _, ok := ps.index[id]; ok {
		delete(ps.index, id)
	}
	ps.Unlock()
}

func (ps *pduIDs) claimID(token tokenCompleter, id string) {
	ps.Lock()
	defer ps.Unlock()
	if _, ok := ps.index[id]; !ok {
		ps.index[id] = token
	} else {
		old := ps.index[id]
		old.flowComplete()
		ps.index[id] = token
	}
}

func (ps *pduIDs) put(id string, t tokenCompleter) {
	ps.Lock()
	defer ps.Unlock()
	ps.index[id] = t
}

func (ps *pduIDs) getToken(id string) tokenCompleter {
	ps.RLock()
	defer ps.RUnlock()
	if token, ok := ps.index[id]; ok {
		return token
	}
	return nil
}
