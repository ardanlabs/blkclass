// Package public maintains the group of handlers for public access.
package public

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
	"github.com/ardanlabs/blockchain/foundation/nameservice"
	"github.com/ardanlabs/blockchain/foundation/web"
	"go.uber.org/zap"
)

// Handlers manages the set of bar ledger endpoints.
type Handlers struct {
	Log   *zap.SugaredLogger
	State *state.State
	NS    *nameservice.NameService
}

// SubmitWalletTransaction adds new user transactions to the mempool.
func (h Handlers) SubmitWalletTransaction(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	var signedTx storage.SignedTx
	if err := web.Decode(r, &signedTx); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	from, err := signedTx.FromAccount()
	if err != nil {
		return fmt.Errorf("unable to get from account address: %w", err)
	}

	h.Log.Infow("add user tran", "traceid", v.TraceID, "nonce", signedTx.Nonce, "from", from, "value", signedTx.Value, "to", signedTx.To, "value", signedTx.Value, "tip", signedTx.Tip)
	if err := h.State.SubmitWalletTransaction(signedTx); err != nil {
		return err
	}

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

	acts := make([]info, 0, len(blkAccounts))
	for account, blkInfo := range blkAccounts {
		act := info{
			Account: account,
			Name:    h.NS.Lookup(account),
			Balance: blkInfo.Balance,
		}
		acts = append(acts, act)
	}

	ai := actInfo{
		Uncommitted: len(h.State.RetrieveMempool()),
		Accounts:    acts,
	}

	return web.Respond(ctx, w, ai, http.StatusOK)
}
