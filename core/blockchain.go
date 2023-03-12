package core

import (
	"fmt"
	"github.com/fzft/crypto-simple-blockchain/types"
	"github.com/go-kit/log"
	"sync"
)

type Blockchain struct {
	logger  log.Logger
	lock    sync.RWMutex
	store   Storage
	blocks  []*Block
	headers []*Header

	accountState *AccountState

	stateLock       sync.RWMutex
	collectionState map[types.Hash]*CollectionTx
	MintState       map[types.Hash]*MintTx
	validator       Validator
	blockStore      map[types.Hash]*Block
	txStore         map[types.Hash]*Transaction

	contractState *State
}

func NewBlockchain(l log.Logger, genesis *Block) (*Blockchain, error) {
	bc := &Blockchain{
		headers:         []*Header{},
		store:           &MemoryStore{},
		logger:          l,
		contractState:   NewState(),
		txStore:         make(map[types.Hash]*Transaction),
		blockStore:      make(map[types.Hash]*Block),
		MintState:       make(map[types.Hash]*MintTx),
		collectionState: make(map[types.Hash]*CollectionTx),
		accountState:    NewAccountState(),
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

	bc.stateLock.RLock()
	defer bc.stateLock.RUnlock()

	for _, tx := range b.Transactions {

		// if we have code, execute it
		if len(tx.Data) > 0 {
			bc.logger.Log("msg", "executing code", "len", len(tx.Data), "hash", tx.Hash(&TxHasher{}))
			vm := NewVM(tx.Data, bc.contractState)
			if err := vm.Run(); err != nil {
				return err
			}

			fmt.Printf("STATE: %+v\n", bc.contractState)
			result := vm.stack.Pop()
			fmt.Printf("RESULT: %+v\n", result)
		}

		// handle native NFT transactions
		if tx.TxInner != nil {
			if err := bc.handleNativeNftTx(tx); err != nil {
				return err
			}
		}

		// handle normal transactions
		if tx.Value > 0 {
			if err := bc.handleNativeTx(tx); err != nil {
				return err
			}
		}
	}

	return bc.addBlockWithoutValidation(b)
}

// handleNativeTx handles native transactions
func (bc *Blockchain) handleNativeTx(tx *Transaction) error {
	bc.logger.Log("msg", "handling native tx", "hash", tx.Hash(&TxHasher{}))

	return bc.accountState.Transfer(tx.From.Address(), tx.To.Address(), tx.Value)

}

// handleNativeNftTx handles native NFT transactions
func (bc *Blockchain) handleNativeNftTx(tx *Transaction) error {
	hash := tx.Hash(&TxHasher{})
	switch t := tx.TxInner.(type) {
	case CollectionTx:
		bc.collectionState[hash] = &t
		bc.logger.Log("msg", "collection tx", "hash", hash)
	case MintTx:
		_, ok := bc.collectionState[t.Collection]
		if !ok {
			return fmt.Errorf("collection not found: %s", t.Collection)
		}
		bc.MintState[hash] = &t
		bc.logger.Log("msg", "mint tx", "hash", hash)
	default:
		return fmt.Errorf("unknown tx type: %T", t)
	}
	return nil
}

func (bc *Blockchain) GetHeader(height uint32) (*Header, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("height (%d) is too high ", height)
	}
	return bc.headers[height], nil
}

func (bc *Blockchain) GetBlock(height uint32) (*Block, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("height (%d) is too high ", height)
	}
	return bc.blocks[height], nil
}

func (bc *Blockchain) HasBlock(height uint32) bool {
	return height <= bc.Height()
}

func (bc *Blockchain) Height() uint32 {
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	return uint32(len(bc.headers) - 1)
}

func (bc *Blockchain) addBlockWithoutValidation(b *Block) error {
	bc.lock.Lock()
	bc.blockStore[b.Hash(BlockHasher{})] = b
	bc.headers = append(bc.headers, b.Header)
	bc.blocks = append(bc.blocks, b)

	for _, tx := range b.Transactions {
		bc.txStore[tx.Hash(&TxHasher{})] = tx
	}
	bc.lock.Unlock()
	bc.logger.Log("msg", "new block",
		"hash", b.Hash(BlockHasher{}),
		"height", b.Height,
		"transactions", len(b.Transactions),
	)
	return bc.store.Put(b)
}

// GetTxByHash returns a transaction by its hash
func (bc *Blockchain) GetTxByHash(hash types.Hash) (*Transaction, error) {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	if tx, ok := bc.txStore[hash]; ok {
		return tx, nil
	}

	return nil, fmt.Errorf("tx not found")
}

// GetBlockByHash returns a block by its hash
func (bc *Blockchain) GetBlockByHash(hash types.Hash) (*Block, error) {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	if block, ok := bc.blockStore[hash]; ok {
		return block, nil
	}

	return nil, fmt.Errorf("block not found")
}
