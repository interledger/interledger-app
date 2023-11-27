package client

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/rafiki"
	"gitlab.com/fynbos/backend/rafiki/external"
	"gitlab.com/fynbos/backend/rafiki/ops"
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
	rafikiExt external.Client
}

func (ob opsBackends) External() external.Client {
	return ob.rafikiExt
}

var _ rafiki.Client = &client{}

type client struct {
	b ops.Backends
}

func New(b Backends) rafiki.Client {
	se := external.New()

	return &client{b: &opsBackends{Backends: b, rafikiExt: se}}
}

func (c *client) CreateWalletAddress(ctx context.Context, w wallets.Wallet) error {
	return ops.CreateWalletAddress(ctx, c.b, w)
}

func (c *client) WebhookHandler() http.HandlerFunc {
	return ops.EventWebhook(c.b)
}

func (c *client) FundOutgoingPayment(ctx context.Context, paymentID string) error {
	return ops.FundOutgoingPayment(ctx, c.b, paymentID)
}
