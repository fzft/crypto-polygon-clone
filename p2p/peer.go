package p2p

type Peer struct {
	Name string
	ListenAddr string
}

func NewPeer(listenAddr string) *Peer {
	return &Peer{
		ListenAddr: listenAddr,
	}
}
