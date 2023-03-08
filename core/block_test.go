package core

import (
	"github.com/fzft/crypto-polygon-clone/crypto"
	"github.com/fzft/crypto-polygon-clone/types"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func randomBlock(t *testing.T, height uint32, prevBlockHash types.Hash) *Block {
	tx := randomTxWithSignature(t)
	return NewBlock(&Header{
		Version:       1,
		PrevBlockHash: prevBlockHash,
		Timestamp:     time.Now().UnixNano(),
		Height:        height,
	}, []Transaction{*tx})
}

func randomBlockWithSignature(t *testing.T, height uint32, prevHash types.Hash) *Block {
	prvKey := crypto.GeneratePrivateKey()
	b := randomBlock(t, height, prevHash)
	tx := randomTxWithSignature(t)
	b.AddTransaction(tx)
	assert.Nil(t, b.Sign(prvKey))
	return b
}

func TestHashBlock(t *testing.T) {
	b := randomBlock(t,0, types.Hash{})
	hasher := &BlockHasher{}
	t.Log(b.Hash(hasher))
}

func TestBlockSign(t *testing.T) {
	prvKey := crypto.GeneratePrivateKey()
	b := randomBlock(t,0, types.Hash{})

	assert.Nil(t, b.Sign(prvKey))
	assert.NotNil(t, b.Signature)
}

func TestBlockVerify(t *testing.T) {
	prvKey := crypto.GeneratePrivateKey()
	b := randomBlock(t, 0, types.Hash{})

	assert.Nil(t, b.Sign(prvKey))
	assert.Nil(t, b.Verify())

	otherPrvKey := crypto.GeneratePrivateKey()
	b.Validator = otherPrvKey.PublicKey()

	assert.NotNil(t, b.Verify())

	b.Height = 10
	assert.NotNil(t, b.Verify())
}

func getPrevBlockHash(t *testing.T, bc *Blockchain, height uint32) types.Hash {
	prevHeader, err := bc.GetHeader(height - 1)
	assert.Nil(t, err)
	return BlockHasher{}.Hash(prevHeader)
}