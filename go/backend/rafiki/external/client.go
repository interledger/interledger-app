package external

import (
	"context"
	"fmt"
	"os"

	"github.com/Khan/genqlient/graphql"
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/rafiki"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/env"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Client interface {
	CreatePaymentPointer(ctx context.Context, wallet wallets.Wallet) (string, error)
	CreatePaymentPointerKey(ctx context.Context, paymentPointerID string, key keys.Key) error
	RevokePaymentPointerKey(ctx context.Context, keyID string) error
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

func (c client) CreatePaymentPointerKey(ctx context.Context, paymentPointerID string, key keys.Key) error {
	r, err := CreatePaymentPointerKey(ctx, c.gcl, CreatePaymentPointerKeyInput{
		PaymentPointerId: paymentPointerID,
		Jwk: JwkInput{
			Kid: key.ID,
			X:   key.PublicKey,
			Alg: "EdDSA",
			Kty: "OKP",
			Crv: "Ed25519",
		},
	})
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}
	if !r.GetCreatePaymentPointerKey().Success {
		return fmt.Errorf("error code (%s) message (%s)", r.GetCreatePaymentPointerKey().Code, r.GetCreatePaymentPointerKey().Message)
	}

	return nil
}

func (c client) RevokePaymentPointerKey(ctx context.Context, keyID string) error {
	r, err := RevokePaymentPointerKey(ctx, c.gcl, RevokePaymentPointerKeyInput{
		Id: keyID,
	})
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}
	if !r.GetRevokePaymentPointerKey().Success {
		return fmt.Errorf("error code (%s) message (%s)", r.GetRevokePaymentPointerKey().Code, r.GetRevokePaymentPointerKey().Message)
	}
	return nil
}
