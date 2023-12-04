package wallets

import (
	"context"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/db"
)

type Client interface {
	Create(ctx context.Context, args CreateArgs) (*Wallet, error)
	ForContext(ctx context.Context) (*Wallet, error)
	Get(ctx context.Context, id string) (*Wallet, error)
	List(ctx context.Context, userID string) ([]Wallet, error)
	ListAll(ctx context.Context, page db.Pagination) ([]Wallet, error)
	SetWalletName(ctx context.Context, id, name string) (*Wallet, error)
	GetFromAddress(ctx context.Context, address string) (*Wallet, error)
	AddAddress(ctx context.Context, id, address string) (*Wallet, error)
	SetCountry(ctx context.Context, id string, country country.Country) (*Wallet, error)
	SetExceededLimits(ctx context.Context, id string, exceeded bool) (*Wallet, error)
}
