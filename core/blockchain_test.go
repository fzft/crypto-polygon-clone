package core

import (
	"github.com/fzft/crypto-simple-blockchain/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func newBlockchainWithGenesis(t *testing.T) *Blockchain {
	bc, err := NewBlockchain(nil, randomBlock( t,0, types.Hash{}))
	assert.Nil(t, err)
	return bc
}

func  TestNewBlockchain(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	assert.NotNil(t, bc.validator)
	assert.Equal(t, uint32(0), bc.Height())
}

func TestHasBlock(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	assert.True(t, bc.HasBlock(0))
}

func TestAddBlock(t *testing.T) {
	bc := newBlockchainWithGenesis(t)

	lenBlocks := 1000
	for i := 0; i < lenBlocks; i++ {
		prevBlockHash := getPrevBlockHash(t, bc, uint32(i + 1))
		b := randomBlock(t,uint32(i + 1), prevBlockHash)
		assert.Nil(t, bc.AddBlock(b))
	}

	assert.Equal(t, bc.Height(), uint32(lenBlocks))
	assert.Equal(t, len(bc.headers), lenBlocks + 1)
	assert.NotNil(t, bc.AddBlock(randomBlockWithSignature(t,78, types.Hash{})))
}

func TestBlockchainGetHeader(t *testing.T) {
		bc := newBlockchainWithGenesis(t)
		lenBlocks := 1000
		for i := 0; i < lenBlocks; i++ {
			prevBlockHash := getPrevBlockHash(t, bc, uint32(i + 1))
			b := randomBlock(t,uint32(i + 1), prevBlockHash)
			assert.Nil(t, bc.AddBlock(b))
			header, err := bc.GetHeader(uint32(i + 1))
			assert.Nil(t, err)
			assert.Equal(t, header, b.Header)
		}
}