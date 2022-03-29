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

// ApplyMiningReward gives the specififed account the mining reward.
func (act *Accounts) ApplyMiningReward(minerAccount string) {
	act.mu.Lock()
	defer act.mu.Unlock()

	info := act.info[minerAccount]
	info.Balance += act.genesis.MiningReward

	act.info[minerAccount] = info
}

// ApplyTransaction performs the business logic for applying a transaction
// to the accounts information.
func (act *Accounts) ApplyTransaction(minerAccount string, tx storage.UserTx) error {
	act.mu.Lock()
	defer act.mu.Unlock()

	from := tx.From

	if from == tx.To {
		return fmt.Errorf("invalid transaction, sending money to yourself, from %s, to %s", from, tx.To)
	}

	fromInfo := act.info[from]
	fee := /*tx.Gas*/ +tx.Tip

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

	return nil
}
