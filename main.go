package main

import (
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/fzft/crypto-polygon-clone/log"
	myp2p "github.com/fzft/crypto-polygon-clone/p2p"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	stopCh := make(chan struct{})
	term := make(chan os.Signal)

	msgHandler := func(msg myp2p.Message, peer *p2p.Peer, ws p2p.MsgReadWriter) {
		log.Printf("get message %s", msg)
	}

	node1  := myp2p.NewNode("node1", ":30300", msgHandler, true)
	node2  := myp2p.NewNode("node2", ":30301", msgHandler, true)
	node1.Run()
	node2.Run()
	node1.SyncAddPeer(node2.Server().Self())
	go func() {
		signal.Notify(term, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-term
		node1.Stop()
		node2.Stop()
		time.Sleep(1 * time.Second)
		log.Info("start to shut down block chain node")
		stopCh <- struct{}{}
	}()
	<-stopCh
}
