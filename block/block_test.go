package block

import (
	"testing"

	"github.com/NikoMalik/blockchain/crypto"
	"github.com/NikoMalik/blockchain/util"
	"github.com/stretchr/testify/assert"
)

func TestSignBlock(t *testing.T) {
	block := util.RandomBlock()
	privKey := crypto.NewPrivateKey()
	pubKey := privKey.Public()
	sing := SignBlock(privKey, block)

	assert.True(t, sing.Verify(pubKey, HashBlock(block)))
	assert.Equal(t, len(sing.Bytes()), 64)

}

func TestHashBlock(t *testing.T) {
	block := util.RandomBlock()
	assert.Equal(t, len(HashBlock(block)), 64)

}
