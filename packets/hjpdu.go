package packets

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// HjPdu .....
type HjPdu struct {
	QN   time.Time
	ST   string
	CN   CNType
	PW   string
	MN   string
	Flag byte
	PNUM uint32
	PNO  uint32
	CP   CPField
}

// CPkv CP key value pair
type CPkv map[string]interface{}

// CPkvg .....
type CPkvg []CPkv

// CPField ...
type CPField struct {
	DataTime time.Time
	CPkvg
	NoDTime bool
}

// NewCPFildFromCPkvs   create CPField from map list
func NewCPFildFromCPkvs(cpkvs ...CPkv) *CPField {
	return &CPField{
		DataTime: time.Now(),
		CPkvg:    cpkvs,
	}
}

// NewCPFildFromCPkvg   create CPField from CPkvg
func NewCPFildFromCPkvg(cpkvg []CPkv) *CPField {
	return &CPField{
		DataTime: time.Now(),
		CPkvg:    cpkvg,
	}
}

func (c *CPField) cpMarshal() (ch string) {
	ch = "&&"
	if !c.NoDTime {
		c.CPkvg = append(c.CPkvg, CPkv{"DataTime": c.DataTime.Format("20061002150405")})
	}
	ch += cpkvg2str(c.CPkvg)
	ch += "&&"
	return
}

func cp2json(cps string) (jstr string) {
	patternCp := regexp.MustCompile(`.*CP=&&(.*)&&`)
	tstr := patternCp.ReplaceAllString(cps, "${1}")
	if tstr == "" {
		return "[]"
	}
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

func cpUnMarshal(cps interface{}) (cp *CPField) {
	cp = new(CPField)
	cp.NoDTime = true
	var cg []CPkv
	cpl, ok := cps.([]interface{})
	if !ok {
		cpl = []interface{}{cps}
	}
	for _, cms := range cpl {
		if cm, ok := cms.(map[string]interface{}); ok {
			if dt, ok := cm["DataTime"].(string); ok {
				if t, err := time.Parse("20060102150405", dt); err == nil {
					cp.DataTime = t
					cp.NoDTime = false
				}
			} else {
				cg = append(cg, CPkv(cm))
			}
		}
	}
	cp.CPkvg = cg
	return
}

// NewHjPdu create HjPdu struct
func NewHjPdu(all map[string]interface{}, cp CPField) *HjPdu {
	hjp := new(HjPdu)
	hjInit(all, hjp)
	hjp.CP = cp
	return hjp
}

func (h *HjPdu) marshal() (dh string) {
	qns := h.QN.Format("QN=20060102150405.000;")
	dh = strings.Replace(qns, ".", "", 1)
	if h.ST != "" {
		dh += fmt.Sprintf("ST=%s;", h.ST)
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
	crc := Crc16CheckoutStr(df)
	hjs = hlenght + df + crc + "\r\n"
	return
}

// Hj2json 2 json format (all data value is string format)
func Hj2json(hjpdu []byte) (jstr string) {
	hjpbs := string(hjpdu)
	if lt, err := strconv.Atoi(hjpbs[2:6]); err == nil &&
		hjpbs[:2] == "##" &&
		len(hjpbs) == lt+12 &&
		(hjpbs[lt+6:lt+10] == fmt.Sprintf("%04x", Crc16Checkout(hjpbs[6:lt+6])) ||
			hjpbs[lt+6:lt+10] == Crc16CheckoutStr(hjpbs[6:lt+6])) {
		re := regexp.MustCompile(`(?P<k>\w+)=(?P<v>\w+);(?:CP.*)?`)
		tstr := re.ReplaceAllString(hjpbs[6:lt+6], "\"${k}\":\"${v}\",")
		jstr = fmt.Sprintf("{%s\"CP\":%s}", tstr, cp2json(hjpbs[6:lt+6]))
		log.Debugf("Hj2json ok !")
	}
	return
}

// UnMarshal HjPdu.UnMarshal
func (h *HjPdu) UnMarshal(hjpdu []byte) (err error) {
	var hp *HjPdu
	hp, err = UnMarshal(hjpdu)
	*h = *hp
	return
}

// Writeto pdu write to io
func (h *HjPdu) Writeto(w io.Writer) (i int, err error) {
	hb := []byte(h.Marshal())
	//log.Debugf("Pduer writeto func out bytes = % x", hb)
	log.Infof("Pduer writeto func out string = %s\n", hb)
	wbuf := bytes.NewBuffer(hb)
	return w.Write(wbuf.Bytes())
}

// Unpack get pdu struct from io
func (h *HjPdu) Unpack(r io.Reader) (err error) {
	bh := make([]byte, 2)
	for {
		var unmb = []byte{'#', '#'}
		if _, err = io.ReadFull(r, bh); err == nil && bh[0] == '#' && bh[1] == '#' {
			bl := make([]byte, 4)
			var lt int
			if _, err = io.ReadFull(r, bl); err == nil {
				if lt, err = strconv.Atoi(string(bl)); err == nil {
					unmb = append(unmb, bl...)
					bt := make([]byte, lt+6)
					if _, err = io.ReadFull(r, bt); err == nil {
						unmb = append(unmb, bt...)
						log.Debugf("io Unpack func hex pdu: %+v", unmb)
						if uerr := h.UnMarshal(unmb); uerr != nil {
							log.Error(uerr)
							continue
						}
						return
					}
				}
			}
		}
		if err != nil {
			log.Error(err)
			return
		}
	}
}

// UnMarshal ....
func UnMarshal(hjpdu []byte) (hjp *HjPdu, err error) {
	hjjstr := Hj2json(hjpdu)
	if len(hjpdu) == 0 && len(hjjstr) == 0 {
		err = errors.New("UnMarshal empty pdu")
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
		if t, err := time.Parse("20060102150405.000", qn); err == nil {
			hjp.QN = t
		}
	}
	if qn, ok := tmap["QN"].(time.Time); ok {
		hjp.QN = qn
	}
	if st, ok := tmap["ST"].(string); ok {
		hjp.ST = st
	}
	//if st, ok := tmap["ST"].(uint32); ok {
	//hjp.ST = st
	//}
	if cn, ok := tmap["CN"].(string); ok {
		if stint, err := strconv.Atoi(cn); err == nil {
			hjp.CN = CNType(stint)
		}
	}
	if cn, ok := tmap["CN"].(int); ok {
		hjp.CN = CNType(cn)
	}
	if pw, ok := tmap["PW"].(string); ok {
		hjp.PW = pw
	}
	if mn, ok := tmap["MN"].(string); ok {
		hjp.MN = mn
	}
	if flag, ok := tmap["Flag"].(int); ok {
		hjp.Flag = uint8(flag)
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
	if cp, ok := tmap["CP"].(interface{}); ok {
		hjp.CP = *cpUnMarshal(cp)
	}
}

// Crc16CheckoutStr fkhj212 Crc16CheckoutStr
func Crc16CheckoutStr(inif interface{}) string {
	s := Crc16Checkout(inif)
	return strings.ToUpper(fmt.Sprintf("%04x", s))
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
