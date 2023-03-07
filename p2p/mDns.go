package p2p

import "net"

type mDnsTopic struct {
	Topic string
	Addr  *net.UDPAddr
}

type mDns struct {
	realAddr string
}

func newMDns(realAddr string) *mDns {
	return &mDns{}
}

func (m *mDns) udpBroadcast(port int) {
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(255, 255, 255, 255),
		Port: port,
	})
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	data := []byte("hello")
	_, err = conn.Write(data)
	return
}


