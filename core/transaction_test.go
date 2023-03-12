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
	tx.From = otherPrvKey.PublicKey()

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

func TestNFTTransaction(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()
	collectionTx := CollectionTx{
		Fee:      200,
		MetaData: []byte("the beginning of a new collection"),
	}

	tx := &Transaction{
		Type:    TxTypeCollection,
		TxInner: collectionTx,
	}

	tx.Sign(privKey)

	buf := new(bytes.Buffer)
	assert.Nil(t, tx.Encode(NewGobTxEncoder(buf)))

	txDecoded := new(Transaction)
	decoder := NewGobTxDecoder(buf)
	assert.Nil(t, txDecoded.Decode(decoder))
	assert.Equal(t, tx, txDecoded)
}

func TestNativeTransferTransaction(t *testing.T) {
	fromPrvKey := crypto.GeneratePrivateKey()
	toPrvKey := crypto.GeneratePrivateKey()

	tx := &Transaction{
		To:    toPrvKey.PublicKey(),
		Value: 100,
	}

	assert.Nil(t, tx.Sign(fromPrvKey))

}
