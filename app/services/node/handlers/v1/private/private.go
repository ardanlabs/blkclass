package private

import (
	"context"
	"net/http"

	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
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

// SubmitNodeTransaction adds new node transactions to the mempool.
func (h Handlers) SubmitNodeTransaction(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	return nil
}

// AddPeerBlock accepts a new mined block from a peer, validates it, then adds it
// to the block chain.
func (h Handlers) AddPeerBlock(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	return nil
}

// Status returns the current status of the node.
func (h Handlers) Status(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	return nil
}

// BlocksByNumber returns all the blocks based on the specified to/from values.
func (h Handlers) BlocksByNumber(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	return nil
}

// Mempool returns the set of uncommitted transactions.
func (h Handlers) Mempool(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	txs := h.State.RetrieveMempool()
	return web.Respond(ctx, w, txs, http.StatusOK)
}
