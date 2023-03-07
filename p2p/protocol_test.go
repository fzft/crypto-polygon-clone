package p2p

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProtocolToBytes(t *testing.T) {
	p := ProtoMessage{
		Proto:    [128]byte{'M', 'y', 'P', 'r', 'o', 't', 'o', 'c', 'o', 'l'},
		Version: [8]byte{'1','.','0'},
		Event:   [1]byte{1},
		Message: []byte("message") ,
	}

	pp := newProtoMessageProcessor()
	encodeP := pp.encode(p)
	bytesP := pp.decode(encodeP)

	assert.Equal(t, bytesP.Proto, p.Proto)
}
