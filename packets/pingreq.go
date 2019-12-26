package packets

import (
	"io"
)

//PingreqPacket is an internal representation of the fields of the
//Pingreq  packet
type PingreqPacket struct {
	HjPdu
}

func (pr *PingreqPacket) Write(w io.Writer) error {

	return nil
}

//Unpack decodes the details of a ControlPacket after the fixed
//header has been read
func (pr *PingreqPacket) Unpack(b io.Reader) error {
	return nil
}
