package util

import (
	"github.com/fzft/crypto-simple-blockchain/types"
	"math/rand"
)

func RandomBytes(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return b
}

func RandomHash() types.Hash {
	return types.HashFromBytes(RandomBytes(32))
}
