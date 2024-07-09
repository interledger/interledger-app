package client

import (
	"context"
	"net/http"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/chimoney"
	"gitlab.com/fynbos/backend/providers/chimoney/external"
	"gitlab.com/fynbos/backend/providers/chimoney/ops"
	httplogger "gitlab.com/fynbos/backend/providers/http"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var _ chimoney.Client = &Client{}

type Client struct {
	b        ops.Backends
	external external.Client
}

func New(b ops.Backends) chimoney.Client {
	return &Client{
		b: b,
		external: external.New(
			&http.Client{
				Transport: otelhttp.NewTransport(
					httplogger.NewTransport(http.DefaultTransport, b, external.Redact),
				),
			},
		),
	}
}

func (c *Client) CreateWallet(ctx context.Context, walletID string) (chimoney.Await, error) {
	return ops.CreateWallet(ctx, c.b, walletID)
}

func (c *Client) AddInterlocEmail(ctx context.Context, walletID, email string) (string, error) {
	return ops.UpsertInteracEmail(ctx, c.b, walletID, email)
}

func (c *Client) GetInterlocEmail(ctx context.Context, walletID string) (string, error) {
	return ops.GetInteracEmail(ctx, c.b, walletID)
}

func (c *Client) CreateDepositLink(ctx context.Context, walletID string, amt currency.Amount) (string, error) {
	return ops.CreateDepositLink(ctx, c.b, c.external, walletID, amt)
}

func (c *Client) Withdraw(ctx context.Context, walletID string, amt currency.Amount) error {
	return ops.Withdraw(ctx, c.b, c.external, walletID, amt)
}

func (c *Client) ReserveBalance(ctx context.Context, linkedAccountID, txID string, amt currency.Amount, timeout time.Duration) (*chimoney.Balance, error) {
	return ops.ReserveBalance(ctx, c.b, linkedAccountID, txID, amt, timeout)
}

func (c *Client) Transfer(ctx context.Context, args chimoney.TransferArgs) error {
	return ops.Transfer(ctx, c.b, c.external, args)
}

func (c *Client) FinaliseReserve(ctx context.Context, txID string) error {
	return ops.FinaliseReserve(ctx, c.b, txID)
}

func (c *Client) AssignBalance(ctx context.Context, linkedAccountID, trxID string, amount currency.Amount) (*chimoney.Balance, error) {
	return ops.AssignBalance(ctx, c.b, linkedAccountID, trxID, amount)
}

func (c *Client) RollbackReserve(ctx context.Context, txID string) error {
	return ops.RollbackReserve(ctx, c.b, txID)
}

func (c *Client) GetKYCWidget(ctx context.Context, walletID string) (string, error) {
	return ops.GetKYCWidget(ctx, c.b, walletID)
}

func (c *Client) WatchForSuccessfulKYC(ctx context.Context, walletID string) error {
	return ops.WatchForSuccessfulKYC(ctx, c.b, walletID)
}

func (c *Client) CreateDeposit(ctx context.Context, walletID, issueID string) (chimoney.Await, error) {
	return ops.CreateDeposit(ctx, c.b, c.external, walletID, issueID)
}
