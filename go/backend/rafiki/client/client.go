package client

import (
	"context"
	"net/http"

	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/providers/chimoney"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/providers/xago"
	"gitlab.com/fynbos/backend/transactions"

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
	Gatehub() gatehub.Client
	Xago() xago.Client
	Chimoney() chimoney.Client
	KYC() kyc.Client
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

func New(b Backends, signingConfig external.AdminSigningConfig) rafiki.Client {
	se := external.New(signingConfig)

	return &client{b: &opsBackends{Backends: b, rafikiExt: se}}
}

func (c *client) GetWalletAddress(ctx context.Context, walletID string) (*rafiki.WalletAddress, error) {
	return ops.GetWalletAddress(ctx, c.b, walletID)
}

func (c *client) CreatePaymentPointer(ctx context.Context, w wallets.Wallet) (string, error) {
	return ops.CreatePaymentPointer(ctx, c.b, w)
}

func (c *client) WebhookHandler() http.HandlerFunc {
	return ops.EventWebhook(c.b)
}

func (c *client) FundOutgoingPayment(ctx context.Context, paymentID string) error {
	return ops.FundOutgoingPayment(ctx, c.b, paymentID)
}

func (c *client) FinalizeWebMonetization(ctx context.Context, paymentID string) error {
	return ops.FinalizeWebMonetization(ctx, c.b, paymentID)
}

func (c *client) RollbackWebMonetization(ctx context.Context, paymentID string) error {
	return ops.RollbackWebMonetization(ctx, c.b, paymentID)
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

func (c *client) ListPendingTransactions(ctx context.Context, walletID string) ([]transactions.Transaction, error) {
	return ops.ListPendingWebMonetization(ctx, c.b, walletID)
}

func (c *client) UpdateWalletAddressStatus(ctx context.Context, walletID rafiki.UpdateAddressStatus, status bool) error {
	return ops.UpdateWalletAddressStatus(ctx, c.b, walletID, status)
}

func (c *client) GetIncomingPayment(ctx context.Context, id string) (*rafiki.IncomingPayment, error) {
	return ops.GetIncomingPayment(ctx, c.b, id)
}

func (c *client) CancelOutgoingPayment(ctx context.Context, paymentPointerID, reason string) error {
	return ops.CancelOutgoingPayment(ctx, c.b, paymentPointerID, reason)
}

func (c *client) WithdrawIncomingPaymentLiquidity(ctx context.Context, incomingPaymentID string) error {
	return ops.WithdrawIncomingPaymentLiquidity(ctx, c.b, incomingPaymentID)
}

func (c *client) WithdrawOutgoingPaymentLiquidity(ctx context.Context, outgoingPaymentID string) error {
	return ops.WithdrawOutgoingPaymentLiquidity(ctx, c.b, outgoingPaymentID)
}
