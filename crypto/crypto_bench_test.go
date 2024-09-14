package crypto

import (
	"crypto/ed25519"
	"testing"
)

func BenchmarkAddress(b *testing.B) {
	pubKey := NewPublicKey(ed25519.NewKeyFromSeed(make([]byte, 32)).Public().(ed25519.PublicKey))

	for i := 0; i < b.N; i++ {
		_ = pubKey.Address()
	}
}

func BenchmarkAddressSlice(b *testing.B) {
	pubKey := NewPublicKeySlice(ed25519.NewKeyFromSeed(make([]byte, 32)).Public().(ed25519.PublicKey))

	for i := 0; i < b.N; i++ {
		_ = pubKey.AddressSlice()
	}
}

func BenchmarkSign(b *testing.B) {
	privateKey := NewPrivateKey()

	msg := make([]byte, 32)
	for i := 0; i < b.N; i++ {
		_ = privateKey.Sign(msg)
	}
}

func BenchmarkSignSlice(b *testing.B) {
	privateKey := NewPrivateKeySlice()

	msg := make([]byte, 32)
	for i := 0; i < b.N; i++ {
		_ = privateKey.signSlice(msg)
	}
}

func BenchmarkVerify(b *testing.B) {
	privateKey := NewPrivateKey()
	pubKey := privateKey.Public()
	msg := make([]byte, 32)
	signature := privateKey.Sign(msg)

	for i := 0; i < b.N; i++ {
		_ = signature.Verify(pubKey, msg)
	}
}

func BenchmarkVerifySlice(b *testing.B) {
	privateKey := NewPrivateKeySlice()
	pubKey := privateKey.PublicSlice()
	msg := make([]byte, 32)
	signature := privateKey.signSlice(msg)

	for i := 0; i < b.N; i++ {
		_ = signature.verifySlice(pubKey, msg)
	}
}

func BenchmarkNewPrivateKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewPrivateKey()
	}
}

func BenchmarkNewPrivateKeySlice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewPrivateKeySlice()
	}
}

func BenchmarkNewPrivateKeyFromSeed(b *testing.B) {
	seed := make([]byte, seedLen)
	for i := 0; i < b.N; i++ {
		_ = NewPrivateKeyFromSeed(seed)
	}
}

func BenchmarkNewPrivateKeyFromSeedSlice(b *testing.B) {
	seed := make([]byte, seedLen)
	for i := 0; i < b.N; i++ {
		_ = newPrivateKeyFromSeedSlice(seed)
	}
}
