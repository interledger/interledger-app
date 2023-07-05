package ops

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/analytics"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/notify"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

func createTransaction(ctx context.Context, dbc sqlx.ExecerContext, b Backends, args transactions.CreateTransactionArgs) (string, error) {
	transID := uuid.NewString()
	if args.ID != "" {
		transID = args.ID
	}
	is := db.NewInsert("transactions").
		Value("id", transID).
		Value("wallet_id", args.WalletID).
		Value("type", args.ForeignType).
		Value("state", args.State).
		Value("provider", args.Provider).
		Value("amount", args.Amount.Value).
		Value("asset_code", args.Amount.Currency).
		Value("asset_scale", args.Amount.Scale).
		Value("grant_id", sql.NullString{
			String: args.GrantID,
			Valid:  args.GrantID != "",
		}).
		Value("linked_account_title", sql.NullString{
			String: args.LinkedAccountTitle,
			Valid:  args.LinkedAccountTitle != "",
		}).
		Value("destination_identity_type", sql.NullString{
			String: args.DestinationIdentityType,
			Valid:  args.DestinationIdentityType != "",
		}).
		Value("destination_identity", sql.NullString{
			String: args.DestinationIdentity,
			Valid:  args.DestinationIdentity != "",
		}).
		Value("reference", sql.NullString{
			String: args.Reference,
			Valid:  args.Reference != "",
		})
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

	err = b.Notify().NotifyWallet(ctx, args.WalletID, notify.NotificationTypeTransaction)
	if err != nil {
		log.Error("notify error", zap.Error(err))
	}

	userID := getWalletUserID(ctx, b, args.WalletID)
	b.Analytics().TrackWalletTransactionCreated(args.WalletID, analytics.WalletTransactionArgs{
		ID:       transID,
		TrxType:  args.ForeignType,
		Provider: args.Provider,
		Amount:   args.Amount,
		UserID:   userID,
	})

	return transID, nil
}

func CreateTransactionTx(ctx context.Context, b Backends, tx *sqlx.Tx, args transactions.CreateTransactionArgs) (string, error) {
	err := b.Validator().Struct(args)
	if err != nil {
		return "", fmt.Errorf("%w %s", transactions.ErrInvalidArgument, err)
	}

	return createTransaction(ctx, tx, b, args)
}

func CreateTransaction(ctx context.Context, b Backends, args transactions.CreateTransactionArgs) (string, error) {
	err := b.Validator().Struct(args)
	if err != nil {
		return "", fmt.Errorf("%w %s", transactions.ErrInvalidArgument, err)
	}

	var trxID = ""
	err = crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		id, err := createTransaction(ctx, tx, b, args)
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
	transactionCols = ` id, foreign_id, type, state, provider, note, source, destination, amount, asset_scale, asset_code, linked_account_title, destination_identity_type, destination_identity, reference, updated_at `
	transferCols    = ` id, foreign_id, linked_acc_id, type, state, amount, asset_scale, asset_code, updated_at `
)

type dbTransaction struct {
	ID                      string                       `db:"id"`
	ForeignID               sql.NullString               `db:"foreign_id"`
	Type                    transactions.TransactionType `db:"type"`
	State                   transactions.State           `db:"state"`
	Provider                transactions.Provider        `db:"provider"`
	Note                    sql.NullString               `db:"note"`
	Source                  sql.NullString               `db:"source"`
	Destination             sql.NullString               `db:"destination"`
	Amount                  uint64                       `db:"amount"`
	Scale                   int                          `db:"asset_scale"`
	Asset                   string                       `db:"asset_code"`
	Timestamp               time.Time                    `db:"updated_at"`
	LinkedAccountTitle      sql.NullString               `db:"linked_account_title"`
	DestinationIdentityType sql.NullString               `db:"destination_identity_type"`
	DestinationIdentity     sql.NullString               `db:"destination_identity"`
	Reference               sql.NullString               `db:"reference"`
}

func List(ctx context.Context, b Backends, walletID string, page db.Pagination) ([]transactions.Transaction, error) {

	sqlStmt := "SELECT %s FROM transactions WHERE wallet_id=$1 ORDER BY updated_at DESC,id  %s"
	sqlArgs := []interface{}{walletID}

	if page.PageToken != "" {
		sqlStmt = "SELECT %s FROM transactions WHERE wallet_id=$1 and updated_at<=(SELECT updated_at FROM transactions WHERE id=$2) ORDER BY updated_at DESC,id %s"
		sqlArgs = []interface{}{walletID, page.PageToken}
	}

	return listTransaction(ctx, b, page, sqlStmt, sqlArgs)
}

