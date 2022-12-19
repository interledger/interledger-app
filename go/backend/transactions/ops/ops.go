package ops

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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
		Value("provider", args.Provider).
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

func updateTransaction(ctx context.Context, dbc *sqlx.Tx, args transactions.UpdateTransactionArgs) error {
	res, err := dbc.ExecContext(ctx, "UPDATE transactions SET state=$1, amount=$2, asset_code=$3, asset_scale=$4, updated_at=now() WHERE foreign_id=$5 AND wallet_id=$6",
		args.State, args.Amount.Value, args.Amount.Asset, args.Amount.AssetScale, args.ForeignID, args.WalletID)
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

	for _, transfer := range args.UpdateTransfers {
		tid, err := getTransactionID(ctx, dbc, transfer.ForeignID, transfer.WalletID)
		if err != nil {
			return err
		}

		err = updateTransfer(ctx, dbc, tid, transfer)
		if err != nil {
			return err
		}
	}

	return nil
}

func UpdateTransaction(ctx context.Context, b Backends, args transactions.UpdateTransactionArgs) error {
	err := b.Validator().Struct(args)
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInvalidArgument, err)
	}

	return crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		return updateTransaction(ctx, tx, args)
	})
}

func UpdateTransactionTx(ctx context.Context, b Backends, tx *sqlx.Tx, args transactions.UpdateTransactionArgs) error {
	err := b.Validator().Struct(args)
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInvalidArgument, err)
	}

	return updateTransaction(ctx, tx, args)
}

