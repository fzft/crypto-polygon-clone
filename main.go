package main

import (
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/fzft/crypto-simple-blockchain/core"
	"github.com/fzft/crypto-simple-blockchain/crypto"
	"github.com/fzft/crypto-simple-blockchain/p2p"
	"time"
)

func main() {

	trLocal := p2p.NewLocalTransport("LOCAL")
	trRemoteA := p2p.NewLocalTransport("REMOTE_A")
	trRemoteB := p2p.NewLocalTransport("REMOTE_B")
	trRemoteC := p2p.NewLocalTransport("REMOTE_C")

	trLocal.Connect(trRemoteA)
	trRemoteA.Connect(trRemoteB)
	trRemoteB.Connect(trRemoteC)

	trRemoteA.Connect(trLocal)
	initRemoteServers([]p2p.Transport{trRemoteA, trRemoteB, trRemoteC})

	go func() {
		for {
			if err := sendTransaction(trRemoteA, trLocal.Addr()); err != nil {
				log.Error(err.Error())
			}
			time.Sleep(2 * time.Second)
		}
	}()

	//go func() {
	//	time.Sleep(7 * time.Second)
	//
	//	trLate := p2p.NewLocalTransport("LATE_REMOTE")
	//	trRemoteC.Connect(trLate)
	//	lateServer := makeServer("LATE_REMOTE", trLate, nil)
	//	go lateServer.Start()
	//}()

	privateKey := crypto.GeneratePrivateKey()

	localServer := makeServer("LOCAL", trLocal, &privateKey)
	localServer.Start()

}

func initRemoteServers(trs []p2p.Transport) {
	for i := 0; i < len(trs); i++ {
		id := fmt.Sprintf("REMOTE_%d", i)
		s := makeServer(id, trs[i], nil)
		go s.Start()
	}
}

func makeServer(id string, tr p2p.Transport, pk *crypto.PrivateKey) *p2p.Server {
	opts := p2p.ServerOpts{
		PrivateKey: pk,
		ID:         id,
		Transports: []p2p.Transport{tr},
	}

	s, err := p2p.NewServer(opts)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	return s
}

func sendTransaction(tr p2p.Transport, to p2p.NetAddr) error {
	prvKey := crypto.GeneratePrivateKey()
	data := contract()
	tx := core.NewTransaction(data)
	tx.Sign(prvKey)
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}
	msg := p2p.NewMessage(p2p.MessageTypeTx, buf.Bytes())
	return tr.SendMessage(to, msg.Bytes())
}

func contract() []byte {
	pushFoo := []byte{0x4f, 0x0c, 0x4f, 0x0c, 0x46, 0x0c, 0x03, 0x0a, 0x0d, 0xae}
	data := []byte{0x02, 0x0a, 0x03, 0x0a, 0x0b, 0x4f, 0x0c, 0x4f, 0x0c, 0x46, 0x0c, 0x03, 0x0a, 0x0d, 0x0f}
	data = append(data, pushFoo...)
	return data
}
