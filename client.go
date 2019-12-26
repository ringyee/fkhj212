package fkhj212

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/yjiong/fkhj212/packets"
)

const (
	disconnected uint32 = iota
	connecting
	reconnecting
	connected
)

type FkhjClient interface {
	IsConnected() bool
	IsConnectionOpen() bool
	Connect() Token
	Disconnect(quiesce uint)
	UploadMsg(phj packets.Pduer) Token
}

// client implements the FkhjClient interface
type client struct {
	lastSent        atomic.Value
	lastReceived    atomic.Value
	pingOutstanding int32
	status          uint32
	sync.RWMutex
	msgID           string
	conn            net.Conn
	ibound          chan packets.Pduer
	obound          chan *packetAndToken
	stopRouter      chan bool
	incomingPubChan chan *packets.Pduer
	errors          chan error
	stop            chan struct{}
	options         ClientOptions
	optionsMu       sync.Mutex // Protects the options in a few limited cases where needed for testing
	workers         sync.WaitGroup
}

// NewFkhjClient will create
func NewFkhjClient(o *ClientOptions) FkhjClient {
	c := &client{}
	c.options = *o

	c.status = disconnected
	return c
}

// the client is connected or not.
func (c *client) IsConnected() bool {
	c.RLock()
	defer c.RUnlock()
	status := atomic.LoadUint32(&c.status)
	switch {
	case status == connected:
		return true
	case c.options.AutoReconnect && status > connecting:
		return true
	case c.options.ConnectRetry && status == connecting:
		return true
	default:
		return false
	}
}

// IsConnectionOpen return a bool signifying whether the client has an active
func (c *client) IsConnectionOpen() bool {
	c.RLock()
	defer c.RUnlock()
	status := atomic.LoadUint32(&c.status)
	switch {
	case status == connected:
		return true
	default:
		return false
	}
}

func (c *client) connectionStatus() uint32 {
	c.RLock()
	defer c.RUnlock()
	status := atomic.LoadUint32(&c.status)
	return status
}

func (c *client) setConnected(status uint32) {
	c.Lock()
	defer c.Unlock()
	atomic.StoreUint32(&c.status, uint32(status))
}

//ErrNotConnected is the error returned from function calls that are
//made when the client is not connected
var ErrNotConnected = errors.New("Not Connected")

// Connect will create a connection to the message broker, by default
func (c *client) Connect() Token {
	t := newToken()
	var err error
	if c.options.ConnectRetry && atomic.LoadUint32(&c.status) != disconnected {
		// if in any state other than disconnected and ConnectRetry is
		// enabled then the connection will come up automatically
		// client can assume connection is up
		log.Warnln("Connect() called but not disconnected")
		t.flowComplete()
		return t
	}

	c.obound = make(chan *packetAndToken)
	c.ibound = make(chan packets.Pduer)

	go func() {
		c.errors = make(chan error, 1)
		c.stop = make(chan struct{})
		if len(c.options.Servers) == 0 {
			t.setError(fmt.Errorf("No servers defined to connect to"))
			return
		}
		c.Lock()
		c.conn, err = openConnection(c.options.Servers[0], c.options.ConnectTimeout)
		c.Unlock()
		if err == nil {
		} else {
			log.Errorln("failed connect to server", err.Error())
			return
		}

		if c.options.ConnectRetry {
			c.reserveStoredPublishIDs() // Reserve IDs to allow publish before connect complete
		}

		if c.options.KeepAlive != 0 {
			atomic.StoreInt32(&c.pingOutstanding, 0)
			c.lastReceived.Store(time.Now())
			c.lastSent.Store(time.Now())
			c.workers.Add(1)
			go keepalive(c)
		}

		c.incomingPubChan = make(chan *packets.Pduer)

		c.setConnected(connected)
		log.Debugln("client is connected")
		if c.options.OnConnect != nil {
			go c.options.OnConnect(c)
		}

		c.workers.Add(4)
		go errorWatch(c)
		go alllogic(c)
		go sendPduing(c)
		go receivePduing(c)

		log.Debugln("exit startFkhjClient")
		t.flowComplete()
	}()
	return t
}

