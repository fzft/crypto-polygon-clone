package p2p

import (
	"bytes"
	"crypto/md5"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"github.com/fzft/crypto-simple-blockchain/core"
	"github.com/fzft/crypto-simple-blockchain/crypto"
	"github.com/fzft/crypto-simple-blockchain/log"
	"net"
	"time"
)

//type Message struct {
//	msg ProtoMessage
//	conn *net.UDPConn
//	addr *net.UDPAddr
//}

const DefaultBlockTime = 5 * time.Second

type HandshakeStep int

const (
	preExchange HandshakeStep = iota
	exchangePubKey
	exchangeSymKey
	exchangeComplete
)

type ServerOpts struct {
	name       string
	listenAddr string

	Transports    []Transport
	PrivateKey    *crypto.PrivateKey
	BlockTime     time.Duration
	RPCDecodeFunc RPCDecodeFunc
	RPCProcessor  RPCProcessor
}

type Server struct {
	ServerOpts
	isValidator bool
	memPool     *TxPool
	blockTime   time.Duration

	uuid       [16]byte
	name       string
	listenAddr string
	realAddr   string
	privateKey *rsa.PrivateKey
	publicKey  []byte
	conn       *net.UDPConn
	hmac       string

	peers                 map[[16]byte]*Peer
	msgRecv               chan Message
	msgSend               chan Message
	msgPreConSend         chan Message
	conns                 map[[16]byte]net.Conn
	handshakeOk           map[[16]byte]HandshakeStep
	peerSymkeys           map[[16]byte][]byte
	peerPubKey            map[[16]byte][294]byte
	protoMessageProcessor *ProtoMessageProcessor
	setConn               map[[16]byte]chan error

	mDns *mDns

	symKey []byte
	stopCh chan struct{}

	rpcCh chan RPC
}

func NewServer(opts ServerOpts) *Server {
	if opts.BlockTime == 0 {
		opts.BlockTime = DefaultBlockTime
	}

	if opts.RPCDecodeFunc == nil {
		opts.RPCDecodeFunc = DefaultRPCDecodeFunc
	}

	privateKey := newkey()
	pubKey, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	uuid := md5.Sum(pubKey)
	s := &Server{
		ServerOpts:            opts,
		isValidator:           opts.PrivateKey != nil,
		blockTime:             opts.BlockTime,
		memPool:               NewTxPool(),
		name:                  opts.name,
		uuid:                  uuid,
		privateKey:            privateKey,
		listenAddr:            opts.listenAddr,
		publicKey:             pubKey,
		msgRecv:               make(chan Message),
		msgSend:               make(chan Message),
		msgPreConSend:         make(chan Message),
		stopCh:                make(chan struct{}),
		peers:                 make(map[[16]byte]*Peer),
		conns:                 make(map[[16]byte]net.Conn),
		handshakeOk:           make(map[[16]byte]HandshakeStep),
		peerSymkeys:           make(map[[16]byte][]byte),
		setConn:               make(map[[16]byte]chan error),
		peerPubKey:            make(map[[16]byte][294]byte),
		protoMessageProcessor: newProtoMessageProcessor(),
		rpcCh:                 make(chan RPC),

		// hard code
		symKey: []byte("example key 1234"),
	}

	if opts.RPCProcessor == nil {
		opts.RPCProcessor = s
	}

	s.ServerOpts = opts

	return s
}

func (srv *Server) initTransports() {
	for _, tr := range srv.Transports {
		go func(tr Transport) {
			for rpc := range tr.Consume() {
				srv.rpcCh <- rpc
			}
		}(tr)
	}
}

func (srv *Server) Start() {
	srv.initTransports()
	ticker := time.NewTicker(srv.blockTime)

LOOP:
	for {
		select {
		case rpc := <-srv.rpcCh:
			msg, err := srv.RPCDecodeFunc(rpc)
			if err != nil {
				log.Errorf("failed to decode rpc: %v", err)
				continue
			}

			if err = srv.RPCProcessor.ProcessMessage(msg); err != nil {
				log.Error(err)
			}

		case <-srv.stopCh:
			break LOOP
		case <-ticker.C:
			if srv.isValidator {
				srv.createNewBlock()
			}
		}
	}

	//addr, err := net.ResolveUDPAddr("udp", srv.listenAddr)
	//if err != nil {
	//	log.Errorf("invalid ip address: %s", srv.listenAddr)
	//	return
	//}
	//
	//srv.conn, err = net.ListenUDP("udp", addr)
	//if err != nil {
	//	log.Errorf("failed to listen on %s: %v", srv.listenAddr, err)
	//	return
	//}
	//
	//realAddr := srv.conn.LocalAddr().(*net.UDPAddr)
	//srv.realAddr = realAddr.String()
	//srv.mDns = newMDns(srv.realAddr)
	//log.Infof("UDP listener up %s\n", realAddr)
	//go srv.handleMessage()
	//
	//go func() {
	//	for {
	//		select {
	//		case <-srv.stopCh:
	//			log.Info("server stopped")
	//			return
	//		default:
	//			buf := make([]byte, 1024)
	//			n, addr, err := srv.conn.ReadFromUDP(buf)
	//			log.Infof("srv %s recv msg : %d", srv.listenAddr, n)
	//
	//			if err != nil {
	//				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
	//					continue
	//				}
	//				if opErr, ok := err.(*net.OpError); ok && opErr.Err != nil && errors.Is(opErr.Err, net.ErrClosed) {
	//					return
	//				}
	//				log.Fatal("Error reading:", err)
	//				continue
	//			}
	//			if !srv.doHandshake(buf[:n], addr) {
	//				continue
	//			}
	//
	//			// msgRecv only handle sendMsg event
	//			srv.msgRecv <- Message{srv.protoMessageProcessor.decode(buf[:n]), srv.conn, addr}
	//		}
	//	}
	//}()
}

