package gatehub

import "context"

type Client interface {
	CreateUser(ctx context.Context, walletID string) (string, error)
	SaveUser(ctx context.Context, walletID, externalID string) error
}
