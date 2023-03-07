package p2p

import (
	"github.com/ethereum/go-ethereum/rlp"
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
	Ipv4 []byte
 	Uuid [16]byte
	Event   [1]byte
	Message []byte
}

func newProtoMessageProcessor() *ProtoMessageProcessor {
	return &ProtoMessageProcessor{}
}

func (p *ProtoMessageProcessor) decode(rawMsg []byte) ProtoMessage {

	var decodedData ProtoMessage
	err := rlp.DecodeBytes(rawMsg, &decodedData)
	if err != nil {
		panic(err)
	}
	//err := binary.Read(bytes.NewReader(rawMsg), binary.LittleEndian, &decodedData)
	//if err != nil {
	//	log.Fatalf("Error decoding data: %v", err)
	//	return ProtoMessageErr
	//}
	return decodedData
}

func (p *ProtoMessageProcessor) encode(msg ProtoMessage) []byte {
	bytes , err := rlp.EncodeToBytes(msg)
	if err !=nil {
		panic(err)
	}
	//buf := new(bytes.Buffer)
	//err := binary.Write(buf, binary.LittleEndian, msg)
	//if err != nil {
	//	log.Fatalf("binary.Write failed:%v", err)
	//	return nil
	//}
	//bytes := buf.Bytes()
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

func (p *ProtoMessageProcessor) encryptMessage(sendMsg string, event PeerEventType, symKey []byte, uuid [16]byte, ipv4 []byte) ProtoMessage {
	msg := ProtoMessage{
		Proto:   [128]byte{'M', 'y', 'P', 'r', 'o', 't', 'o', 'c', 'o', 'l'},
		Version: [8]byte{'1', '.', '0'},
		Ipv4: ipv4,
		Uuid: uuid,
		Event:   [1]byte{byte(event)},
		Message: aesEncrypt(symKey, []byte(sendMsg)),
	}

	return msg
}

func (p *ProtoMessageProcessor) encryptSendMessage(sendMsg string, symKey []byte, uuid [16]byte, ipv4 []byte) ProtoMessage {
	return p.encryptMessage(sendMsg, PeerEventTypeSendMsg, symKey, uuid, ipv4)
}

func (p *ProtoMessageProcessor) encryptBeforeAddPubKey(pubKey []byte, uuid [16]byte, ipv4 []byte) ProtoMessage {
	msg := ProtoMessage{
		Proto:   [128]byte{'M', 'y', 'P', 'r', 'o', 't', 'o', 'c', 'o', 'l'},
		Version: [8]byte{'1', '.', '0'},
		Uuid: uuid,
		Ipv4: ipv4,
		Event:   [1]byte{byte(PeerEventTypeBeforeAddPubKey)},
		Message: pubKey,
	}
	return msg
}

func (p *ProtoMessageProcessor) encryptAfterAddPubKey(pubKey []byte, uuid [16]byte, ipv4 []byte) ProtoMessage {
	msg := ProtoMessage{
		Proto:   [128]byte{'M', 'y', 'P', 'r', 'o', 't', 'o', 'c', 'o', 'l'},
		Version: [8]byte{'1', '.', '0'},
		Uuid: uuid,
		Ipv4: ipv4,
		Event:   [1]byte{byte(PeerEventTypeAfterAddPubKey)},
		Message: pubKey,
	}
	return msg
}

func (p *ProtoMessageProcessor) encryptAfterAddSymKey(pubKey []byte, symKey []byte, uuid [16]byte, ipv4 []byte) ProtoMessage {
	msg := ProtoMessage{
		Proto:   [128]byte{'M', 'y', 'P', 'r', 'o', 't', 'o', 'c', 'o', 'l'},
		Version: [8]byte{'1', '.', '0'},
		Uuid: uuid,
		Ipv4: ipv4,
		Event:   [1]byte{byte(PeerEventTypeAfterAddSymKey)},
		Message: rsaEncrypt(symKey, pubKey),
	}
	return msg
}

func (p *ProtoMessageProcessor) encryptBeforeAddSymKey(pubKey []byte, symKey []byte, uuid [16]byte, ipv4 []byte) ProtoMessage {
	msg := ProtoMessage{
		Proto:   [128]byte{'M', 'y', 'P', 'r', 'o', 't', 'o', 'c', 'o', 'l'},
		Version: [8]byte{'1', '.', '0'},
		Uuid: uuid,
		Ipv4: ipv4,
		Event:   [1]byte{byte(PeerEventTypeBeforeAddSymKey)},
		Message: rsaEncrypt(symKey, pubKey),
	}
	return msg
}
