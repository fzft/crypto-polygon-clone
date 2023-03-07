package p2p

type Peer struct {
	Name string
	ListenAddr string
	UUID [16]byte
}

func NewPeer(listenAddr string, uuid [16]byte) *Peer {
	return &Peer{
		ListenAddr: listenAddr,
		UUID: uuid,
	}
}
