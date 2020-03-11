package device

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
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
		m, _ := Dev.GetMD(clientapp.CF.Factors[0])
		log.Debugf("%#v", m)
		So(m, ShouldNotBeNil)
		val, err := m.ReadDev()
		log.Debugf("%+v,err :%s", val, err)
		So(err, ShouldBeNil)
		mapv, err := m.GetChkMap()
		log.Debugf("%+v,%d ", mapv, len(mapv))
		So(len(mapv), ShouldEqual, len(mrule.codes))
		///////////////////////////////
		db, err := sqlx.Open("postgres", "postgres://postgres:yj12345@2.59.151.166:1932/postgres?sslmode=disable")
		if err != nil {
			t.Fatal(err)
		}
		So(err, ShouldBeNil)
		for {
			if err := db.Ping(); err != nil {
				t.Errorf("ping database error, will retry in 2s: %s", err)
				time.Sleep(2 * time.Second)
			} else {
				break
			}
		}
		err = m.StoreVal(db)
		log.Error(err)
		So(err, ShouldBeNil)
	})
}
