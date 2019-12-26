package fkhj212

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"os/signal"
	"syscall"
	"testing"
)

func TestDirve(t *testing.T) {
	Convey("==================测试=====================\n", t, func() {
		log.SetLevel(log.DebugLevel)
		opt := NewClientOptions()
		opt.AddTargetServer(":9898")
		fc := NewFkhjClient(opt)
		fc.Connect()

		sinChan := make(chan os.Signal)
		signal.Notify(sinChan, os.Interrupt, syscall.SIGTERM)
		select {
		case <-sinChan:
			fmt.Println("exiting")
		}

	})
}