// Disconnect will end the connection with the server, but not before waiting
// the specified number of milliseconds to wait for existing work to be
// completed.
func (c *client) Disconnect(quiesce uint) {
	status := atomic.LoadUint32(&c.status)
	if status == connected {
		log.Debugln("disconnecting")
		c.setConnected(disconnected)

		c.disconnect()
	}
}

// ForceDisconnect will end the connection with the mqtt broker immediately.
func (c *client) forceDisconnect() {
	if !c.IsConnected() {
		log.Warnln("already disconnected")
		return
	}
	c.setConnected(disconnected)
	c.conn.Close()
	log.Debugln("forcefully disconnecting")
	c.disconnect()
}

func (c *client) internalConnLost(err error) {
	// Only do anything if this was called and we are still "connected"
	// forceDisconnect can cause incoming/outgoing/alllogic to end with
	// error from closing the socket but state will be "disconnected"
	if c.IsConnected() {
		c.closeStop()
		c.conn.Close()
		c.workers.Wait()
		if c.options.CleanSession && !c.options.AutoReconnect {
			//c.messageIds.cleanUp()
		}
		if c.options.AutoReconnect {
			c.setConnected(reconnecting)
			//go c.reconnect()
		} else {
			c.setConnected(disconnected)
		}
		if c.options.OnConnectionLost != nil {
			go c.options.OnConnectionLost(c, err)
		}
	}
}

func (c *client) closeStop() {
	c.Lock()
	defer c.Unlock()
	select {
	case <-c.stop:
		log.Debugln("In disconnect and stop channel is already closed")
	default:
		if c.stop != nil {
			close(c.stop)
		}
	}
}

func (c *client) closeStopRouter() {
	c.Lock()
	defer c.Unlock()
	select {
	case <-c.stopRouter:
		log.Debugln("In disconnect and stop channel is already closed")
	default:
		if c.stopRouter != nil {
			close(c.stopRouter)
		}
	}
}

func (c *client) closeConn() {
	c.Lock()
	defer c.Unlock()
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *client) disconnect() {
	c.closeStop()
	c.closeConn()
	c.workers.Wait()
	//c.messageIds.cleanUp()
	c.closeStopRouter()
	log.Debugln("disconnected")
}

// Publish will publish a message with the specified QoS and content
// to the specified topic.
// Returns a token to track delivery of the message to the broker
func (c *client) UploadMsg(phj packets.Pduer) Token {
	token := newToken()
	log.Debugln("enter Publish")
	switch {
	case !c.IsConnected():
		token.setError(ErrNotConnected)
		return token
	case c.connectionStatus() == reconnecting:
		token.flowComplete()
		return token
	default:
		waitTimeout := c.options.WriteTimeout
		if waitTimeout == 0 {
			waitTimeout = time.Second * 30
			pub := &packetAndToken{
				p: phj,
				t: token,
			}
			select {
			case c.obound <- pub:
			case <-time.After(waitTimeout):
				token.setError(errors.New("send msg by timeout"))
			}
		}
	}
	return token
}

// reserveStoredPublishIDs reserves the ids for publish packets in the persistant store to ensure these are not duplicated
func (c *client) reserveStoredPublishIDs() {
	// The resume function sets the stored id for publish packets only (some other packets
	// will get new ids in net code). This means that the only keys we need to ensure are
	// unique are the publish ones (and these will completed/replaced in resume() )
	if !c.options.CleanSession {
	}
}

// Unsubscribe will end the subscription from each of the topics provided.
// Messages published to those topics from other clients will no longer be
// received.
// OptionsReader returns a FkhjClientOptionsReader which is a copy of the clientoptions
// in use by the client.
func (c *client) OptionsReader() ClientOptionsReader {
	r := ClientOptionsReader{options: &c.options}
	return r
}

//DefaultConnectionLostHandler is a definition of a function that simply
//reports to the DEBUG log the reason for the client losing a connection.
func DefaultConnectionLostHandler(client FkhjClient, reason error) {
	log.Debugln("Connection lost:", reason.Error())
}

func chkerr(e error) {
	if e != nil {
		panic(e)
	}
}
