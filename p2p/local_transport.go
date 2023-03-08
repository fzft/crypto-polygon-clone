package p2p

import (
	"fmt"
	"sync"
)

type LocalTransport struct {
	addr      NetAddr
	consumeCh chan RPC
	lock      sync.RWMutex
	peers     map[NetAddr]*LocalTransport
}

func NewLocalTransport(addr NetAddr) Transport {
	return &LocalTransport{
		addr:      addr,
		consumeCh: make(chan RPC),
		peers:     make(map[NetAddr]*LocalTransport),
	}
}

func (t *LocalTransport) Consume() <-chan RPC {
	return t.consumeCh
}

func (t *LocalTransport) Connect(transport Transport) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.peers[transport.Addr()] = transport.(*LocalTransport)
	return nil
}

func (t *LocalTransport) SendMessage(addr NetAddr, bytes []byte) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	peer, ok := t.peers[addr]
	if !ok {
		return fmt.Errorf("peer %s not found", addr)
	}

	peer.consumeCh <- RPC{From: addr, Payload: bytes}
	return nil
}

func (t *LocalTransport) Addr() NetAddr {
	return t.addr
}
