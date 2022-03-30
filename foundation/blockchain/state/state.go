package state

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ardanlabs/blockchain/foundation/blockchain/accounts"
	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
	"github.com/ardanlabs/blockchain/foundation/blockchain/mempool"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

// ErrNotEnoughTransactions is returned when a block is requested to be created
// and there are not enough transactions.
var ErrNotEnoughTransactions = errors.New("not enough transactions in mempool")

// EventHandler defines a function that is called when events
// occur in the processing of persisting blocks.
type EventHandler func(v string, args ...interface{})

// Config represents the configuration required to start
// the blockchain node.
type Config struct {
	MinerAccount storage.Account
	Host         string
	DBPath       string
	EvHandler    EventHandler
}

// State manages the blockchain database.
type State struct {
	minerAccount storage.Account
	host         string
	dbPath       string

	evHandler EventHandler

	genesis     genesis.Genesis
	storage     *storage.Storage
	mempool     *mempool.Mempool
	accounts    *accounts.Accounts
	latestBlock storage.Block
	mu          sync.Mutex

	worker *worker
}

// New constructs a new blockchain for data management.
func New(cfg Config) (*State, error) {

	// Load the genesis file to get starting balances for
	// founders of the block chain.
	genesis, err := genesis.Load()
	if err != nil {
		return nil, err
	}

	// Access the storage for the blockchain.
	strg, err := storage.New(cfg.DBPath)
	if err != nil {
		return nil, err
	}

	// Load all existing blocks from storage into memory for processing. This
	// won't work in a system like Ethereum.
	blocks, err := strg.ReadAllBlocks()
	if err != nil {
		return nil, err
	}

	// Keep the latest block from the blockchain.
	var latestBlock storage.Block
	if len(blocks) > 0 {
		latestBlock = blocks[len(blocks)-1]
	}

	// Create a new accounts value to manage accounts who transact on
	// the blockchain.
	accounts := accounts.New(genesis)

	// Process the blocks and transactions for each account.
	for _, block := range blocks {
		for _, tx := range block.Transactions {

			// Apply the balance changes based for this transaction.
			accounts.ApplyTransaction(block.Header.MinerAccount, tx)
		}

		// Apply the mining reward for this block.
		accounts.ApplyMiningReward(block.Header.MinerAccount)
	}

	// Construct a mempool with the specified sort strategy.
	mempool, err := mempool.New()
	if err != nil {
		return nil, err
	}

	// Build a safe event handler function for use.
	ev := func(v string, args ...interface{}) {
		if cfg.EvHandler != nil {
			cfg.EvHandler(v, args...)
		}
	}

	// Create the State to provide support for managing the blockchain.
	state := State{
		minerAccount: cfg.MinerAccount,
		host:         cfg.Host,
		dbPath:       cfg.DBPath,
		evHandler:    ev,

		genesis:     genesis,
		storage:     strg,
		mempool:     mempool,
		accounts:    accounts,
		latestBlock: latestBlock,
	}

	// Run the worker which will assign itself to this state.
	runWorker(&state, cfg.EvHandler)

	return &state, nil
}

// Shutdown cleanly brings the node down.
func (s *State) Shutdown() error {

	// Make sure the database file is properly closed.
	defer func() {
		s.storage.Close()
	}()

	// Stop all blockchain writing activity.
	s.worker.shutdown()

	return nil
}

// =============================================================================

// SubmitWalletTransaction accepts a transaction from a wallet for inclusion.
func (s *State) SubmitWalletTransaction(signedTx storage.SignedTx) error {
	if err := s.validateTransaction(signedTx); err != nil {
		return err
	}

	tx := storage.NewBlockTx(signedTx, s.genesis.GasPrice)

	n, err := s.mempool.Upsert(tx)
	if err != nil {
		return err
	}

	if n >= s.genesis.TransPerBlock {
		s.worker.signalStartMining()
	}

	return nil
}

