package discord

import (
	"context"
)

type Client interface {
	CreateAuthURL(ctx context.Context, args CreateAuthURLArgs) (string, error)
	CreateConnection(ctx context.Context, args CreateConnectionArgs) (*Connection, error)
	GetWalletConnections(ctx context.Context, id string) ([]Connection, error)
}
