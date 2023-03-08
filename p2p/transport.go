package p2p

type NetAddr string

// Transport tcp., udp
type Transport interface {
	Consume() <-chan RPC
	Connect(Transport) error
	SendMessage(NetAddr, []byte) error
	Broadcast([]byte) error
	Addr() NetAddr
}
