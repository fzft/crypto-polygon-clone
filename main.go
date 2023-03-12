package main

import (
	"bytes"
	"encoding/json"
	"github.com/ethereum/go-ethereum/log"
	"github.com/fzft/crypto-simple-blockchain/core"
	"github.com/fzft/crypto-simple-blockchain/crypto"
	"github.com/fzft/crypto-simple-blockchain/p2p"
	"github.com/fzft/crypto-simple-blockchain/types"
	"github.com/fzft/crypto-simple-blockchain/util"
	"net/http"
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

	time.Sleep(time.Second * 1)

	if err := sendTransaction(); err != nil {
		panic(err)
	}

	//txSendTicker := time.NewTicker(500 * time.Microsecond)
	//collectionOwnerPrvKey := crypto.GeneratePrivateKey()
	//
	//// 1. Create a collection
	//collectionHash := createCollectionTx(collectionOwnerPrvKey)
	//go func() {
	//	for i := 0; i < 1; i++ {
	//		nftMinter(collectionOwnerPrvKey, collectionHash)
	//		<-txSendTicker.C
	//	}
	//}()

	select {}
}

func sendTransaction() error {
	prvKey := crypto.GeneratePrivateKey()
	toPrvKey := crypto.GeneratePrivateKey()
	tx := core.NewTransaction(nil)
	tx.To = toPrvKey.PublicKey()
	tx.Value = 100
	if err := tx.Sign(prvKey); err != nil {
		return err
	}

	tx.Sign(crypto.GeneratePrivateKey())

	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", "http://localhost:9090/tx", buf)
	if err != nil {
		panic(err)
	}

	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		panic(err)
	}

	return nil
}

func createCollectionTx(prvKey crypto.PrivateKey) types.Hash {
	tx := core.NewTransaction(nil)
	tx.TxInner = core.CollectionTx{
		Fee:      100,
		MetaData: []byte("My first NFT"),
	}

	tx.Sign(prvKey)

	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", "http://localhost:9090/tx", buf)
	if err != nil {
		panic(err)
	}

	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		panic(err)
	}

	return tx.Hash(core.TxHasher{})
}

func nftMinter(prvKey crypto.PrivateKey, collection types.Hash) {
	metaData := map[string]any{
		"Name":   "My first NFT",
		"color":  "red",
		"random": "random",
		"health": 100,
	}

	metaBuf := new(bytes.Buffer)
	if err := json.NewEncoder(metaBuf).Encode(metaData); err != nil {
		panic(err)
	}

	tx := core.NewTransaction(nil)
	tx.TxInner = core.MintTx{
		Fee:             100,
		MetaData:        metaBuf.Bytes(),
		Collection:      collection,
		CollectionOwner: prvKey.PublicKey(),
		NFT:             util.RandomHash(),
	}

	tx.Sign(prvKey)

	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", "http://localhost:9090/tx", buf)
	if err != nil {
		panic(err)
	}

	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		panic(err)
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

func contract() []byte {
	pushFoo := []byte{0x4f, 0x0c, 0x4f, 0x0c, 0x46, 0x0c, 0x03, 0x0a, 0x0d, 0xae}
	data := []byte{0x02, 0x0a, 0x03, 0x0a, 0x0b, 0x4f, 0x0c, 0x4f, 0x0c, 0x46, 0x0c, 0x03, 0x0a, 0x0d, 0x0f}
	data = append(data, pushFoo...)
	return data
}
