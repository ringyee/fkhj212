package packets

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// HjPdu .....
type HjPdu struct {
	QN   time.Time
	ST   uint32
	CN   uint32
	PW   string
	MN   string
	Flag byte
	PNUM uint32
	PNO  uint32
	CP   CPField
}

type cpkv map[string]interface{}

type cpkvg []cpkv

// CPField ...
type CPField struct {
	DataTime time.Time
	cpkvg
	NoDTime bool
}

// NewCPFild   create CPField from map list
func NewCPFild(dt time.Time, cpkvs ...cpkv) *CPField {
	return &CPField{
		DataTime: dt,
		cpkvg:    cpkvs,
	}
}

func (c *CPField) cpMarshal() (ch string) {
	ch = "&&"
	if !c.NoDTime {
		ch += c.DataTime.Format("DataTime=20060102150405;")
	}
	ch += cpkvg2str(c.cpkvg)
	ch += "&&"
	return
}

func cp2json(cps string) (jstr string) {
	patternCp := regexp.MustCompile(`.*CP=&&(.*)&&`)
	tstr := fmt.Sprintf("%s", patternCp.ReplaceAllString(cps, "${1}"))
	splittstr := strings.Split(tstr, ";")
	if len(splittstr) > 1 {
		jstr = "["
	}
	for _, kvs := range splittstr {
		jstr += "{"
		for _, kv := range strings.Split(kvs, ",") {
			jstr += "\""
			jstr += strings.ReplaceAll(kv, "=", "\":\"")
			jstr += "\","
		}
		jstr = strings.TrimRight(jstr, ",")
		jstr += "},"
	}
	jstr = strings.TrimRight(jstr, ",")
	if len(splittstr) > 1 {
		jstr += "]"
	}
	return
}

func cpUnMarshal(cpl []interface{}) (cp *CPField) {
	cp = new(CPField)
	var cg []cpkv
	for _, cms := range cpl {
		if cm, ok := cms.(map[string]interface{}); ok {
			if dt, ok := cm["DataTime"].(string); ok {
				if t, err := time.Parse("20060102150405", dt); err == nil {
					cp.DataTime = t
				} else {
					cp.NoDTime = true
				}
			} else {
				cg = append(cg, cpkv(cm))
			}
		}
	}
	cp.cpkvg = cg
	return
}

// NewHjPdu create HjPdu struct
// ???????????????????????????????????????????????????????????????
func NewHjPdu(LocalSet map[string]interface{}, cp CPField) *HjPdu {
	hjp := new(HjPdu)
	hjInit(LocalSet, hjp)
	hjp.CP = cp
	return hjp
}

// ???????????????????????????????????????????????????????????????

func (h *HjPdu) marshal() (dh string) {
	qns := h.QN.Format("QN=20060102150405.999;")
	dh = strings.Replace(qns, ".", "", 1)
	if h.ST != 0 {
		dh += fmt.Sprintf("ST=%d;", h.ST)
	}
	if h.CN != 0 {
		dh += fmt.Sprintf("CN=%d;", h.CN)
	}
	if h.PW != "" {
		dh += fmt.Sprintf("PW=%s;", h.PW)
	}
	if h.MN != "" {
		dh += fmt.Sprintf("MN=%s;", h.MN)
	}
	if h.Flag != 0 {
		dh += fmt.Sprintf("Flag=%d;", h.Flag)
	}
	if h.PNUM != 0 {
		dh += fmt.Sprintf("PNUM=%d;", h.PNUM)
	}
	if h.PNO != 0 {
		dh += fmt.Sprintf("PNO=%d;", h.PNO)
	}
	dh += fmt.Sprintf("CP=%s", h.CP.cpMarshal())
	return
}

// Marshal ....
func (h *HjPdu) Marshal() (hjs string) {
	df := h.marshal()
	hlenght := "##" + fmt.Sprintf("%04d", len(df))
	crc := Crc16Checkout(df)
	hjs = hlenght + df + fmt.Sprintf("%04x\r\n", crc)
	return
}

