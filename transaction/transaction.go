package transaction

import (
	"crypto/sha512"

	"github.com/NikoMalik/blockchain/crypto"
	"github.com/NikoMalik/blockchain/proto"
	pb "google.golang.org/protobuf/proto"
)

func SignTransaction(pk *crypto.PrivateKey, tx *proto.Transaction) *crypto.Signature {
	sign := pk.Sign(HashTransaction(tx))
	tx.Inputs[0].Signature = sign.Bytes()[:]

	return sign

}

func HashTransaction(tx *proto.Transaction) []byte {
	b, err := pb.Marshal(tx)
	if err != nil {
		panic(err)
	}
	hash := sha512.Sum512(b)
	return hash[:]
}
func VerifyTransaction(tx *proto.Transaction) bool {
	if len(tx.Inputs) == 0 {
		return false
	}
	for i := range tx.Inputs {
		pubKey := crypto.NewPublicKey(tx.Inputs[i].PublicKey)
		signature := &crypto.Signature{}

		tx.Inputs[i].Signature = nil

		if !signature.Verify(pubKey, HashTransaction(tx)) {
			return false
		}
	}
	return true
}
