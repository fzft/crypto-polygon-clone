package core

import (
	"github.com/fzft/crypto-simple-blockchain/crypto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTransactionSign(t *testing.T) {
	prvKey := crypto.GeneratePrivateKey()
	data := []byte("foo")
	tx := &Transaction{Data: data}

	assert.Nil(t, tx.Sign(prvKey))
	assert.NotNil(t, tx.Signature)
}

func TestTransactionVerify(t *testing.T) {
	prvKey := crypto.GeneratePrivateKey()
	data := []byte("foo")
	tx := &Transaction{Data: data}

	assert.Nil(t, tx.Sign(prvKey))
	assert.Nil(t, tx.Verify())

	otherPrvKey := crypto.GeneratePrivateKey()
	tx.PublicKey = otherPrvKey.PublicKey()

	assert.NotNil(t, tx.Verify())

}

func randomTxWithSignature(t *testing.T) *Transaction {
	prvKey := crypto.GeneratePrivateKey()
	tx := &Transaction{
		Data: []byte("foo"),
	}
	assert.Nil(t, tx.Sign(prvKey))
	return tx
}
