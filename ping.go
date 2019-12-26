package fkhj212

/*
 * Copyright (c) 2013 IBM Corp.
 *
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v1.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v10.html
 *
 * Contributors:
 *    Seth Hoenig
 *    Allan Stockdill-Mander
 *    Mike Robertson
 */

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

	if c.options.KeepAlive > 10 {
		checkInterval = 5
	} else {
		checkInterval = c.options.KeepAlive / 2
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
					log.Debugln("keepalive sending ping")
					ping := packets.NewHjPdu(map[string]interface{}{"xxxx": "xxxx"}, packets.CPField{})
					//We don't want to wait behind large messages being sent, the Write call
					//will block until it it able to send the packet.
					atomic.StoreInt32(&c.pingOutstanding, 1)
					ping.Writeto(c.conn)
					c.lastSent.Store(time.Now())
					pingSent = time.Now()
				}
			}
			if atomic.LoadInt32(&c.pingOutstanding) > 0 && time.Since(pingSent) >= c.options.PingTimeout {
				log.Infoln("pingresp not received, disconnecting")
				c.errors <- errors.New("pingresp not received, disconnecting")
				return
			}
		}
	}
}
