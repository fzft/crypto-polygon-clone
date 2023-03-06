package p2p

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
)

func TestProtocolToBytes(t *testing.T) {
	p := ProtoMessage{
		Proto:    [128]byte{'M', 'y', 'P', 'r', 'o', 't', 'o', 'c', 'o', 'l'},
		Version: [8]byte{'1','.','0'},
		Event:   [1]byte{1},
		Message: []byte("message") ,
	}

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
		return
	}
	bytes := buf.Bytes()
	var bytesName [128]byte
	copy(bytesName[:], bytes)

	t.Log(bytesName)

}