func addTransfer(ctx context.Context, dbc sqlx.ExecerContext, transactionID string, args transactions.TransferArgs) error {
	is := db.NewInsert("transfers").
		Value("transaction_id", transactionID).
		Value("foreign_id", args.ForeignID).
		Value("type", args.Type).
		Value("state", args.State).
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

func AddTransfersTx(ctx context.Context, b Backends, tx *sqlx.Tx, args []transactions.TransferArgs) error {
	if len(args) <= 0 {
		return nil
	}

	err := b.Validator().Struct(args)
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInvalidArgument, err)
	}

	for _, a := range args {
		tid, err := getTransactionID(ctx, tx, a.TransactionForeignID, a.WalletID)
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

func AddTransfers(ctx context.Context, b Backends, args []transactions.TransferArgs) error {
	if len(args) <= 0 {
		return nil
	}

	err := b.Validator().Struct(args)
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInvalidArgument, err)
	}

	return crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		for _, a := range args {
			tid, err := getTransactionID(ctx, tx, a.TransactionForeignID, a.WalletID)
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

func UpdateTransfersTx(ctx context.Context, b Backends, tx *sqlx.Tx, args []transactions.TransferArgs) error {
	if len(args) <= 0 {
		return nil
	}

	err := b.Validator().Var(args, "dive")
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInvalidArgument, err)
	}

	for _, a := range args {
		tid, err := getTransactionID(ctx, tx, a.TransactionForeignID, a.WalletID)
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

func UpdateTransfers(ctx context.Context, b Backends, args []transactions.TransferArgs) error {
	if len(args) <= 0 {
		return nil
	}

	err := b.Validator().Var(args, "dive")
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInvalidArgument, err)
	}

	return crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		for _, a := range args {
			tid, err := getTransactionID(ctx, tx, a.TransactionForeignID, a.WalletID)
			if err != nil {
				return err
			}

			err = updateTransfer(ctx, tx, tid, a)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func updateTransfer(ctx context.Context, dbc sqlx.ExecerContext, transactionID string, args transactions.TransferArgs) error {
	res, err := dbc.ExecContext(ctx, "UPDATE transfers SET state=$1, amount=$2, asset_code=$3, asset_scale=$4, updated_at=now() WHERE foreign_id=$5 AND transaction_id=$6 AND type=$7",
		args.State, args.Amount.Value, args.Amount.Asset, args.Amount.AssetScale, args.ForeignID, transactionID, args.Type)
	if err != nil {
		return fmt.Errorf("%w failed to update transfer %s", transactions.ErrInternal, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w failed to get affected transfer rows %s", transactions.ErrInternal, err)
	}
	if rows != 1 {
		return fmt.Errorf("%w wrong number of transfer rows updated (%d)", transactions.ErrInternal, rows)
	}

	return nil
}

func getTransactionID(ctx context.Context, dbc sqlx.QueryerContext, fid, walletID string) (string, error) {
	var transID string
	row := dbc.QueryRowxContext(ctx, "SELECT id FROM  transactions WHERE foreign_id=$1 AND wallet_id=$2", fid, walletID)
	if row.Err() != nil {
		return "", fmt.Errorf("%w %s", transactions.ErrInternal, row.Err())
	}
	err := row.Scan(&transID)
	if err != nil {
		return "", fmt.Errorf("%w %s", transactions.ErrInternal, row.Err())
	}

	return transID, nil
}

const (
	transactionCols = ` id, foreign_id, type, state, provider, note, source, destination, amount, asset_scale, asset_code, updated_at `
	transferCols    = ` foreign_id, type, state, amount, asset_scale, asset_code, updated_at `
)

type dbTransaction struct {
	ID          string                       `db:"id"`
	ForeignID   string                       `db:"foreign_id"`
	Type        transactions.TransactionType `db:"type"`
	State       transactions.State           `db:"state"`
	Provider    transactions.Provider        `db:"provider"`
	Note        sql.NullString               `db:"note"`
	Source      sql.NullString               `db:"source"`
	Destination sql.NullString               `db:"destination"`
	Amount      uint64                       `db:"amount"`
	Scale       int                          `db:"asset_scale"`
	Asset       string                       `db:"asset_code"`
	Timestamp   time.Time                    `db:"updated_at"`
}

func ListTransactions(ctx context.Context, b Backends, walletID string, page db.Pagination) ([]transactions.Transaction, error) {
	var txs []dbTransaction
	err := b.DB().SelectContext(ctx, &txs,
		fmt.Sprintf("SELECT %s FROM transactions WHERE wallet_id=$1 and (state in ($2,$3) or (state=$4 and type<>$5)) ORDER BY updated_at DESC %s", transactionCols, page.SQL()),
		walletID, transactions.StateCompleted, transactions.StateFailed, transactions.StatePending, transactions.TransactionTypeOpenPaymentIncoming)
	if err != nil {
		return nil, fmt.Errorf("%w %s", transactions.ErrInternal, err)
	}

	if len(txs) == 0 {
		return nil, nil
	}

	resp := make([]transactions.Transaction, len(txs))

	for i, t := range txs {
		trs, err := getTransfers(ctx, b, t.ID)
		if err != nil {
			return nil, err
		}
		resp[i] = transactions.Transaction{
			Source:      t.Source.String,
			Destination: t.Destination.String,
			Type:        t.Type,
			Timestamp:   t.Timestamp,
			Note:        t.Note.String,
			State:       t.State,
			Provider:    t.Provider,
			Amount: transactions.Amount{
				Value:      t.Amount,
				Asset:      t.Asset,
				AssetScale: t.Scale,
			},
			Transfers: trs,
		}
	}

	return resp, err
}

type dbTransfer struct {
	TransactionID string                    `db:"transaction_id"`
	ForeignID     string                    `db:"foreign_id"`
	Type          transactions.TransferType `db:"type"`
	State         transactions.State        `db:"state"`
	Amount        uint64                    `db:"amount"`
	Scale         int                       `db:"asset_scale"`
	Asset         string                    `db:"asset_code"`
	Timestamp     time.Time                 `db:"updated_at"`
}

func getTransfers(ctx context.Context, b Backends, txID string) ([]transactions.Transfer, error) {
	var trs []dbTransfer
	err := b.DB().SelectContext(ctx, &trs, fmt.Sprintf("SELECT %s FROM transfers WHERE transaction_id=$1", transferCols), txID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", transactions.ErrInternal, err)
	}

	if len(trs) == 0 {
		return nil, nil
	}

	res := make([]transactions.Transfer, len(trs))
	for i, t := range trs {
		res[i] = transactions.Transfer{
			Type: t.Type,
			Amount: transactions.Amount{
				Value:      t.Amount,
				Asset:      t.Asset,
				AssetScale: t.Scale,
			},
			State: t.State,
		}
	}

	return res, nil
}
