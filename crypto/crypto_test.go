package crypto

import (
	"crypto/ed25519"
	"testing"

	lowlevelfunctions "github.com/NikoMalik/low-level-functions"
	"github.com/stretchr/testify/assert"
)

func TestBytes(t *testing.T) {

	p := &PrivateKey{
		key: ed25519.PrivateKey{}}

	p.Bytes()

}

func TestGenerateKeys(t *testing.T) {
	privKey := NewPrivateKey()

	assert.Equal(t, len(privKey.Bytes()), privateKeyLen)
}

func TestPrivateKeySign(t *testing.T) {
	privKey := NewPrivateKey()
	pubKey := privKey.Public()
	msg := lowlevelfunctions.StringToBytes("nigga balls")
	sign := privKey.Sign(msg)
	assert.True(t, sign.Verify(pubKey, msg))

	assert.False(t, sign.Verify(pubKey, lowlevelfunctions.StringToBytes("niggas balls")))
}

func TestPublicKeyToAddress(t *testing.T) {
	privKey := NewPrivateKey()
	pubKey := privKey.Public()

	address := pubKey.Address()
	assert.Equal(t, addressLen, len(address.Bytes()))
}

func TestNewPrivateKeyFromString(t *testing.T) {
	seed := "59dd435daae31b6ddf058e0f0f2f6ae045d356e5ab5346293f34e0e3c13d0c85"
	privKey := NewPrivateKeyFromSeed(seed)

	assert.Equal(t, len(privKey.Bytes()), privateKeyLen)
	address := privKey.Public().Address()
	assert.Equal(t, len(address.Bytes()), addressLen)

	t.Log(address)

}
