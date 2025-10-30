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
			os.Getenv("GATEHUB_CARD_APP_ID"),
			os.Getenv("GATEHUB_GATEWAY_ID"),
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

func (c Client) GetUser(ctx context.Context, walletID string) (*gatehub.User, error) {
	return ops.GetUser(ctx, c.b, c.external, walletID)
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

func (c Client) AssignBalance(ctx context.Context, linkedAccountID string, txID string, amt currency.Amount) (*gatehub.Balance, error) {
	return ops.AssignBalance(ctx, c.b, linkedAccountID, txID, amt)
}

func (c Client) CreateTransfer(ctx context.Context, args gatehub.CreateTransferArgs) (*external.Transaction, error) {
	return ops.CreateTransfer(ctx, c.b, c.external, args)
}

func (c Client) GetTransaction(ctx context.Context, walletID, id string) (*external.Transaction, error) {
	return ops.GetTransaction(ctx, c.b, c.external, walletID, id)
}

func (c Client) ListDeliveryAddresses(ctx context.Context, walletID string) ([]external.CustomerDeliveryAddress, error) {
	return ops.ListDeliveryAddresses(ctx, c.b, c.external, walletID)
}

func (c Client) ListCards(ctx context.Context, externalIDs gatehub.ExternalIDs) ([]external.Card, error) {
	return ops.ListCards(ctx, c.b, c.external, externalIDs)
}

func (c Client) GetCardApplicationProducts(ctx context.Context) ([]external.CardApplicationProduct, error) {
	return ops.GetCardApplicationProducts(ctx, c.b, c.external)
}

func (c Client) OrderCard(ctx context.Context, args gatehub.OrderCardArgs) error {
	return ops.OrderCard(ctx, c.b, c.external, args)
}

func (c Client) GetExternalIDs(ctx context.Context, walletID string) (*gatehub.ExternalIDs, error) {
	return ops.GetExternalIDs(ctx, c.b, walletID)
}

func (c Client) LinkUserToGatewayByWalletID(ctx context.Context, walletID string) error {
	return ops.LinkUserToGatewayByWalletID(ctx, c.b, c.external, walletID)
}

func (c Client) LinkUserToGatewayByExternalID(ctx context.Context, externalID string) error {
	return ops.LinkUserToGatewayByExternalID(ctx, c.external, externalID)
}

func (c Client) GetCardToken(ctx context.Context, args gatehub.GetCardTokenArgs) (*external.TokenResponse, error) {
	return ops.GetCardToken(ctx, c.b, c.external, args)
}

func (c Client) FreezeCard(ctx context.Context, args gatehub.FreezeCardArgs) error {
	return ops.FreezeCard(ctx, c.b, c.external, args)
}

func (c Client) UnfreezeCard(ctx context.Context, args gatehub.UnfreezeCardArgs) error {
	return ops.UnfreezeCard(ctx, c.b, c.external, args)
}

func (c Client) BlockCard(ctx context.Context, args gatehub.BlockCardArgs) error {
	return ops.BlockCard(ctx, c.b, c.external, args)
}

func (c Client) ValidateCardProductCode(ctx context.Context, cardProductCode string) error {
	return ops.ValidateCardProductCode(ctx, c.external, cardProductCode)
}
