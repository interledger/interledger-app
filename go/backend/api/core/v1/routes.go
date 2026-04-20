package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
)

type Backends interface {
	Users() user.Client
	Wallets() wallets.Client
	Gatehub() gatehub.Client
}

type handlers struct {
	backends Backends
}

func NewRouter(b Backends) http.Handler {
	r := chi.NewRouter()
	h := &handlers{backends: b}

	r.Route("/statements", func(r chi.Router) {
		r.Get("/account", h.getAccountStatement)
	})

	return r
}
