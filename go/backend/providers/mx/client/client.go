package client

import (
	"context"
	"time"

	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/mx/external"
	external_client "gitlab.com/fynbos/backend/providers/mx/external/client"
	"gitlab.com/fynbos/backend/providers/mx/ops"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
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

// This will perform an initial call. If that fails, then it will retry up to four times.
func (c *Client) GetWidget(ctx context.Context, walletID string) (string, error) {
	var widget string
	var err error
	backoff := []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 400 * time.Millisecond, time.Second}
	var retry int
	for {
		widget, err = ops.GetWidget(ctx, c.b, walletID)
		if err != nil && retry < 4 {
			log.Warn("Failed getting mx widget.", zap.Int("retry", retry), zap.Duration("backoff", backoff[retry]), zap.Error(err))
			time.Sleep(backoff[retry])
			retry++
		} else {
			break
		}
	}

	return widget, err
}

func (c *Client) CreateBankAccounts(ctx context.Context, args mx.CreateBankAccountsArgs) ([]linkedaccounts.LinkedAccount, error) {
	return ops.CreateBankAccounts(ctx, c.b, args)
}

func (c *Client) GetAccount(ctx context.Context, walletID, accountGuid string) (*mx.Account, error) {
	return ops.GetAccount(ctx, c.b, walletID, accountGuid)
}
