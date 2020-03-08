package packets

import (
	"io"
	"strings"
	"time"
)

// Pduer .......
type Pduer interface {
	GetID() string
	GetPW() string
	GetMN() string
	GetST() string
	GetCN() CNType
	GetQN() time.Time
	GetPNO() uint32
	GetPNUM() uint32
	GetFlag() byte
	GetCPkvg() CPkvg
	NeedAck() bool
	Writeto(io.Writer) (int, error)
	Unpack(io.Reader) error
}

// GetID get hjpdu QN
func (h *HjPdu) GetID() string {
	qn := strings.Replace(h.QN.Format(`20060102150405.000`), ".", "", 1)
	return qn
}

// GetST get hjpdu ST
func (h *HjPdu) GetST() string {
	return h.ST
}

// GetCN get hjpdu CN
func (h *HjPdu) GetCN() CNType {
	return h.CN
}

// GetPW get hjpdu PW
func (h *HjPdu) GetPW() string {
	return h.PW
}

// GetQN get hjpdu QN
func (h *HjPdu) GetQN() time.Time {
	return h.QN
}

// GetMN get hjpdu MN
func (h *HjPdu) GetMN() string {
	return h.MN
}

// GetFlag get hjpdu flag
func (h *HjPdu) GetFlag() byte {
	return h.Flag
}

// GetPNO get hjpdu PNO
func (h *HjPdu) GetPNO() uint32 {
	return h.PNO
}

// GetPNUM get hjpdu PNUM
func (h *HjPdu) GetPNUM() uint32 {
	return h.PNUM
}

// GetCPkvg get hjpdu CPkvg
func (h *HjPdu) GetCPkvg() CPkvg {
	return h.CP.CPkvg
}

// NeedAck .....
func (h *HjPdu) NeedAck() bool {
	return h.Flag&0x1 == 0x1
}

// ReadPduer ....
func ReadPduer(r io.Reader) (Pduer, error) {
	h := &HjPdu{}
	err := h.Unpack(r)
	return h, err
}
