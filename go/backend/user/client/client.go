package client

import (
	"context"
	"time"

	"gitlab.com/fynbos/backend/db"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	kratos "github.com/ory/kratos-client-go"

	"go.uber.org/zap"

	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/user/ops"
	"gitlab.com/fynbos/log"
)

var _ user.Client = &client{}

type client struct {
	b ops.Backends
}

func New(b Backends, kratosURL, kratosAdminURL string) user.Client {

	configuration := kratos.NewConfiguration()
	configuration.HTTPClient = otelhttp.DefaultClient
	configuration.Servers = kratos.ServerConfigurations{
		{
			URL:         kratosURL,
			Description: "Public Kratos",
		},
		{
			URL:         kratosAdminURL,
			Description: "Admin Kratos",
		},
	}

	kratosClient := kratos.NewAPIClient(configuration)

	ob := &opsBackends{
		Backends: b,
		kratos:   kratosClient,
	}

	return &client{
		b: ob,
	}
}

func (c *client) UserForCookie(ctx context.Context, cookie string) (usr *user.User, err error) {
	defer func(begin time.Time) {
		if err != nil {
			log.Debug("failed to parse user cookie", zap.Error(err))
		}
	}(time.Now())

	return ops.UserForCookie(ctx, c.b, cookie)
}

func (c *client) UserForContext(ctx context.Context) (usr *user.User, err error) {
	defer func(begin time.Time) {
		if err != nil {
			log.Info("no user in context", zap.Error(err))
		}
	}(time.Now())

	return ops.UserForContext(ctx)
}

func (c *client) GetUser(ctx context.Context, userID string) (*user.User, error) {
	return ops.GetUser(ctx, c.b, userID)
}

func (c *client) ListUsers(ctx context.Context, walletID string) ([]user.User, error) {
	return ops.ListUsers(ctx, c.b, walletID)
}

func (c *client) WalletForContext(ctx context.Context) (uw *user.Wallet, err error) {
	defer func(begin time.Time) {
		if err != nil {
			log.Info("no wallet in context", zap.Error(err))
		}
	}(time.Now())

	return ops.WalletForContext(ctx)
}

func (c *client) CreateNewWallet(ctx context.Context, args user.CreateWalletArgs) (*user.Wallet, error) {
	return ops.CreateWallet(ctx, c.b, args)
}

func (c *client) ListWallets(ctx context.Context, userID string) ([]user.Wallet, error) {
	return ops.ListWallets(ctx, c.b, userID)
}

func (c *client) GetWallet(ctx context.Context, id string) (*user.Wallet, error) {
	return ops.GetWallet(ctx, c.b, id)
}

func (c *client) ListAllWallets(ctx context.Context, pagination db.Pagination) ([]user.Wallet, error) {
	return ops.ListAllWallets(ctx, c.b, pagination)
}

func (c *client) SetWalletName(ctx context.Context, id, name string) error {
	return ops.SetWalletName(ctx, c.b, id, name)
}
