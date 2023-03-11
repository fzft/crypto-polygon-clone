package p2p

import "github.com/fzft/crypto-simple-blockchain/core"

type StatusMessage struct {
	CurrentHeight uint32
	ID            string
}

type GetStatusMessage struct {
}

type GetBlocksMessage struct {
	From uint32
	To   uint32
}

type BlocksMessage struct {
	Blocks []*core.Block
}
