package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/yjiong/fkhj212/clientapp"
	"github.com/yjiong/fkhj212/device"
	"github.com/yjiong/fkhj212/packets"
	"reflect"
	"strconv"
	"time"
)

func setLogLevel() error {
	log.SetLevel(log.Level(logLevel))
	return nil
}

func printStartMessage() error {
	log.WithFields(log.Fields{
		"version": version,
		"docs":    "http://www.jslcinfo.com",
	}).Info("starting fkhj212 process")
	return nil
}

func setPostgreSQLConnection() error {
	log.Info("connecting to postgresql")
	//db, err := storage.OpenDatabase(config.C.PostgreSQL.DSN)
	//if err != nil {
	//return errors.Wrap(err, "database connection error")
	//}
	//config.C.PostgreSQL.DB = db
	return nil
}

func setHjClient() error {
	ops := clientapp.NewClientOptions()
	clients = make([]clientapp.Fkhjer, 0)
	for _, ups := range clientapp.UpS.UpServers {
		port := strconv.Itoa(int(ups.Port))
		ops.SetTargetServer(ups.Address + ":" + port)
		ops.SetMN(ups.MN)
		c := clientapp.NewFkhj(ops)
		go connect(c)
		clients = append(clients, c)
	}
	<-time.After(2 * time.Second)
	if len(clients) == 0 {
		c := clientapp.NewFkhj(ops)
		go connect(c)
		clients = append(clients, c)
	}
	regDev()
	return nil
}

func chkRestart() error {
	go func() {
		for {
			<-clientapp.ReConfig
			for _, c := range clients {
				c.Disconnect()
			}
			initConfig()
			<-time.After(100 * time.Millisecond)
			setHjClient()
		}
	}()
	return nil
}

func regDev() error {
	for _, f := range clientapp.CF.Factors {
		log.Debugf("dev %s", f.PC)
		if d, err := device.Dev.GetMD(f); err == nil {
			devs = append(devs, d)
		} else {
			log.Error(err)
		}
	}
	return nil
}

func readDev() error {
	go func() {
		for {
			for _, dev := range devs {
				dev.GetCP()
			}
			<-time.After(2 * time.Second)
		}
	}()
	return nil
}

func autoUpload() error {
	for _, dc := range clients {
		go func(c clientapp.Fkhjer) {
			for {
				<-time.After(c.RtdInterval())
				for _, dev := range devs {
					c.UploadPdu(packets.SupRealTimeData, (dev.GetValue()))
				}
			}
		}(dc)
	}
	return nil
}

func connect(c clientapp.Fkhjer) {
	t := c.Connect()
	var exist bool
	for t.Wait() || t.Error() != nil {
		<-time.After(2 * time.Second)
		t = c.Connect()
		log.Infoln("retry......")
		exist = false
		for _, ec := range clients {
			if reflect.DeepEqual(ec, c) {
				exist = true
				break
			}
		}
		if !exist {
			return
		}
	}
}

var clients []clientapp.Fkhjer
var devs []*device.ModbusDev

//func setRedisPool() error {
//log.Info("setup redis connection pool")
//config.C.Redis.Pool = storage.NewRedisPool(config.C.Redis.URL)
//return nil
//}
