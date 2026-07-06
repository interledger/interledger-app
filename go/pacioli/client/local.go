package client

import (
	"context"

	"github.com/go-playground/validator/v10"

	"github.com/jmoiron/sqlx"

	"github.com/interledger/interledger-app/go/pacioli"
	"github.com/interledger/interledger-app/go/pacioli/ledger"
)

var _ pacioli.Client = localClient{}

type localClient struct {
	b ledger.Backends
}

type backends struct {
	db  *sqlx.DB
	val *validator.Validate
}

func (b backends) DB() *sqlx.DB {
	return b.db
}

func (b backends) Validator() *validator.Validate {
	return b.val
}

func NewLocal(db *sqlx.DB) pacioli.Client {
	b := &backends{
		db:  db,
		val: validator.New()}

	go ledger.TimeoutTransfersForever(b)

	return localClient{
		b: b,
	}
}

func (l localClient) ConfigureLedgers(ctx context.Context, args []pacioli.ConfigureLedgerArgs) ([]pacioli.LedgerResult, error) {
	return ledger.ConfigureLedgers(ctx, l.b, args)
}

func (l localClient) GetLedgers(ctx context.Context, ledgerIDs []uint32) ([]pacioli.Ledger, error) {
	return ledger.GetLedgers(ctx, l.b, ledgerIDs)
}

func (l localClient) ConfigureAccounts(ctx context.Context, args []pacioli.ConfigureAccountArgs) ([]pacioli.AccountResult, error) {
	return ledger.ConfigureAccounts(ctx, l.b, args)
}

func (l localClient) GetAccounts(ctx context.Context, accountIDs []string) ([]pacioli.Account, error) {
	return ledger.GetAccounts(ctx, l.b, accountIDs)
}

func (l localClient) CreateTransfers(ctx context.Context, args []pacioli.CreateTransferArgs) ([]pacioli.TransferResult, error) {
	return ledger.CreateTransfers(ctx, l.b, args)
}

func (l localClient) GetTransfers(ctx context.Context, transferIDs []string) ([]pacioli.Transfer, error) {
	return ledger.GetTransfers(ctx, l.b, transferIDs)
}

func (l localClient) PostTransfers(ctx context.Context, transferIDs []string) ([]pacioli.TransferResult, error) {
	return ledger.PostTransfers(ctx, l.b, transferIDs)
}

func (l localClient) VoidTransfers(ctx context.Context, transferIDs []string) ([]pacioli.TransferResult, error) {
	return ledger.VoidTransfers(ctx, l.b, transferIDs)
}
