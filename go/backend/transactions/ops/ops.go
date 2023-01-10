package ops

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/currency"

	"github.com/google/uuid"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/transactions"
)

func createTransaction(ctx context.Context, dbc sqlx.ExecerContext, args transactions.CreateTransactionArgs) (string, error) {
	transID := uuid.NewString()
	is := db.NewInsert("transactions").
		Value("id", transID).
		Value("wallet_id", args.WalletID).
		Value("type", args.ForeignType).
		Value("state", args.State).
		Value("provider", args.Provider).
		Value("amount", args.Amount.Value).
		Value("asset_code", args.Amount.Currency).
		Value("asset_scale", args.Amount.Scale)
	if args.Source != "" {
		is.Value("source", args.Source)
	}
	if args.Destination != "" {
		is.Value("destination", args.Destination)
	}
	if args.Note != "" {
		is.Value("note", args.Note)
	}
	if args.ForeignID != "" {
		is.Value("foreign_id", args.ForeignID)
	}

	stmt, qargs, err := is.GetStatement()
	if err != nil {
		return "", fmt.Errorf("%w %s", transactions.ErrInternal, err)
	}

	_, err = dbc.ExecContext(ctx, stmt, qargs...)
	if err != nil && !db.IsErrorCode(err, db.UniqueViolationError) {
		return "", fmt.Errorf("%w %s", transactions.ErrInternal, err)
	}

	for _, transfer := range args.Transfers {
		err = addTransfer(ctx, dbc, transID, transfer)
		if err != nil {
			return "", err
		}
	}

	return transID, nil
}

func CreateTransactionTx(ctx context.Context, b Backends, tx *sqlx.Tx, args transactions.CreateTransactionArgs) (string, error) {
	err := b.Validator().Struct(args)
	if err != nil {
		return "", fmt.Errorf("%w %s", transactions.ErrInvalidArgument, err)
	}

	return createTransaction(ctx, tx, args)
}

func CreateTransaction(ctx context.Context, b Backends, args transactions.CreateTransactionArgs) (string, error) {
	err := b.Validator().Struct(args)
	if err != nil {
		return "", fmt.Errorf("%w %s", transactions.ErrInvalidArgument, err)
	}

	var trxID = ""
	err = crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		id, err := createTransaction(ctx, tx, args)
		trxID = id
		return err
	})
	if err != nil {
		return "", err
	}

	return trxID, nil
}

func addTransfer(ctx context.Context, dbc sqlx.ExecerContext, transactionID string, args transactions.TransferArgs) error {
	is := db.NewInsert("transfers").
		Value("transaction_id", transactionID).
		Value("linked_acc_id", args.LinkedAccountID).
		Value("type", args.Type).
		Value("state", args.State).
		Value("amount", args.Amount.Value).
		Value("asset_code", args.Amount.Currency).
		Value("asset_scale", args.Amount.Scale)
	if args.ForeignID != "" {
		is.Value("foreign_id", args.ForeignID)
	}

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

func AddTransfersTx(ctx context.Context, b Backends, tx *sqlx.Tx, trxID string, args []transactions.TransferArgs) error {
	if len(args) <= 0 {
		return nil
	}

	// Run validation
	err := b.Validator().Var(args, "dive")
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInvalidArgument, err)
	}

	for _, a := range args {
		err = addTransfer(ctx, tx, trxID, a)
		if err != nil {
			return err
		}
	}

	return nil
}

func AddTransfers(ctx context.Context, b Backends, trxID string, transferArgs []transactions.TransferArgs) error {
	if len(transferArgs) <= 0 {
		return nil
	}

	// Run validation
	err := b.Validator().Var(transferArgs, "dive")
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInvalidArgument, err)
	}

	return crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		for _, a := range transferArgs {
			err = addTransfer(ctx, tx, trxID, a)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

const (
	transactionCols = ` id, foreign_id, type, state, provider, note, source, destination, amount, asset_scale, asset_code, updated_at `
	transferCols    = ` id, foreign_id, linked_acc_id, type, state, amount, asset_scale, asset_code, updated_at `
)

type dbTransaction struct {
	ID          string                       `db:"id"`
	ForeignID   sql.NullString               `db:"foreign_id"`
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
			ID:          t.ID,
			ForeignID:   t.ForeignID.String,
			Source:      t.Source.String,
			Destination: t.Destination.String,
			Type:        t.Type,
			Timestamp:   t.Timestamp,
			Note:        t.Note.String,
			State:       t.State,
			Provider:    t.Provider,
			Amount: currency.Amount{
				Value:    t.Amount,
				Currency: currency.ParseCurrency(t.Asset),
				Scale:    t.Scale,
			},
			Transfers: trs,
		}
	}

	return resp, err
}

func ListTransactionsInRange(ctx context.Context, b Backends, walletID string, inRange transactions.TransactionRangeFilter) ([]transactions.Transaction, error) {
	selectByWallet := fmt.Sprintf("SELECT %s FROM transactions WHERE wallet_id=$1", transactionCols)
	andByState := "and (state in ($2,$3) or (state=$4 and type<>$5))"
	andByDateRange := "and ($6 <= updated_at and updated_at <= $7)"
	orderAndPaginate := "ORDER BY updated_at DESC"
	query := strings.Join([]string{selectByWallet, andByState, andByDateRange, orderAndPaginate}, " ")

	var txs []dbTransaction
	err := b.DB().SelectContext(ctx, &txs, query,
		walletID, transactions.StateCompleted, transactions.StateFailed, transactions.StatePending, transactions.TransactionTypeOpenPaymentIncoming, inRange.StartTimestamp, inRange.EndTimestamp)
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
			ID:          t.ID,
			ForeignID:   t.ForeignID,
			Source:      t.Source.String,
			Destination: t.Destination.String,
			Type:        t.Type,
			Timestamp:   t.Timestamp,
			Note:        t.Note.String,
			State:       t.State,
			Provider:    t.Provider,
			Amount: currency.Amount{
				Value:    t.Amount,
				Currency: currency.ParseCurrency(t.Asset),
				Scale:    t.Scale,
			},
			Transfers: trs,
		}
	}

	return resp, err
}

