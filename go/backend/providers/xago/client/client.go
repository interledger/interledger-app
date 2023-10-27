package client

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/xago"
	"gitlab.com/fynbos/backend/providers/xago/external"
	"gitlab.com/fynbos/backend/providers/xago/ops"
	"gitlab.com/fynbos/backend/wallets"
	temporal "go.temporal.io/sdk/client"
)

type Backends interface {
	DB() *sqlx.DB
	Payments() payments.Client
	Temporal() temporal.Client
	LinkedAccounts() linkedaccounts.Client
	Wallets() wallets.Client
}

var _ ops.Backends = opsBackends{}

type opsBackends struct {
	Backends
	xagoExt external.Client
}

func (o opsBackends) External() external.Client {
	return o.xagoExt
}

var _ xago.Client = &client{}

type client struct {
	b ops.Backends
}

func New(b Backends) xago.Client {
	ex := external.New()

	return &client{b: &opsBackends{
		Backends: b,
		xagoExt:  ex,
	}}
}

func (c *client) WebhookHandler() http.HandlerFunc {
	return ops.EventWebhook(c.b)
}
