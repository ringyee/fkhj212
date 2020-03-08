package clientapp

import (
	"errors"
	"net"
	"net/url"
	"os"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/yjiong/fkhj212/packets"
	"github.com/yjiong/iotgateway/serial"
	"golang.org/x/net/proxy"
)

// Dialer is connection
type Dialer interface {
	Dial(network, addr string) (c net.Conn, err error)
	SerialDial(serPort string) (serial.Port, error)
}

func signalError(c chan<- error, err error) {
	select {
	case c <- err:
	default:
	}
}

func openConnection(uri *url.URL, timeout time.Duration) (net.Conn, error) {
	switch uri.Scheme {
	case "tcp":
		allProxy := os.Getenv("all_proxy")
		if len(allProxy) == 0 {
			conn, err := net.DialTimeout("tcp", uri.Host, timeout)
			if err != nil {
				return nil, err
			}
			log.Debugln(uri.Host, "OnConnect ...")
			return conn, nil
		}
		proxyDialer := proxy.FromEnvironment()
		conn, err := proxyDialer.Dial("tcp", uri.Host)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
	return nil, errors.New("Unknown protocol")
}

func receivePduing(c *client) {
	var err error
	var pdue packets.Pduer
	defer c.workers.Done()
	log.Infoln("receivePduing started")
	for {
		if pdue, err = packets.ReadPduer(c.conn); err != nil {
			log.Errorln("in receivePduing", err)
			break
		}
		log.Debugf("Received in receivePduing get pdu %T ok !", pdue)
		if pdue.GetMN() != c.options.MN {
			continue
		}
		select {
		case c.ibound <- pdue:
			if c.options.KeepAlive != 0 {
				c.lastReceived.Store(time.Now())
			}
		case <-c.stop:
			log.Warning("receivePdu dropped a received message during shutdown")
			break
		}
	}
	select {
	case <-c.stop:
		log.Infoln("receivePdu stopped")
		return
	default:
		log.Errorln("receivePdu stopped with error", err)
		signalError(c.errors, err)
		return
	}
}

func sendPduing(c *client) {
	defer c.workers.Done()
	log.Infoln("sendPduing started")
	for {
		log.Infoln("sendPduing waiting for an outbound message")
		select {
		case <-c.stop:
			log.Infoln("sendPduing stopped")
			return
		case opdu := <-c.obound:
			log.Printf("sendPduing out pdu id = %s\n", opdu.p.GetID())
			if c.options.OverTime > 0 {
				c.conn.SetWriteDeadline(time.Now().Add(c.options.OverTime))
			}
			if _, err := opdu.p.Writeto(c.conn); err != nil {
				log.Errorln("sendPduing stopped with error", err)
				opdu.t.setError(err)
				signalError(c.errors, err)
				return
			}

			if c.options.OverTime > 0 {
				c.conn.SetWriteDeadline(time.Time{})
			}
		}
		if c.options.KeepAlive != 0 {
			c.lastSent.Store(time.Now())
		}
	}
}

// receive pdu on ibound
func handleRece(c *client) {
	defer c.workers.Done()
	log.Infoln("handleRece started")

	for {
		log.Infoln("handleRece waiting for msg on ibound")
		select {
		case iPdu := <-c.ibound:
			log.Infoln("handleRece got pdu on ibound")
			log.Debugf("ibound iPdu type = %+v ", iPdu)
			if iPdu.GetMN() != c.options.MN {
				continue
			}
			id := iPdu.GetID()
			if iPdu.NeedAck() {
				cpkv := packets.CPkv{"QnRtn": 1}
				if iPdu.GetPW() != c.options.PW {
					cpkv = packets.CPkv{"QnRtn": 3}
				}
				cpfield := packets.NewCPFildFromCPkvs(cpkv)
				cpfield.NoDTime = true
				ack := NewPdu(packets.Sresponse, c, *cpfield)
				ack.QN = iPdu.GetQN()
				c.upLoadHjPdu(ack)
				//if iPdu.GetPW() != c.options.PW {
				//continue
				//}
			}
			switch iPdu.GetCN() {
			case packets.HsetSlaveTime:
				log.Debugln("received HsetSlaveTime pdu id:", id)
				if atomic.LoadInt32(&c.pingOutstanding) > 0 {
					atomic.StoreInt32(&c.pingOutstanding, 0)
				}
			case packets.HdataACK:
				log.Debugln("received HdataACK pdu id:", id)
				tk := c.getToken(id)
				if tk != nil {
					tk.flowComplete()
					c.pduIDs.free(id)
				}
			case packets.HnoticeACK:
				log.Debugln("received HnoticeACK pdu id:", id)
				tk := c.getToken(id)
				if tk != nil {
					tk.flowComplete()
					c.pduIDs.free(id)
				}
			case packets.HGetConfig:
				log.Debugln("received HGetConfig pdu id:", id)
				c.handleGetConfig(iPdu)
			case packets.HSetConfig:
				log.Debugln("received HSetConfig pdu id:", id)
				c.handleSetConfig(iPdu)
			default:
				log.Debugln("received HSetConfig pdu id for default handle:", id)
				c.defalutHandle(iPdu)
			}
		case <-c.stop:
			log.Warnln("logic stopped")
			return
		}
	}
}

func errorWatch(c *client) {
	defer c.workers.Done()
	select {
	case <-c.stop:
		log.Warnln("errorWatch stopped")
		return
	case err := <-c.errors:
		log.Errorln("error triggered, stopping")
		go c.internalConnLost(err)
		return
	}
}