func GetTransaction(ctx context.Context, b Backends, walletID string, trxID string) (*transactions.Transaction, error) {
	var tx dbTransaction
	err := b.DB().GetContext(ctx, &tx,
		fmt.Sprintf("SELECT %s FROM transactions WHERE wallet_id=$1 and id=$2", transactionCols),
		walletID, trxID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", transactions.ErrInternal, err)
	}

	trs, err := getTransfers(ctx, b, tx.ID)
	if err != nil {
		return nil, err
	}

	return &transactions.Transaction{
		ID:          tx.ID,
		ForeignID:   tx.ForeignID.String,
		Source:      tx.Source.String,
		Destination: tx.Destination.String,
		Type:        tx.Type,
		Timestamp:   tx.Timestamp,
		Note:        tx.Note.String,
		State:       tx.State,
		Provider:    tx.Provider,
		Amount: currency.Amount{
			Value:    tx.Amount,
			Currency: currency.ParseCurrency(tx.Asset),
			Scale:    tx.Scale,
		},
		Transfers: trs,
	}, nil
}

func GetTransactionByForeignID(ctx context.Context, b Backends, walletID string, foreignID string) (*transactions.Transaction, error) {
	var tx dbTransaction
	err := b.DB().GetContext(ctx, &tx,
		fmt.Sprintf("SELECT %s FROM transactions WHERE wallet_id=$1 and foreign_id=$2", transactionCols),
		walletID, foreignID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", transactions.ErrInternal, err)
	}

	trs, err := getTransfers(ctx, b, tx.ID)
	if err != nil {
		return nil, err
	}

	return &transactions.Transaction{
		ID:          tx.ID,
		ForeignID:   tx.ForeignID.String,
		Source:      tx.Source.String,
		Destination: tx.Destination.String,
		Type:        tx.Type,
		Timestamp:   tx.Timestamp,
		Note:        tx.Note.String,
		State:       tx.State,
		Provider:    tx.Provider,
		Amount: currency.Amount{
			Value:    tx.Amount,
			Currency: currency.ParseCurrency(tx.Asset),
			Scale:    tx.Scale,
		},
		Transfers: trs,
	}, nil
}

type dbTransfer struct {
	ID              string                    `db:"id"`
	TransactionID   string                    `db:"transaction_id"`
	ForeignID       sql.NullString            `db:"foreign_id"`
	LinkedAccountID sql.NullString            `db:"linked_acc_id"`
	Type            transactions.TransferType `db:"type"`
	State           transactions.State        `db:"state"`
	Amount          uint64                    `db:"amount"`
	Scale           int                       `db:"asset_scale"`
	Asset           string                    `db:"asset_code"`
	Timestamp       time.Time                 `db:"updated_at"`
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
			ID:              t.ID,
			Type:            t.Type,
			ForeignID:       t.ForeignID.String,
			LinkedAccountID: t.LinkedAccountID.String,
			Amount: currency.Amount{
				Value:    t.Amount,
				Currency: currency.ParseCurrency(t.Asset),
				Scale:    t.Scale,
			},
			State: t.State,
		}
	}

	return res, nil
}

func SetTransactionForeignID(ctx context.Context, b Backends, ID string, foreignID string) error {
	_, err := b.DB().ExecContext(ctx, "UPDATE transactions SET foreign_id=$1, updated_at=now() WHERE id=$2",
		foreignID, ID)
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInternal, err)
	}

	return nil
}

func SetTransferForeignID(ctx context.Context, b Backends, ID string, foreignID string) error {
	_, err := b.DB().ExecContext(ctx, "UPDATE transfers SET foreign_id=$1, updated_at=now() WHERE id=$2",
		foreignID, ID)
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInternal, err)
	}

	return nil
}

func SetTransactionState(ctx context.Context, b Backends, ID string, state transactions.State) error {
	_, err := b.DB().ExecContext(ctx, "UPDATE transactions SET state=$1, updated_at=now() WHERE id=$2",
		state, ID)
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInternal, err)
	}

	return nil
}

func SetTransactionStateTx(ctx context.Context, b Backends, tx *sqlx.Tx, ID string, state transactions.State) error {
	_, err := tx.ExecContext(ctx, "UPDATE transactions SET state=$1, updated_at=now() WHERE id=$2",
		state, ID)
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInternal, err)
	}

	return nil
}

func SetTransferState(ctx context.Context, b Backends, ID string, state transactions.State) error {
	_, err := b.DB().ExecContext(ctx, "UPDATE transfers SET state=$1, updated_at=now() WHERE id=$2",
		state, ID)
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInternal, err)
	}

	return nil
}

func SetTransactionAmountTx(ctx context.Context, b Backends, tx *sqlx.Tx, ID string, amount currency.Amount) error {
	_, err := tx.ExecContext(ctx, "UPDATE transactions SET amount=$1, asset_code=$2, asset_scale=$3, updated_at=now() WHERE id=$4",
		amount.Value, amount.Currency.String(), amount.Scale, ID)
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInternal, err)
	}

	return nil
}
