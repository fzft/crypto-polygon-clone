package p2p

import (
	"github.com/fzft/crypto-simple-blockchain/core"
	"github.com/fzft/crypto-simple-blockchain/types"
	"sync"
)

type TxMapSorter struct {
	lock         sync.RWMutex
	transactions *types.List[*core.Transaction]
	lookup       map[types.Hash]*core.Transaction
}

func NewTxMapSorter() *TxMapSorter {
	return &TxMapSorter{
		transactions: types.NewList[*core.Transaction](),
		lookup:       make(map[types.Hash]*core.Transaction),
	}
}

func (s *TxMapSorter) First() *core.Transaction {
	s.lock.RLock()
	defer s.lock.RUnlock()
	first := s.transactions.Get(0)
	return s.lookup[first.Hash(core.TxHasher{})]
}

func (s *TxMapSorter) Get(h types.Hash) *core.Transaction {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.lookup[h]
}

func (s *TxMapSorter) Add(tx *core.Transaction) {
	hash := tx.Hash(core.TxHasher{})
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.lookup[hash]; !ok {
		s.lookup[hash] = tx
		s.transactions.Insert(tx)
	}
}

func (s *TxMapSorter) Remove(h types.Hash) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.transactions.Remove(s.lookup[h])
	delete(s.lookup, h)
}

func (s *TxMapSorter) Count() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.lookup)
}

func (s *TxMapSorter) Contains(h types.Hash) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, ok := s.lookup[h]
	return ok
}

func (s *TxMapSorter) Clear() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.lookup = make(map[types.Hash]*core.Transaction)
	s.transactions.Clear()
}

type TxPool struct {
	all     *TxMapSorter
	pending *TxMapSorter

	maxLength int
}

func NewTxPool(maxLength int) *TxPool {
	return &TxPool{
		all:       NewTxMapSorter(),
		pending:   NewTxMapSorter(),
		maxLength: maxLength,
	}
}

func (p *TxPool) Add(tx *core.Transaction) {
	if p.all.Count() == p.maxLength {
		oldest := p.all.First()
		p.all.Remove(oldest.Hash(core.TxHasher{}))
	}

	if !p.all.Contains(tx.Hash(core.TxHasher{})) {
		p.all.Add(tx)
		p.pending.Add(tx)
	}
}

func (p *TxPool) Contains(h types.Hash) bool {
	return p.all.Contains(h)
}

func (p *TxPool) Pending() []*core.Transaction {
	return p.pending.transactions.Data
}

func (p *TxPool) ClearPending() {
	p.pending.Clear()
}

func (p *TxPool) PendingCount() int {
	return p.pending.Count()
}

func (p *TxPool) All() []*core.Transaction {
	return p.all.transactions.Data
}

