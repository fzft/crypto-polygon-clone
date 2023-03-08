package p2p

import (
	"github.com/fzft/crypto-simple-blockchain/core"
	"github.com/fzft/crypto-simple-blockchain/log"
	"github.com/fzft/crypto-simple-blockchain/types"
	"sort"
)

type TxMapSorter struct {
	transactions []*core.Transaction
}

func NewTxMapSorter(txMap map[types.Hash]*core.Transaction) *TxMapSorter {
	txx := make([]*core.Transaction, len(txMap))

	i := 0
	for _, tx := range txMap {
		txx[i] = tx
		i++
	}
	s := &TxMapSorter{transactions: txx}
	sort.Sort(s)
	return s
}

func (s *TxMapSorter) Len() int { return len(s.transactions) }
func (s *TxMapSorter) Swap(i, j int) {
	s.transactions[i], s.transactions[j] = s.transactions[j], s.transactions[i]
}

func (s *TxMapSorter) Less(i, j int) bool {
	return s.transactions[i].FirstSeen() < s.transactions[j].FirstSeen()
}

type TxPool struct {
	transactions map[types.Hash]*core.Transaction
}

func NewTxPool() *TxPool {
	return &TxPool{
		transactions: make(map[types.Hash]*core.Transaction),
	}
}

func (p *TxPool) Len() int {
	return len(p.transactions)
}

func (p *TxPool) Flush() {
	p.transactions = make(map[types.Hash]*core.Transaction)
}

// Has return true if the transaction is in the pool.
func (p *TxPool) Has(hash types.Hash) bool {
	_, ok := p.transactions[hash]
	return ok
}

func (p *TxPool) Get(hash types.Hash) *core.Transaction {
	return p.transactions[hash]
}

// Add adds a transaction to the pool.
func (p *TxPool) Add(tx *core.Transaction) error {
	hash := tx.Hash(core.TxHasher{})
	p.transactions[hash] = tx
	log.Infof("Added transaction %s to pool", hash)
	return nil
}

func (p *TxPool) Transactions() []*core.Transaction {
	s := NewTxMapSorter(p.transactions)
	return s.transactions
}
