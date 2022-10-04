package machnet

import (
	"context"

	"gitlab.com/fynbos/backend/providers/machnet/external"
)

type Client interface {
	GetUser(ctx context.Context, walletID string) (*User, error)
	CreateUser(ctx context.Context, args CreateArgs) (*User, error)
	GetWidgetToken(ctx context.Context, walletID string) (*WidgetToken, error)
	HandleEvent(ctx context.Context, event external.Event) error
}
