package clientapp

import (
	"github.com/yjiong/fkhj212/packets"
	"time"
)

// Fkhjer hj interface
type Fkhjer interface {
	IsConnected() bool
	IsConnectionOpen() bool
	Connect() Token
	Disconnect()
	UploadPdu(cntype packets.CNType, ucp packets.CPField) Token
	RtdInterval() time.Duration
}

// NewFkhj will create
func NewFkhj(o *ClientOptions) Fkhjer {
	c := &client{}
	c.options = *o
	c.status = disconnected
	c.Pt = pt
	return c
}
