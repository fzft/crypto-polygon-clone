package p2p

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/fzft/crypto-simple-blockchain/core"
	"io"
)

type MessageType byte

const (
	MessageTypeTx        MessageType = 0x1
	MessageTypeBlock     MessageType = 0x2
	MessageTypeGetBlocks MessageType = 0x3
	MessageTypeStatus    MessageType = 0x4
	MessageTypeGetStatus MessageType = 0x5
)

type RPC struct {
	From    NetAddr
	Payload io.Reader
}

type Message struct {
	Header MessageType
	Data   []byte
}

func NewMessage(t MessageType, data []byte) *Message {
	return &Message{
		Header: t,
		Data:   data,
	}
}

func (msg *Message) Bytes() []byte {
	buf := &bytes.Buffer{}
	gob.NewEncoder(buf).Encode(msg)
	return buf.Bytes()
}

type DecodedMessage struct {
	From NetAddr
	Data any
}

type RPCDecodeFunc func(RPC) (*DecodedMessage, error)

func DefaultRPCDecodeFunc(rpc RPC) (*DecodedMessage, error) {
	msg := Message{}
	if err := gob.NewDecoder(rpc.Payload).Decode(&msg); err != nil {
		return nil, fmt.Errorf("failed to decode message from %s: %v", rpc.From, err)
	}

	switch msg.Header {
	case MessageTypeTx:
		tx := new(core.Transaction)
		if err := tx.Decode(core.NewGobTxDecoder(bytes.NewReader(msg.Data))); err != nil {
			return nil, err
		}
		return &DecodedMessage{
			From: rpc.From,
			Data: tx,
		}, nil
	case MessageTypeBlock:
		block := new(core.Block)
		if err := block.Decode(core.NewGobBlockDecoder(bytes.NewReader(msg.Data))); err != nil {
			return nil, err
		}
		return &DecodedMessage{
			From: rpc.From,
			Data: block,
		}, nil
	case MessageTypeGetStatus:
		getStatusMessage := new(GetStatusMessage)
		if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(&GetStatusMessage{}); err != nil {
			return nil, fmt.Errorf("failed to decode get status message from %s: %v", rpc.From, err)
		}
		return &DecodedMessage{
			From: rpc.From,
			Data: getStatusMessage,
		}, nil
	case MessageTypeStatus:
		status := new(StatusMessage)
		if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(status); err != nil {
			return nil, fmt.Errorf("failed to decode status message from %s: %v", rpc.From, err)
		}
		return &DecodedMessage{
			From: rpc.From,
			Data: status,
		}, nil

	default:
		return nil, fmt.Errorf("invalid message type %v", msg.Header)
	}
}

type RPCHandler interface {
	HandleRPC(RPC) error
}

type DefaultRPCHandler struct {
	p RPCProcessor
}

func NewDefaultRPCHandler(p RPCProcessor) *DefaultRPCHandler {
	return &DefaultRPCHandler{p: p}
}

type RPCProcessor interface {
	ProcessMessage(message *DecodedMessage) error
}
