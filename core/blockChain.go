package core

import (
	"fmt"
	"sync"

	"github.com/Kevin2cqs/Go-OriginChain/crypto"
	"github.com/Kevin2cqs/Go-OriginChain/types"
	"github.com/go-kit/log"
)

type BlockChain struct {
	logger log.Logger
	store  Storage

	lock       sync.RWMutex
	headers    []*Header
	blocks     []*Block
	txStore    map[types.Hash]*Transaction
	blockStore map[types.Hash]*Block

	accountState *AccountState

	stateLock     sync.RWMutex
	validator     Validator
	contractState *State
}

func NewBlockChain(l log.Logger, genesis *Block) (*BlockChain, error) {
	accountState := NewAccountState()

	coinbase := crypto.PublicKey{}
	accountState.CreateAccount(coinbase.Address())

	bc := &BlockChain{
		contractState: NewState(),
		headers:       []*Header{},
		store:         NewMemorystore(),
		logger:        l,
		accountState:  accountState,
		blockStore:    make(map[types.Hash]*Block),
		txStore:       make(map[types.Hash]*Transaction),
	}
	bc.validator = NewBlockValidator(bc)
	err := bc.addBlockWithoutValidation(genesis)

	return bc, err
}

func (bc *BlockChain) HasBlock(height uint32) bool {
	return height <= bc.Height()
}

func (bc *BlockChain) Height() uint32 {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	return uint32(len(bc.headers) - 1)
}

func (bc *BlockChain) GetHeader(height uint32) (*Header, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("given height (%d) too high", height)
	}
	bc.lock.Lock()
	defer bc.lock.Unlock()

	return bc.headers[height], nil
}

func (bc *BlockChain) handleNativeTransfer(tx *Transaction) error {
	bc.logger.Log(
		"msg", "handle native token transfer",
		"from", tx.From,
		"to", tx.To,
		"value", tx.Value,
	)
	return bc.accountState.Transfer(tx.From.Address(), tx.To.Address(), tx.Value)
}

func (bc *BlockChain) handleTransaction(tx *Transaction) error {
	if len(tx.Data) > 0 {
		bc.logger.Log("msg", "executing code", "len", len(tx.Data), "hash", tx.Hash(&TxHasher{}))
		vm := NewVM(tx.Data, bc.contractState)
		if err := vm.Run(); err != nil {
			return err
		}
	}

	// if tx.TxInner != nil {
	// 	if err := bc.handleNativeNFT(tx); err != nil {
	// 		return err
	// 	}
	// }

	if tx.Value > 0 {
		if err := bc.handleNativeTransfer(tx); err != nil {
			return err
		}
	}

	return nil
}

func (bc *BlockChain) addBlockWithoutValidation(b *Block) error {
	bc.stateLock.Lock()
	validTransactions := []*Transaction{}
	for i := 0; i < len(b.Transactions); i++ {
		if err := bc.handleTransaction(b.Transactions[i]); err != nil {
			bc.logger.Log("error", err.Error())
		}
		validTransactions = append(validTransactions, b.Transactions[i])
	}
	b.Transactions = validTransactions
	bc.stateLock.Unlock()

	bc.lock.Lock()
	bc.headers = append(bc.headers, b.Header)
	bc.blocks = append(bc.blocks, b)
	bc.blockStore[b.Hash(BlockHasher{})] = b
	for _, tx := range b.Transactions {
		bc.txStore[tx.Hash(TxHasher{})] = tx
	}
	bc.lock.Unlock()

	bc.logger.Log(
		"msg", "new block",
		"hash", b.Hash(BlockHasher{}),
		"height", b.Height,
		"transactions", len(b.Transactions),
	)

	return bc.store.Put(b)
}
