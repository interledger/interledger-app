package client

import (
	"context"
	"time"

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

func (c *client) UserForToken(ctx context.Context, token string) (usr *user.User, err error) {
	defer func(begin time.Time) {
		if err != nil {
			log.Debug("failed to parse user x-session-token", zap.Error(err))
		}
	}(time.Now())

	return ops.UserForToken(ctx, c.b, token)
}

func (c *client) UserForContext(ctx context.Context) (usr *user.User, err error) {
	return ops.UserForContext(ctx)
}

func (c *client) GetUser(ctx context.Context, userID string) (*user.User, error) {
	return ops.GetUser(ctx, c.b, userID)
}

func (c *client) ListUsers(ctx context.Context, walletID string) ([]user.User, error) {
	return ops.ListUsers(ctx, c.b, walletID)
}

func (c *client) CheckUserTotpEnabled(ctx context.Context, identityID string) (bool, error) {
	return ops.CheckUserTotpEnabled(ctx, c.b, identityID)
}

func (c *client) Delete2FATotpEnrollment(ctx context.Context, identityID string) error {
	return ops.Delete2FATotpEnrollment(ctx, c.b, identityID)
}

func (c *client) GetTotpURL(ctx context.Context, userID string) (string, error) {
	return ops.GetTotpURL(ctx, c.b, userID)
}

func (c *client) ValidateTotpCode(ctx context.Context, userID, code string, now time.Time) error {
	return ops.ValidateTotpCode(ctx, c.b, userID, code, now)
}

func (c *client) GetUserIDForWallet(ctx context.Context, walletID string) (string, error) {
	return ops.GetUserIDForWallet(ctx, c.b, walletID)
}
