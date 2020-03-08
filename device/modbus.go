package device

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/yjiong/fkhj212/clientapp"
	"github.com/yjiong/fkhj212/packets"
	"github.com/yjiong/iotgateway/modbus"
)

// Dev defined
var (
	mrule = mustGetRule()
	v     = clientapp.GetConf()
	Dev   *ModbusDev
)

func init() {
	Dev = &ModbusDev{}
}

// ModbusDev Modbus device
type ModbusDev struct {
	name      string
	iface     string
	baudrate  int
	databits  int
	parity    string
	stopbits  int
	timeout   time.Duration
	slaveid   byte
	startaddr uint16
	quantity  uint16
	fs        []clientapp.ConfFS
	rule      *valueRule
	value     *packets.CPField
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

// GetMD .....
func (m *ModbusDev) GetMD(f clientapp.ConfFactor) (*ModbusDev, error) {
	//sl := viper.GetStringMapString("modbus")
	log.Debugf("in new modbus dev")
	ts := ModbusDev{
		name:  "",
		iface: `/dev/ttyS` + f.CO,
		//iface:     `/dev/ttyUSB0`,
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

// GetCP ......
func (m *ModbusDev) GetCP() (cp *packets.CPField, err error) {
	var md []byte
	if md, err = m.gesmdRet(); err != nil {
		log.Error(errors.WithMessagef(err, "read modbus device at serail prot %s", m.iface))
		return
	}
	m.value, err = m.mdBytes2CPField(md)
	return m.value, err
}

// GetChkMap ......
func (m *ModbusDev) GetChkMap() (ckv []map[string]interface{}, err error) {
	var md []byte
	if md, err = m.gesmdRet(); err != nil {
		return
	}
	return m.getChkMap(md)
}

func (m *ModbusDev) gesmdRet() (mdret []byte, err error) {
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
	mdret, err = c.ReadHoldingRegisters(m.startaddr, m.quantity)
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
	if len(key) != len(val) {
		err = errors.New("modbus value number error")
		return
	}
	for ifs, fs := range m.fs {
		if fs.NM == key[ifs] {
			id[ifs] = fs.ID
		}
	}
	for i, k := range code {
		kvp = append(kvp, make(packets.CPkv, 1))
		kvp[i][k+"-Rtd"] = fmt.Sprintf("%0.2f", val[i])
		if len(id[i]) > 0 {
			kvp[i][k+"-ID"] = fmt.Sprintf("%s", id[i])
		}
	}
	return
}

func (m *ModbusDev) getChkMap(mdret []byte) (ckv []map[string]interface{}, err error) {
	ckey, _ := m.keyAndID()
	fv := getFloat(mdret, m.rule)
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

func (m *ModbusDev) mdBytes2CPField(mdret []byte) (cg *packets.CPField, err error) {
	var cpkvs []packets.CPkv
	fv := getFloat(mdret, m.rule)
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
