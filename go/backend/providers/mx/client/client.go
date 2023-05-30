package client

import (
	"context"

	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/mx/external"
	external_client "gitlab.com/fynbos/backend/providers/mx/external/client"
	"gitlab.com/fynbos/backend/providers/mx/ops"
)

type Backends interface {
	KYC() kyc.Client
	LinkedAccounts() linkedaccounts.Client
}

var _ mx.Client = &Client{}

type opsBackends struct {
	external external.Client
	b        Backends
}

func (ob *opsBackends) External() external.Client {
	return ob.external
}

func (ob *opsBackends) KYC() kyc.Client {
	return ob.b.KYC()
}

func (ob *opsBackends) LinkedAccounts() linkedaccounts.Client {
	return ob.b.LinkedAccounts()
}

func New(clientID, apiKey string, b Backends) *Client {
	return &Client{
		b: &opsBackends{
			external: external_client.New(clientID, apiKey),
			b:        b,
		},
	}
}

type Client struct {
	b ops.Backends
}

func (c *Client) GetWidget(ctx context.Context, walletID string) (string, error) {
	return ops.GetWidget(ctx, c.b, walletID)
}

func (c *Client) CreateBankAccounts(ctx context.Context, args mx.CreateBankAccountsArgs) ([]linkedaccounts.LinkedAccount, error) {
	return ops.CreateBankAccounts(ctx, c.b, args)
}

func (c *Client) GetAccount(ctx context.Context, walletID, accountGuid string) (*mx.Account, error) {
	return ops.GetAccount(ctx, c.b, walletID, accountGuid)
}

func (c *Client) ListUsers(ctx context.Context) ([]mx.User, error) {
	return ops.ListUsers(ctx, c.b)
}

func (c *Client) DeleteExternalUser(ctx context.Context, guid string) error {
	return ops.DeleteUser(ctx, c.b, guid)
}
