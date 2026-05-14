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

// ExchangePublicToken trades the short-lived public_token from Plaid Link for
// a long-lived access_token + item_id.
func (c *Client) ExchangePublicToken(ctx context.Context, publicToken string) (string, string, error) {
	req := plaidsdk.NewItemPublicTokenExchangeRequest(publicToken)
	resp, _, err := c.sdk.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(*req).Execute()
	if err != nil {
		return "", "", fmt.Errorf("plaid: ItemPublicTokenExchange: %w", wrapPlaidError(err))
	}
	return resp.AccessToken, resp.ItemId, nil
}

// GetInstitutionForItem resolves the institution name attached to an Item.
// Returns empty strings (no error) for Items created without an institution
// link (e.g. Same Day Micro-deposits).
func (c *Client) GetInstitutionForItem(ctx context.Context, accessToken string) (string, string, error) {
	itemResp, _, err := c.sdk.PlaidApi.ItemGet(ctx).ItemGetRequest(*plaidsdk.NewItemGetRequest(accessToken)).Execute()
	if err != nil {
		return "", "", fmt.Errorf("plaid: ItemGet: %w", wrapPlaidError(err))
	}
	institutionID := itemResp.Item.GetInstitutionId()
	if institutionID == "" {
		return "", "", nil
	}
	instResp, _, err := c.sdk.PlaidApi.InstitutionsGetById(ctx).
		InstitutionsGetByIdRequest(*plaidsdk.NewInstitutionsGetByIdRequest(institutionID, c.countryCodes)).
		Execute()
	if err != nil {
		return institutionID, "", fmt.Errorf("plaid: InstitutionsGetById: %w", wrapPlaidError(err))
	}
	return institutionID, instResp.Institution.Name, nil
}

// GetAccounts calls Plaid `/accounts/get` and returns the full SDK response so
// the handler can serialize it verbatim.
func (c *Client) GetAccounts(ctx context.Context, accessToken string) (*plaidsdk.AccountsGetResponse, error) {
	resp, _, err := c.sdk.PlaidApi.AccountsGet(ctx).
		AccountsGetRequest(*plaidsdk.NewAccountsGetRequest(accessToken)).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("plaid: AccountsGet: %w", wrapPlaidError(err))
	}
	return &resp, nil
}

// GetAuth calls Plaid `/auth/get` to return ACH routing + account numbers.
func (c *Client) GetAuth(ctx context.Context, accessToken string) (*plaidsdk.AuthGetResponse, error) {
	resp, _, err := c.sdk.PlaidApi.AuthGet(ctx).
		AuthGetRequest(*plaidsdk.NewAuthGetRequest(accessToken)).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("plaid: AuthGet: %w", wrapPlaidError(err))
	}
	return &resp, nil
}

// GetBalance calls Plaid `/accounts/balance/get` (real-time balance refresh,
// distinct from cached values returned by /accounts/get).
func (c *Client) GetBalance(ctx context.Context, accessToken string) (*plaidsdk.AccountsGetResponse, error) {
	resp, _, err := c.sdk.PlaidApi.AccountsBalanceGet(ctx).
		AccountsBalanceGetRequest(*plaidsdk.NewAccountsBalanceGetRequest(accessToken)).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("plaid: AccountsBalanceGet: %w", wrapPlaidError(err))
	}
	return &resp, nil
}

// GetIdentity calls Plaid `/identity/get` for account-holder names, addresses,
// emails, phones.
func (c *Client) GetIdentity(ctx context.Context, accessToken string) (*plaidsdk.IdentityGetResponse, error) {
	resp, _, err := c.sdk.PlaidApi.IdentityGet(ctx).
		IdentityGetRequest(*plaidsdk.NewIdentityGetRequest(accessToken)).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("plaid: IdentityGet: %w", wrapPlaidError(err))
	}
	return &resp, nil
}

// SyncTransactions walks the Plaid `/transactions/sync` cursor stream from
// the beginning of the Item's history and accumulates added / modified /
// removed transactions. Returns the final `next_cursor` so a future caller
// could resume incrementally; we don't persist it in the POC.
//
// Safety cap: 50 pages × 100 items = 5,000 transactions. Sandbox + most real
// items finish in 1–3 pages; the cap exists to bound runaway loops on
// malformed cursor responses.
func (c *Client) SyncTransactions(ctx context.Context, accessToken string) (*plaid.TransactionsSyncResult, error) {
	const maxPages = 50

	result := &plaid.TransactionsSyncResult{}
	cursor := ""
	for page := range maxPages {
		req := plaidsdk.NewTransactionsSyncRequest(accessToken)
		if cursor != "" {
			req.SetCursor(cursor)
		}
		resp, _, err := c.sdk.PlaidApi.TransactionsSync(ctx).TransactionsSyncRequest(*req).Execute()
		if err != nil {
			return nil, fmt.Errorf("plaid: TransactionsSync (page %d): %w", page, wrapPlaidError(err))
		}
		
		fmt.Printf("plaid: ✅ transactions sync page %d: %d added, %d modified, %d removed, has_more=%v",
			page, len(resp.Added), len(resp.Modified), len(resp.Removed), resp.HasMore)

		result.Added = append(result.Added, resp.Added...)
		result.Modified = append(result.Modified, resp.Modified...)
		result.Removed = append(result.Removed, resp.Removed...)
		cursor = resp.NextCursor
		if !resp.HasMore {
			result.NextCursor = cursor
			return result, nil
		}
	}
	return nil, fmt.Errorf("plaid: TransactionsSync exceeded %d pages", maxPages)
}

// CreateProcessorToken calls Plaid `/processor/token/create` to mint a
// long-lived, account-scoped credential for a partner processor (e.g. "fiant").
// The returned processor_token is bound to exactly one (item, account_id,
// processor) triple — the partner (Fiant) can use it repeatedly to query
// Plaid for that one account until the Item is removed or access revoked.
// N accounts → N processor tokens. Phase 2 forwards it to Fiant via
// /users/{externalId}/payment-information.
func (c *Client) CreateProcessorToken(ctx context.Context, accessToken, accountID, processor string) (string, error) {
	req := plaidsdk.NewProcessorTokenCreateRequest(accessToken, accountID, processor)
	resp, _, err := c.sdk.PlaidApi.ProcessorTokenCreate(ctx).ProcessorTokenCreateRequest(*req).Execute()
	if err != nil {
		return "", fmt.Errorf("plaid: ProcessorTokenCreate: %w", wrapPlaidError(err))
	}
	return resp.ProcessorToken, nil
}

// RemoveItem invalidates an Item on Plaid via `/item/remove`. After this call
// the access_token can no longer be used. Local TokenStore deletion is the
// caller's responsibility (see ops.Handlers.Disconnect).
func (c *Client) RemoveItem(ctx context.Context, accessToken string) error {
	_, _, err := c.sdk.PlaidApi.ItemRemove(ctx).
		ItemRemoveRequest(*plaidsdk.NewItemRemoveRequest(accessToken)).
		Execute()
	if err != nil {
		return fmt.Errorf("plaid: ItemRemove: %w", wrapPlaidError(err))
	}
	return nil
}
