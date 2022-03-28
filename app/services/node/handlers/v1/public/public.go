// Package public maintains the group of handlers for public access.
package public

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
	"github.com/ardanlabs/blockchain/foundation/web"
	"go.uber.org/zap"
)

// Handlers manages the set of bar ledger endpoints.
type Handlers struct {
	Log   *zap.SugaredLogger
	State *state.State
}

// SubmitWalletTransaction adds new user transactions to the mempool.
func (h Handlers) SubmitWalletTransaction(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	var userTX storage.UserTx
	if err := web.Decode(r, &userTX); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	h.Log.Infow("add user tran", "traceid", v.TraceID, "nonce", userTX.Nonce, "from", userTX.From, "value", "to", userTX.To, "value", userTX.Value, "tip", userTX.Tip)
	h.State.SubmitWalletTransaction(userTX)

	resp := struct {
		Status string `json:"status"`
	}{
		Status: "WE HAVE THE TRAN",
	}

	return web.Respond(ctx, w, resp, http.StatusOK)
}

// Mempool returns the set of uncommitted transactions.
func (h Handlers) Mempool(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	mempool := h.State.RetrieveMempool()
	return web.Respond(ctx, w, mempool, http.StatusOK)
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
