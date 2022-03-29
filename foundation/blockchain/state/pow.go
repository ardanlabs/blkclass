package state

import (
	"context"
	"crypto/rand"
	"math"
	"math/big"
	"time"

	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

// performPOW does the work of mining to find a valid hash for a specified
// block and returns a BlockFS ready to be written to disk.
func performPOW(ctx context.Context, difficulty int, b storage.Block, ev EventHandler) (storage.BlockFS, time.Duration, error) {
	ev("worker: runMiningOperation: MINING: POW: started")
	defer ev("worker: runMiningOperation: MINING: POW: completed")

	for _, tx := range b.Transactions {
		ev("worker: runMiningOperation: MINING: POW: tx[%s]", tx)
	}

	t := time.Now()

	// Choose a random starting point for the nonce.
	nBig, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return storage.BlockFS{}, time.Since(t), ctx.Err()
	}
	b.Header.Nonce = nBig.Uint64()

	var attempts uint64
	for {
		attempts++
		if attempts%1_000_000 == 0 {
			ev("worker: runMiningOperation: MINING: POW: attempts[%d]", attempts)
		}

		// Did we timeout trying to solve the problem.
		if ctx.Err() != nil {
			ev("worker: runMiningOperation: MINING: POW: CANCELLED")
			return storage.BlockFS{}, time.Since(t), ctx.Err()
		}

		// Hash the block and check if we have solved the puzzle.
		hash := b.Hash()
		if !isHashSolved(difficulty, hash) {
			b.Header.Nonce++
			continue
		}

		// Did we timeout trying to solve the problem.
		if ctx.Err() != nil {
			ev("worker: runMiningOperation: MINING: POW: CANCELLED")
			return storage.BlockFS{}, time.Since(t), ctx.Err()
		}

		ev("worker: runMiningOperation: MINING: POW: SOLVED: prevBlk[%s]: newBlk[%s]", b.Header.ParentHash, hash)
		ev("worker: runMiningOperation: MINING: POW: attempts[%d]", attempts)

		// We found a solution to the POW.
		bfs := storage.BlockFS{
			Hash:  hash,
			Block: b,
		}
		return bfs, time.Since(t), nil
	}
}

// isHashSolved checks the hash to make sure it complies with
// the POW rules. We need to match a difficulty number of 0's.
func isHashSolved(difficulty int, hash string) bool {
	const match = "00000000000000000"

	if len(hash) != 64 {
		return false
	}

	return hash[:difficulty] == match[:difficulty]
}
