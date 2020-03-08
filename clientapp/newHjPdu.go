package clientapp

import (
	"time"

	"github.com/yjiong/fkhj212/packets"
)

// ACKBIT falg ack bit
const ACKBIT uint8 = 5

// NewPdu make pdu
func NewPdu(pCNtype packets.CNType, c *client, cp packets.CPField) (hp *packets.HjPdu) {
	//cp := packets.NewCPFild(time.Now(), cps)
	hp = packets.NewHjPdu(map[string]interface{}{
		"QN":   time.Now(),
		"MN":   c.options.MN,
		"ST":   c.options.ST,
		"PW":   c.options.PW,
		"Flag": 4,
	}, cp)
	hp.CN = pCNtype
	hp.CP = cp
	switch pCNtype {
	case packets.SsetTimeReq:
		hp.Flag |= ACKBIT
		hp.CP.NoDTime = true
	case packets.SPutConfig:
		hp.CP.NoDTime = true
		hp.ST = "LC"
	default:
		/////////
	}
	return
}
