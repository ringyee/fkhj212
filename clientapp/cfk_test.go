package clientapp

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
	//"github.com/yjiong/fkhj212/packets"
)

func TestDirve(t *testing.T) {
	Convey("==================测试=====================\n", t, func() {
		log.SetLevel(log.DebugLevel)
		//viper.SetConfigName("config")
		//viper.SetConfigType("yml")
		viper.SetConfigFile(".")
		viper.AddConfigPath("$HOME/GOPATH/src/github.com/yjiong/fkhj212/clientapp")
		//if err := viper.ReadInConfig(); err == nil {
		//v := viper.AllSettings()
		//log.Debugln(v)
		//}
		opt := NewClientOptions()
		//opt.SetTargetServer("58.215.28.118:8901")
		opt.SetTargetServer(":8901")
		//sen := NewSensors()
		//cf, err := sen["m1"].GetCP(opt.CPKcID)
		//ckmap, err := sen["m1"].GetChkMap(opt.CPKcID)
		//log.Debugf("ckmap = %+v,%v", ckmap, err)

		fc := NewFkhjClient(opt)
		fc.Connect()
		//////////////////////////////////////////
		/*
						rtd := packets.CPkv{
							"006-Rtd": 0.77,
							"006-ID":  0,
						}
						rtd1 := packets.CPkv{
							"006-Rtd": 0.48,
							"006-ID":  1,
						}
						rtd2 := packets.CPkv{
							"007-Rtd": 35,
							"007-ID":  0,
						}
						rtd3 := packets.CPkv{
							"008-Rtd": 36,
			    with open("/home/yjiong/ipaddr", "r") as f:
			        fip = f.read()
			        fips = fip.strip('\n')
			    ipaddr = requests.get("http://ipv4.icanhazip.com")
							"008-ID":  0,
						}
						rtd4 := packets.CPkv{
							"012-Rtd": 0.00,
							"012-ID":  1,
						}
						rtd5 := packets.CPkv{
							"011-Rtd": 0,
							"011-ID":  1,
						}
						rtd6 := packets.CPkv{
							"012-Rtd": 0.00,
							"012-ID":  2,
						}
						rtd7 := packets.CPkv{
							"011-Rtd": 0,
							"011-ID":  2,
						}
						rtd8 := packets.CPkv{
							"009-Rtd": 0,
							"009-ID":  0,
						}
						rtd9 := packets.CPkv{
							"010-Rtd": 0,
							"010-ID":  0,
						}
						cpf := packets.NewCPFildFromCPkvs(time.Now(), rtd, rtd1, rtd2, rtd3, rtd4,
							rtd5, rtd6, rtd7, rtd8, rtd9)
						<-time.After(1 * time.Second)
						<-time.NewTimer(1 * time.Second).C
						t := fc.UploadRealTimedata(*cpf)
						t1 := fc.UploadRealTimedata(*cpf)
						t.Wait()
						t1.Wait()
						log.Debugln(t.Error())
						log.Debugln(t1.Error())
		*/
		<-time.After(3 * time.Second)
		//t := fc.UploadPdu(packets.SupRealTimeData, *cf)
		//t.Wait()
		//log.Debugln(t.Error())
		sinChan := make(chan os.Signal)
		signal.Notify(sinChan, os.Interrupt, syscall.SIGTERM)
		select {
		case <-sinChan:
			fmt.Println("exiting")
		}
	})
}
