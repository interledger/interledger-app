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
	CreatePaymentPointer(ctx context.Context, wallet wallets.Wallet, assetCode string) (string, error)
	CreatePaymentPointerKey(ctx context.Context, paymentPointerID string, key keys.Key) error
	RevokePaymentPointerKey(ctx context.Context, keyID string) error
	FundOutgoingPayment(ctx context.Context, eventID string) error
}

type client struct {
	gcl   graphql.Client
	usdID string
	zarID string
}

func New() Client {
	baseURL := "http://rafiki-rafiki-backend.rafiki:3001/graphql"
	cl := graphql.NewClient(baseURL, otelhttp.DefaultClient) // TODO: set auth headers maybe

	// Default value for eu1
	assetUSD := os.Getenv("RAFIKI_USD_ASSET")
	if assetUSD == "" && env.IsDev() {
		assetUSD = "80d80585-5341-413a-acaf-169779b4642c"
	}

	assetZAR := os.Getenv("RAFIKI_ZAR_ASSET")
	if assetZAR == "" && env.IsDev() {
		assetZAR = "9913bb55-64a2-41c8-a20a-1d607ef8267a"
	}

	return &client{gcl: cl, usdID: assetUSD, zarID: assetZAR}
}

func (c client) CreatePaymentPointer(ctx context.Context, w wallets.Wallet, assetCode string) (string, error) {
	var assetID string
	switch assetCode {
	case "USD":
		assetID = c.usdID
	case "ZAR":
		assetID = c.zarID
	default:
		return "", fmt.Errorf("%w: asset code (%s) not found", rafiki.ErrInternal, assetCode)
	}

	pp, err := CreateWalletAddress(ctx, c.gcl, CreateWalletAddressInput{
		AssetId:        assetID,
		Url:            w.AddressString(),
		PublicName:     w.Name,
		IdempotencyKey: w.ID,
	})
	if err != nil {
		return "", err
	}
	if !pp.GetCreateWalletAddress().Success {
		return "", fmt.Errorf("error code (%s) message (%s)", pp.GetCreateWalletAddress().Code, pp.GetCreateWalletAddress().Message)
	}

	return pp.CreateWalletAddress.WalletAddress.Id, nil
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

func (c client) CreatePaymentPointerKey(ctx context.Context, walletAddressID string, key keys.Key) error {
	fmt.Println("CreatePaymentPointerKey Key ID:" + key.ID)
	r, err := CreateWalletAddressKey(ctx, c.gcl, CreateWalletAddressKeyInput{
		WalletAddressId: walletAddressID,
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
	if !r.GetCreateWalletAddressKey().Success {
		return fmt.Errorf("error code (%s) message (%s)", r.GetCreateWalletAddressKey().Code, r.GetCreateWalletAddressKey().Message)
	}

	fmt.Println(r.CreateWalletAddressKey.GetWalletAddressKey())

	return nil
}

func (c client) RevokePaymentPointerKey(ctx context.Context, keyID string) error {
	r, err := RevokeWalletAddressKey(ctx, c.gcl, RevokeWalletAddressKeyInput{
		Id: keyID,
	})
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	if !r.GetRevokeWalletAddressKey().Success {
		return fmt.Errorf("error code (%s) message (%s)", r.GetRevokeWalletAddressKey().Code, r.GetRevokeWalletAddressKey().Message)
	}
	return nil
}
