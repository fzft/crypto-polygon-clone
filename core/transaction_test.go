package core

import (
	"bytes"
	"github.com/fzft/crypto-simple-blockchain/crypto"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
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

func RandomTxWithSignature(t *testing.T) *Transaction {
	prvKey := crypto.GeneratePrivateKey()
	tx := &Transaction{
		Data: []byte(strconv.FormatInt(int64(rand.Intn(10000000000)), 10)),
	}
	assert.Nil(t, tx.Sign(prvKey))
	return tx
}

func TestNewGobTxDecoder(t *testing.T) {
	tx := RandomTxWithSignature(t)
	buf := &bytes.Buffer{}
	assert.Nil(t, tx.Encode(NewGobTxEncoder(buf)))

	txDecoded := new(Transaction)
	decoder := NewGobTxDecoder(buf)
	assert.Nil(t, txDecoded.Decode(decoder))
	assert.Equal(t, tx, txDecoded)
}
