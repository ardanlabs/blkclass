// Package public maintains the group of handlers for public access.
package public

import (
	"context"
	"net/http"

	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
	"github.com/ardanlabs/blockchain/foundation/web"
	"go.uber.org/zap"
)

// Handlers manages the set of bar ledger endpoints.
type Handlers struct {
	Log *zap.SugaredLogger
}

// Test adds new user transactions to the mempool.
func (h Handlers) Genesis(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	gen, err := genesis.Load()
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, gen, http.StatusOK)
}
