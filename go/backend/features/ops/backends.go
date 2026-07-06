package ops

import (
	"github.com/interledger/interledger-app/go/backend/wallets"
	"testing"

	"github.com/interledger/interledger-app/go/backend/linkedaccounts"

	"github.com/interledger/interledger-app/go/backend/kyc"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	DB() *sqlx.DB
	KYC() kyc.Client
	LinkedAccounts() linkedaccounts.Client
	Wallets() wallets.Client
}

type testBackends struct {
	db  *sqlx.DB
	kyc kyc.Client
	la  linkedaccounts.Client
	w   wallets.Client
}

func (t testBackends) LinkedAccounts() linkedaccounts.Client {
	return t.la
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func (t testBackends) KYC() kyc.Client {
	return t.kyc
}

func (t testBackends) Wallets() wallets.Client {
	return t.w
}

func NewTestBackends(t *testing.T, db *sqlx.DB, kyc kyc.Client, la linkedaccounts.Client, w wallets.Client) Backends {
	return &testBackends{db: db, kyc: kyc, la: la, w: w}
}
