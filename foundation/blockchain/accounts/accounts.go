package accounts

import (
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
)

// Info represents information stored for an individual account.
type Info struct {
	Balance uint
}

// Accounts manages data related to accounts who have transacted on
// the blockchain.
type Accounts struct {
	genesis genesis.Genesis
	info    map[string]Info
	mu      sync.RWMutex
}

func New(genesis genesis.Genesis) *Accounts {
	accounts := Accounts{
		genesis: genesis,
		info:    make(map[string]Info),
	}

	for account, balance := range genesis.Balances {
		accounts.info[account] = Info{Balance: balance}
	}

	return &accounts
}

// Copy makes a copy of the current information for all accounts.
func (act *Accounts) Copy() map[string]Info {
	act.mu.RLock()
	defer act.mu.RUnlock()

	accounts := make(map[string]Info)
	for account, info := range act.info {
		accounts[account] = info
	}
	return accounts
}
