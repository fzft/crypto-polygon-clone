package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"github.com/fzft/crypto-simple-blockchain/types"
	"math/rand"
)

type Hasher[T any] interface {
	Hash(T) types.Hash
}

type BlockHasher struct {
}

func (BlockHasher) Hash(b *Header) types.Hash {
	h := sha256.Sum256(b.Bytes())
	return h
}

type TxHasher struct {
}

func (TxHasher) Hash(tx *Transaction) types.Hash {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, tx.Data)
	binary.Write(buf, binary.LittleEndian, rand.Intn(10000000000))
	return sha256.Sum256(buf.Bytes())
}
