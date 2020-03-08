package clientapp

import (
	"errors"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/yjiong/fkhj212/packets"
)

func keepalive(c *client) {
	defer c.workers.Done()
	log.Debugln("keepalive starting")
	var checkInterval int64
	var pingSent time.Time

	if c.options.KeepAlive > 300 {
		checkInterval = 300
	} else {
		checkInterval = c.options.KeepAlive
	}

	intervalTicker := time.NewTicker(time.Duration(checkInterval * int64(time.Second)))
	defer intervalTicker.Stop()

	for {
		select {
		case <-c.stop:
			log.Debugln("keepalive stopped")
			return
		case <-intervalTicker.C:
			lastSent := c.lastSent.Load().(time.Time)
			lastReceived := c.lastReceived.Load().(time.Time)

			log.Debugln("ping check", time.Since(lastSent).Seconds())
			if time.Since(lastSent) >= time.Duration(c.options.KeepAlive*int64(time.Second)) || time.Since(lastReceived) >= time.Duration(c.options.KeepAlive*int64(time.Second)) {
				if atomic.LoadInt32(&c.pingOutstanding) == 0 {
					log.Debugln("keepalive sending time-correct request")
					timeCorrect := NewPdu(packets.SsetTimeReq, c, packets.CPField{})
					atomic.StoreInt32(&c.pingOutstanding, 1)
					timeCorrect.Writeto(c.conn)
					c.lastSent.Store(time.Now())
					pingSent = time.Now()
				}
			}
			if atomic.LoadInt32(&c.pingOutstanding) > 0 && time.Since(pingSent) >= c.options.PingTimeout {
				log.Infoln("keepalive response not received, disconnecting")
				c.errors <- errors.New("keepalive response not received, disconnecting")
				return
			}
		}
	}
}
