package accounts

import (
	"fmt"
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

// Info represents information stored for an individual account.
type Info struct {
	Balance uint
}

// Accounts manages data related to accounts who have transacted on
// the blockchain.
type Accounts struct {
	genesis genesis.Genesis
	info    map[storage.Account]Info
	mu      sync.RWMutex
}

func New(genesis genesis.Genesis) *Accounts {
	accounts := Accounts{
		genesis: genesis,
		info:    make(map[storage.Account]Info),
	}

	for account, balance := range genesis.Balances {
		accounts.info[account] = Info{Balance: balance}
	}

	return &accounts
}

// Copy makes a copy of the current information for all accounts.
func (act *Accounts) Copy() map[storage.Account]Info {
	act.mu.RLock()
	defer act.mu.RUnlock()

	accounts := make(map[storage.Account]Info)
	for account, info := range act.info {
		accounts[account] = info
	}
	return accounts
}

// ApplyMiningReward gives the specififed account the mining reward.
func (act *Accounts) ApplyMiningReward(minerAccount storage.Account) {
	act.mu.Lock()
	defer act.mu.Unlock()

	info := act.info[minerAccount]
	info.Balance += act.genesis.MiningReward

	act.info[minerAccount] = info
}

// ApplyTransaction performs the business logic for applying a transaction
// to the accounts information.
func (act *Accounts) ApplyTransaction(minerAccount storage.Account, tx storage.BlockTx) error {
	from, err := tx.FromAccount()
	if err != nil {
		return fmt.Errorf("invalid signature, %s", err)
	}

	act.mu.Lock()
	defer act.mu.Unlock()
	{
		if from == tx.To {
			return fmt.Errorf("invalid transaction, sending money to yourself, from %s, to %s", from, tx.To)
		}

		fromInfo := act.info[from]
		fee := tx.Gas + tx.Tip

		if tx.Value+fee > act.info[from].Balance {
			return fmt.Errorf("%s has an insufficient balance", from)
		}

		toInfo := act.info[tx.To]
		minerInfo := act.info[minerAccount]

		fromInfo.Balance -= tx.Value
		toInfo.Balance += tx.Value

		minerInfo.Balance += fee
		fromInfo.Balance -= fee

		act.info[from] = fromInfo
		act.info[tx.To] = toInfo
		act.info[minerAccount] = minerInfo
	}

	return nil
}
