package ops

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/openpayments"
)

type transactionState int

const (
	transactionStatePending   transactionState = 1
	transactionStateCompleted transactionState = 2
	transactionStateFailed    transactionState = 3
)

type createTransactionArgs struct {
	WalletID    string // Fynbos wallet ID
	ForeignID   string
	ForeignType openpayments.TransactionType
	Note        string
	State       transactionState
	Source      string // Usually the sending payment pointer
	Destination string // Usually the receiving payment pointer
	Amount      openpayments.Amount
}

type updateTransactionArgs struct {
	ForeignID string
	State     transactionState
	Amount    openpayments.Amount
}

type dbTransaction struct {
	ForeignID   string                       `db:"foreign_id"`
	ForeignType openpayments.TransactionType `db:"foreign_type"`
	Note        sql.NullString               `db:"note"`
	Source      sql.NullString               `db:"source"`
	Destination sql.NullString               `db:"destination"`
	Amount      uint64                       `db:"amount"`
	Scale       int                          `db:"asset_scale"`
	Asset       string                       `db:"asset_code"`
	Timestamp   time.Time                    `db:"updated_at"`
}

const (
	transactionCols = ` foreign_id, foreign_type, note, source, destination, amount, asset_scale, asset_code, updated_at `
)

func failTransaction(ctx context.Context, tx *sqlx.Tx, foreignID string) error {
	res, err := tx.ExecContext(ctx, "UPDATE openpayments_transactions SET state=$1, updated_at=now() WHERE foreign_id=$2",
		transactionStateFailed, foreignID)
	if err != nil {
		return fmt.Errorf("%w failed to update transaction %s", openpayments.ErrInternal, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w failed to get affected transaction rows %s", openpayments.ErrInternal, err)
	}
	if rows != 1 {
		return fmt.Errorf("%w wrong number of transaciton rows updated (%d)", openpayments.ErrInternal, rows)
	}

	return nil
}

func updateTransaction(ctx context.Context, tx *sqlx.Tx, args updateTransactionArgs) error {
	res, err := tx.ExecContext(ctx, "UPDATE openpayments_transactions SET state=$1, amount=$2, asset_code=$3, asset_scale=$4, updated_at=now() WHERE foreign_id=$5",
		args.State, args.Amount.Value, args.Amount.Asset, args.Amount.AssetScale, args.ForeignID)
	if err != nil {
		return fmt.Errorf("%w failed to update transaction %s", openpayments.ErrInternal, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w failed to get affected transaction rows %s", openpayments.ErrInternal, err)
	}
	if rows != 1 {
		return fmt.Errorf("%w wrong number of transaciton rows updated (%d)", openpayments.ErrInternal, rows)
	}

	return nil
}

func createTransaction(ctx context.Context, tx *sqlx.Tx, args createTransactionArgs) error {
	is := db.NewInsert("openpayments_transactions").
		Value("wallet_id", args.WalletID).
		Value("foreign_id", args.ForeignID).
		Value("foreign_type", args.ForeignType).
		Value("state", args.State).
		Value("amount", args.Amount.Value).
		Value("asset_code", args.Amount.Asset).
		Value("asset_scale", args.Amount.AssetScale)
	if args.Source != "" {
		is.Value("source", args.Source)
	}
	if args.Destination != "" {
		is.Value("destination", args.Destination)
	}
	if args.Note != "" {
		is.Value("note", args.Note)
	}

	stmt, qargs, err := is.GetStatement()
	if err != nil {
		return fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	_, err = tx.ExecContext(ctx, stmt, qargs...)
	if err != nil {
		return fmt.Errorf("%w failed to insert transaction %s", openpayments.ErrInternal, err)
	}

	return nil
}

func ListTransactions(ctx context.Context, b Backends, walletID string, page db.Pagination) ([]openpayments.Transaction, error) {
	return listTransactions(ctx, b, walletID, transactionStateCompleted, page)
}

func ListPendingTransactions(ctx context.Context, b Backends, walletID string, page db.Pagination) ([]openpayments.Transaction, error) {
	return listTransactions(ctx, b, walletID, transactionStatePending, page)
}

func listTransactions(ctx context.Context, b Backends, walletID string, state transactionState, page db.Pagination) ([]openpayments.Transaction, error) {
	var transactions []dbTransaction
	err := b.DB().SelectContext(ctx, &transactions,
		fmt.Sprintf("SELECT %s FROM openpayments_transactions WHERE wallet_id=$1 and state=$2 ORDER BY updated_at DESC ", transactionCols)+page.SQL(),
		walletID, state)
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	if len(transactions) == 0 {
		return nil, nil
	}

	resp := make([]openpayments.Transaction, len(transactions))

	for i, t := range transactions {
		resp[i] = openpayments.Transaction{
			ID:          t.ForeignID,
			Source:      t.Source.String,
			Destination: t.Destination.String,
			Type:        t.ForeignType,
			Timestamp:   t.Timestamp,
			Note:        t.Note.String,
			Amount: openpayments.Amount{
				Value:      t.Amount,
				Asset:      t.Asset,
				AssetScale: t.Scale,
			},
		}
	}

	return resp, err
}
