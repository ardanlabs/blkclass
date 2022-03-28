package state

import (
	"github.com/ardanlabs/blockchain/foundation/blockchain/accounts"
	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
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

	// Create the State to provide support for managing the blockchain.
	state := State{
		minerAccount: cfg.MinerAccount,
		host:         cfg.Host,
		dbPath:       cfg.DBPath,

		genesis:  genesis,
		accounts: accounts,
	}

	return &state, nil
}

// =============================================================================

// RetrieveGenesis returns a copy of the genesis information.
func (s *State) RetrieveGenesis() genesis.Genesis {
	return s.genesis
}

// RetrieveAccounts returns a copy of the set of account information.
func (s *State) RetrieveAccounts() map[string]accounts.Info {
	return s.accounts.Copy()
}
