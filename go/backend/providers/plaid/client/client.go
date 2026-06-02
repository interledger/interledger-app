package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	plaidsdk "github.com/plaid/plaid-go/v42/plaid"

	"gitlab.com/fynbos/backend/providers/plaid"
)

const maxPages = 50

type Client struct {
	sdk          *plaidsdk.APIClient
	products     []plaidsdk.Products
	countryCodes []plaidsdk.CountryCode
}

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

func (c *Client) CreateLinkToken(ctx context.Context, userID string) (string, time.Time, error) {
	req := plaidsdk.NewLinkTokenCreateRequest("Interledger Wallet", "en", c.countryCodes)
	req.SetUser(plaidsdk.LinkTokenCreateRequestUser{ClientUserId: userID})
	req.SetProducts(c.products)

	resp, _, err := c.sdk.PlaidApi.LinkTokenCreate(ctx).LinkTokenCreateRequest(*req).Execute()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("plaid: LinkTokenCreate: %w", wrapPlaidError(err))
	}
	return resp.LinkToken, resp.Expiration, nil
}

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

func (c *Client) ExchangePublicToken(ctx context.Context, publicToken string) (string, string, error) {
	req := plaidsdk.NewItemPublicTokenExchangeRequest(publicToken)
	resp, _, err := c.sdk.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(*req).Execute()
	if err != nil {
		return "", "", fmt.Errorf("plaid: ItemPublicTokenExchange: %w", wrapPlaidError(err))
	}
	return resp.AccessToken, resp.ItemId, nil
}

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

func (c *Client) GetAccounts(ctx context.Context, accessToken string) (*plaidsdk.AccountsGetResponse, error) {
	resp, _, err := c.sdk.PlaidApi.AccountsGet(ctx).
		AccountsGetRequest(*plaidsdk.NewAccountsGetRequest(accessToken)).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("plaid: AccountsGet: %w", wrapPlaidError(err))
	}
	return &resp, nil
}

func (c *Client) GetAuth(ctx context.Context, accessToken string) (*plaidsdk.AuthGetResponse, error) {
	resp, _, err := c.sdk.PlaidApi.AuthGet(ctx).
		AuthGetRequest(*plaidsdk.NewAuthGetRequest(accessToken)).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("plaid: AuthGet: %w", wrapPlaidError(err))
	}
	return &resp, nil
}

func (c *Client) GetBalance(ctx context.Context, accessToken string) (*plaidsdk.AccountsGetResponse, error) {
	resp, _, err := c.sdk.PlaidApi.AccountsBalanceGet(ctx).
		AccountsBalanceGetRequest(*plaidsdk.NewAccountsBalanceGetRequest(accessToken)).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("plaid: AccountsBalanceGet: %w", wrapPlaidError(err))
	}
	return &resp, nil
}

func (c *Client) GetIdentity(ctx context.Context, accessToken string) (*plaidsdk.IdentityGetResponse, error) {
	resp, _, err := c.sdk.PlaidApi.IdentityGet(ctx).
		IdentityGetRequest(*plaidsdk.NewIdentityGetRequest(accessToken)).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("plaid: IdentityGet: %w", wrapPlaidError(err))
	}
	return &resp, nil
}

func (c *Client) SyncTransactions(ctx context.Context, accessToken string) (*plaid.TransactionsSyncResult, error) {
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
// N accounts → N processor tokens.
func (c *Client) CreateProcessorToken(ctx context.Context, accessToken, accountID, processor string) (string, error) {
	req := plaidsdk.NewProcessorTokenCreateRequest(accessToken, accountID, processor)
	resp, _, err := c.sdk.PlaidApi.ProcessorTokenCreate(ctx).ProcessorTokenCreateRequest(*req).Execute()
	if err != nil {
		return "", fmt.Errorf("plaid: ProcessorTokenCreate: %w", wrapPlaidError(err))
	}
	return resp.ProcessorToken, nil
}

// Invalidates an Item on Plaid. After this call the access_token can no longer be used.
func (c *Client) RemoveItem(ctx context.Context, accessToken string) error {
	_, _, err := c.sdk.PlaidApi.ItemRemove(ctx).
		ItemRemoveRequest(*plaidsdk.NewItemRemoveRequest(accessToken)).
		Execute()
	if err != nil {
		return fmt.Errorf("plaid: ItemRemove: %w", wrapPlaidError(err))
	}

	return nil
}
