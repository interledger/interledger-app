package plaid

import (
	"context"
	"time"

	plaidsdk "github.com/plaid/plaid-go/v42/plaid"
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

// TokenSet is the per-user state persisted after a successful Plaid Link.
type TokenSet struct {
	AccessToken     string
	ItemID          string
	InstitutionID   string
	InstitutionName string
	LinkedAt        time.Time
}
type State struct {
	Linked          bool       `json:"linked"`
	ItemID          string     `json:"item_id,omitempty"`
	InstitutionName string     `json:"institution_name,omitempty"`
	LinkedAt        *time.Time `json:"linked_at,omitempty"`
}

// TransactionsSyncResult is the accumulated output of a TransactionsSync
// pagination loop.
type TransactionsSyncResult struct {
	Added      []plaidsdk.Transaction        `json:"added"`
	Modified   []plaidsdk.Transaction        `json:"modified"`
	Removed    []plaidsdk.RemovedTransaction `json:"removed"`
	NextCursor string                        `json:"next_cursor"`
}

type Client interface {
	CreateLinkToken(ctx context.Context, userID string) (linkToken string, expiration time.Time, err error)
	ExchangePublicToken(ctx context.Context, publicToken string) (accessToken, itemID string, err error)
	GetInstitutionForItem(ctx context.Context, accessToken string) (institutionID, institutionName string, err error)
	GetAccounts(ctx context.Context, accessToken string) (*plaidsdk.AccountsGetResponse, error)
	GetAuth(ctx context.Context, accessToken string) (*plaidsdk.AuthGetResponse, error)
	GetBalance(ctx context.Context, accessToken string) (*plaidsdk.AccountsGetResponse, error)
	GetIdentity(ctx context.Context, accessToken string) (*plaidsdk.IdentityGetResponse, error)
	SyncTransactions(ctx context.Context, accessToken string) (*TransactionsSyncResult, error)
	RemoveItem(ctx context.Context, accessToken string) error
	CreateProcessorToken(ctx context.Context, accessToken, accountID, processor string) (string, error)
}

type TokenStore interface {
	Get(ctx context.Context, userID string) (TokenSet, bool, error)
	Put(ctx context.Context, userID string, t TokenSet) error
	Delete(ctx context.Context, userID string) error
}
