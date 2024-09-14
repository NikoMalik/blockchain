package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

type seedSlice interface {
	~[]byte | ~string
}

const (
	slicePrivateKeyLen = ed25519.PrivateKeySize
	slicePublicKeyLen  = ed25519.PublicKeySize
	sliceSeedLen       = ed25519.SeedSize
	sliceAddressLen    = 20
)

type signatureSlice struct {
	value []byte
}

func (s *signatureSlice) resetSlice() {
	for i := 0; i < len(s.value); i++ {
		s.value[i] = 0
	}
}

type addressSlice struct {
	value []byte
}

func (a *addressSlice) resetSlice() {
	for i := 0; i < len(a.value); i++ {
		a.value[i] = 0
	}
}

func (a *addressSlice) bytesSlice() []byte { return a.value }
func (a *addressSlice) stringSlice() string {
	return hex.EncodeToString(a.value)
}

func (s *signatureSlice) bytesSlice() []byte { return s.value }

func (s *signatureSlice) verifySlice(pubKey *PublicKeySlice, msg []byte) bool {
	return ed25519.Verify(pubKey.key, msg, s.value)
}

type privateKeySlice struct {
	key ed25519.PrivateKey
}

func (p *privateKeySlice) bytesSlice() ed25519.PrivateKey { return p.key }

func (p *privateKeySlice) signSlice(msg []byte) *signatureSlice {
	if msg == nil {
		panic("ed25519: message cannot be nil")
	}

	if len(msg) > 64 {
		panic("ed25519: message too long")
	}
	sig := ed25519.Sign(p.key, msg)

	s := &signatureSlice{value: make([]byte, len(sig))}
	copy(s.value, sig)

	return s
}

func newPrivateKeyFromSeedSlice[T seedSlice](seed T) *privateKeySlice {
	var seedBytes []byte
	switch v := any(seed).(type) {
	case []byte:
		seedBytes = v
	case string:
		var err error
		seedBytes, err = hex.DecodeString(v)
		if err != nil {
			panic(fmt.Sprintf("ed25519: invalid seed string: %v", err))
		}
	default:
		panic("ed25519: unsupported seed type")
	}

	if len(seedBytes) != sliceSeedLen {
		panic(fmt.Sprintf("ed25519: bad seed length: %d", len(seedBytes)))
	}

	return &privateKeySlice{
		key: ed25519.NewKeyFromSeed(seedBytes),
	}
}

func NewPrivateKeySlice() *privateKeySlice {
	seed := make([]byte, sliceSeedLen)
	if _, err := io.ReadFull(rand.Reader, seed); err != nil {
		panic(err)
	}
	privateKey := ed25519.NewKeyFromSeed(seed)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	if len(publicKey) != slicePublicKeyLen {
		panic(fmt.Sprintf("ed25519: bad public key length: %d", len(publicKey)))
	}

	return &privateKeySlice{key: privateKey}
}

func (p *privateKeySlice) PublicSlice() *PublicKeySlice {
	b := make([]byte, slicePublicKeyLen)
	copy(b, p.key[32:])
	return &PublicKeySlice{key: b}
}

type PublicKeySlice struct {
	key ed25519.PublicKey
}

func (p *PublicKeySlice) BytesSlice() ed25519.PublicKey { return p.key }

func (p *PublicKeySlice) AddressSlice() *addressSlice {
	a := &addressSlice{value: make([]byte, sliceAddressLen)}
	copy(a.value, p.key[len(p.key)-sliceAddressLen:])
	return a
}

func NewPublicKeySlice(key ed25519.PublicKey) *PublicKeySlice {
	return &PublicKeySlice{key: key}
}
