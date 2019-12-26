package packets

import (
	"bytes"
	"errors"
	//log "github.com/sirupsen/logrus"
	"io"
	"strconv"
	"strings"
)

// Pduer .......
type Pduer interface {
	Writeto(io.Writer) (int, error)
	Unpack(io.Reader) error
	GetID() string
}

// Writeto pdu write to io
func (h *HjPdu) Writeto(w io.Writer) (i int, err error) {
	hb := []byte(h.Marshal())
	wbuf := bytes.NewBuffer(hb)
	return w.Write(wbuf.Bytes())
}

// Unpack get pdu struct from io
func (h *HjPdu) Unpack(r io.Reader) (err error) {
	bh := make([]byte, 6)
	var lt int
	if _, err = io.ReadFull(r, bh); err == nil {
		if bh[0] == bh[1] && bh[1] == '#' {
			if lt, err = strconv.Atoi(string(bh[2:6])); err == nil {
				tb := make([]byte, lt+6)
				if _, err = io.ReadFull(r, tb); err == nil {
					return h.UnMarshal(append(bh, tb...))
				}
			}
		}
	}
	if err == nil {
		err = errors.New("Unpack to HjPdu failed")
	}
	return
}

// GetID get hjpdu QN
func (h *HjPdu) GetID() string {
	qn := strings.Replace(h.QN.Format(`20060102150405.999`), ".", "", 1)
	return qn
}

// ReadPduer ....
func ReadPduer(r io.Reader) (Pduer, error) {
	h := &HjPdu{}
	err := h.Unpack(r)
	return h, err
}
