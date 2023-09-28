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
	CreatePaymentPointer(ctx context.Context, wallet wallets.Wallet) (string, error)
}

type client struct {
	gcl   graphql.Client
	usdID string
}

func New() Client {
	baseURL := "http://rafiki-rafiki-backend.rafiki:3001/"
	cl := graphql.NewClient(baseURL, otelhttp.DefaultClient) // TODO: set auth headers maybe

	// Default value for eu1
	assetUSD := os.Getenv("RAFIKI_USD_ASSET")
	if assetUSD == "" && env.IsDev() {
		assetUSD = "80d80585-5341-413a-acaf-169779b4642c"
	}

	return &client{gcl: cl, usdID: assetUSD}
}

func (c client) CreatePaymentPointer(ctx context.Context, w wallets.Wallet) (string, error) {
	pp, err := CreatePaymentPointer(ctx, c.gcl, CreatePaymentPointerInput{
		AssetId:        c.usdID,
		Url:            w.AddressString(),
		PublicName:     w.Name,
		IdempotencyKey: w.ID,
	})
	if err != nil {
		return "", err
	}
	if !pp.GetCreatePaymentPointer().Success {
		return "", fmt.Errorf("error code (%s) message (%s)", pp.GetCreatePaymentPointer().Code, pp.GetCreatePaymentPointer().Message)
	}

	return pp.CreatePaymentPointer.PaymentPointer.Id, nil

}
