package ops

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/transactions"
)

func createTransaction(ctx context.Context, dbc sqlx.ExecerContext, args transactions.CreateTransactionArgs) error {
	transID := uuid.NewString()
	is := db.NewInsert("transactions").
		Value("id", transID).
		Value("wallet_id", args.WalletID).
		Value("foreign_id", args.ForeignID).
		Value("type", args.ForeignType).
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
		return fmt.Errorf("%w %s", transactions.ErrInternal, err)
	}

	_, err = dbc.ExecContext(ctx, stmt, qargs...)
	if err != nil && !db.IsErrorCode(err, db.UniqueViolationError) {
		return fmt.Errorf("%w %s", transactions.ErrInternal, err)
	}

	for _, transfer := range args.Transfers {
		err = addTransfer(ctx, dbc, transID, transfer)
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateTransactionTx(ctx context.Context, b Backends, tx *sqlx.Tx, args transactions.CreateTransactionArgs) error {
	err := b.Validator().Struct(args)
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInvalidArgument, err)
	}

	return createTransaction(ctx, tx, args)
}

func CreateTransaction(ctx context.Context, b Backends, args transactions.CreateTransactionArgs) error {
	err := b.Validator().Struct(args)
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInvalidArgument, err)
	}

	err = crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		return createTransaction(ctx, tx, args)
	})

	return err
}

func updateTransaction(ctx context.Context, dbc sqlx.ExecerContext, args transactions.UpdateTransactionArgs) error {
	res, err := dbc.ExecContext(ctx, "UPDATE transactions SET state=$1, amount=$2, asset_code=$3, asset_scale=$4, updated_at=now() WHERE foreign_id=$5",
		args.State, args.Amount.Value, args.Amount.Asset, args.Amount.AssetScale, args.ForeignID)
	if err != nil {
		return fmt.Errorf("%w failed to update transaction %s", transactions.ErrInternal, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w failed to get affected transaction rows %s", transactions.ErrInternal, err)
	}
	if rows != 1 {
		return fmt.Errorf("%w wrong number of transaciton rows updated (%d)", transactions.ErrInternal, rows)
	}

	return nil
}

func UpdateTransaction(ctx context.Context, b Backends, args transactions.UpdateTransactionArgs) error {
	err := b.Validator().Struct(args)
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInvalidArgument, err)
	}

	return updateTransaction(ctx, b.DB(), args)
}

func UpdateTransactionTx(ctx context.Context, b Backends, tx *sqlx.Tx, args transactions.UpdateTransactionArgs) error {
	err := b.Validator().Struct(args)
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInvalidArgument, err)
	}

	return updateTransaction(ctx, tx, args)
}

func addTransfer(ctx context.Context, dbc sqlx.ExecerContext, transactionID string, args transactions.CreateTransferArgs) error {
	is := db.NewInsert("transfers").
		Value("transaction_id", transactionID).
		Value("foreign_id", args.ForeignID).
		Value("type", args.Type).
		Value("amount", args.Amount.Value).
		Value("asset_code", args.Amount.Asset).
		Value("asset_scale", args.Amount.AssetScale)

	stmt, qargs, err := is.GetStatement()
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInternal, err)
	}

	_, err = dbc.ExecContext(ctx, stmt, qargs...)
	if err != nil && !db.IsErrorCode(err, db.UniqueViolationError) {
		return fmt.Errorf("%w %s", transactions.ErrInternal, err)
	}

	return nil
}

func AddTransfersTx(ctx context.Context, b Backends, tx *sqlx.Tx, args []transactions.CreateTransferArgs) error {
	if len(args) <= 0 {
		return nil
	}

	err := b.Validator().Struct(args)
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInvalidArgument, err)
	}

	for _, a := range args {
		tid, err := getTransactionID(ctx, b.DB(), a.TransactionForeignID)
		if err != nil {
			return err
		}

		err = addTransfer(ctx, tx, tid, a)
		if err != nil {
			return err
		}
	}

	return nil
}

func AddTransfers(ctx context.Context, b Backends, args []transactions.CreateTransferArgs) error {
	if len(args) <= 0 {
		return nil
	}

	err := b.Validator().Struct(args)
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInvalidArgument, err)
	}

	return crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		for _, a := range args {
			tid, err := getTransactionID(ctx, tx, a.TransactionForeignID)
			if err != nil {
				return err
			}

			err = addTransfer(ctx, tx, tid, a)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func getTransactionID(ctx context.Context, dbc sqlx.QueryerContext, fid string) (string, error) {
	var transID string
	row := dbc.QueryRowxContext(ctx, "SELECT id FROM  transactions WHERE foreign_id=$1", fid)
	if row.Err() != nil {
		return "", fmt.Errorf("%w %s", transactions.ErrInternal, row.Err())
	}
	err := row.Scan(&transID)
	if err != nil {
		return "", fmt.Errorf("%w %s", transactions.ErrInternal, row.Err())
	}

	return transID, nil
}
