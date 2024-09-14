package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

type Seed interface {
	~[]byte | ~string
}

const (
	privateKeyLen = ed25519.PrivateKeySize
	publicKeyLen  = ed25519.PublicKeySize
	seedLen       = ed25519.SeedSize
	addressLen    = 20
)

type Signature struct {
	value [64]byte
}

func (s *Signature) Reset() {
	for i := 0; i < len(s.value); i++ {
		s.value[i] = 0
	}
}

type Address struct {
	value [addressLen]byte
}

func (a *Address) Reset() {
	for i := 0; i < len(a.value); i++ {
		a.value[i] = 0
	}
}

func (a *Address) Bytes() *[20]byte { return &a.value }
func (a *Address) String() string {
	return hex.EncodeToString(a.value[:])
}

func (s *Signature) Bytes() *[64]byte { return &s.value }

func (s *Signature) Verify(pubKey *PublicKey, msg []byte) bool {
	return ed25519.Verify(pubKey.key, msg, s.value[:])
}

type PrivateKey struct {
	key ed25519.PrivateKey
}

func (p *PrivateKey) Bytes() ed25519.PrivateKey { return p.key }

func (p *PrivateKey) Sign(msg []byte) *Signature {
	if msg == nil {
		panic("ed25519: message cannot be nil")
	}

	if len(msg) > 64 {
		panic("ed25519: message too long")
	}
	sig := ed25519.Sign(p.key, msg)

	s := [64]byte{}
	copy(s[:], sig)

	return &Signature{value: s}
}

func NewPrivateKeyFromSeed[T Seed](seed T) *PrivateKey {
	var seedBytes [seedLen]byte
	switch v := any(seed).(type) {
	case []byte:
		if len(v) != seedLen {
			panic("ed25519: bad seed length")
		}
		copy(seedBytes[:], v)
	case string:
		decoded, err := hex.DecodeString(v)
		if err != nil || len(decoded) != seedLen {
			panic("ed25519: invalid seed string")
		}
		copy(seedBytes[:], decoded)
	default:
		panic("ed25519: unsupported seed type")
	}
	return &PrivateKey{key: ed25519.NewKeyFromSeed(seedBytes[:])}
}

func NewPrivateKey() *PrivateKey {
	seed := make([]byte, seedLen)
	if _, err := io.ReadFull(rand.Reader, seed); err != nil {
		panic(err)
	}
	privateKey := ed25519.NewKeyFromSeed(seed)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	if len(publicKey) != publicKeyLen {
		panic(fmt.Sprintf("ed25519: bad public key length: %d", len(publicKey)))
	}

	return &PrivateKey{key: privateKey}
}

func (p *PrivateKey) Public() *PublicKey {
	var b [publicKeyLen]byte
	copy(b[:], p.key[32:])
	return &PublicKey{key: b[:]}
}

type PublicKey struct {
	key ed25519.PublicKey
}

func (p *PublicKey) Bytes() ed25519.PublicKey { return p.key }

func (p *PublicKey) Address() *Address {
	a := [addressLen]byte{}
	copy(a[:], p.key[len(p.key)-addressLen:])

	return &Address{value: a}
}

func NewPublicKey(key ed25519.PublicKey) *PublicKey {
	if len(key) != publicKeyLen {
		panic(fmt.Sprintf("ed25519: bad public key length: %d", len(key)))
	}
	return &PublicKey{key: key}
}
