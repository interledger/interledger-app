package client

import (
	"context"

	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
	external_client "gitlab.com/fynbos/backend/providers/tabapay/external/client"
	"gitlab.com/fynbos/backend/providers/tabapay/ops"
	temporal "go.temporal.io/sdk/client"
)

var _ tabapay.Client = &Client{}

type NewClientArgs = external_client.NewClientArgs

type Backends interface {
	KYC() kyc.Client
	LinkedAccounts() linkedaccounts.Client
	Temproal() temporal.Client
}

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

func (ob *opsBackends) Temporal() temporal.Client {
	return ob.b.Temproal()
}

func New(args NewClientArgs, b Backends) (*Client, error) {
	externalClient, err := external_client.New(args)
	if err != nil {
		return nil, err
	}

	return &Client{
		b: &opsBackends{
			external: externalClient,
			b:        b,
		},
	}, nil
}

type Client struct {
	b ops.Backends
}

func (c *Client) CreateCard(ctx context.Context, args tabapay.CreateCardArgs) (tabapay.Await, error) {
	return ops.CreateCard(ctx, c.b, args)
}
