package p2p

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

type TCPPeer struct {
	conn     net.Conn
	Outgoing bool
}

func (peer *TCPPeer) SendMessage(data []byte) error {
	_, err := peer.conn.Write(data)
	return err

}

func (peer *TCPPeer) readLoop(rpcCh chan RPC) {
	buffer := make([]byte, 4096)
	for {
		n, err := peer.conn.Read(buffer)
		if err == io.EOF {
			continue
		}
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			continue
		}

		msg := buffer[:n]
		rpcCh <- RPC{
			From:    peer.conn.RemoteAddr(),
			Payload: bytes.NewReader(msg),
		}
	}
}

type TCPTransport struct {
	ListenAddr string
	Listener   net.Listener
	peerCh     chan *TCPPeer
}

func NewTCPTransport(addr string, peerCh chan *TCPPeer) *TCPTransport {
	return &TCPTransport{
		ListenAddr: addr,
		peerCh:     peerCh,
	}
}

func (t *TCPTransport) accept() {
	for {
		conn, err := t.Listener.Accept()
		if err != nil {
			fmt.Printf("accept error from %+v\n", conn)
			continue
		}

		peer := &TCPPeer{
			conn: conn,
		}

		t.peerCh <- peer
	}
}

func (t *TCPTransport) Start() error {
	ln, err := net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	t.Listener = ln
	go t.accept()
	return nil
}
