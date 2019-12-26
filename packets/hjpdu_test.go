package packets

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestHjPdu(t *testing.T) {
	Convey("==================测试=====================\n", t, func() {
		//ts := "QN=20160801085857223;ST=32;CN=1062;PW=100000;MN=010000A8900016F000169DC0;Flag=5;CP=&&RtdInterval=30&&"
		//fmt.Printf("%x", Crc16Checkout(ts))
		//fmt.Println(time.Now().Format("20060102150405.99"))
		////////////////////////////////
		//cpk := cpkv{"RtdInterval": 30}
		rtd := cpkv{
			"006-Rtd": 0.77,
			"006-ID":  0,
		}
		rtd1 := cpkv{
			"006-Rtd": 0.48,
			"006-ID":  1,
		}
		rtd2 := cpkv{
			"007-Rtd": 35,
			"007-ID":  0,
		}
		rtd3 := cpkv{
			"008-Rtd": 36,
			"008-ID":  0,
		}
		rtd4 := cpkv{
			"012-Rtd": 0.00,
			"012-ID":  1,
		}
		rtd5 := cpkv{
			"011-Rtd": 0,
			"011-ID":  1,
		}
		rtd6 := cpkv{
			"012-Rtd": 0.00,
			"012-ID":  2,
		}
		rtd7 := cpkv{
			"011-Rtd": 0,
			"011-ID":  2,
		}
		rtd8 := cpkv{
			"009-Rtd": 0,
			"009-ID":  0,
		}
		rtd9 := cpkv{
			"010-Rtd": 0,
			"010-ID":  0,
		}
		cpf := NewCPFild(time.Now(), rtd, rtd1, rtd2, rtd3, rtd4,
			rtd5, rtd6, rtd7, rtd8, rtd9)
		//cpf.NoDTime = true
		//t.Log(cpf.cpMarshal())
		//t.Log("% x", []byte(cpf.cpMarshal()))
		qn, _ := time.Parse("20060102150405.999", "20191217173712.123")
		hjps := &HjPdu{
			QN:   qn,
			ST:   99,
			CN:   2011,
			PW:   "123456",
			MN:   "20181025170013",
			Flag: 5,
		}
		hjps.CP = *cpf
		fmt.Printf("hjpdu struct=%+v\n", *hjps)
		////////////////////////////////////
		cpb := hjps.Marshal()
		fmt.Printf("hjpdu struct Marshal=%s\n\n", cpb)
		hjjson := Hj2json([]byte(cpb))
		fmt.Printf("hjjson=%s\n\n", hjjson)
		////////////////////////////////////
		uv, _ := UnMarshal([]byte(cpb))
		fmt.Printf("UnMarshal hjpdu struct = %+v\n\n", *uv)
		fmt.Printf("hjpdu strct Marshal=%s\n\n", uv.Marshal())
		fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	})
}
