package external

import (
	"context"
	"fmt"

	"github.com/Khan/genqlient/graphql"
	"gitlab.com/fynbos/backend/wallets"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	assetUSD = "4a0e9bec-7a57-44b3-aab7-e450886b8a43"
)

type Client interface {
	CreatePaymentPointer(ctx context.Context, wallet wallets.Wallet) (string, error)
}

type client struct {
	gcl graphql.Client
}

func New() Client {
	baseURL := "https://localhost:8080/"
	cl := graphql.NewClient(baseURL, otelhttp.DefaultClient) // TODO: set auth headers maybe

	return &client{gcl: cl}
}

func (c client) CreatePaymentPointer(ctx context.Context, w wallets.Wallet) (string, error) {
	pp, err := CreatePaymentPointer(ctx, c.gcl, CreatePaymentPointerInput{
		AssetId:        assetUSD,
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
