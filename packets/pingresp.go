package packets

import (
	"io"
)

//PingrespPacket is an internal representation of the fields of the
//Pingresp  packet
type PingrespPacket struct {
	HjPdu
}

func (pr *PingrespPacket) Write(w io.Writer) error {

	return nil
}

//Unpack decodes the details of a ControlPacket after the fixed
//header has been read
func (pr *PingrespPacket) Unpack(b io.Reader) error {
	return nil
}
