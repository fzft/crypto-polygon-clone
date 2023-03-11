package main

import (
	"bytes"
	"github.com/ethereum/go-ethereum/log"
	"github.com/fzft/crypto-simple-blockchain/core"
	"github.com/fzft/crypto-simple-blockchain/crypto"
	"github.com/fzft/crypto-simple-blockchain/p2p"
	"net"
	"time"
)

func main() {

	pk := crypto.GeneratePrivateKey()
	localNode := makeServer("LOCAL_NODE", ":3000", &pk, []string{":3001"}, ":9090")
	go localNode.Start()

	remoteNode := makeServer("REMOTE_NODE", ":3001", nil, []string{":3002"}, "")
	go remoteNode.Start()

	remoteNodeB := makeServer("REMOTE_NODE_B", ":3002", nil, []string{}, "")
	go remoteNodeB.Start()

	go func() {
		time.Sleep(7 * time.Second)
		lateNode := makeServer("LATE_NODE", ":3003", nil, []string{":3001"}, "")
		go lateNode.Start()
	}()

	//go tcpTester()
	select {}
}

func tcpTester() {
	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		log.Error(err.Error())
		return
	}

	tx, _ := sendTransaction()
	_, err = conn.Write(tx)
	if err != nil {
		log.Error(err.Error())
	}

}

func makeServer(id string, listenAddr string, pk *crypto.PrivateKey, seedNodes []string, apiListenAddr string) *p2p.Server {
	opts := p2p.ServerOpts{
		PrivateKey:    pk,
		ListenAddr:    listenAddr,
		SeedNodes:     seedNodes,
		ID:            id,
		APIListenAddr: apiListenAddr,
	}

	s, err := p2p.NewServer(opts)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	return s
}

func sendTransaction() ([]byte, error) {
	prvKey := crypto.GeneratePrivateKey()
	data := contract()
	tx := core.NewTransaction(data)
	tx.Sign(prvKey)
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return nil, err
	}
	msg := p2p.NewMessage(p2p.MessageTypeTx, buf.Bytes())
	return msg.Bytes(), nil
}

func contract() []byte {
	pushFoo := []byte{0x4f, 0x0c, 0x4f, 0x0c, 0x46, 0x0c, 0x03, 0x0a, 0x0d, 0xae}
	data := []byte{0x02, 0x0a, 0x03, 0x0a, 0x0b, 0x4f, 0x0c, 0x4f, 0x0c, 0x46, 0x0c, 0x03, 0x0a, 0x0d, 0x0f}
	data = append(data, pushFoo...)
	return data
}
