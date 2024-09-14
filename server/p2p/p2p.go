package p2p

import (
	"bytes"
	"errors"
	"io"
	"time"
)

const (
	handshakeMsg = 0x00
	discMsg      = 0x01
	pingMsg      = 0x02
	pongMsg      = 0x03
)

// Unexported devp2p protocol lengths from p2p package.
const (
	baseProtoLen        = 16
	ethProtoLen         = 17
	snapProtoLen        = 8
	message_type_size   = 1
	message_option_size = 4
)

// Unexported handshake structure from p2p/peer.go.
type protoHandshake struct {
	Version    uint64
	Name       string
	Caps       []byte
	ListenPort uint64
	ID         []byte
}

type H = protoHandshake

// Proto is an enum representing devp2p protocol types.
type Proto int

type Cap struct {
	Name    string
	Version int32
}

type Msg struct {
	Code       byte
	Size       uint32 // Size of the raw payload
	Payload    *bytes.Reader
	ReceivedAt time.Time

	Reply chan Msg
}

func NewMessage(id byte) *Msg {

	return &Msg{Code: id}
}

func (m *Msg) MarshalBinary() ([]byte, error) {

	buf := new(bytes.Buffer)

	buf.WriteByte(m.Code)

	sizeBytes := [4]byte{
		byte(m.Size >> 24),
		byte(m.Size >> 16),
		byte(m.Size >> 8),
		byte(m.Size),
	}
	buf.Write(sizeBytes[:])

	if m.Payload != nil {

		data, err := io.ReadAll(m.Payload)
		if err != nil {
			return nil, err
		}
		buf.Write(data)
	}

	return buf.Bytes(), nil
}

func (m *Msg) UnmarshalBinary(d []byte) error {

	if len(d) < 5 {
		return errors.New("insufficient data length")
	}

	m.Code = d[0]

	m.Size = uint32(d[1])<<24 | uint32(d[2])<<16 | uint32(d[3])<<8 | uint32(d[4])

	if len(d) < 5+int(m.Size) {
		return errors.New("insufficient data for payload")
	}

	m.Payload = bytes.NewReader(d[5 : 5+int(m.Size)])

	return nil
}
