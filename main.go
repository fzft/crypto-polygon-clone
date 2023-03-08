package main

import (
	"bytes"
	"github.com/ethereum/go-ethereum/log"
	"github.com/fzft/crypto-simple-blockchain/core"
	"github.com/fzft/crypto-simple-blockchain/crypto"
	"github.com/fzft/crypto-simple-blockchain/p2p"
	"math/rand"
	"strconv"
	"time"
)

func main() {

	trLocal := p2p.NewLocalTransport("LOCAL")
	trRemote := p2p.NewLocalTransport("REMOTE")

	trLocal.Connect(trRemote)
	trRemote.Connect(trLocal)

	go func() {
		for {
			if err := sendTransaction(trRemote, trLocal.Addr()); err != nil {
				log.Error(err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	}()

	opts := p2p.ServerOpts{
		Transports: []p2p.Transport{trLocal},
	}

	s := p2p.NewServer(opts)
	s.Start()

}

func sendTransaction(tr p2p.Transport, to p2p.NetAddr) error {
	prvKey := crypto.GeneratePrivateKey()
	data := []byte(strconv.FormatInt(int64(rand.Intn(1000000000000)), 10))
	tx := core.NewTransaction(data)
	tx.Sign(prvKey)
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}
	msg := p2p.NewMessage(p2p.MessageTypeTx, buf.Bytes())
	return tr.SendMessage(to, msg.Bytes())
}
