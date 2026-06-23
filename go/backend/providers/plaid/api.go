package plaid

import (
	"context"
	"time"
)

// Config is the Plaid provider configuration assembled from env vars at startup.
type Config struct {
	Enabled      bool
	ClientID     string
	Secret       string
	Env          string
	Products     []string
	CountryCodes []string
	Processor    string
	// APIURL, when non-empty, overrides the SDK base URL (e.g. point at mockplaid
	// in local dev). Empty = use the real Plaid environment selected by Env.
	APIURL string
}

type Client interface {
	CreateLinkToken(ctx context.Context, userID string) (linkToken string, expiration time.Time, err error)
	ExchangePublicToken(ctx context.Context, publicToken string) (accessToken, itemID string, err error)
	CreateProcessorToken(ctx context.Context, accessToken, accountID, processor string) (string, error)
}
