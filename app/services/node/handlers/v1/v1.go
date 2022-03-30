// Package v1 contains the full set of handler functions and routes
// supported by the v1 web api.
package v1

import (
	"net/http"

	"github.com/ardanlabs/blockchain/app/services/node/handlers/v1/public"
	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
	"github.com/ardanlabs/blockchain/foundation/nameservice"
	"github.com/ardanlabs/blockchain/foundation/web"
	"go.uber.org/zap"
)

const version = "v1"

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log   *zap.SugaredLogger
	State *state.State
	NS    *nameservice.NameService
}

// PublicRoutes binds all the version 1 public routes.
func PublicRoutes(app *web.App, cfg Config) {
	pbl := public.Handlers{
		Log:   cfg.Log,
		State: cfg.State,
		NS:    cfg.NS,
	}

	app.Handle(http.MethodGet, version, "/tx/uncommitted/list", pbl.Mempool)
	app.Handle(http.MethodPost, version, "/tx/submit", pbl.SubmitWalletTransaction)
	app.Handle(http.MethodGet, version, "/genesis", pbl.Genesis)
	app.Handle(http.MethodGet, version, "/accounts/list", pbl.Accounts)
}
