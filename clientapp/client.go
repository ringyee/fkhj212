package clientapp

import (
	"errors"
	"fmt"
	"net"
	"strings"
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

// client implements the Fkhj interface
type client struct {
	lastSent        atomic.Value
	lastReceived    atomic.Value
	pingOutstanding int32
	status          uint32
	sync.RWMutex
	pduIDs
	conn      net.Conn
	ibound    chan packets.Pduer
	obound    chan *packetAndToken
	errors    chan error
	stop      chan struct{}
	options   ClientOptions
	optionsMu sync.Mutex // Protects the options in a few limited cases where needed for testing
	workers   sync.WaitGroup
	Pt        Persister
	pbuf      []string
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

// Connect will create a connection to the UpServer by default
func (c *client) Connect() Token {
	t := newToken()
	var err error
	if c.options.ConnectRetry && atomic.LoadUint32(&c.status) != disconnected {
		log.Warnln("Connect() called but not disconnected")
		t.flowComplete()
		return t
	}

	c.obound = make(chan *packetAndToken)
	c.ibound = make(chan packets.Pduer)
	c.pduIDs.index = make(map[string]tokenCompleter)

	go func() {
		c.errors = make(chan error, 1)
		c.stop = make(chan struct{})
		if c.options.Server == nil {
			t.setError(fmt.Errorf("No servers defined to connect to"))
			return
		}
		c.Lock()
		c.conn, err = openConnection(c.options.Server, c.options.ConnectTimeout)
		c.Unlock()
		if err == nil {
			log.Infof("socket connected to %s", c.options.Server)
		} else {
			log.Errorln("failed connect to server", err.Error())
			t.setError(err)
			return
		}

		//if c.options.ConnectRetry {
		//c.reserveStoredIDs() // Reserve IDs to allow before connect complete
		//}

		if c.options.KeepAlive != 0 {
			atomic.StoreInt32(&c.pingOutstanding, 0)
			c.lastReceived.Store(time.Now())
			c.lastSent.Store(time.Now())
			c.workers.Add(1)
			go keepalive(c)
		}

		c.setConnected(connected)
		log.Debugln("client is connected")
		if c.options.OnConnect != nil {
			go c.options.OnConnect(c)
		}

		c.workers.Add(4)
		go errorWatch(c)
		go handleRece(c)
		go sendPduing(c)
		go receivePduing(c)
	}()
	<-time.After(10 * time.Millisecond)
	return t
}

func (c *client) reconnect() {
	var (
		err   error
		rc    = byte(1)
		sleep = time.Duration(1 * time.Second)
	)

	for rc != 0 && atomic.LoadUint32(&c.status) != disconnected {
		if nil != c.options.OnReconnecting {
			c.options.OnReconnecting(c, &c.options)
		}
		c.Lock()
		c.conn, err = openConnection(c.options.Server, c.options.ConnectTimeout)
		c.Unlock()
		if err == nil {
			log.Infoln("socket connected to UpServer")
			rc = 0
		} else {
			log.Errorln("failed connect to server", err.Error())
		}
		if rc != 0 {
			log.Infoln("Reconnect failed, sleeping for", int(sleep.Seconds()), "seconds")
			time.Sleep(sleep)
			if sleep < c.options.MaxReconnectInterval {
				sleep *= 2
			}
			if sleep > c.options.MaxReconnectInterval {
				sleep = c.options.MaxReconnectInterval
			}
		}
	}
	// Disconnect() must have been called while we were trying to reconnect.
	if c.connectionStatus() == disconnected {
		log.Infoln("Client moved to disconnected state while reconnecting, abandoning reconnect")
		return
	}

	c.stop = make(chan struct{})

	if c.options.KeepAlive != 0 {
		atomic.StoreInt32(&c.pingOutstanding, 0)
		c.lastReceived.Store(time.Now())
		c.lastSent.Store(time.Now())
		c.workers.Add(1)
		go keepalive(c)
	}

	c.setConnected(connected)
	log.Infoln("client is reconnected")
	if c.options.OnConnect != nil {
		go c.options.OnConnect(c)
	}

	c.workers.Add(4)
	go errorWatch(c)
	go handleRece(c)
	go sendPduing(c)
	go receivePduing(c)
}

// Disconnect will end the connection with the server, but not before waiting
// the specified number of milliseconds to wait for existing work to be
// completed.
func (c *client) Disconnect() {
	status := atomic.LoadUint32(&c.status)
	if status == connected {
		log.Debugln("disconnecting")
		c.setConnected(disconnected)
		c.disconnect()
	}
}

// ForceDisconnect will end the connection immediately.
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
	if c.IsConnected() {
		c.closeStop()
		c.conn.Close()
		c.workers.Wait()
		if c.options.CleanSession && !c.options.AutoReconnect {
			c.pduIDs.cleanUp()
		}
		if c.options.AutoReconnect {
			c.setConnected(reconnecting)
			go c.reconnect()
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
	c.pduIDs.cleanUp()
	log.Debugln("disconnected")
}

//UploadPdu send up load real time hjpdu to target
func (c *client) UploadPdu(cntype packets.CNType, ucp packets.CPField) Token {
	rphj := NewPdu(packets.SupRealTimeData, c, ucp)
	rphj.CN = cntype
	return c.upLoadHjPdu(rphj)
}

func (c *client) upLoadHjPdu(rphj *packets.HjPdu) Token {
	token := newToken()
	log.Debugln("send uploadRealTimedata")
	switch {
	case !c.IsConnected():
		token.setError(ErrNotConnected)
		return token
	case c.connectionStatus() == reconnecting:
		token.flowComplete()
		return token
	default:
		updut := &packetAndToken{
			p: rphj,
			t: token,
		}
		c.pduIDs.put(updut.p.GetID(), updut.t)
		go c.sendUpPdu(updut)
	}
	//make QN only one
	<-time.After(1 * time.Millisecond)
	if rphj.Flag&0x1 != 0x1 {
		token.flowComplete()
	}
	return token
}

func (c *client) sendUpPdu(pt *packetAndToken) {
	waitTimeout := c.options.OverTime
	if waitTimeout == 0 {
		waitTimeout = time.Second * 30
	}
	for pt.t.getReCounts() < c.options.ReCount {
		select {
		case c.obound <- pt:
			if t := pt.t.WaitTimeout(waitTimeout); t {
				log.Debugf("in c.obound channel %v", t)
				return
			}
		case <-time.After(waitTimeout):
			if pt.t.getReCounts() >= c.options.ReCount {
				pt.t.setError(errors.New("send pdu by timeout"))
				return
			}
		}
	}
	pt.t.setError(errors.New("send pdu by timeout"))
	c.pduIDs.free(pt.p.GetID())
	log.Debugf("%+v", c.pduIDs.index)
}

func (c *client) reserveStoredIDs() {
	if !c.options.CleanSession {
		//......
	}
}

//DefaultConnectionLostHandler is a definition of a function that simply
//reports to the DEBUG log the reason for the client losing a connection.
func DefaultConnectionLostHandler(client Fkhjer, reason error) {
	log.Debugln("Connection lost:", reason.Error())
}

func chkerr(e error) {
	if e != nil {
		panic(e)
	}
}

func (c *client) handleGetConfig(gp packets.Pduer) {
	var b64str string
	var berr error
	for _, cp := range gp.GetCPkvg() {
		if v, ok := cp["R"].(string); ok {
			switch v {
			case "1":
				b64str, berr = c.Pt.If2B64(CF)
			case "2":
				b64str, berr = c.Pt.If2B64(UpS)
			}
			if berr == nil {
				cpkv := packets.CPkv{"R1": b64str}
				upcp := packets.NewCPFildFromCPkvs(cpkv)
				upp := NewPdu(packets.SPutConfig, c, *upcp)
				upp.QN = gp.GetQN()
				c.upLoadHjPdu(upp)
			}
		}
	}
}

func (c *client) handleSetConfig(gp packets.Pduer) {
	path := ConfPath
	var b64, js []byte
	var fname string
	var err error
	for _, cp := range gp.GetCPkvg() {
		if gps, ok := cp["W1"].(string); ok {
			if v := c.handlePbuf(gps, gp); v != "" {
				if b64, err = c.Pt.B642json(v); err == nil {
					fname = "factors"
					path += "factors.json"
					js = addPrefix(fname, b64)
				}
			}
		}
		if gps, ok := cp["W2"].(string); ok {
			if v := c.handlePbuf(gps, gp); v != "" {
				if b64, err = c.Pt.B642json(v); err == nil {
					fname = "upservers"
					path += "updataconf.json"
					js = addPrefix(fname, b64)
				}
			}
		}
		if len(js) > 0 && len(fname) > 0 {
			if err = c.Pt.JSONPut(path, js); err == nil {
				ReConfig <- struct{}{}
			}
		}
		if err != nil {
			log.Error(err)
		}
	}
}

func addPrefix(pfname string, b64 []byte) (js []byte) {
	enstr := string(b64)
	if strings.HasPrefix(enstr, "[") && strings.HasSuffix(enstr, "]") {
		enstr = `{"` + pfname + `":` + enstr + `}`
		js = []byte(enstr)
	}
	return
}

func (c *client) defalutHandle(gp packets.Pduer) {
	go c.options.ExCmdHandle(c, gp)
}

func (c *client) ResponseValue(iPdu packets.Pduer, kvs ...packets.CPkv) {
	cpfield := packets.NewCPFildFromCPkvg(kvs)
	cpfield.NoDTime = true
	rsp := NewPdu(packets.SexecResult, c, *cpfield)
	rsp.QN = iPdu.GetQN()
	c.upLoadHjPdu(rsp)
}

func (c *client) ResponseExeRtn(iPdu packets.Pduer, ern bool) {
	var v int
	if ern {
		v = 1
	}
	cpkv := packets.CPkv{"QnRtn": v}
	cpfield := packets.NewCPFildFromCPkvs(cpkv)
	cpfield.NoDTime = true
	rsp := NewPdu(packets.SexecResult, c, *cpfield)
	rsp.QN = iPdu.GetQN()
	c.upLoadHjPdu(rsp)
}

func (c *client) RtdInterval() time.Duration {
	return c.options.RtdInterval
}

func (c *client) handlePbuf(pbs string, gp packets.Pduer) string {
	if gp.GetFlag()&0x2 != 0x2 {
		return pbs
	}
	if gp.GetPNUM() > 0 && gp.GetPNO() == 1 && (gp.GetFlag()&0x2) == 0x2 {
		c.pbuf = make([]string, int(gp.GetPNUM()))
	}
	c.pbuf[int(gp.GetPNO()-1)] = pbs
	if gp.GetPNO() == gp.GetPNUM() {
		rs := ""
		for _, s := range c.pbuf {
			rs += s
		}
		return rs
	}
	return ""
}
