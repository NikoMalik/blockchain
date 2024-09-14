package transaction

import (
	"testing"
	"time"

	"github.com/NikoMalik/blockchain/crypto"
	"github.com/NikoMalik/blockchain/proto"
	"github.com/NikoMalik/blockchain/util"
	"github.com/stretchr/testify/assert"
)

func TestNewTransaction(t *testing.T) {
	privKey := crypto.NewPrivateKey()
	toPrivKey := crypto.NewPrivateKey()
	fromAddress := privKey.Public().Address().Bytes()
	toAddress := toPrivKey.Public().Address().Bytes()
	input := &proto.TxInput{
		PrevTxHash:  util.RandomHash(),
		OutputIndex: 1,
		Signature:   util.RandomHash(),
		PublicKey:   privKey.Public().Bytes(),
	}

	output := &proto.TxOutput{
		Address: toAddress[:],
		Amount:  5,
	}
	output2 := &proto.TxOutput{
		Address: fromAddress[:],
		Amount:  95,
	}
	tx := &proto.Transaction{
		Version:   1,
		Timestamp: time.Now().Unix(),
		Inputs:    []*proto.TxInput{input},
		Outputs:   []*proto.TxOutput{output, output2},
	}
	sign := SignTransaction(privKey, tx)
	t.Log(sign)

	v := VerifyTransaction(tx)
	assert.True(t, v)

}
