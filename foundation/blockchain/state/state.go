package state

import (
	"github.com/ardanlabs/blockchain/foundation/blockchain/accounts"
	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
	"github.com/ardanlabs/blockchain/foundation/blockchain/mempool"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

// EventHandler defines a function that is called when events
// occur in the processing of persisting blocks.
type EventHandler func(v string, args ...interface{})

// Config represents the configuration required to start
// the blockchain node.
type Config struct {
	MinerAccount string
	Host         string
	DBPath       string
	EvHandler    EventHandler
}

// State manages the blockchain database.
type State struct {
	minerAccount string
	host         string
	dbPath       string

	evHandler EventHandler

	genesis  genesis.Genesis
	storage  *storage.Storage
	mempool  *mempool.Mempool
	accounts *accounts.Accounts
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

	// Create a new accounts value to manage accounts who transact on
	// the blockchain.
	accounts := accounts.New(genesis)

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

		genesis:  genesis,
		storage:  strg,
		mempool:  mempool,
		accounts: accounts,
	}

	return &state, nil
}

// =============================================================================

// SubmitWalletTransaction accepts a transaction from a wallet for inclusion.
func (s *State) SubmitWalletTransaction(tx storage.UserTx) error {
	n, err := s.mempool.Upsert(tx)
	if err != nil {
		return err
	}

	if n >= s.genesis.TransPerBlock {
		if err := s.MineNextBlock(); err != nil {
			return err
		}
	}

	return nil
}

// MineNextBlock will perform a block creation.
func (s *State) MineNextBlock() error {
	trans := s.mempool.PickBest(2)
	nb := storage.NewBlock(s.minerAccount, s.genesis.Difficulty, s.genesis.TransPerBlock, trans)
	blockFS := storage.BlockFS{
		Hash:  "my hash",
		Block: nb,
	}

	s.evHandler("worker: MineNextBlock: MINING: write block to disk")

	// Write the new block to the chain on disk.
	if err := s.storage.Write(blockFS); err != nil {
		return err
	}

	s.evHandler("worker: MineNextBlock: MINING: remove trans from mempool")

	for _, tx := range trans {
		s.evHandler("worker: MineNextBlock: MINING: REMOVE: %s:%d", tx.From, tx.Nonce)
		s.mempool.Delete(tx)
	}

	//    ------- POW
	// UPDATE ACCOUNTS
	// START RELOAD

	return nil
}

// =============================================================================

// RetrieveMempool returns a copy of the mempool.
func (s *State) RetrieveMempool() []storage.UserTx {
	return s.mempool.Copy()
}

// RetrieveGenesis returns a copy of the genesis information.
func (s *State) RetrieveGenesis() genesis.Genesis {
	return s.genesis
}

// RetrieveAccounts returns a copy of the set of account information.
func (s *State) RetrieveAccounts() map[string]accounts.Info {
	return s.accounts.Copy()
}
