package p2p

import (
	"bytes"
	"encoding/binary"
	"github.com/fzft/crypto-polygon-clone/log"
)

type PeerEventType int8

const (
	PeerEventTypeError PeerEventType = iota
	PeerEventTypeBeforeAddPubKey
	PeerEventTypeAfterAddPubKey
	PeerEventTypeBeforeAddSymKey
	PeerEventTypeAfterAddSymKey
	PeerEventTypeSendMsg
)

type ProtoMessageProcessor struct {
}

var ProtoMessageErr = ProtoMessage{
	Proto:   [128]byte{'M', 'y', 'P', 'r', 'o', 't', 'o', 'c', 'o', 'l'},
	Version: [8]byte{'1', '.', '0'},
	Event:   [1]byte{byte(PeerEventTypeError)},
	Message: []byte("Error"),
}

type ProtoMessage struct {
	Proto   [128]byte
	Version [8]byte
	Event   [1]byte
	Message []byte
}

func newProtoMessageProcessor() *ProtoMessageProcessor {
	return &ProtoMessageProcessor{}
}

func (p *ProtoMessageProcessor) decode(rawMsg []byte) ProtoMessage {
	var decodedData ProtoMessage
	err := binary.Read(bytes.NewReader(rawMsg), binary.LittleEndian, &decodedData)
	if err != nil {
		log.Fatalf("Error decoding data: %v", err)
		return ProtoMessageErr
	}
	return decodedData
}

func (p *ProtoMessageProcessor) encode(msg ProtoMessage) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, msg)
	if err != nil {
		log.Fatalf("binary.Write failed:%v", err)
		return nil
	}
	bytes := buf.Bytes()
	return bytes
}

func (p *ProtoMessageProcessor) decryptMessage(symKey []byte, msg ProtoMessage) []byte {
	decryptMsg := aesDecrypt(symKey, msg.Message)
	return decryptMsg
}

func (p *ProtoMessageProcessor) getPeerEvent(rawMsg []byte) PeerEventType {
	decodedData := p.decode(rawMsg)
	return PeerEventType(decodedData.Event[0])
}

func (p *ProtoMessageProcessor) encryptMessage(sendMsg string, event PeerEventType, symKey []byte) ProtoMessage {
	msg := ProtoMessage{
		Proto:   [128]byte{'M', 'y', 'P', 'r', 'o', 't', 'o', 'c', 'o', 'l'},
		Version: [8]byte{'1', '.', '0'},
		Event:   [1]byte{byte(event)},
		Message: aesEncrypt([]byte(sendMsg), symKey),
	}

	return msg
}

func (p *ProtoMessageProcessor) encryptSendMessage(sendMsg string, symKey []byte) ProtoMessage {
	return p.encryptMessage(sendMsg, PeerEventTypeSendMsg, symKey)
}

func (p *ProtoMessageProcessor) encryptBeforeAddPubKey(pubKey []byte) ProtoMessage {
	msg := ProtoMessage{
		Proto:   [128]byte{'M', 'y', 'P', 'r', 'o', 't', 'o', 'c', 'o', 'l'},
		Version: [8]byte{'1', '.', '0'},
		Event:   [1]byte{byte(PeerEventTypeBeforeAddPubKey)},
		Message: pubKey,
	}
	return msg
}

func (p *ProtoMessageProcessor) encryptAfterAddPubKey(pubKey []byte) ProtoMessage {
	msg := ProtoMessage{
		Proto:   [128]byte{'M', 'y', 'P', 'r', 'o', 't', 'o', 'c', 'o', 'l'},
		Version: [8]byte{'1', '.', '0'},
		Event:   [1]byte{byte(PeerEventTypeAfterAddPubKey)},
		Message: pubKey,
	}
	return msg
}

func (p *ProtoMessageProcessor) encryptAfterAddSymKey(pubKey []byte, symKey []byte) ProtoMessage {
	msg := ProtoMessage{
		Proto:   [128]byte{'M', 'y', 'P', 'r', 'o', 't', 'o', 'c', 'o', 'l'},
		Version: [8]byte{'1', '.', '0'},
		Event:   [1]byte{byte(PeerEventTypeAfterAddSymKey)},
		Message: rsaEncrypt(symKey, pubKey),
	}
	return msg
}

func (p *ProtoMessageProcessor) encryptBeforeAddSymKey(pubKey []byte, symKey []byte) ProtoMessage {
	msg := ProtoMessage{
		Proto:   [128]byte{'M', 'y', 'P', 'r', 'o', 't', 'o', 'c', 'o', 'l'},
		Version: [8]byte{'1', '.', '0'},
		Event:   [1]byte{byte(PeerEventTypeBeforeAddSymKey)},
		Message: rsaEncrypt(symKey, pubKey),
	}
	return msg
}
