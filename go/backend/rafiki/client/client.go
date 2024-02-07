package client

import (
	"context"
	"gitlab.com/fynbos/backend/providers/pti"
	"net/http"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/keys"
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
	Keys() keys.Client
	PTI() pti.Client
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

func (c *client) CreatePaymentPointer(ctx context.Context, w wallets.Wallet, assetCode string) error {
	return ops.CreatePaymentPointer(ctx, c.b, w, assetCode)
}

func (c *client) WebhookHandler() http.HandlerFunc {
	return ops.EventWebhook(c.b)
}

func (c *client) FundOutgoingPayment(ctx context.Context, paymentID string) error {
	return ops.FundOutgoingPayment(ctx, c.b, paymentID)
}

func (c *client) CreatePaymentPointerKey(ctx context.Context, keyID string, walletID string) error {
	return ops.CreatePaymentPointerKey(ctx, c.b, keyID, walletID)
}

func (c *client) RevokePaymentPointerKey(ctx context.Context, keyID string) error {
	return ops.RevokePaymentPointerKey(ctx, c.b, keyID)
}

func (c *client) ListGrants(ctx context.Context, walletID string) ([]rafiki.Grant, error) {
	return ops.ListGrants(ctx, c.b, walletID)
}

func (c *client) GetGrant(ctx context.Context, grantID string) (*rafiki.Grant, error) {
	return ops.GetGrant(ctx, c.b, grantID)
}

func (c *client) RevokeGrant(ctx context.Context, grantID string) error {
	return ops.RevokeGrant(ctx, c.b, grantID)
}
