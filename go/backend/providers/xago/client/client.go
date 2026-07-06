package client

import (
	"context"
	"net/http"
	"time"

	"github.com/interledger/interledger-app/go/backend/email"

	httplogger "github.com/interledger/interledger-app/go/backend/providers/http"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/kyc"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	"github.com/interledger/interledger-app/go/backend/payments"
	"github.com/interledger/interledger-app/go/backend/providers/xago"
	"github.com/interledger/interledger-app/go/backend/providers/xago/external"
	mock_client "github.com/interledger/interledger-app/go/backend/providers/xago/external/mock"
	"github.com/interledger/interledger-app/go/backend/providers/xago/ops"
	"github.com/interledger/interledger-app/go/backend/transactions"
	"github.com/interledger/interledger-app/go/backend/user"
	"github.com/interledger/interledger-app/go/backend/wallets"
	"github.com/interledger/interledger-app/go/env"
	"github.com/interledger/interledger-app/go/pacioli"
	"github.com/jmoiron/sqlx"
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

func New(b Backends, cfg external.Config) xago.Client {
	ex := external.New(&http.Client{
		Transport: otelhttp.NewTransport(
			httplogger.NewTransport(http.DefaultTransport, b, Redact),
		),
	}, b.DB(), cfg)
	if env.IsTest() {
		ex = mock_client.SetupDevMock(nil)
	}

	return &client{b: &opsBackends{
		Backends: b,
		xagoExt:  ex,
	}}
}

func (c *client) WebhookHandler() http.HandlerFunc {
	return ops.EventWebhook(c.b)
}

func (c *client) LookupSubAccount(ctx context.Context, walletID string) (*xago.SubAccount, error) {
	return ops.LookupSubAccount(ctx, c.b, walletID)
}

func (c *client) CreateBeneficiary(ctx context.Context, bankAcc xago.CreateBankAccountArgs) (xago.Await, error) {
	return ops.CreateBeneficiary(ctx, c.b, bankAcc)
}

func (c *client) CreateTransaction(ctx context.Context, args xago.CreateTransactionArgs) (*xago.Transaction, error) {
	return ops.CreateTransaction(ctx, c.b, args)
}

func (c *client) GetBalance(ctx context.Context, linkedAccountID string) (*xago.Balance, error) {
	return ops.GetBalance(ctx, c.b, linkedAccountID)
}

func (c *client) CreateBalanceAccount(ctx context.Context, args xago.CreateBalanceAccArgs) (xago.Await, error) {
	return ops.CreateBalanceAccount(ctx, c.b, args)
}

func (c *client) ReserveBalance(ctx context.Context, linkedAccountID, txID string, amt currency.Amount, timeout time.Duration) (*xago.Balance, error) {
	return ops.ReserveBalance(ctx, c.b, linkedAccountID, txID, amt, timeout)
}

func (c *client) FinaliseReserve(ctx context.Context, txID string) error {
	return ops.FinaliseReserve(ctx, c.b, txID)
}

func (c *client) RollbackReserve(ctx context.Context, txID string) error {
	return ops.RollbackReserve(ctx, c.b, txID)
}

func (c *client) AssignBalance(ctx context.Context, linkedAccountID, txID string, amt currency.Amount) (*xago.Balance, error) {
	return ops.AssignBalance(ctx, c.b, linkedAccountID, txID, amt)
}

func (c *client) LookupWithdrawal(ctx context.Context, id string) (*xago.Withdrawal, error) {
	return ops.LookupWithdrawal(ctx, c.b, id)
}

func (c *client) TestDeposit(ctx context.Context, sa xago.SubAccount) error {
	return ops.TestDeposit(ctx, c.b, sa)
}
func (c *client) UpdateInquiryLink(ctx context.Context, accountID string, walletID string) error {
	return ops.UpdateInquiryLink(ctx, c.b, xago.InquiryLinkUpdate{
		WalletID:  walletID,
		AccountID: accountID,
	})
}

func (c *client) GetBankAccount(ctx context.Context) (*xago.DepositDetails, error) {
	return ops.GetBankAccount(ctx, c.b)
}
