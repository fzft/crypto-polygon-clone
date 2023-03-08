package p2p
//
//import (
//	"crypto/ecdsa"
//	"github.com/ethereum/go-ethereum/crypto"
//	"github.com/ethereum/go-ethereum/p2p"
//	"github.com/ethereum/go-ethereum/p2p/enode"
//	"net"
//	"time"
//
//	"github.com/fzft/crypto-simple-blockchain/log"
//)
//
//type connFlag int32
//
//const (
//	dynDialedConn connFlag = 1 << iota
//	staticDialedConn
//	inboundConn
//	trustedConn
//)
//
//type Message string
//
//type Node struct {
//	srv     *p2p.Server
//	timeout bool
//	// blockP2pProtocolMsgHandler is the handler for block p2p protocol message
//	blockP2pProtocolMsgHandler func(msg Message, peer *p2p.Peer, ws p2p.MsgReadWriter)
//
//	con net.Conn
//	peers []*p2p.Peer
//	// whisper
//	//shh     *whisper.Whisper
//	//symKey  []byte
//	//aSymKey *ecdsa.PrivateKey
//	//topic   whisper.TopicType
//	//filterID string
//}
//
//func NewNode(name, listenAddr string, msgHandler func(msg Message, peer *p2p.Peer, ws p2p.MsgReadWriter), timeout bool) *Node {
//	var noDial bool
//	if name == "node2" {
//		noDial = true
//	}
//
//	//shh := whisper.StandaloneWhisperService(&whisper.DefaultConfig)
//	//asymKeyID, err := shh.NewKeyPair()
//	//if err != nil {
//	//	fmt.Println("Failed to generate new key pair: ", err)
//	//	return nil
//	//}
//	//
//	//privateKey, err := shh.GetPrivateKey(asymKeyID)
//	//if err != nil {
//	//	fmt.Println("Failed to get private key: ", err)
//	//	return nil
//	//}
//
//	server := &p2p.Server{
//		Config: p2p.Config{
//			PrivateKey:  newkey(),
//			MaxPeers:    1,
//			Name:        name,
//			ListenAddr:  listenAddr,
//			NoDial:      noDial,
//			NoDiscovery: true,
//			Protocols:   []p2p.Protocol{discard},
//		},
//	}
//
//	n := &Node{
//		timeout: timeout,
//		srv:     server,
//	}
//
//	return n
//}
//
//func newkey() *ecdsa.PrivateKey {
//	key, err := crypto.GenerateKey()
//	if err != nil {
//		panic("couldn't generate key: " + err.Error())
//	}
//	return key
//}
//
//func (n *Node) msgHandler(peer *p2p.Peer, ws p2p.MsgReadWriter) error {
//	for {
//		msg, err := ws.ReadMsg()
//		if err != nil {
//			log.Errorf("server read msg got error:%v", err)
//			return err
//		}
//		var myMessage [1]Message
//		err = msg.Decode(&myMessage)
//		if err != nil {
//			log.Errorf("msg handler decode message got error:%v", err)
//			continue
//		}
//		log.Infof("server read msg:%v", myMessage)
//		n.blockP2pProtocolMsgHandler(myMessage[0], peer, ws)
//	}
//}
//
//func (n *Node) SyncAddPeer(node *enode.Node) {
//	var (
//		ch      = make(chan *p2p.PeerEvent)
//		sub     = n.srv.SubscribeEvents(ch)
//		timeout = time.After(20 * time.Second)
//	)
//	defer sub.Unsubscribe()
//	n.srv.AddPeer(node)
//	for {
//		select {
//		case ev := <-ch:
//			switch ev.Type {
//			case p2p.PeerEventTypeAdd:
//				if ev.Peer == node.ID() {
//					log.Infof("add peer:%v", ev.Peer)
//					n.con, _ = net.Pipe()
//					n.srv.SetupConn(n.con, 1, node)
//				}
//			case p2p.PeerEventTypeMsgSend:
//				log.Infof("add peer:%v", ev.Peer)
//			case p2p.PeerEventTypeDrop:
//				log.Infof("drop peer:%v", ev.Peer)
//			case p2p.PeerEventTypeMsgRecv:
//				log.Infof("recv peer:%v", ev.Peer)
//			}
//
//		case <-timeout:
//			return
//		}
//	}
//}
//
//func (n *Node) Server() *p2p.Server {
//	return n.srv
//}
//
//func (n *Node) Stop() {
//	if n.con != nil {
//		n.con.Close()
//	}
//	n.srv.Stop()
//}
//
//func (n *Node) Run() {
//	n.srv.Start()
//	log.Printf("start to node: %s, listen on: %v", n.srv.Name, n.srv.ListenAddr)
//	go func() {
//		var connected bool
//		var cnt int
//		for !connected {
//			time.Sleep(500 * time.Millisecond)
//			connected = n.srv.PeerCount() > 0
//			if n.timeout {
//				cnt++
//				if cnt > 1000 {
//					log.Errorf("Timeout expired, failed to connect")
//					n.Stop()
//					return
//				}
//			}
//		}
//
//		for _, peer := range n.srv.Peers() {
//			log.Printf("[NODE: %s] connect to peer [NODE: %s] success ", n.srv.Name, peer.Name())
//		}
//
//
//
//		//symkeyId, err := n.shh.AddSymKeyFromPassword("test")
//		//if err != nil {
//		//	log.Fatalf("Failed to create symmetric key: %v", err)
//		//	return
//		//}
//		//
//		//n.symKey, err = n.shh.GetSymKey(symkeyId)
//		//if err != nil {
//		//	log.Fatalf("Failed to get symmetric key: %v", err)
//		//	return
//		//}
//		//
//		//copy(n.topic[:], common.FromHex("test"))
//		//
//		//filter := whisper.Filter{
//		//	KeySym:   n.symKey,
//		//	//KeyAsym:  n.aSymKey,
//		//	Topics:   [][]byte{n.topic[:]},
//		//	AllowP2P: true,
//		//}
//		//
//		//n.filterID, err = n.shh.Subscribe(&filter)
//		//if err != nil {
//		//	log.Fatalf("Failed to subscribe to filter: %v", err)
//		//	return
//		//}
//		//
//		//go n.messageLoop()
//		//
//		//for {
//		//	time.Sleep(1 * time.Second)
//		//	n.sendMsg([]byte("hello world"))
//		//}
//
//	}()
//
//}
//
////func (n *Node) sendMsg(payload []byte) common.Hash {
////	params := whisper.MessageParams{
////		Src:      n.aSymKey,
////		KeySym:   n.symKey,
////		Payload:  payload,
////		Topic:    n.topic,
////		TTL:      whisper.DefaultTTL,
////		PoW:      whisper.DefaultMinimumPoW,
////		WorkTime: 5,
////	}
////
////	msg, err := whisper.NewSentMessage(&params)
////	if err != nil {
////		log.Fatalf("failed to create new message: %s", err)
////	}
////	envelope, err := msg.Wrap(&params)
////	if err != nil {
////		log.Printf("failed to seal message: %v \n", err)
////		return common.Hash{}
////	}
////
////	err = n.shh.Send(envelope)
////	if err != nil {
////		log.Printf("failed to send message: %v \n", err)
////		return common.Hash{}
////	}
////
////	return envelope.Hash()
////}
////
////func (n *Node) messageLoop() {
////	f := n.shh.GetFilter(n.filterID)
////	if f == nil {
////		utils.Fatalf("filter is not installed")
////	}
////
////	ticker := time.NewTicker(time.Millisecond * 50)
////
////	for {
////		select {
////		case <-ticker.C:
////			messages := f.Retrieve()
////			for _, msg := range messages {
////				n.printMessageInfo(msg)
////			}
////		}
////	}
////}
////
////func (n *Node) printMessageInfo(msg *whisper.ReceivedMessage) {
////	text := string(msg.Payload)
////	timestamp := time.Unix(int64(msg.Sent), 0).Format("2006-01-02 15:04:05")
////	var address common.Address
////	if msg.Src != nil {
////		address = crypto.PubkeyToAddress(*msg.Src)
////	}
////	if whisper.IsPubKeyEqual(msg.Src, &n.aSymKey.PublicKey) {
////		fmt.Printf("\n%s <mine>: %s\n", timestamp, text) // message from myself
////	} else {
////		fmt.Printf("\n%s [%x]: %s\n", timestamp, address, text) // message from a peer
////	}
////}
//
//var discard = p2p.Protocol{
//	Name:   "discard",
//	Length: 1,
//	Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
//		for {
//			msg, err := rw.ReadMsg()
//			if err != nil {
//				log.Fatalf("read msg error: %v", err)
//				return err
//			}
//			log.Printf("discarding %d\n", msg.Code)
//			if err = msg.Discard(); err != nil {
//				return err
//			}
//		}
//	},
//}
