package p2p

import (
	"bytes"
	"testing"
)

func TestMessageMarshalling(t *testing.T) {
	payload := bytes.NewReader([]byte("test"))
	mes := &Msg{Code: 20, Size: uint32(payload.Len()), Payload: payload}
	bs, err := mes.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	newMes := new(Msg)
	err = newMes.UnmarshalBinary(bs)
	if err != nil {
		t.Error(err)
	}

	// Ensure the original and unmarshaled messages are equal
	if newMes.Code != mes.Code || newMes.Size != mes.Size {
		t.Errorf("Marshall unmarshall message error, got %v, want %v", newMes, mes)
	}

	// Compare payloads
	expectedPayload := make([]byte, mes.Size)
	_, err = mes.Payload.Read(expectedPayload)
	if err != nil {
		t.Error(err)
	}

	actualPayload := make([]byte, newMes.Size)
	_, err = newMes.Payload.Read(actualPayload)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(expectedPayload, actualPayload) {
		t.Errorf("Payload mismatch: got %v, want %v", actualPayload, expectedPayload)
	}
}
