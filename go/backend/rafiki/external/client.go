package external

import (
	"context"
	"fmt"
	"os"

	"github.com/Khan/genqlient/graphql"
	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/rafiki"
	"gitlab.com/fynbos/backend/slack"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

type Client interface {
	CreatePaymentPointer(ctx context.Context, wallet wallets.Wallet) (string, error)
	GetWalletAddress(ctx context.Context, id string) (*GetWalletAddressWalletAddress, error)
	CreatePaymentPointerKey(ctx context.Context, paymentPointerID string, key keys.Key) (string, error)
	RevokePaymentPointerKey(ctx context.Context, keyID string) error
	FundOutgoingPayment(ctx context.Context, outgoingPaymentID string) error
	ListGrants(ctx context.Context, paymentPointer string) ([]ListGrantsGrantsGrantsConnectionEdgesGrantEdgeNodeGrant, error)
	GetGrant(ctx context.Context, grantID string) (*GetGrantGrant, error)
	RevokeGrant(ctx context.Context, grantID string) error
	GetIncomingPayment(ctx context.Context, id string) (*GetIncomingPaymentIncomingPayment, error)
}

type client struct {
	backendClient graphql.Client
	authClient    graphql.Client
	usdID         string
	zarID         string
	eurID         string
	cadID         string
}

func New() Client {
	backendGraphql := "http://localhost:3001/graphql"
	if os.Getenv("RAFIKI_BACKEND_GRAPHQL_URL") != "" {
		backendGraphql = os.Getenv("RAFIKI_BACKEND_GRAPHQL_URL")
	}
	cl := graphql.NewClient(backendGraphql, otelhttp.DefaultClient) // TODO: set auth headers maybe

	// Default value for eu1
	// TODO: Load assets at startup/cache them.
	// We might have to have a temporal and sometimes reach out to Rafiki to
	// retrieve all the assets - in case a new one has been added.
	assetUSD := os.Getenv("RAFIKI_USD_ASSET")
	if assetUSD == "" && env.IsLocal() {
		assetUSD = "<replace-me>"
	} else if assetUSD == "" && env.IsDev() {
		assetUSD = "80d80585-5341-413a-acaf-169779b4642c"
	} else if assetUSD == "" && env.IsProd() {
		assetUSD = "22fd68aa-d9b3-40eb-a69d-6a45f4b9cbeb"
	}

	assetZAR := os.Getenv("RAFIKI_ZAR_ASSET")
	if assetZAR == "" && env.IsLocal() {
		assetZAR = "<replace-me>"
	} else if assetZAR == "" && env.IsDev() {
		assetZAR = "9913bb55-64a2-41c8-a20a-1d607ef8267a"
	} else if assetZAR == "" && env.IsProd() {
		assetZAR = "622c7646-a8aa-491b-b260-40e33a433d1c"
	}

	assetEUR := os.Getenv("RAFIKI_EUR_ASSET")
	if assetEUR == "" && env.IsLocal() {
		assetEUR = "<replace-me>"
	} else if assetEUR == "" && env.IsDev() {
		assetEUR = "108e1cc9-a3b0-4d33-a876-b4c992ddbaed"
	} else if assetEUR == "" && env.IsProd() {
		assetEUR = "9c73e88a-be59-4246-b2fe-dfa8b657b4b5"
	}

	assetCAD := os.Getenv("RAFIKI_CAD_ASSET")
	if assetCAD == "" && env.IsLocal() {
		assetCAD = "<replace-me>"
	} else if assetCAD == "" && env.IsDev() {
		assetCAD = "7e09ec86-dc19-445b-8483-ca4d91362605"
	} else if assetCAD == "" && env.IsProd() {
		assetCAD = "e254ae75-a520-42e0-8045-badf09c24ece"
	}

	authGraphql := "http://localhost:3003/graphql"
	if os.Getenv("RAFIKI_AUTH_GRAPHQL_URL") != "" {
		authGraphql = os.Getenv("RAFIKI_AUTH_GRAPHQL_URL")
	}
	authCl := graphql.NewClient(authGraphql, otelhttp.DefaultClient)

	return &client{backendClient: cl, usdID: assetUSD, authClient: authCl, zarID: assetZAR, eurID: assetEUR, cadID: assetCAD}
}

func (c client) GetWalletAddress(ctx context.Context, id string) (*GetWalletAddressWalletAddress, error) {
	pp, err := GetWalletAddress(ctx, c.backendClient, id)
	if err != nil {
		return nil, err
	}

	return &pp.WalletAddress, nil
}

func (c client) CreatePaymentPointer(ctx context.Context, w wallets.Wallet) (string, error) {
	var assetID string
	if w.Country == country.US {
		assetID = c.usdID
	} else if w.Country == country.ZA {
		assetID = c.zarID
	} else if country.EUCountries[w.Country] {
		assetID = c.eurID
	} else if w.Country == country.CA {
		assetID = c.cadID
	} else {
		assetID = c.usdID
		slack.SendToChannel(ctx, slack.ChannelNotifyEvents, "Fynbot", fmt.Sprintf("Rafiki external: Asset not configured for country=%s. Defaulting to USD. walletID=%s", w.Country, w.ID))
	}

	log.Info("Creating payment pointer in rafiki", zap.String("url", w.AddressString()))
	pp, err := CreateWalletAddress(ctx, c.backendClient, CreateWalletAddressInput{
		AssetId:        assetID,
		Url:            w.AddressString(),
		PublicName:     w.Name,
		IdempotencyKey: w.ID,
	})
	if err != nil {
		return "", err
	}

	return pp.CreateWalletAddress.WalletAddress.Id, nil
}

func (c client) FundOutgoingPayment(ctx context.Context, outgoingPaymentID string) error {
	_, err := DepositOutgoingPaymentLiquidity(ctx, c.backendClient, DepositOutgoingPaymentLiquidityInput{
		OutgoingPaymentId: outgoingPaymentID,
		IdempotencyKey:    outgoingPaymentID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c client) CreatePaymentPointerKey(ctx context.Context, walletAddressID string, key keys.Key) (string, error) {
	r, err := CreateWalletAddressKey(ctx, c.backendClient, CreateWalletAddressKeyInput{
		WalletAddressId: walletAddressID,
		Jwk: JwkInput{
			Kid: key.KeyID,
			X:   key.PublicKey,
			Alg: "EdDSA",
			Kty: "OKP",
			Crv: "Ed25519",
		},
		IdempotencyKey: key.KeyID,
	})
	if err != nil {
		return "", fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	return r.CreateWalletAddressKey.GetWalletAddressKey().Id, nil
}

func (c client) RevokePaymentPointerKey(ctx context.Context, keyID string) error {
	_, err := RevokeWalletAddressKey(ctx, c.backendClient, RevokeWalletAddressKeyInput{
		Id: keyID,
	})
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	return nil
}

func (c client) ListGrants(ctx context.Context, paymentPointer string) ([]ListGrantsGrantsGrantsConnectionEdgesGrantEdgeNodeGrant, error) {
	r, err := ListGrants(ctx, c.authClient, GrantFilter{
		Identifier: FilterString{In: []string{paymentPointer}},
		State:      FilterGrantState{In: []GrantState{GrantStateApproved, GrantStatePending, GrantStateApproved}},
	})
	if err != nil {
		return nil, err
	}

	var resp []ListGrantsGrantsGrantsConnectionEdgesGrantEdgeNodeGrant
	for _, ge := range r.GetGrants().Edges {
		resp = append(resp, ge.GetNode())
	}

	return resp, nil
}

func (c client) GetGrant(ctx context.Context, grantID string) (*GetGrantGrant, error) {
	r, err := GetGrant(ctx, c.authClient, grantID)
	if err != nil {
		return nil, err
	}

	return &r.Grant, nil
}

func (c client) RevokeGrant(ctx context.Context, grantID string) error {
	_, err := RevokeGrant(ctx, c.authClient, RevokeGrantInput{GrantId: grantID})
	if err != nil {
		return err
	}

	return nil
}

func (c client) GetIncomingPayment(ctx context.Context, id string) (*GetIncomingPaymentIncomingPayment, error) {
	r, err := GetIncomingPayment(ctx, c.backendClient, id)
	if err != nil {
		return nil, err
	}

	return &r.IncomingPayment, nil
}
