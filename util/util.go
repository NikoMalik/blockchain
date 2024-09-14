package util

import (
	randc "crypto/rand"
	"io"
	"math/rand"
	"time"

	"github.com/NikoMalik/blockchain/proto"
)

func RandomHash() []byte {
	hash := make([]byte, 64)

	_, err := io.ReadFull(randc.Reader, hash)
	if err != nil {
		panic(err)
	}

	return hash
}

func RandomBlock() *proto.Block {
	header := &proto.Header{
		Version:      1,
		Height:       int32(rand.Intn(1000)),
		PreviousHash: RandomHash(),
		MerkleRoot:   RandomHash(),
		Timestamp:    time.Now().Unix(),
	}
	return &proto.Block{Header: header}
}
