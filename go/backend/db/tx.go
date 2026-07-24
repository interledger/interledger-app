package db

import (
	"context"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/jmoiron/sqlx"
)

type TxRunner struct {
	db *sqlx.DB
}

func NewTxRunner(db *sqlx.DB) *TxRunner {
	return &TxRunner{db: db}
}

func (r *TxRunner) WithTx(ctx context.Context, fn func(*sqlx.Tx) error) error {
	return crdbsqlx.ExecuteTx(ctx, r.db, nil, fn)
}
