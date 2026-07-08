package client

import (
	"context"

	"github.com/interledger/interledger-app/go/backend/contacts"
	"github.com/interledger/interledger-app/go/backend/contacts/ops"
	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/interledger/interledger-app/go/backend/wallets"
)

var _ contacts.Client = client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) contacts.Client {
	return &client{
		b: b,
	}
}

func (s client) Create(ctx context.Context, args contacts.CreateContactArgs) (*contacts.Contact, error) {
	return ops.Create(ctx, s.b, args)
}

func (s client) List(ctx context.Context, walletID string, page db.Pagination, orderBy string) ([]contacts.Contact, error) {
	return ops.List(ctx, s.b, walletID, page, orderBy)
}

func (s client) Get(ctx context.Context, walletID string, wa wallets.Address) (*contacts.Contact, error) {
	return ops.Get(ctx, s.b, walletID, wa)
}

func (s client) SetLastPaidAtNow(ctx context.Context, walletID string, wa wallets.Address) error {
	return ops.SetLastPaidAtNow(ctx, s.b, walletID, wa)
}