func ListCompleted(ctx context.Context, b Backends, walletID string, page db.Pagination) ([]transactions.Transaction, error) {

	sqlStmt := "SELECT %s FROM transactions WHERE wallet_id=$1 and state=$2 ORDER BY updated_at DESC,id %s"
	sqlArgs := []interface{}{walletID, transactions.StateCompleted}

	if page.PageToken != "" {
		sqlStmt = "SELECT %s FROM transactions WHERE wallet_id=$1 and state=$2 and updated_at<=(SELECT updated_at FROM transactions WHERE id=$3) ORDER BY updated_at DESC,id %s"
		sqlArgs = []interface{}{walletID, transactions.StateCompleted, page.PageToken}
	}

	return listTransaction(ctx, b, page, sqlStmt, sqlArgs)
}

func ListWithPending(ctx context.Context, b Backends, walletID string, page db.Pagination) ([]transactions.Transaction, error) {

	sqlStmt := "SELECT %s FROM transactions WHERE wallet_id=$1 and (state in ($2) or (state=$3 and type<>$4)) ORDER BY updated_at DESC,id %s"
	sqlArgs := []interface{}{walletID, transactions.StateCompleted, transactions.StatePending, transactions.TransactionTypeOpenPaymentIncoming}

	if page.PageToken != "" {
		sqlStmt = "SELECT %s FROM transactions WHERE wallet_id=$1 and (state in ($2) or (state=$3 and type<>$4)) and updated_at<=(SELECT updated_at FROM transactions WHERE id=$5) ORDER BY updated_at DESC,id %s"
		sqlArgs = []interface{}{walletID, transactions.StateCompleted, transactions.StatePending, transactions.TransactionTypeOpenPaymentIncoming, page.PageToken}
	}

	return listTransaction(ctx, b, page, sqlStmt, sqlArgs)
}

func listTransaction(ctx context.Context, b Backends, page db.Pagination, sqlStmt string, sqlArgs []interface{}) ([]transactions.Transaction, error) {
	var txs []dbTransaction
	err := b.DB().SelectContext(ctx, &txs,
		fmt.Sprintf(sqlStmt, transactionCols, page.SQL()),
		sqlArgs...)
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
			ID:                      t.ID,
			ForeignID:               t.ForeignID.String,
			Source:                  t.Source.String,
			Destination:             t.Destination.String,
			Type:                    t.Type,
			Timestamp:               t.Timestamp,
			Note:                    t.Note.String,
			State:                   t.State,
			Provider:                t.Provider,
			LinkedAccountTitle:      t.LinkedAccountTitle.String,
			DestinationIdentity:     t.DestinationIdentity.String,
			DestinationIdentityType: t.DestinationIdentityType.String,
			Reference:               t.Reference.String,
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
			ForeignID:   t.ForeignID.String,
			Source:      t.Source.String,
			Destination: t.Destination.String,
			Note:        t.Note.String,
			Type:        t.Type,
			Timestamp:   t.Timestamp,
			Provider:    t.Provider,
			State:       t.State,
			Amount: currency.Amount{
				Value:    t.Amount,
				Currency: currency.ParseCurrency(t.Asset),
				Scale:    t.Scale,
			},
			LinkedAccountTitle:      t.LinkedAccountTitle.String,
			DestinationIdentity:     t.DestinationIdentity.String,
			DestinationIdentityType: t.DestinationIdentityType.String,
			Transfers:               trs,
			Reference:               t.Reference.String,
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
		ID:                      tx.ID,
		ForeignID:               tx.ForeignID.String,
		Source:                  tx.Source.String,
		Destination:             tx.Destination.String,
		Type:                    tx.Type,
		Timestamp:               tx.Timestamp,
		Note:                    tx.Note.String,
		State:                   tx.State,
		Provider:                tx.Provider,
		LinkedAccountTitle:      tx.LinkedAccountTitle.String,
		DestinationIdentity:     tx.DestinationIdentity.String,
		DestinationIdentityType: tx.DestinationIdentityType.String,
		Reference:               tx.Reference.String,
		Amount: currency.Amount{
			Value:    tx.Amount,
			Currency: currency.ParseCurrency(tx.Asset),
			Scale:    tx.Scale,
		},
		Transfers: trs,
	}, nil
}

func GetTransactionByForeignID(ctx context.Context, b Backends, walletID string, foreignID string) (*transactions.Transaction, error) {
	id := foreignID
	idxSlash := strings.LastIndex(id, "/")
	if idxSlash > 0 {
		id = foreignID[idxSlash+1:]
	}

	var tx dbTransaction
	err := b.DB().GetContext(ctx, &tx,
		fmt.Sprintf("SELECT %s FROM transactions WHERE wallet_id=$1 and foreign_id=$2", transactionCols),
		walletID, id)
	if err != nil {
		return nil, fmt.Errorf("%w %s", transactions.ErrInternal, err)
	}

	trs, err := getTransfers(ctx, b, tx.ID)
	if err != nil {
		return nil, err
	}

	return &transactions.Transaction{
		ID:                      tx.ID,
		ForeignID:               tx.ForeignID.String,
		Source:                  tx.Source.String,
		Destination:             tx.Destination.String,
		Type:                    tx.Type,
		Timestamp:               tx.Timestamp,
		Note:                    tx.Note.String,
		State:                   tx.State,
		Provider:                tx.Provider,
		LinkedAccountTitle:      tx.LinkedAccountTitle.String,
		DestinationIdentity:     tx.DestinationIdentity.String,
		DestinationIdentityType: tx.DestinationIdentityType.String,
		Reference:               tx.Reference.String,
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
			State:     t.State,
			Timestamp: t.Timestamp,
		}
	}

	return res, nil
}

