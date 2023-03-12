package core

import (
	"encoding/gob"
	"fmt"
	"github.com/fzft/crypto-simple-blockchain/crypto"
	"github.com/fzft/crypto-simple-blockchain/types"
	"math/rand"
)

type TxType byte

const (
	TxTypeCollection TxType = iota
	TxTypeMint
)

type CollectionTx struct {
	Fee      int64
	MetaData []byte
}

type MintTx struct {
	Fee             int64
	NFT             types.Hash
	Collection      types.Hash
	MetaData        []byte
	CollectionOwner crypto.PublicKey
	Signature       *crypto.Signature
}

type Transaction struct {
	Type TxType

	//
	TxInner any
	Data    []byte

	From      crypto.PublicKey
	Signature *crypto.Signature
	To        crypto.PublicKey
	Value     uint64

	// cached version
	hash      types.Hash
	firstSeen int64
	Nonce     int64
}

func NewTransaction(data []byte) *Transaction {
	return &Transaction{
		Data:  data,
		Nonce: rand.Int63n(1000000),
	}
}

func (tx *Transaction) Sign(prv crypto.PrivateKey) error {
	sig, err := prv.Sign(tx.Data)
	if err != nil {
		return err
	}

	tx.From = prv.PublicKey()
	tx.Signature = sig
	return nil
}

func (tx *Transaction) Verify() error {
	if tx.Signature == nil {
		return fmt.Errorf("no signature")
	}

	if !tx.Signature.Verify(tx.From, tx.Data) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

func (tx *Transaction) Hash(hasher Hasher[*Transaction]) types.Hash {
	if tx.hash.IsZero() {
		tx.hash = hasher.Hash(tx)
	}
	return tx.hash
}

func (tx *Transaction) Decode(dec Decoder[*Transaction]) error {
	return dec.Decode(tx)
}

func (tx *Transaction) Encode(enc Encoder[*Transaction]) error {
	return enc.Encode(tx)
}

func (tx *Transaction) SetFirstSeen(t int64) {
	tx.firstSeen = t
}

func (tx *Transaction) FirstSeen() int64 {
	return tx.firstSeen
}

func init() {
	gob.Register(CollectionTx{})
	gob.Register(MintTx{})
}
