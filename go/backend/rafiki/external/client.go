package external

import (
	"context"
	"fmt"
	"os"

	"github.com/Khan/genqlient/graphql"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/env"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Client interface {
	CreateWalletAddress(ctx context.Context, wallet wallets.Wallet) (string, error)
	FundOutgoingPayment(ctx context.Context, eventID string) error
}

type client struct {
	gcl   graphql.Client
	usdID string
}

func New() Client {
	baseURL := "http://rafiki-rafiki-backend.rafiki:3001/graphql"
	cl := graphql.NewClient(baseURL, otelhttp.DefaultClient) // TODO: set auth headers maybe

	// Default value for eu1
	assetUSD := os.Getenv("RAFIKI_USD_ASSET")
	if assetUSD == "" && env.IsDev() {
		assetUSD = "80d80585-5341-413a-acaf-169779b4642c"
	}

	return &client{gcl: cl, usdID: assetUSD}
}

func (c client) CreateWalletAddress(ctx context.Context, w wallets.Wallet) (string, error) {
	wa, err := CreateWalletAddress(ctx, c.gcl, CreateWalletAddressInput{
		AssetId:        c.usdID,
		Url:            w.AddressString(),
		PublicName:     w.Name,
		IdempotencyKey: w.ID,
	})
	if err != nil {
		return "", err
	}
	if !wa.GetCreateWalletAddress().Success {
		return "", fmt.Errorf("error code (%s) message (%s)", wa.GetCreateWalletAddress().Code, wa.GetCreateWalletAddress().Message)
	}

	return wa.CreateWalletAddress.WalletAddress.Id, nil
}

func (c client) FundOutgoingPayment(ctx context.Context, eventID string) error {
	r, err := DepositEventLiquidity(ctx, c.gcl, DepositEventLiquidityInput{
		EventId: eventID,
	})
	if err != nil {
		return err
	}
	if !r.GetDepositEventLiquidity().Success {
		return fmt.Errorf("error code (%s) message (%s)", r.GetDepositEventLiquidity().Code, r.GetDepositEventLiquidity().Message)
	}
	return nil
}
