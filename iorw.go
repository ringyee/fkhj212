package fkhj212

import (
	"errors"
	"net"
	"net/url"
	"os"
	//"reflect"
	//"sync/atomic"
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
	var hjcp packets.Pduer
	defer c.workers.Done()
	for {
		if hjcp, err = packets.ReadPduer(c.conn); err != nil {
			log.Errorln("in receivePduing", err)
			break
		}
		log.Infof("Received hjpdu :%+v", hjcp)
		select {
		case c.ibound <- hjcp:
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
		log.Println("receivePdu stopped")
		return
	default:
		log.Errorln("receivePdu stopped with error", err)
		signalError(c.errors, err)
		return
	}
}

func sendPduing(c *client) {
	defer c.workers.Done()
	log.Println("sendPduing started")

	for {
		log.Println("sendPduing waiting for an outbound message")
		select {
		case <-c.stop:
			log.Println("sendPduing stopped")
			return
		case pub := <-c.obound:
			msg := pub

			if c.options.WriteTimeout > 0 {
				c.conn.SetWriteDeadline(time.Now().Add(c.options.WriteTimeout))
			}

			if _, err := msg.p.Writeto(c.conn); err != nil {
				log.Errorln("sendPduing stopped with error", err)
				pub.t.setError(err)
				signalError(c.errors, err)
				return
			}

			if c.options.WriteTimeout > 0 {
				// If we successfully wrote, we don't want the timeout to happen during an idle period
				// so we reset it to infinite.
				c.conn.SetWriteDeadline(time.Time{})
			}

			pub.t.flowComplete()
			//log.Println("obound wrote msg, id:", msg.MessageID)
		}
		// Reset ping timer after sending control packet.
		if c.options.KeepAlive != 0 {
			c.lastSent.Store(time.Now())
		}
	}
}

// receive pdu on ibound
func alllogic(c *client) {
	defer c.workers.Done()
	log.Println("logic started")

	for {
		log.Println("logic waiting for msg on ibound")

		select {
		case msg := <-c.ibound:
			log.Println("logic got pdu on ibound")
			switch m := msg.(type) {
			//case *packets.PingrespPacket:
			//log.Println("received pingresp")
			//atomic.StoreInt32(&c.pingOutstanding, 0)
			case *packets.HjPdu:
				log.Debugln("received hjpdu id:", m.GetID())
				//token := c.getToken(m.MessageID)
				//switch t := token.(type) {
				//case *SubscribeToken:
				//log.Println("granted qoss", m.ReturnCodes)
				//for i, qos := range m.ReturnCodes {
				//t.subResult[t.subs[i]] = qos
				//}
				//}
				//token.flowComplete()
				//c.freeID(m.MessageID)
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
