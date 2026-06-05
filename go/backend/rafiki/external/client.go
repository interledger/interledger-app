package external

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/Khan/genqlient/graphql"
	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/rafiki"
	"gitlab.com/fynbos/backend/slack"
	"gitlab.com/fynbos/backend/wallets"
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
	UpdateWalletAddressStatus(ctx context.Context, walletID rafiki.UpdateAddressStatus, status bool) error
	CancelOutgoingPayment(ctx context.Context, paymentPointerID, reason string) error
	WithdrawOutgoingPaymentLiquidity(ctx context.Context, outgoingPaymentID string, timeoutSeconds uint64) error
	WithdrawIncomingPaymentLiquidity(ctx context.Context, incomingPaymentID string, timeoutSeconds uint64) error
}

type assets struct {
	data map[string]string
	mu   sync.RWMutex
}

type client struct {
	backendClient graphql.Client
	authClient    graphql.Client
	a             *assets
}

type AdminSigningConfig struct {
	OperatorTenantID  string
	AdminAPISecret    string
	SignatureVersion  string
	BackendGraphQLURL string
	AuthGraphQLURL    string
}

func New(signingConfig AdminSigningConfig) Client {
	cl := graphql.NewClient(signingConfig.BackendGraphQLURL, newSignedAdminHTTPClient(signingConfig))
	authCl := graphql.NewClient(signingConfig.AuthGraphQLURL, newSignedAdminHTTPClient(signingConfig))
	return &client{backendClient: cl, authClient: authCl, a: &assets{data: nil}}
}

func (c client) GetWalletAddress(ctx context.Context, id string) (*GetWalletAddressWalletAddress, error) {
	pp, err := GetWalletAddress(ctx, c.backendClient, id)
	if err != nil {
		return nil, err
	}

	return &pp.WalletAddress, nil
}

func (c client) CreatePaymentPointer(ctx context.Context, w wallets.Wallet) (string, error) {
	var assetCode string

	if w.Country == country.US {
		assetCode = "USD"
	} else if w.Country == country.ZA {
		assetCode = "ZAR"
	} else if country.EUCountries[w.Country] {
		assetCode = "EUR"
	} else if w.Country == country.CA {
		assetCode = "CAD"
	} else {
		assetCode = "USD"
		slack.SendToChannel(ctx, slack.ChannelError, "wallet-info-bot", fmt.Sprintf("Rafiki external: Asset not configured for country=%s. Defaulting to USD. walletID=%s", w.Country, w.ID))
	}

	assetID, err := c.a.Get(ctx, c.backendClient, assetCode)

	if err != nil {
		log.Info(err.Error())
		return "", err
	}

	log.Info("Creating payment pointer in rafiki", zap.String("url", w.AddressString()))
	pp, err := CreateWalletAddress(ctx, c.backendClient, CreateWalletAddressInput{
		AssetId:        assetID,
		Address:        w.AddressString(),
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

func (c client) CancelOutgoingPayment(ctx context.Context, outgoingPaymentID, reason string) error {
	_, err := CancelOutgoingPayment(ctx, c.backendClient, CancelOutgoingPaymentInput{
		Id:     outgoingPaymentID,
		Reason: reason,
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

func newSignedAdminHTTPClient(signingConfig AdminSigningConfig) *http.Client {
	base := otelhttp.DefaultClient
	baseTransport := base.Transport
	if baseTransport == nil {
		baseTransport = http.DefaultTransport
	}

	return &http.Client{
		Transport: &adminSigningRoundTripper{
			base:             baseTransport,
			operatorTenantID: signingConfig.OperatorTenantID,
			adminAPISecret:   signingConfig.AdminAPISecret,
			signatureVersion: signingConfig.SignatureVersion,
		},
		Timeout: base.Timeout,
	}
}

func (a *assets) Get(ctx context.Context, backendClient graphql.Client, assetCode string) (string, error) {
	a.mu.RLock()
	if a.data != nil {
		if val, ok := a.data[assetCode]; ok {
			defer a.mu.RUnlock()
			return val, nil
		}
	}
	a.mu.RUnlock()

	a.mu.Lock()
	defer a.mu.Unlock()

	// Double checked locking pattern
	if a.data != nil {
		if val, ok := a.data[assetCode]; ok {
			return val, nil
		}
	}

	assets, err := GetAssets(ctx, backendClient)
	if err != nil {
		return "", fmt.Errorf("failed to fetch assets from Rafiki: %w", err)
	}

	data := make(map[string]string)
	for _, v := range assets.Assets.Edges {
		data[v.Node.Code] = v.Node.Id
	}

	a.data = data

	if val, ok := a.data[assetCode]; ok {
		return val, nil
	}

	return "", fmt.Errorf("asset %s not found after fetching from Rafiki", assetCode)
}

func (c client) UpdateWalletAddressStatus(ctx context.Context, wallet rafiki.UpdateAddressStatus, status bool) error {
	statusVal := WalletAddressStatusInactive
	if status {
		statusVal = WalletAddressStatusActive
	}
	_, err := UpdateWalletAddress(ctx, c.backendClient, UpdateWalletAddressInput{
		Id:         wallet.ID,
		Status:     statusVal,
		PublicName: wallet.Name,
	})
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	return nil
}

func (c client) WithdrawOutgoingPaymentLiquidity(ctx context.Context, outgoingPaymentID string, timeoutSeconds uint64) error {
	_, err := CreateOutgoingPaymentWithdrawal(ctx, c.backendClient, CreateOutgoingPaymentWithdrawalInput{
		OutgoingPaymentId: outgoingPaymentID,
		IdempotencyKey:    outgoingPaymentID + "_withdrawal",
		TimeoutSeconds:    strconv.FormatUint(timeoutSeconds, 10),
	})
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	return nil
}

func (c client) WithdrawIncomingPaymentLiquidity(ctx context.Context, incomingPaymentID string, timeoutSeconds uint64) error {
	_, err := CreateIncomingPaymentWithdrawal(ctx, c.backendClient, CreateIncomingPaymentWithdrawalInput{
		IncomingPaymentId: incomingPaymentID,
		IdempotencyKey:    incomingPaymentID + "_withdrawal",
		TimeoutSeconds:    strconv.FormatUint(timeoutSeconds, 10),
	})
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	return nil
}
