package core

import (
	"fmt"
)


type Blockchain struct {
	store Storage
	headers []*Header
	validator Validator
}

func NewBlockchain(genesis *Block) (*Blockchain, error) {
	bc := &Blockchain{
		headers: []*Header{},
		store: &MemoryStore{},
	}
	bc.validator = NewBlockValidator(bc)
	err := bc.addBlockWithoutValidation(genesis)
	return bc, err
}

func (bc *Blockchain) SetValidator(v Validator) {
	bc.validator = v
}

func (bc *Blockchain) AddBlock(b *Block) error {
	if err := bc.validator.ValidateBlock(b); err != nil {
		return err
	}
	return bc.addBlockWithoutValidation(b)
}

func (bc *Blockchain) GetHeader(height uint32) (*Header, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("height (%d) is too high", height)
	}
	return bc.headers[height], nil
}

func (bc *Blockchain) HasBlock(height uint32) bool {
	return height <= bc.Height()
}

func (bc *Blockchain) Height() uint32 {
	return uint32(len(bc.headers) - 1)
}

func (bc *Blockchain) addGenesisBlock(b *Block) {
}

func (bc *Blockchain) addBlockWithoutValidation(b *Block) error {
	bc.headers = append(bc.headers, b.Header)

	//log.Infof("adding block (%d) with hash (%s)", b.Header.Height, b.Hash(BlockHasher{}))
	return bc.store.Put(b)
}