package client

import (
	"context"
	"gitlab.com/fynbos/backend/contacts"
	"gitlab.com/fynbos/backend/contacts/ops"
	"gitlab.com/fynbos/backend/db"
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

func (s client) List(ctx context.Context, walletID string, page db.Pagination) ([]contacts.Contact, error) {
	return ops.List(ctx, s.b, walletID, page)
}