// Hj2json 2 json format (all data value is string format)
func Hj2json(hjpdu []byte) (jstr string) {
	hjpbs := string(hjpdu)
	if lt, err := strconv.Atoi(hjpbs[2:6]); err == nil &&
		hjpbs[:2] == "##" &&
		len(hjpbs) == lt+12 &&
		hjpbs[lt+6:lt+10] == fmt.Sprintf("%04x", Crc16Checkout(hjpbs[6:lt+6])) {
		re := regexp.MustCompile(`(?P<k>\w+)=(?P<v>\w+);(?:CP.*)?`)
		tstr := fmt.Sprintf("%s", re.ReplaceAllString(hjpbs[6:lt+6], "\"${k}\":\"${v}\","))
		jstr = fmt.Sprintf("{%s\"CP\":%s}", tstr, cp2json(hjpbs[6:lt+6]))
	}
	log.Debugf("Hj2json jstr=%s", jstr)
	return
}

// UnMarshal HjPdu.UnMarshal
func (h *HjPdu) UnMarshal(hjpdu []byte) (err error) {
	var hp *HjPdu
	hp, err = UnMarshal(hjpdu)
	*h = *hp
	return
}

// UnMarshal ....
func UnMarshal(hjpdu []byte) (hjp *HjPdu, err error) {
	hjjstr := Hj2json(hjpdu)
	if len(hjpdu) == 0 && len(hjjstr) == 0 {
		err = errors.New("UnMarshal empty hjpdu")
		return
	}
	hjp = new(HjPdu)
	tmap := make(map[string]interface{})
	if err = json.Unmarshal([]byte(hjjstr), &tmap); err != nil {
		return
	}
	hjInit(tmap, hjp)
	return
}

func hjInit(tmap map[string]interface{}, hjp *HjPdu) {
	if qn, ok := tmap["QN"].(string); ok {
		qn = qn[:14] + "." + qn[14:]
		if t, err := time.Parse("20060102150405.999", qn); err == nil {
			hjp.QN = t
		}
	}
	if st, ok := tmap["ST"].(string); ok {
		if stint, err := strconv.Atoi(st); err == nil {
			hjp.ST = uint32(stint)
		}
	}
	if cn, ok := tmap["CN"].(string); ok {
		if stint, err := strconv.Atoi(cn); err == nil {
			hjp.CN = uint32(stint)
		}
	}
	if pw, ok := tmap["PW"].(string); ok {
		hjp.PW = pw
	}
	if mn, ok := tmap["MN"].(string); ok {
		hjp.MN = mn
	}
	if flag, ok := tmap["Flag"].(string); ok {
		if stint, err := strconv.Atoi(flag); err == nil {
			hjp.Flag = uint8(stint)
		}
	}
	if pnum, ok := tmap["PNUM"].(string); ok {
		if stint, err := strconv.Atoi(pnum); err == nil {
			hjp.PNUM = uint32(stint)
		}
	}
	if pno, ok := tmap["PNO"].(string); ok {
		if stint, err := strconv.Atoi(pno); err == nil {
			hjp.PNO = uint32(stint)
		}
	}
	if cp, ok := tmap["CP"].([]interface{}); ok {
		hjp.CP = *cpUnMarshal(cp)
	}
	return
}

// Crc16Checkout fkhj212 Crc16Checkout
func Crc16Checkout(inif interface{}) uint16 {
	var bs []rune
	switch inif.(type) {
	case string:
		bs = []rune(inif.(string))
	case []rune:
		bs = inif.([]rune)
	case []byte:
		for bb := range inif.([]byte) {
			bs = append(bs, rune(bb))
		}
	}
	var crcReg uint16 = 0xFFFF
	for i := 0; i < len(bs); i++ {
		crcReg = (crcReg >> 8) ^ (uint16(bs[i]))
		for j := 0; j < 8; j++ {
			check := crcReg & 1
			crcReg >>= 0x1
			if check == 0x1 {
				crcReg ^= 0xA001
			}
		}
	}
	return crcReg
}
