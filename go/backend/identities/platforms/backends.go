package platforms

import (
	"testing"

	"github.com/interledger/interledger-app/go/backend/wallets"

	"github.com/interledger/interledger-app/go/backend/images"

	temporal "go.temporal.io/sdk/client"

	"github.com/interledger/interledger-app/go/backend/analytics"
	"github.com/interledger/interledger-app/go/backend/keys"
	"github.com/interledger/interledger-app/go/backend/twitter"

	"github.com/go-playground/validator/v10"

	"github.com/jmoiron/sqlx"
)

type Backends interface {
	Twitter() twitter.Client
	Validator() *validator.Validate
	DB() *sqlx.DB
	Keys() keys.Client
	Analytics() analytics.Client
	Temporal() temporal.Client
	Images() images.Client
	Wallets() wallets.Client
}

type testBackends struct {
	db      *sqlx.DB
	val     *validator.Validate
	an      analytics.Client
	keys    keys.Client
	twitter twitter.Client
	img     images.Client
}

func (t testBackends) Wallets() wallets.Client {
	return nil
}

func (t testBackends) Temporal() temporal.Client {
	//TODO implement me
	panic("implement me")
}

func (t testBackends) Analytics() analytics.Client {
	return t.an
}

func (t testBackends) Twitter() twitter.Client {
	return t.twitter
}

func (t testBackends) Validator() *validator.Validate {
	return t.val
}

func (t testBackends) Keys() keys.Client {
	return t.keys
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func (t testBackends) Images() images.Client {
	return t.img
}

func NewTestBackends(_ *testing.T, db *sqlx.DB) Backends {
	return &testBackends{db: db, val: validator.New()}
}