func (srv *Server) ProcessMessage(msg *DecodedMessage) error {
	switch t := msg.Data.(type) {
	case *core.Transaction:
		return srv.ProcessTransaction(t)
	default:
		return fmt.Errorf("invalid message type: %s", t)
	}
}

func (srv *Server) broadcast(msg []byte) error {
	for _, t := range srv.Transports {
		if err := t.Broadcast(msg); err != nil {
			return err
		}
	}
	return nil
}

func (srv *Server) ProcessTransaction(tx *core.Transaction) error {
	hash := tx.Hash(core.TxHasher{})
	if srv.memPool.Has(hash) {
		return fmt.Errorf("transaction already exists in mempool: %s", hash)
	}
	if err := tx.Verify(); err != nil {
		return err
	}

	tx.SetFirstSeen(time.Now().UnixNano())

	go srv.broadcastTx(tx)

	return srv.memPool.Add(tx)
}

func (srv *Server) broadcastTx(tx *core.Transaction) error {
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeTx, buf.Bytes())
	return srv.broadcast(msg.Bytes())
}

func (srv *Server) createNewBlock() error {
	log.Info("create new block")
	return nil
}

//func (srv *Server) handleMessage() {
//	for {
//		select {
//		case message := <-srv.msgRecv:
//			symKey := srv.peerSymkeys[message.msg.Uuid]
//			deMsg := srv.protoMessageProcessor.decryptMessage(symKey, message.msg)
//			log.Infof("received %s", string(deMsg))
//		case message := <-srv.msgSend:
//			log.Infof("srv %s send msg", srv.name)
//			_, err := message.conn.Write(srv.protoMessageProcessor.encode(message.msg))
//			if err != nil {
//				log.Fatalf("failed to write: %v", err)
//			}
//		case message := <-srv.msgPreConSend:
//			log.Infof("srv %s send pre conn msg to %s", srv.name, message.addr.String())
//			_, err := message.conn.WriteToUDP(srv.protoMessageProcessor.encode(message.msg), message.addr)
//			if err != nil {
//				log.Fatalf("failed to write: %v", err)
//			}
//		case <-srv.stopCh:
//			log.Info("stop handle message")
//			return
//		}
//	}
//}

