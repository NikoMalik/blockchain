package block

import (
	"crypto/sha512"
	"time"

	"github.com/NikoMalik/blockchain/crypto"
	"github.com/NikoMalik/blockchain/proto"
	pb "google.golang.org/protobuf/proto"
)

func SignBlock(pk *crypto.PrivateKey, block *proto.Block) *crypto.Signature {
	return pk.Sign(HashBlock(block))
}

// sha256
func HashBlock(block *proto.Block) []byte {
	b, err := pb.Marshal(block)
	if err != nil {
		panic(err)
	}
	hash := sha512.Sum512(b)
	return hash[:]
}

type Block struct {
	block *proto.Block
}

func NewBlock(transactions []*proto.Transaction, prevHash []byte, height int) *Block {
	return &Block{
		block: &proto.Block{
			Header: &proto.Header{
				PreviousHash: prevHash,
				Timestamp:    time.Now().Unix(),
			}},
	}
}

func (b *Block) AddBlock(p *proto.Block) error {
	return nil
}

func (b *Block) GetBlock() *proto.Block {
	return b.block
}

func (b *Block) GetBlockByHash(hash []byte) (*proto.Block, error) {
	hexHash, err := b.Get(hash)

	if err != nil {
		return nil, err
	}

	return hexHash, nil

}

func (b *Block) GetBlockByHeight(height int) (*proto.Block, error) {
	return nil, nil
}
func (b *Block) Get(bytes []byte) (*proto.Block, error) {
	return nil, nil
}

// func (b *Block) GenerateNonce(prefix []byte) uint32 {

// 	for {

// 		if CheckProofOfWork(prefix, b.Hash()) {
// 			break
// 		}

// 		b.block.Header.Nonce++
// 	}

// 	return b.block.Header.Nonce
// }
