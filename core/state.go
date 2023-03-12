package core

import (
	"fmt"
	"github.com/fzft/crypto-simple-blockchain/types"
	"sync"
)

type AccountState struct {
	mu   sync.RWMutex
	data map[types.Address]uint64
}

func NewAccountState() *AccountState {
	return &AccountState{
		data: make(map[types.Address]uint64),
	}
}

func (s *AccountState) AddBalance(to types.Address, amount uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[to] += amount
	return nil
}

// Transfer transfers amount from from to to.
func (s *AccountState) Transfer(from, to types.Address, amount uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[from]; !ok {
		return fmt.Errorf("account (%s) not found", from.String())
	}

	if s.data[from] < amount {
		return fmt.Errorf("insufficient funds (%d < %d)", s.data[from], amount)
	}
	s.data[from] -= amount
	s.data[to] += amount
	return nil
}

func (s *AccountState) GetBalance(to types.Address) (uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.data[to]; !ok {
		return 0, fmt.Errorf("account not found")
	}
	return s.data[to], nil
}

// SubBalance subtracts amount from the account associated with addr.
func (s *AccountState) SubBalance(to types.Address, amount uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[to]; !ok {
		return fmt.Errorf("account (%s) not found", to)
	}

	if s.data[to] < amount {
		return fmt.Errorf("insufficient funds (%d < %d)", s.data[to], amount)
	}
	s.data[to] -= amount
	return nil
}

type State struct {
	data map[string][]byte
}

func NewState() *State {
	return &State{
		data: make(map[string][]byte),
	}
}

func (s *State) Put(k, v []byte) error {
	s.data[string(k)] = v
	return nil
}

func (s *State) Delete(k []byte) error {
	delete(s.data, string(k))
	return nil
}

func (s *State) Get(k []byte) ([]byte, error) {
	if _, ok := s.data[string(k)]; !ok {
		return nil, fmt.Errorf("key not found")
	}
	return s.data[string(k)], nil
}
