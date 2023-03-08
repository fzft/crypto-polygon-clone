package core

import (
	"bytes"
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

func TestNewGobTxDecoder(t *testing.T) {
	tx := randomTxWithSignature(t)
	buf := &bytes.Buffer{}
	assert.Nil(t, tx.Encode(NewGobTxEncoder(buf)))

	txDecoded := new(Transaction)
	decoder := NewGobTxDecoder(buf)
	assert.Nil(t, txDecoded.Decode(decoder))
	assert.Equal(t, tx, txDecoded)
}