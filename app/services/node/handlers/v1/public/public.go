// Package public maintains the group of handlers for public access.
package public

import (
	"context"
	"net/http"

	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
	"github.com/ardanlabs/blockchain/foundation/web"
	"go.uber.org/zap"
)

// Handlers manages the set of bar ledger endpoints.
type Handlers struct {
	Log   *zap.SugaredLogger
	State *state.State
}

// Test adds new user transactions to the mempool.
func (h Handlers) Genesis(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	gen := h.State.RetrieveGenesis()
	return web.Respond(ctx, w, gen, http.StatusOK)
}

// Accounts returns the current balances for all users.
func (h Handlers) Accounts(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	blkAccounts := h.State.RetrieveAccounts()
	return web.Respond(ctx, w, blkAccounts, http.StatusOK)
}
