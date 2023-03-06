package p2p

import (
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"github.com/fzft/crypto-polygon-clone/log"
	"net"
)

type Message struct {
	msg ProtoMessage
	conn *net.UDPConn
}

type HandshakeStep int

const (
	preExchange HandshakeStep = iota
	exchangePubKey
	exchangeSymKey
	exchangeComplete
)

type Server struct {
	name       string
	listenAddr string
	privateKey *rsa.PrivateKey
	publicKey []byte
	conn       *net.UDPConn

	peers    map[string]*Peer
	msgRecv chan Message
	msgSend chan Message
	conns    map[string]net.Conn
	handshakeOk map[string]HandshakeStep
	peerSymkeys map[string][]byte
	peerPubKey map[string][65]byte
	protoMessageProcessor *ProtoMessageProcessor

	symKey []byte
	stopCh chan struct{}
}

func NewServer(name, listenAddr string) *Server {

	privateKey := newkey()
	pubKey, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)

	return &Server{
		name:       name,
		listenAddr: listenAddr,
		privateKey: privateKey,
		publicKey: pubKey,
		msgRecv:   make(chan Message),
		msgSend:   make(chan Message),
		stopCh:     make(chan struct{}),
		peers:      make(map[string]*Peer),
		conns:      make(map[string]net.Conn),
		handshakeOk: make(map[string]HandshakeStep),
		peerPubKey: make(map[string][65]byte),
		protoMessageProcessor: newProtoMessageProcessor(),

		// hard code
		symKey: []byte("example key 1234"),
	}
}

func (srv *Server) Start() {
	addr, err := net.ResolveUDPAddr("udp", srv.listenAddr)
	if err != nil {
		log.Errorf("invalid ip address: %s", srv.listenAddr)
		return
	}

	srv.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		log.Errorf("failed to listen on %s: %v", srv.listenAddr, err)
		return
	}

	realAddr := srv.conn.LocalAddr().(*net.UDPAddr)
	log.Infof("UDP listener up %s\n", realAddr)
	go srv.handleMessage()

	go func() {
		for {
			select {
			case <-srv.stopCh:
				log.Info("server stopped")
				return
			default:
				buf := make([]byte, 1024)
				n, addr, err := srv.conn.ReadFromUDP(buf)
				if err != nil {
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						continue
					}
					if opErr, ok := err.(*net.OpError); ok && opErr.Err != nil && errors.Is(opErr.Err, net.ErrClosed) {
						return
					}
					log.Fatal("Error reading:", err)
					continue
				}
				srv.addPeer(addr.String())
				if !srv.decryptHandshake(buf[:n], srv.conn) {
					continue
				}

				// msgRecv only handle sendMsg event
				srv.msgRecv <- Message{srv.protoMessageProcessor.decode(buf[:n]), srv.conn}
			}
		}
	}()
}

func (srv *Server) handleMessage() {
	for {
		select {
		case message := <-srv.msgRecv:
			symKey := srv.peerSymkeys[message.conn.RemoteAddr().String()]
			deMsg := srv.protoMessageProcessor.decryptMessage(symKey, message.msg)
			log.Infof("received %s", string(deMsg))
		case message := <-srv.msgSend:
			_, err := message.conn.Write(srv.protoMessageProcessor.encode(message.msg))
			if err != nil {
				log.Fatalf("failed to write: %v", err)
				continue
			}
		case <-srv.stopCh:
			log.Info("stop handle message")
			return
		}
	}
}

func (srv *Server) decryptHandshake(rawMsg []byte, conn *net.UDPConn) bool {

	protoMsg := srv.protoMessageProcessor.decode(rawMsg)
	eventType := PeerEventType(protoMsg.Event[0])

	if PeerEventTypeError ==  eventType{
		// invalid message format
		return false
	}


	// check handshake is ok?
	remoteAddr := conn.RemoteAddr().String()
	step, ok := srv.handshakeOk[remoteAddr]
	if !ok {
		step = preExchange
		srv.handshakeOk[remoteAddr] = step
	}

	// exchange public key
	if step == preExchange && eventType == PeerEventTypeBeforeAddPubKey{
		// validate the pub key
		if len(protoMsg.Message) != 65 {
			log.Fatalf("invalid public key")
			return false
		}
		var publicKey [65]byte
		copy(publicKey[:], protoMsg.Message)
		srv.peerPubKey[remoteAddr] = publicKey
		srv.handshakeOk[remoteAddr] = exchangePubKey

		srv.msgSend <- Message{ srv.protoMessageProcessor.encryptAfterAddPubKey(srv.publicKey) , conn}
		return false
	}

	if step == exchangePubKey && eventType == PeerEventTypeAfterAddPubKey {
		// validate the pub key
		if len(protoMsg.Message) != 65 {
			log.Fatalf("invalid public key")
			return false
		}
		var publicKey [65]byte
		copy(publicKey[:], protoMsg.Message)
		srv.peerPubKey[remoteAddr] = publicKey
		srv.handshakeOk[remoteAddr] = exchangeSymKey
		srv.msgSend <- Message{ srv.protoMessageProcessor.encryptBeforeAddSymKey(publicKey[:], srv.symKey) , conn}
		return false
	}

	if step == exchangeSymKey && eventType == PeerEventTypeBeforeAddSymKey {
		publicKey := srv.peerPubKey[remoteAddr]
		peerSymKey := rsaDecrypt(protoMsg.Message, srv.privateKey)
		srv.peerSymkeys[remoteAddr] = peerSymKey
		srv.handshakeOk[remoteAddr] = exchangeComplete
		srv.msgSend <- Message{ srv.protoMessageProcessor.encryptAfterAddSymKey(publicKey[:], srv.symKey) , conn}
		return false
	}

	if step == exchangeSymKey && eventType == PeerEventTypeAfterAddSymKey {
		peerSymKey := rsaDecrypt(protoMsg.Message, srv.privateKey)
		srv.peerSymkeys[remoteAddr] = peerSymKey
		srv.handshakeOk[remoteAddr] = exchangeComplete
		return false
	}

	if step == exchangeComplete && eventType == PeerEventTypeSendMsg {
		return true
	}

	return false

}

// handshake with peer, send symkey with asym crypto process
func (srv *Server) handshake(peer *Peer, message string) {
	conn, err := net.Dial("udp", peer.ListenAddr)
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
		return
	}
	srv.msgSend <- Message{srv.protoMessageProcessor.encryptBeforeAddPubKey(srv.publicKey), conn.(*net.UDPConn)}
	log.Printf("sent %s to %s", message, peer.Name)
}

//func (srv *Server) sendMessage(peer *Peer, message string) {
//	if conn, ok := srv.conns[peer.ListenAddr]; ok {
//		_, err := conn.Write(encrypt(srv.symKey, []byte(message)))
//		if err != nil {
//			log.Fatalf("failed to write: %v", err)
//			return
//		}
//
//		log.Printf("sent %s to %s", message, peer.Name)
//		return
//	}
//
//
//}

func (srv *Server) addPeer(addr string) {
	srv.peers[addr] = NewPeer(addr)
}

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
		ListenAddr: srv.listenAddr,
	}
}
