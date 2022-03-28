// Package public maintains the group of handlers for public access.
package public

import (
	"context"
	"net/http"

	"github.com/ardanlabs/blockchain/foundation/web"
	"go.uber.org/zap"
)

// Handlers manages the set of bar ledger endpoints.
type Handlers struct {
	Log *zap.SugaredLogger
}

// Test adds new user transactions to the mempool.
func (h Handlers) Test(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	status := struct {
		Status string
	}{
		Status: "OK",
	}
	return web.Respond(ctx, w, status, http.StatusOK)
}
