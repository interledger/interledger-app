package plaid

import (
	"context"
	"errors"
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
}

// TokenSet is the per-user state persisted after a successful Plaid Link.
type TokenSet struct {
	AccessToken     string
	ItemID          string
	InstitutionID   string
	InstitutionName string
	LinkedAt        time.Time
}

// State is the public view of a user's current Plaid linkage. Never includes
// the raw access token.
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

// Client is the surface our HTTP handlers consume. Tests mock this; the
// production implementation in providers/plaid/client wraps the Plaid SDK.
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
}

// TokenStore persists TokenSets keyed by user ID. POC uses Redis (see
// providers/plaid/store/redis.go); an in-memory implementation is kept for
// tests. Production path is encrypted Postgres — see
// documentation/poc/plaid/architecture.md §7.
type TokenStore interface {
	Get(ctx context.Context, userID string) (TokenSet, bool, error)
	Put(ctx context.Context, userID string, t TokenSet) error
	Delete(ctx context.Context, userID string) error
}

// ErrNotImplemented is returned by scaffolded methods until their owning
// task (B5*) fills them in.
var ErrNotImplemented = errors.New("plaid: not implemented")

// ErrNoLinkedItem indicates the caller looked up a user that has no stored
// TokenSet. Handlers translate this to HTTP 404.
var ErrNoLinkedItem = errors.New("plaid: no linked item for user")
