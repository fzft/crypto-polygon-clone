package p2p

import (
	"github.com/fzft/crypto-simple-blockchain/core"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestMaxLength(t *testing.T) {
	p := NewTxPool(1)
	p.Add(&core.Transaction{})
	assert.Equal(t, 1, p.all.Count())

	p.Add(&core.Transaction{})
	p.Add(&core.Transaction{})
	p.Add(&core.Transaction{})
	p.Add(&core.Transaction{})

	assert.Equal(t, 1, p.all.Count())

}

func TestTxPoolAdd(t *testing.T) {
	p := NewTxPool(11)
	n := 10
	for i := 1; i < n; i++ {
		tx := &core.Transaction{Data: []byte(strconv.Itoa(i))}
		p.Add(tx)
		p.Add(tx)

		assert.Equal(t, i, p.PendingCount())
		assert.Equal(t, i, p.all.Count())
	}

}

func TestTxPoolMxLength(t *testing.T) {
	maxLength := 10
	p := NewTxPool(maxLength)
	n := 100
	txx := []*core.Transaction{}

	for i := 1; i < n; i++ {
		tx := &core.Transaction{Data: []byte(strconv.Itoa(i))}
		p.Add(tx)
		if i > n-(p.maxLength+1) {
			txx = append(txx, tx)
		}
	}

	assert.Equal(t, maxLength, p.all.Count())
	assert.Equal(t, len(txx), maxLength)

	for _, tx := range txx {
		assert.True(t, p.Contains(tx.Hash(core.TxHasher{})))
	}

}

func TestTxSortedMapFirst(t *testing.T) {
	m := NewTxMapSorter()
	tx := &core.Transaction{Data: []byte{1}}
	m.Add(tx)
	m.Add(&core.Transaction{Data: []byte{1}})
	m.Add(&core.Transaction{Data: []byte{1}})
	m.Add(&core.Transaction{Data: []byte{1}})
	m.Add(&core.Transaction{Data: []byte{1}})

	assert.Equal(t, tx, m.First())

}

func TestTxSortedMapAdd(t *testing.T) {
	m := NewTxMapSorter()
	n := 100
	for i := 0; i < n; i++ {
		tx := &core.Transaction{Data: []byte(strconv.Itoa(i))}
		m.Add(tx)
		assert.Equal(t, i+1, m.Count())
		assert.True(t, m.Contains(tx.Hash(core.TxHasher{})))
		assert.Equal(t, len(m.lookup), m.transactions.Len())
		assert.Equal(t, tx, m.Get(tx.Hash(core.TxHasher{})))
	}

	m.Clear()
	assert.Equal(t, 0, m.Count())
	assert.Equal(t, 0, len(m.lookup))
	assert.Equal(t, 0, m.transactions.Len())
}

func TestTxMapSorterRemove(t *testing.T) {
	m := NewTxMapSorter()

	tx := &core.Transaction{Data: []byte{1}}
	m.Add(tx)
	assert.Equal(t, 1, m.Count())
	m.Remove(tx.Hash(core.TxHasher{}))
	assert.Equal(t, 0, m.Count())
	assert.False(t, m.Contains(tx.Hash(core.TxHasher{})))
}
