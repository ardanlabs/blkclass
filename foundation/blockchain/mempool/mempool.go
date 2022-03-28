package mempool

import (
	"fmt"
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

// Mempool represents a cache of transactions organized by account:nonce.
type Mempool struct {
	pool map[string]storage.UserTx
	mu   sync.RWMutex
}

// New constructs a new mempool with specified sort strategy.
func New() (*Mempool, error) {
	mp := Mempool{
		pool: make(map[string]storage.UserTx),
	}

	return &mp, nil
}

// Count returns the current number of transaction in the pool.
func (mp *Mempool) Count() int {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return len(mp.pool)
}

// Upsert adds or replaces a transaction from the mempool.
func (mp *Mempool) Upsert(tx storage.UserTx) (int, error) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	key, err := mapKey(tx)
	if err != nil {
		return 0, err
	}

	mp.pool[key] = tx

	return len(mp.pool), nil
}

// Delete removed a transaction from the mempool.
func (mp *Mempool) Delete(tx storage.UserTx) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	key, err := mapKey(tx)
	if err != nil {
		return err
	}

	delete(mp.pool, key)

	return nil
}

// Copy uses the configured sort strategy to return the next set
// of transactions for the next block.
func (mp *Mempool) Copy() []storage.UserTx {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	cpy := []storage.UserTx{}
	for _, tx := range mp.pool {
		cpy = append(cpy, tx)
	}
	return cpy
}

// PickBest uses the configured sort strategy to return the next set
// of transactions for the next block.
func (mp *Mempool) PickBest(howMany int) []storage.UserTx {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	cpy := []storage.UserTx{}
	for _, tx := range mp.pool {
		cpy = append(cpy, tx)
		if len(cpy) == howMany {
			break
		}
	}

	return cpy
}

// =============================================================================

// mapKey is used to generate the map key.
func mapKey(tx storage.UserTx) (string, error) {
	return fmt.Sprintf("%s:%d", tx.From, tx.Nonce), nil
}