// MineNewBlock writes the published transaction from the memory pool to disk.
func (s *State) MineNewBlock(ctx context.Context) (storage.Block, time.Duration, error) {
	s.evHandler("worker: runMiningOperation: MINING: check mempool count")

	// Are there enough transactions in the pool.
	if s.mempool.Count() < s.genesis.TransPerBlock {
		return storage.Block{}, 0, ErrNotEnoughTransactions
	}

	s.evHandler("worker: runMiningOperation: MINING: create new block: pick %d", s.genesis.TransPerBlock)

	trans := s.mempool.PickBest(2)
	nb := storage.NewBlock(s.minerAccount, s.genesis.Difficulty, s.genesis.TransPerBlock, s.RetrieveLatestBlock(), trans)

	s.evHandler("worker: runMiningOperation: MINING: copy accounts and update")

	// Process the transactions against a copy of the accounts.
	accounts := s.accounts.Clone()
	for _, tx := range nb.Transactions {

		// Apply the balance changes based on this transaction.
		if err := accounts.ApplyTransaction(s.minerAccount, tx); err != nil {
			s.evHandler("worker: runMiningOperation: MINING: WARNING : %s", err)
			continue
		}

		// Update the total gas and tip fees.
		// nb.Header.TotalGas += tx.Gas
		nb.Header.TotalTip += tx.Tip
	}

	// Apply the mining reward for this block.
	accounts.ApplyMiningReward(s.minerAccount)

	s.evHandler("worker: runMiningOperation: MINING: perform POW")

	// Attempt to create a new BlockFS by solving the POW puzzle.
	// This can be cancelled.
	blockFS, duration, err := performPOW(ctx, s.genesis.Difficulty, nb, s.evHandler)
	if err != nil {
		return storage.Block{}, duration, err
	}

	// Just check one more time we were not cancelled.
	if ctx.Err() != nil {
		return storage.Block{}, duration, ctx.Err()
	}

	// I want to make sure all these state changes are done atomically.
	s.mu.Lock()
	defer s.mu.Unlock()
	{
		s.evHandler("worker: runMiningOperation: MINING: write to disk")

		// Write the new block to the chain on disk.
		if err := s.storage.Write(blockFS); err != nil {
			return storage.Block{}, duration, err
		}

		s.evHandler("worker: runMiningOperation: MINING: apply new account updates")

		s.accounts.Replace(accounts)
		s.latestBlock = blockFS.Block

		// Remove the transactions from this block.
		for _, tx := range nb.Transactions {
			s.evHandler("worker: runMiningOperation: MINING: remove from mempool: tx[%s]", tx)
			s.mempool.Delete(tx)
		}
	}

	return blockFS.Block, duration, nil
}

// =============================================================================

// RetrieveMempool returns a copy of the mempool.
func (s *State) RetrieveMempool() []storage.BlockTx {
	return s.mempool.Copy()
}

// RetrieveGenesis returns a copy of the genesis information.
func (s *State) RetrieveGenesis() genesis.Genesis {
	return s.genesis
}

// RetrieveAccounts returns a copy of the set of account information.
func (s *State) RetrieveAccounts() map[storage.Account]accounts.Info {
	return s.accounts.Copy()
}

// RetrieveLatestBlock returns a copy the current latest block.
func (s *State) RetrieveLatestBlock() storage.Block {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.latestBlock
}

// =============================================================================

// QueryMempoolLength returns the current length of the mempool.
func (s *State) QueryMempoolLength() int {
	return s.mempool.Count()
}

// =============================================================================

// validateTransaction takes the signed transaction and validates it has
// a proper signature and other aspects of the data.
func (s *State) validateTransaction(signedTx storage.SignedTx) error {
	if err := signedTx.Validate(); err != nil {
		return err
	}

	return nil
}

// =============================================================================

// MineNextBlock will perform a block creation.
// func (s *State) xMineNextBlock() error {
// 	trans := s.mempool.PickBest(2)
// 	block := storage.NewBlock(s.minerAccount, s.genesis.Difficulty, s.genesis.TransPerBlock, s.latestBlock, trans)

// 	s.evHandler("worker: MineNextBlock: MINING: find hash")

// 	blockFS, _, err := performPOW(context.Background(), s.genesis.Difficulty, block, s.evHandler)
// 	if err != nil {
// 		return err
// 	}

// 	s.evHandler("worker: MineNextBlock: MINING: write block to disk")

// 	// Write the new block to the chain on disk.
// 	if err := s.storage.Write(blockFS); err != nil {
// 		return err
// 	}

// 	s.evHandler("worker: MineNextBlock: MINING: remove trans from mempool")

// 	s.accounts.ApplyMiningReward(s.minerAccount)

// 	for _, tx := range trans {
// 		from, _ := tx.FromAccount()

// 		s.evHandler("worker: MineNextBlock: MINING: UPDATE ACCOUNTS: %s:%d", from, tx.Nonce)
// 		s.accounts.ApplyTransaction(s.minerAccount, tx)

// 		s.evHandler("worker: MineNextBlock: MINING: REMOVE: %s:%d", from, tx.Nonce)
// 		s.mempool.Delete(tx)
// 	}

// 	// Save this as the latest block.
// 	s.latestBlock = blockFS.Block

// 	return nil
// }
