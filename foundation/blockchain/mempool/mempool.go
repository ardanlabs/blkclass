package mempool

import (
	"fmt"
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

// Mempool represents a cache of transactions organized by account:nonce.
type Mempool struct {
	pool map[string]storage.SignedTx
	mu   sync.RWMutex
}

// New constructs a new mempool with specified sort strategy.
func New() (*Mempool, error) {
	mp := Mempool{
		pool: make(map[string]storage.SignedTx),
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
func (mp *Mempool) Upsert(tx storage.SignedTx) (int, error) {
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
func (mp *Mempool) Delete(tx storage.SignedTx) error {
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
func (mp *Mempool) Copy() []storage.SignedTx {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	cpy := []storage.SignedTx{}
	for _, tx := range mp.pool {
		cpy = append(cpy, tx)
	}
	return cpy
}

// PickBest uses the configured sort strategy to return the next set
// of transactions for the next block.
func (mp *Mempool) PickBest(howMany int) []storage.SignedTx {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	cpy := []storage.SignedTx{}
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
func mapKey(tx storage.SignedTx) (string, error) {
	account, err := tx.FromAccount()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%d", account, tx.Nonce), nil
}
