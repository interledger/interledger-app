// Package client wraps the upstream Plaid SDK with the methods our handlers
// need. It contains no HTTP / chi concerns — those live in providers/plaid/ops.
package client

import (
	"context"
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

// CreateLinkToken is implemented in B5a.
func (c *Client) CreateLinkToken(_ context.Context, _ string) (string, time.Time, error) {
	return "", time.Time{}, plaid.ErrNotImplemented
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
