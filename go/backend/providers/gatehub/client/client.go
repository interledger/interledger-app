package client

import (
	"context"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/gatehub/external"
	ops "gitlab.com/fynbos/backend/providers/gatehub/ops"
	httplogger "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/backend/user"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var _ gatehub.Client = Client{}

type Client struct {
	b  ops.Backends
	ec external.Client
}

type Backends interface {
	DB() *sqlx.DB
	Users() user.Client
}

func New(b Backends) *Client {
	ec := external.NewClient(
		os.Getenv("GATEHUB_APP_ID"),
		os.Getenv("GATEHUB_SECRET"),
		&http.Client{
			Transport: otelhttp.NewTransport(
				httplogger.NewTransport(http.DefaultTransport, b, nil),
			),
		},
	)

	return &Client{
		b:  b,
		ec: ec,
	}
}

func (c Client) CreateUser(ctx context.Context, walletID string) (string, error) {
	return ops.CreateUser(ctx, c.b, c.ec, walletID)
}