func SetTransactionForeignID(ctx context.Context, b Backends, ID string, foreignID string) error {
	var walletID string
	err := b.DB().GetContext(ctx, &walletID, "UPDATE transactions SET foreign_id=$1, updated_at=now() WHERE id=$2 returning wallet_id",
		foreignID, ID)
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInternal, err)
	}

	err = b.Notify().NotifyWallet(ctx, walletID, notify.NotificationTypeTransaction)
	if err != nil {
		log.Error("error sending notification", zap.Error(err))
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
	var trxDetails struct {
		ID       string                       `db:"id"`
		WalletID string                       `db:"wallet_id"`
		Amount   uint64                       `db:"amount"`
		Type     transactions.TransactionType `db:"type"`
		Provider transactions.Provider        `db:"provider"`
	}
	err := b.DB().GetContext(ctx, &trxDetails, "UPDATE transactions SET state=$1, updated_at=now() WHERE id=$2 returning id, wallet_id, amount, type, provider",
		state, ID)
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInternal, err)
	}

	err = b.Notify().NotifyWallet(ctx, trxDetails.WalletID, notify.NotificationTypeTransaction)
	if err != nil {
		log.Error("error sending notification", zap.Error(err))
	}

	userID := getWalletUserID(ctx, b, trxDetails.WalletID)
	analyticsArgs := analytics.WalletTransactionArgs{
		ID:       trxDetails.ID,
		TrxType:  trxDetails.Type,
		Provider: trxDetails.Provider,
		Amount:   currency.FromUInt64(trxDetails.Amount, currency.USD),
		UserID:   userID,
	}

	if state == transactions.StateCompleted {
		b.Analytics().TrackWalletTransactionCompleted(trxDetails.WalletID, analyticsArgs)
	} else if state == transactions.StateFailed {
		b.Analytics().TrackWalletTransactionFailed(trxDetails.WalletID, analyticsArgs)
	}

	return nil
}

func SetTransactionStateTx(ctx context.Context, b Backends, tx *sqlx.Tx, ID string, state transactions.State) error {
	var trxDetails struct {
		ID       string                       `db:"id"`
		WalletID string                       `db:"wallet_id"`
		Amount   uint64                       `db:"amount"`
		Type     transactions.TransactionType `db:"type"`
		Provider transactions.Provider        `db:"provider"`
	}
	err := tx.GetContext(ctx, &trxDetails, "UPDATE transactions SET state=$1, updated_at=now() WHERE id=$2 returning id, wallet_id, amount, type, provider",
		state, ID)
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInternal, err)
	}

	err = b.Notify().NotifyWallet(ctx, trxDetails.WalletID, notify.NotificationTypeTransaction)
	if err != nil {
		log.Error("error sending notification", zap.Error(err))
	}

	userID := getWalletUserID(ctx, b, trxDetails.WalletID)
	analyticsArgs := analytics.WalletTransactionArgs{
		ID:       trxDetails.ID,
		TrxType:  trxDetails.Type,
		Provider: trxDetails.Provider,
		Amount:   currency.FromUInt64(trxDetails.Amount, currency.USD),
		UserID:   userID,
	}

	if state == transactions.StateCompleted {
		b.Analytics().TrackWalletTransactionCompleted(trxDetails.WalletID, analyticsArgs)
	} else if state == transactions.StateFailed {
		b.Analytics().TrackWalletTransactionFailed(trxDetails.WalletID, analyticsArgs)
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
	var walletID string
	err := tx.GetContext(ctx, &walletID, "UPDATE transactions SET amount=$1, asset_code=$2, asset_scale=$3, updated_at=now() WHERE id=$4 returning wallet_id",
		amount.Value, amount.Currency.String(), amount.Scale, ID)
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInternal, err)
	}

	err = b.Notify().NotifyWallet(ctx, walletID, notify.NotificationTypeTransaction)
	if err != nil {
		log.Error("error sending notification", zap.Error(err))
	}

	return nil
}

func getWalletUserID(ctx context.Context, b Backends, walletID string) string {
	if b.Users() == nil {
		return ""
	}

	users, err := b.Users().ListUsers(ctx, walletID)
	if err != nil {
		return ""
	}

	// if there are more than 1 user or no users don't return anything
	if len(users) != 1 {
		return ""
	}

	firstUser := users[0]
	return firstUser.ID
}

func ListTransfers(ctx context.Context, b Backends, txID string) ([]transactions.Transfer, error) {
	return getTransfers(ctx, b, txID)
}
