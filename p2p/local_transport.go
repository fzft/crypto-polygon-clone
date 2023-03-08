package p2p

import (
	"bytes"
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
		consumeCh: make(chan RPC, 1024),
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

func (t *LocalTransport) SendMessage(to NetAddr, data []byte) error {
	t.lock.RLock()
	defer t.lock.RUnlock()

	peer, ok := t.peers[to]
	if !ok {
		return fmt.Errorf("peer %s not found", to)
	}

	peer.consumeCh <- RPC{From: to, Payload: bytes.NewReader(data)}
	return nil
}

func (t *LocalTransport) Addr() NetAddr {
	return t.addr
}

func (t *LocalTransport) Broadcast(data []byte) error {
	for _, peer := range t.peers {
		if err := t.SendMessage(peer.Addr(), data); err != nil {
			return err
		}
	}
	return nil
}
