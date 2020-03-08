package device

import (
	"testing"

	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/yjiong/fkhj212/clientapp"
)

func TestDirve(t *testing.T) {
	Convey("==================测试=====================\n", t, func() {
		log.SetLevel(log.DebugLevel)
		vr := mustGetRule()
		So(vr.vlength, ShouldEqual, 4)
		v := clientapp.GetConf()
		v.SetConfigType("json")
		v.SetConfigName("factors")
		v.ReadInConfig()
		v.Unmarshal(&clientapp.CF)
		So(len(clientapp.CF.Factors), ShouldBeGreaterThan, 0)
		m, _ := DevMap["Modbus"].NewSensor(clientapp.CF.Factors[0])
		log.Debugf("%#v", m)
		So(m, ShouldNotBeNil)
		val, err := m.GetCP()
		log.Debugf("%+v,err :%s", val, err)
		mv := m.GetValue()
		log.Debugf("%+v", mv)
	})
}
