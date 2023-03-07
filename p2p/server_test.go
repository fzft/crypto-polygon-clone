package p2p

import (
	"testing"
	"time"
)

func TestServer_Handshake(t *testing.T) {
	s1 := NewServer("s1", ":30031")
	s2 := NewServer("s2", ":30032")

	s1.Start()
	s2.Start()

	err := s1.AddPeer(s2.Self())
	if err!=nil {
		return
	}
	s1.sendMessage(s2.Self(), "hello")
	time.Sleep(5* time.Second)
	s1.Stop()
	s2.Stop()
}

func TestServer_Handshake2(t *testing.T) {
	s1 := NewServer("s1", ":30031")
	s2 := NewServer("s2", ":30032")

	s1.Start()
	s2.Start()

	err := s1.AddPeer(s2.Self())
	if err!=nil {
		return
	}
	s2.sendMessage(s1.Self(), "hello")
	time.Sleep(5* time.Second)
	s1.Stop()
	s2.Stop()
}

func TestServer_mDns(t *testing.T) {
	s1 := NewServer("s1", ":30031")
	s2 := NewServer("s2", ":30032")

	s1.Start()
	s2.Start()
	s1.Broadcast()

	time.Sleep(5* time.Second)
	s1.Stop()
	s2.Stop()
}