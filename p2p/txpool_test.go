package p2p

import (
	"github.com/fzft/crypto-simple-blockchain/core"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"testing"
)

func TestTxPool(t *testing.T) {
	p := NewTxPool()
	assert.Equal(t, 0, p.Len())
}

func TestPoolAddTx(t *testing.T) {
	p := NewTxPool()
	tx := core.NewTransaction([]byte("foo"))
	assert.Nil(t, p.Add(tx))
	assert.Equal(t, 1, p.Len())
}

func TestSortTransactions(t *testing.T) {
	p := NewTxPool()
	txLen := 1000

	for i := 0; i < txLen; i++ {
		tx := core.NewTransaction([]byte(strconv.FormatInt(int64(i), 10)))
		tx.SetFirstSeen(int64(i * rand.Intn(10000)))
		assert.Nil(t, p.Add(tx))
	}

	assert.Equal(t, txLen, p.Len())

	txx := p.Transactions()
	for i := 0; i < len(txx) -1 ; i++ {
		assert.True(t, txx[i].FirstSeen() <= txx[i+1].FirstSeen())
	}

}