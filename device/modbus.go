package device

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/yjiong/fkhj212/clientapp"
	"github.com/yjiong/fkhj212/packets"
	"github.com/yjiong/fkhj212/storage"
	"github.com/yjiong/iotgateway/modbus"
)

// Dev defined
var (
	mrule         = mustGetRule()
	v             = clientapp.GetConf()
	StoreInterval = getStoreInterval()
	Dev           *ModbusDev
)

func init() {
	Dev = &ModbusDev{}
}

func getStoreInterval() time.Duration {
	v.ReadInConfig()
	val := v.GetInt("StoreInterval")
	return time.Duration(val) * time.Second
}

// ModbusDev Modbus device
type ModbusDev struct {
	name        string
	iface       string
	baudrate    int
	databits    int
	parity      string
	stopbits    int
	timeout     time.Duration
	slaveid     byte
	startaddr   uint16
	quantity    uint16
	fs          []clientapp.ConfFS
	rule        *valueRule
	value       *packets.CPField
	modbusValue []byte
}

type valueRule struct {
	vlength  int
	multiply float64
	codes    []map[string]string
}

func mustGetRule() *valueRule {
	v.ReadInConfig()
	vl := v.GetInt("modbus.rule.vlength")
	if vl <= 0 {
		log.Fatalf("%s", "get rule vlength failed")
	}
	var cpk []map[string]string
	if cpl := v.GetStringSlice("modbus.keycode"); cpl != nil {
		for _, cp := range cpl {
			kv := strings.Split(cp, "-")
			cpk = append(cpk, map[string]string{kv[0]: kv[1]})
			//cpid = append(cpid, map[string]string{kv[0]: kv[2]})
		}
	}
	mu := v.GetFloat64("modbus.rule.multiply")
	return &valueRule{
		vlength:  vl,
		multiply: mu,
		codes:    cpk,
	}
}

var deviceFile string

// GetMD ....
func (m *ModbusDev) GetMD(f clientapp.ConfFactor) (*ModbusDev, error) {
	//sl := viper.GetStringMapString("modbus")
	log.Debugf("in new modbus dev")
	ts := ModbusDev{
		name:      "",
		iface:     `/dev/ttyS` + f.CO,
		baudrate:  f.BR,
		databits:  f.DB,
		parity:    f.PB,
		stopbits:  f.SB,
		timeout:   1 * time.Second,
		slaveid:   f.AR,
		startaddr: 0,
		fs:        f.FS,
		quantity:  uint16(len(mrule.codes) * 2),
		rule:      mrule,
		value:     &packets.CPField{NoDTime: true},
	}
	return &ts, nil
}

// ReadDev ......
func (m *ModbusDev) ReadDev() (cp *packets.CPField, err error) {
	if err = m.gesmdRet(); err != nil {
		log.Error(errors.WithMessagef(err, "read modbus device at serail prot %s", m.iface))
		return
	}
	m.value, err = m.mdBytes2CPField()
	log.Debugf("ModbusDev ReadDev :value=%v", m.value)
	return m.value, err
}

func (m *ModbusDev) gesmdRet() (err error) {
	if int(m.quantity)%m.rule.vlength != 0 {
		err = errors.New("modbus register quantity error")
		return
	}
	sh := modbus.NewRTUClientHandler(m.iface)
	sh.BaudRate = m.baudrate
	sh.DataBits = m.databits
	sh.StopBits = m.stopbits
	sh.Parity = m.parity
	sh.Timeout = m.timeout
	sh.SlaveId = m.slaveid
	if err = sh.Connect(); err != nil {
		return
	}
	defer sh.Close()
	c := modbus.NewClient(sh)
	m.modbusValue, err = c.ReadHoldingRegisters(m.startaddr, m.quantity)
	return
}

func getFloat(results []byte, rule *valueRule) []float32 {
	//log.Debugf("byte value = % x", results)
	var retlistFloat32 []float32
	for i := 0; i < len(results); i += rule.vlength {
		var ti int
		for k := 0; k < rule.vlength; k++ {
			ti += (int(results[k+i]) << (8 * uint(rule.vlength-k-1)))
		}
		retlistFloat32 = append(retlistFloat32, float32(ti)*float32(rule.multiply))
	}
	return retlistFloat32
}

func (m *ModbusDev) getKVRtdWithID(val []float32) (kvp []packets.CPkv, err error) {
	key, code := m.keyAndID()
	id := make([]string, len(key))
	rh := make([]float32, len(key))
	rl := make([]float32, len(key))

	if len(key) != len(val) {
		err = errors.New("modbus value number error")
		return
	}
	for ifs, fs := range m.fs {
		if fs.NM == key[ifs] {
			id[ifs] = fs.ID
			rh[ifs] = fs.RH
			rl[ifs] = fs.RL
		}
	}
	for i, k := range code {
		kvp = append(kvp, make(packets.CPkv, 1))
		kvp[i][k+"-Rtd"] = fmt.Sprintf("%0.2f", val[i])
		if len(id[i]) > 0 {
			kvp[i][k+"-ID"] = fmt.Sprintf("%s", id[i])
		}
		//if val[i] < rl[i] {
		//kvp[i][k+"-Flag"] = "RL"
		//}
		//if val[i] > rh[i] {
		//kvp[i][k+"-Flag"] = "RH"
		//}
	}
	return
}

// GetChkMap ....
func (m *ModbusDev) GetChkMap() (ckv []map[string]interface{}, err error) {
	if len(m.modbusValue) != len(mrule.codes)*4 {
		if _, err = m.ReadDev(); err != nil {
			return
		}
	}
	ckey, _ := m.keyAndID()
	fv := getFloat(m.modbusValue, m.rule)
	if len(ckey) != len(fv) {
		err = errors.New("modbus value number error")
		return
	}
	for i, k := range ckey {
		ckv = append(ckv, make(map[string]interface{}, 1))
		ckv[i][k] = fv[i]
	}
	return
}

func (m *ModbusDev) keyAndID() (chkey, code []string) {
	for _, ckm := range m.rule.codes {
		for k, v := range ckm {
			chkey = append(chkey, k)
			code = append(code, v)
		}
	}
	return
}

func (m *ModbusDev) mdBytes2CPField() (cg *packets.CPField, err error) {
	var cpkvs []packets.CPkv
	fv := getFloat(m.modbusValue, m.rule)
	cpkvs, err = m.getKVRtdWithID(fv)
	if err != nil {
		return
	}
	return packets.NewCPFildFromCPkvg(cpkvs), nil
}

// GetValue ....
func (m *ModbusDev) GetValue() packets.CPField {
	return *m.value
}

// StoreVal ....
func (m *ModbusDev) StoreVal(db sqlx.Ext) (err error) {
	if len(m.modbusValue) != len(mrule.codes)*4 {
		if _, err = m.ReadDev(); err != nil {
			return
		}
	}
	fv := getFloat(m.modbusValue, m.rule)
	f := &storage.Factors{
		ValueTimeStamp: m.value.DataTime,
		DeviceID:       fmt.Sprintf("%s-%d", m.iface, m.slaveid),
		Value1:         float64(fv[0]),
		Value2:         float64(fv[1]),
		Value3:         float64(fv[2]),
		Value4:         float64(fv[3]),
		Value5:         float64(fv[4]),
		Value6:         float64(fv[5]),
		Value7:         float64(fv[6]),
		Value8:         float64(fv[7]),
		Value9:         float64(fv[8]),
		Value10:        float64(fv[9]),
	}
	return storage.InsertFactorValue(db, f)
}
