package core

import (
	"github.com/fzft/crypto-simple-blockchain/crypto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAccountStateTransferNoBalance(t *testing.T) {
	state := NewAccountState()
	from := crypto.GeneratePrivateKey().PublicKey().Address()
	to := crypto.GeneratePrivateKey().PublicKey().Address()
	assert.NotNil(t, state.Transfer(from, to, 100))
}

func TestAccountStateTransferSuccess(t *testing.T) {
	state := NewAccountState()
	from := crypto.GeneratePrivateKey().PublicKey().Address()
	to := crypto.GeneratePrivateKey().PublicKey().Address()
	assert.Nil(t, state.AddBalance(from, 100))
	assert.Nil(t, state.Transfer(from, to, 100))
	assert.Equal(t, uint64(0), state.data[from])
	assert.Equal(t, uint64(100), state.data[to])
}
