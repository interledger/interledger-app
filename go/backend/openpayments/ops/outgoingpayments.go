package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/transactions"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/openpayments"
)

const outgoingPaymentCols = ` id, to_payment_pointer_id, quote_id, failed, description, sent_amount, sent_asset, sent_scale, created_at, updated_at `

type dbOutgoingPayments struct {
	ID                 string    `db:"id"`
	QuoteID            string    `db:"quote_id"`
	ToPaymentPointerID string    `db:"to_payment_pointer_id"`
	Failed             bool      `db:"failed"`
	Description        string    `db:"description"`
	AssetCode          string    `db:"sent_asset"`
	AssetScale         int       `db:"sent_scale"`
	SentAmount         uint64    `db:"sent_amount"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}

func CreateOutgoingPayment(ctx context.Context, b Backends, args openpayments.CreateOutgoingPaymentArgs) (string, error) {
	qid := args.QuoteID
	idxSlash := strings.LastIndex(qid, "/")
	if idxSlash > 0 {
		qid = qid[idxSlash+1:]
	}

	q, err := getDBQuote(ctx, b, "id=$1", qid)
	if err != nil {
		return "", err
	}

	if q.ExpiresAt.Before(time.Now()) {
		return "", fmt.Errorf("%w %s", openpayments.ErrInvalidArgument, "quote has expired")
	}

	ip, err := GetIncomingPayment(ctx, b, q.IncomingPaymentID)
	if err != nil {
		return "", err
	}

	if args.Description == "" {
		args.Description = ip.Description
	}

	id := uuid.NewString()

	stmt, qargs, err := db.NewInsert("openpayments_outgoing_payment").
		Value("id", id).
		Value("quote_id", q.ID[strings.LastIndex(q.ID, "/")+1:]).
		Value("to_payment_pointer_id", q.ReceivePaymentPointer).
		Value("failed", false).
		Value("description", args.Description).
		Value("sent_amount", 0).
		Value("sent_asset", q.SendAsset).
		Value("sent_scale", q.SendAssetScale).GetStatement()
	if err != nil {
		return "", fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	toPP, err := getPaymentPointerByID(ctx, b, q.ReceivePaymentPointer)
	if err != nil {
		return "", err
	}

	fromPP, err := getPaymentPointerByID(ctx, b, q.SendPaymentPointer)
	if err != nil {
		return "", err
	}

	err = crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx, stmt, qargs...)
		if err != nil {
			return fmt.Errorf("%w %s", openpayments.ErrInternal, err)
		}

		_, err = tx.ExecContext(ctx,
			"UPDATE openpayments_incoming_payment SET external_ref=$1 WHERE id=$2",
			args.ExternalRef,
			q.IncomingPaymentID)
		if err != nil {
			return fmt.Errorf("%w %s", openpayments.ErrInternal, err)
		}

		return b.Transactions().CreateTransactionTx(ctx, tx, transactions.CreateTransactionArgs{
			WalletID:    fromPP.WalletID,
			ForeignID:   id,
			ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
			Provider:    transactions.ProviderMachnet,
			Note:        args.Description,
			State:       transactions.StatePending,
			Source:      fromPP.URL,
			Destination: toPP.URL,
			Amount: transactions.Amount{
				Value:      q.SendAmount,
				Asset:      q.SendAsset,
				AssetScale: q.SendAssetScale,
			},
		})
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/outgoing-payments/%s", fromPP.URL, id), nil
}

func GetOutgoingPayment(ctx context.Context, b Backends, id string) (*openpayments.OutgoingPayment, error) {
	// Our friends may have provided the full ID with the payment pointer and the `incoming-payments` prefix.
	idxSlash := strings.LastIndex(id, "/")
	if idxSlash > 0 {
		id = id[idxSlash+1:]
	}

	var op dbOutgoingPayments
	err := b.DB().GetContext(ctx, &op,
		fmt.Sprintf("SELECT %s FROM openpayments_outgoing_payment WHERE id=$1", outgoingPaymentCols),
		id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", openpayments.ErrNotFound, err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	return transformOutgoingPayment(ctx, b, op)
}

func transformOutgoingPayment(ctx context.Context, b Backends, op dbOutgoingPayments) (*openpayments.OutgoingPayment, error) {
	dbq, err := getDBQuote(ctx, b, "id=$1", op.QuoteID)
	if err != nil {
		return nil, err
	}

	q, err := transformQuote(ctx, b, *dbq)
	if err != nil {
		return nil, err
	}

	toPP, err := getPaymentPointerByID(ctx, b, dbq.ReceivePaymentPointer)
	if err != nil {
		return nil, err
	}

	return &openpayments.OutgoingPayment{
		ID:                fmt.Sprintf("%s/outgoing-payments/%s", q.PaymentPointer, op.ID),
		PaymentPointer:    q.PaymentPointer,
		FromLinkedAccount: q.FromLinkedAccount,
		ToPaymentPointer:  toPP.URL,
		Failed:            op.Failed,
		Receiver:          q.IncomingPayment,
		SendAmount:        q.SendAmount,
		ReceiveAmount:     q.ReceiveAmount,
		SentAmount: openpayments.Amount{
			Value:      op.SentAmount,
			Asset:      op.AssetCode,
			AssetScale: op.AssetScale,
		},
		Description: op.Description,
		CreatedAt:   op.CreatedAt,
		UpdatedAt:   op.UpdatedAt,
	}, nil
}

func FailOutgoingPayment(ctx context.Context, b Backends, id string) error {
	// Our friends may have provided the full ID with the payment pointer and the `incoming-payments` prefix.
	idxSlash := strings.LastIndex(id, "/")
	if idxSlash > 0 {
		id = id[idxSlash+1:]
	}

	return crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		res, err := tx.ExecContext(ctx, "UPDATE openpayments_outgoing_payment SET failed=true, completed=true, updated_at=now() WHERE id=$1", id)
		if err != nil {
			return fmt.Errorf("%w %s", openpayments.ErrInternal, err)
		}
		rowCnt, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("%w %s", openpayments.ErrInternal, err)
		}
		if rowCnt != 1 {
			return fmt.Errorf("%w outoing payment (%s) not found", openpayments.ErrNotFound, id)
		}

		return failTransaction(ctx, tx, id)
	})
}

func CompleteOutgoingPayment(ctx context.Context, b Backends, args openpayments.CompleteOutgoingPaymentArgs) error {
	// Our friends may have provided the full ID with the payment pointer and the `incoming-payments` prefix.
	opID := args.ID
	idxSlash := strings.LastIndex(args.ID, "/")
	if idxSlash > 0 {
		opID = opID[idxSlash+1:]
	}

	op, err := GetOutgoingPayment(ctx, b, opID)
	if err != nil {
		return err
	}

	ipID := op.Receiver
	idxSlash = strings.LastIndex(ipID, "/")
	if idxSlash > 0 {
		ipID = ipID[idxSlash+1:]
	}

	toPP, err := GetPaymentPointer(ctx, b, op.ToPaymentPointer)
	if err != nil {
		return err
	}

	fromPP, err := GetPaymentPointer(ctx, b, op.PaymentPointer)
	if err != nil {
		return err
	}

	return crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		res, err := tx.ExecContext(ctx, "UPDATE openpayments_outgoing_payment SET updated_at=now(), completed=true, sent_amount=$1, sent_asset=$2, sent_scale=$3 WHERE id=$4 AND failed=false",
			args.SentAmount.Value, args.SentAmount.Asset, args.SentAmount.AssetScale, opID)
		if err != nil {
			return fmt.Errorf("%w %s", openpayments.ErrInternal, err)
		}
		rowCnt, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("%w %s", openpayments.ErrInternal, err)
		}
		if rowCnt != 1 {
			return fmt.Errorf("%w outoing payment (%s) not found", openpayments.ErrNotFound, opID)
		}

		err = updateTransaction(ctx, tx, updateTransactionArgs{
			ForeignID: opID,
			State:     transactionStateCompleted,
			Amount:    args.SentAmount,
		})
		if err != nil {
			return err
		}

		res, err = tx.ExecContext(ctx, "UPDATE openpayments_incoming_payment SET updated_at=now(), received_amount=$1, asset_code=$2, asset_scale=$3, completed=true WHERE id=$4 AND received_amount<=$1",
			args.SentAmount.Value, args.SentAmount.Asset, args.SentAmount.AssetScale, ipID)
		if err != nil {
			return fmt.Errorf("%w %s", openpayments.ErrInternal, err)
		}
		rowCnt, err = res.RowsAffected()
		if err != nil {
			return fmt.Errorf("%w %s", openpayments.ErrInternal, err)
		}
		if rowCnt != 1 {
			return fmt.Errorf("%w incoming payment (%s) not found", openpayments.ErrNotFound, ipID)
		}

		return createTransaction(ctx, tx, createTransactionArgs{
			WalletID:    toPP.WalletID,
			ForeignID:   ipID,
			ForeignType: openpayments.TransactionTypeIncomingPayment,
			Note:        op.Description,
			State:       transactionStateCompleted,
			Source:      fromPP.URL,
			Destination: toPP.URL,
			Amount:      args.SentAmount,
		})
	})
}
