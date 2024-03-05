package client

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/astra"
	"gitlab.com/fynbos/backend/providers/astra/external"
	mock_client "gitlab.com/fynbos/backend/providers/astra/external/mock"
	"gitlab.com/fynbos/backend/providers/astra/ops"
	"gitlab.com/fynbos/backend/providers/basistheory"
	httplogger "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/pacioli"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	temporal "go.temporal.io/sdk/client"
)

type Backends interface {
	DB() *sqlx.DB
	Payments() payments.Client
	Temporal() temporal.Client
	LinkedAccounts() linkedaccounts.Client
	Wallets() wallets.Client
	Users() user.Client
	KYC() kyc.Client
	Pacioli() pacioli.Client
	Transactions() transactions.Client
	Email() email.Client
	BasisTheory() basistheory.Client
	Twilio() twilio.Service
}

var _ ops.Backends = opsBackends{}

type opsBackends struct {
	Backends
	astraExt external.Client
}

func (o opsBackends) External() external.Client {
	return o.astraExt
}

var _ astra.Client = &client{}

type client struct {
	b ops.Backends
}

func New(b Backends) astra.Client {
	ex := external.New(&http.Client{
		Transport: otelhttp.NewTransport(
			httplogger.NewTransport(http.DefaultTransport, b, external.Redact),
		),
	})
	if env.IsTest() {
		ex = mock_client.SetupDevMock(nil)
	}
	return &client{b: opsBackends{
		Backends: b,
		astraExt: ex,
	}}
}

func (c client) WebhookHandler() http.HandlerFunc {
	return ops.EventWebhook(c.b)
}

func (c client) TrustedAuthInfoWebhook() http.HandlerFunc {
	return ops.GetTrustedAuthenticationInfo(c.b)
}

func (c client) StartKYC(ctx context.Context, walletID string) error {
	return ops.CreateIntent(ctx, c.b, walletID)
}

func (c client) CreateCard(ctx context.Context, args astra.CreateCardArgs) (astra.Await, error) {
	return ops.CreateCard(ctx, c.b, args)
}

func (c client) DebitCard(ctx context.Context, args astra.CardToAccountArgs) (string, error) {
	return ops.DebitCard(ctx, c.b, args)
}

func (c client) CreditCard(ctx context.Context, args astra.AccountToCardsArgs) (string, error) {
	return ops.CreditCard(ctx, c.b, args)
}

func (c client) LookupTransfer(ctx context.Context, walletID, txID string) (*astra.Transfer, error) {
	return ops.LookupTransfer(ctx, c.b, walletID, txID)
}

func (c client) LookupRoutine(ctx context.Context, walletID, routineID string) (*astra.Routine, error) {
	return ops.LookupRoutine(ctx, c.b, walletID, routineID)
}
