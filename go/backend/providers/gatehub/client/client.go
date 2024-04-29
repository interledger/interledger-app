package client

import (
	"context"
	"net/http"
	"os"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/gatehub/external"
	ops "gitlab.com/fynbos/backend/providers/gatehub/ops"
	httplogger "gitlab.com/fynbos/backend/providers/http"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var _ gatehub.Client = Client{}

type Client struct {
	b        ops.Backends
	external external.Client
}

func New(b ops.Backends) *Client {
	return &Client{
		b: b,
		external: external.NewClient(
			os.Getenv("GATEHUB_APP_ID"),
			os.Getenv("GATEHUB_SECRET"),
			&http.Client{
				Transport: otelhttp.NewTransport(
					httplogger.NewTransport(http.DefaultTransport, b, nil),
				),
			},
		),
	}
}

func (c Client) CreateUser(ctx context.Context, walletID string) (gatehub.Await, error) {
	return ops.CreateUser(ctx, c.b, walletID)
}

func (c Client) GetOnboardingWidget(ctx context.Context, walletID string) (string, error) {
	return ops.GetOnboardingWidget(ctx, c.b, c.external, walletID)
}

func (c Client) GetOnOffRampWidget(ctx context.Context, walletID string, isDeposit bool) (string, error) {
	return ops.GetOnOffRampWidget(ctx, c.b, c.external, walletID, isDeposit)
}

func (c Client) GetBalance(ctx context.Context, linkedAccountID string) (*gatehub.Balance, error) {
	return ops.GetBalance(ctx, c.b, linkedAccountID)
}

func (c Client) CreateWithdrawal(ctx context.Context, walletID, externalTransactionID string) (string, error) {
	return ops.CreateWithdrawal(ctx, c.b, c.external, walletID, externalTransactionID)
}

func (c Client) ReserveBalance(ctx context.Context, linkedAccountID, txID string, amt currency.Amount, timeout time.Duration) (*gatehub.Balance, error) {
	return ops.ReserveBalance(ctx, c.b, linkedAccountID, txID, amt, timeout)
}

func (c Client) FinaliseReserve(ctx context.Context, trxID string) error {
	return ops.FinaliseReserve(ctx, c.b, trxID)
}

func (c Client) RollbackReserve(ctx context.Context, trxID string) error {
	return ops.RollbackReserve(ctx, c.b, trxID)
}
