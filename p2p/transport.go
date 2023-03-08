package p2p

type NetAddr string

type RPC struct {
	From NetAddr
	Payload []byte
}

// Transport tcp., udp
type Transport interface {
	Consume() <-chan RPC
	Connect(Transport) error
	SendMessage(NetAddr, []byte) error
	Addr() NetAddr
}
