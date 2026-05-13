// Package client wraps the upstream Plaid SDK with the methods our handlers
// need. It contains no HTTP / chi concerns — those live in providers/plaid/ops.
package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	plaidsdk "github.com/plaid/plaid-go/v42/plaid"

	"gitlab.com/fynbos/backend/providers/plaid"
)

// Client is the production implementation of plaid.Client. It holds a single
// SDK ApiClient initialised at startup.
type Client struct {
	sdk          *plaidsdk.APIClient
	products     []plaidsdk.Products
	countryCodes []plaidsdk.CountryCode
}

// New constructs a Client from the parsed Config. Returns an error if the
// environment selector is unknown.
func New(cfg plaid.Config) (*Client, error) {
	sdkCfg := plaidsdk.NewConfiguration()
	sdkCfg.AddDefaultHeader("PLAID-CLIENT-ID", cfg.ClientID)
	sdkCfg.AddDefaultHeader("PLAID-SECRET", cfg.Secret)

	switch cfg.Env {
	case "sandbox":
		sdkCfg.UseEnvironment(plaidsdk.Sandbox)
	case "production":
		sdkCfg.UseEnvironment(plaidsdk.Production)
	default:
		return nil, fmt.Errorf("plaid client: unknown PLAID_ENV %q (want sandbox|production)", cfg.Env)
	}

	products := make([]plaidsdk.Products, 0, len(cfg.Products))
	for _, p := range cfg.Products {
		products = append(products, plaidsdk.Products(p))
	}
	countryCodes := make([]plaidsdk.CountryCode, 0, len(cfg.CountryCodes))
	for _, c := range cfg.CountryCodes {
		countryCodes = append(countryCodes, plaidsdk.CountryCode(c))
	}

	return &Client{
		sdk:          plaidsdk.NewAPIClient(sdkCfg),
		products:     products,
		countryCodes: countryCodes,
	}, nil
}

// CreateLinkToken calls Plaid `/link/token/create` and returns the link token
// and its expiry. `userID` is sent as `client_user_id` so Plaid Dashboard logs
// can be filtered per app user.
func (c *Client) CreateLinkToken(ctx context.Context, userID string) (string, time.Time, error) {
	req := plaidsdk.NewLinkTokenCreateRequest("Interledger Wallet", "en", c.countryCodes)
	req.SetUser(plaidsdk.LinkTokenCreateRequestUser{ClientUserId: userID})
	req.SetProducts(c.products)

	resp, _, err := c.sdk.PlaidApi.LinkTokenCreate(ctx).LinkTokenCreateRequest(*req).Execute()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("plaid: LinkTokenCreate: %w", wrapPlaidError(err))
	}
	fmt.Println("plaid: ✅ link token created for user: ", userID, ": ", resp.LinkToken)
	return resp.LinkToken, resp.Expiration, nil
}

// wrapPlaidError surfaces Plaid's JSON error_code / error_message / display_message
// alongside the generic SDK error string. Without this the caller only sees
// `400 Bad Request` with no diagnostic detail.
func wrapPlaidError(err error) error {
	var openAPIErr plaidsdk.GenericOpenAPIError
	if !errors.As(err, &openAPIErr) {
		return err
	}
	body := openAPIErr.Body()
	if len(body) == 0 {
		return err
	}
	return fmt.Errorf("%s: %s", err.Error(), string(body))
}

// ExchangePublicToken is implemented in B5b.
func (c *Client) ExchangePublicToken(_ context.Context, _ string) (string, string, error) {
	return "", "", plaid.ErrNotImplemented
}

// GetInstitutionForItem is implemented in B5b.
func (c *Client) GetInstitutionForItem(_ context.Context, _ string) (string, string, error) {
	return "", "", plaid.ErrNotImplemented
}

// GetAccounts is implemented in B5d.
func (c *Client) GetAccounts(_ context.Context, _ string) (*plaidsdk.AccountsGetResponse, error) {
	return nil, plaid.ErrNotImplemented
}

// GetAuth is implemented in B5d.
func (c *Client) GetAuth(_ context.Context, _ string) (*plaidsdk.AuthGetResponse, error) {
	return nil, plaid.ErrNotImplemented
}

// GetBalance is implemented in B5d.
func (c *Client) GetBalance(_ context.Context, _ string) (*plaidsdk.AccountsGetResponse, error) {
	return nil, plaid.ErrNotImplemented
}

// GetIdentity is implemented in B5d.
func (c *Client) GetIdentity(_ context.Context, _ string) (*plaidsdk.IdentityGetResponse, error) {
	return nil, plaid.ErrNotImplemented
}

// SyncTransactions is implemented in B5e.
func (c *Client) SyncTransactions(_ context.Context, _ string) (*plaid.TransactionsSyncResult, error) {
	return nil, plaid.ErrNotImplemented
}

// RemoveItem is implemented in B5f.
func (c *Client) RemoveItem(_ context.Context, _ string) error {
	return plaid.ErrNotImplemented
}
