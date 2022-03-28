package state

import (
	"github.com/ardanlabs/blockchain/foundation/blockchain/accounts"
	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
	"github.com/ardanlabs/blockchain/foundation/blockchain/mempool"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

// Config represents the configuration required to start
// the blockchain node.
type Config struct {
	MinerAccount string
	Host         string
	DBPath       string
}

// State manages the blockchain database.
type State struct {
	minerAccount string
	host         string
	dbPath       string

	genesis  genesis.Genesis
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

	// Create a new accounts value to manage accounts who transact on
	// the blockchain.
	accounts := accounts.New(genesis)

	// Construct a mempool with the specified sort strategy.
	mempool, err := mempool.New()
	if err != nil {
		return nil, err
	}

	// Create the State to provide support for managing the blockchain.
	state := State{
		minerAccount: cfg.MinerAccount,
		host:         cfg.Host,
		dbPath:       cfg.DBPath,

		genesis:  genesis,
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
	// txs := s.mempool.PickBest(2)

	// CREATE A BLOCK
	// POW
	// WRITE TO DISK
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
