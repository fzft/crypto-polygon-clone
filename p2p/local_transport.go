package p2p

import (
	"bytes"
	"fmt"
	"net"
	"sync"
)

type LocalTransport struct {
	addr      net.Addr
	consumeCh chan RPC
	lock      sync.RWMutex
	peers     map[net.Addr]*LocalTransport
}

func NewLocalTransport(addr net.Addr) Transport {
	return &LocalTransport{
		addr:      addr,
		consumeCh: make(chan RPC, 1024),
		peers:     make(map[net.Addr]*LocalTransport),
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

func (t *LocalTransport) SendMessage(to net.Addr, data []byte) error {
	t.lock.RLock()
	defer t.lock.RUnlock()

	if to == t.addr {
		return nil
	}

	peer, ok := t.peers[to]
	if !ok {
		return fmt.Errorf("peer %s not found", to)
	}

	peer.consumeCh <- RPC{From: t.addr, Payload: bytes.NewReader(data)}
	return nil
}

func (t *LocalTransport) Addr() net.Addr {
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
