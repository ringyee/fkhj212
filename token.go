package fkhj212

import (
	"sync"
	"time"

	"github.com/yjiong/fkhj212/packets"
)

type packetAndToken struct {
	p packets.Pduer
	t tokenCompletor
}

// Token defines the interface for the tokens used to indicate when
// actions have completed.
type Token interface {
	Wait() bool
	WaitTimeout(time.Duration) bool
	Error() error
}

type TokenErrorSetter interface {
	setError(error)
}

type tokenCompletor interface {
	Token
	TokenErrorSetter
	flowComplete()
}

type baseToken struct {
	m                   sync.RWMutex
	retransmissionTimes uint8
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

	return false
}

func (b *baseToken) flowComplete() {
	select {
	case <-b.complete:
	default:
		close(b.complete)
	}
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

func newToken() tokenCompletor {
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