//func (srv *Server) doHandshake(rawMsg []byte, addr *net.UDPAddr) bool {
//
//	protoMsg := srv.protoMessageProcessor.decode(rawMsg)
//	eventType := PeerEventType(protoMsg.Event[0])
//	remoteUUID := protoMsg.Uuid
//	ipv4 := string(protoMsg.Ipv4)
//
//
//
//	if PeerEventTypeError ==  eventType{
//		// invalid message format
//		return false
//	}
//
//	// check handshake is ok?
//	step, ok := srv.handshakeOk[remoteUUID]
//
//	log.Infof("recv handshake msg on %s : %d, and current step %d", srv.name, eventType, step )
//
//	if !ok {
//		step = preExchange
//		srv.handshakeOk[remoteUUID] = step
//	}
//
//	if _, ok = srv.setConn[remoteUUID]; !ok {
//		srv.setConn[remoteUUID] = make(chan error)
//	}
//
//	var remoteCon net.Conn
//	var err error
//	if remoteCon, ok = srv.conns[remoteUUID]; !ok {
//		remoteCon, err = net.Dial("udp", ipv4)
//		if err!=nil {
//			err = fmt.Errorf("failed to dial %s", ipv4)
//			srv.setConn[remoteUUID] <- err
//			return false
//		}
//		srv.conns[remoteUUID] = remoteCon
//	}
//
//	// exchange public key
//	if step == preExchange && eventType == PeerEventTypeBeforeAddPubKey{
//		// validate the pub key
//
//		if len(protoMsg.Message) != 294 {
//			err = fmt.Errorf("invalid public key")
//			srv.setConn[remoteUUID] <- err
//			return false
//		}
//		var publicKey [294]byte
//		copy(publicKey[:], protoMsg.Message)
//		srv.peerPubKey[remoteUUID] = publicKey
//		srv.handshakeOk[remoteUUID] = exchangePubKey
//		srv.msgSend <- Message{ srv.protoMessageProcessor.encryptAfterAddPubKey(srv.publicKey, srv.uuid, []byte(srv.realAddr)) , remoteCon.(*net.UDPConn), addr}
//		return false
//	}
//
//	if step == exchangePubKey && eventType == PeerEventTypeAfterAddPubKey {
//		// validate the pub key
//		if len(protoMsg.Message) != 294 {
//			err = fmt.Errorf("invalid public key")
//			srv.setConn[remoteUUID] <- err
//			return false
//		}
//		var publicKey [294]byte
//		copy(publicKey[:], protoMsg.Message)
//		srv.peerPubKey[remoteUUID] = publicKey
//		srv.handshakeOk[remoteUUID] = exchangeSymKey
//		srv.msgSend <- Message{ srv.protoMessageProcessor.encryptBeforeAddSymKey(publicKey[:], srv.symKey, srv.uuid, []byte(srv.realAddr)) , remoteCon.(*net.UDPConn), addr}
//		return false
//	}
//
//	if step == exchangePubKey && eventType == PeerEventTypeBeforeAddSymKey {
//		publicKey := srv.peerPubKey[remoteUUID]
//		peerSymKey := rsaDecrypt(protoMsg.Message, srv.privateKey)
//		srv.peerSymkeys[remoteUUID] = peerSymKey
//		srv.handshakeOk[remoteUUID] = exchangeComplete
//		srv.msgSend <- Message{ srv.protoMessageProcessor.encryptAfterAddSymKey(publicKey[:], srv.symKey, srv.uuid, []byte(srv.realAddr)) , remoteCon.(*net.UDPConn), addr}
//		return false
//	}
//
//	if step == exchangeSymKey && eventType == PeerEventTypeAfterAddSymKey {
//		peerSymKey := rsaDecrypt(protoMsg.Message, srv.privateKey)
//		srv.peerSymkeys[remoteUUID] = peerSymKey
//		srv.handshakeOk[remoteUUID] = exchangeComplete
//		srv.setConn[remoteUUID] <- nil
//		return false
//	}
//
//	if step == exchangeComplete && eventType == PeerEventTypeSendMsg {
//		return true
//	}
//
//	return false
//
//}

//func (srv *Server) AddPeer(peer *Peer) error {
//	conn, err := net.Dial("udp", peer.ListenAddr)
//	if err != nil {
//		log.Fatalf("failed to dial: %v", err)
//		return err
//	}
//
//	srv.conns[peer.UUID] = conn
//	srv.handshakeOk[peer.UUID] = exchangePubKey
//	srv.msgSend <- Message{srv.protoMessageProcessor.encryptBeforeAddPubKey(srv.publicKey, srv.uuid, []byte(srv.realAddr)), conn.(*net.UDPConn), nil}
//
//	if _, ok := srv.setConn[peer.UUID]; !ok {
//		srv.setConn[peer.UUID] = make(chan error)
//	}
//
//	select {
//		case err = <-srv.setConn[peer.UUID]:
//			if err!=nil {
//				log.Fatalf("failed to setup conn: %v", err)
//				return err
//			} else {
//				log.Printf("setup conn success to %s", peer.Name)
//			}
//		case <-time.After(time.Second *5):
//			log.Fatal("failed to add peer: timeout")
//	}
//	return nil
//}

func (srv *Server) Broadcast() {
	srv.mDns.udpBroadcast(12345)
}

//func (srv *Server) sendMessage(peer *Peer, message string) {
//	if conn, ok := srv.conns[peer.UUID]; ok {
//		srv.msgSend <- Message{srv.protoMessageProcessor.encryptSendMessage(message, srv.peerSymkeys[peer.UUID], srv.uuid, []byte(srv.realAddr)), conn.(*net.UDPConn), nil}
//		log.Printf("sent %s to %s", message, peer.Name)
//		return
//	}
//}

//func (srv *Server) addPeer(addr string) {
//	srv.peers[addr] = NewPeer(addr)
//}

func (srv *Server) Stop() {
	close(srv.stopCh)
	if srv.conn != nil {
		srv.conn.Close()
	}

	for _, conn := range srv.conns {
		if conn != nil {
			conn.Close()
		}
	}
}

func (srv *Server) Self() *Peer {
	return &Peer{
		Name:       srv.name,
		ListenAddr: srv.realAddr,
		UUID:       srv.uuid,
	}
}
