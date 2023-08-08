package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"gitlab.com/fynbos/env"

	"github.com/google/uuid"

	"gitlab.com/fynbos/backend/currency"

	"gitlab.com/fynbos/backend/transactions"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/openpayments"
)

const outgoingPaymentCols = ` id, quote_id, failed, description, sent_amount, sent_asset, sent_scale, created_at, updated_at, created_by, sender_wallet_address, receiver_wallet_address `

type dbOutgoingPayments struct {
	ID                    string         `db:"id"`
	QuoteID               string         `db:"quote_id"`
	Failed                bool           `db:"failed"`
	Description           string         `db:"description"`
	AssetCode             string         `db:"sent_asset"`
	AssetScale            int            `db:"sent_scale"`
	SentAmount            uint64         `db:"sent_amount"`
	CreatedAt             time.Time      `db:"created_at"`
	UpdatedAt             time.Time      `db:"updated_at"`
	CreatedBy             sql.NullString `db:"created_by"`
	SenderWalletAddress   sql.NullString `db:"sender_wallet_address"`
	ReceiverWalletAddress sql.NullString `db:"receiver_wallet_address"`
}

func CreateOutgoingPayment(ctx context.Context, b Backends, args openpayments.CreateOutgoingPaymentArgs) (string, string, error) {
	qid := args.QuoteID
	idxSlash := strings.LastIndex(qid, "/")
	if idxSlash > 0 {
		qid = qid[idxSlash+1:]
	}

	q, err := getDBQuote(ctx, b, "id=$1", qid)
	if err != nil {
		return "", "", err
	}

	if q.ExpiresAt.Before(time.Now()) {
		return "", "", fmt.Errorf("%w %s", openpayments.ErrInvalidArgument, "quote has expired")
	}

	ip, err := GetIncomingPayment(ctx, b, q.IncomingPaymentID)
	if err != nil {
		return "", "", err
	}

	if args.Description == "" {
		args.Description = ip.Description
	}

	id := uuid.NewString()

	stmt, qargs, err := db.NewInsert("openpayments_outgoing_payment").
		Value("id", id).
		Value("quote_id", q.ID[strings.LastIndex(q.ID, "/")+1:]).
		Value("failed", false).
		Value("description", args.Description).
		Value("sent_amount", 0).
		Value("sent_asset", q.SendAsset).
		Value("sent_scale", q.SendAssetScale).
		Value("created_by", sql.NullString{
			String: args.CreatedBy,
			Valid:  args.CreatedBy != "",
		}).
		Value("receiver_wallet_address", q.ReceiverWalletAddress).
		Value("sender_wallet_address", q.SenderWalletAddress).GetStatement()
	if err != nil {
		return "", "", fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	if args.DestinationIdentity == "" {
		args.DestinationIdentity = q.ReceiverWalletAddress.String
		args.DestinationIdentityType = "wallet"
	} else if q.RecvIdentity.Valid && q.RecvIdentityType.Valid {
		args.DestinationIdentity = q.RecvIdentity.String
		args.DestinationIdentityType = q.RecvIdentityType.String
	}

	senderWallet, err := b.Wallets().GetFromAddress(ctx, q.SenderWalletAddress.String)
	if err != nil {
		return "", "", err
	}

	var trxID string
	err = crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx, stmt, qargs...)
		if err != nil {
			return fmt.Errorf("%w %s", openpayments.ErrInternal, err)
		}

		_, err = tx.ExecContext(ctx,
			"UPDATE openpayments_incoming_payment SET external_ref=$1 WHERE id=$2",
			args.Description,
			q.IncomingPaymentID)
		if err != nil {
			return fmt.Errorf("%w %s", openpayments.ErrInternal, err)
		}

		id, err := b.Transactions().CreateTransactionTx(ctx, tx, transactions.CreateTransactionArgs{
			WalletID:    senderWallet.ID,
			ForeignID:   id,
			ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
			Provider:    transactions.ProviderGMT,
			State:       transactions.StatePending,
			Note:        args.Description,
			Source:      q.SenderWalletAddress.String,
			Destination: q.ReceiverWalletAddress.String,
			Amount: currency.Amount{
				Value:    q.SendAmount,
				Currency: currency.ParseCurrency(q.SendAsset),
				Scale:    q.SendAssetScale,
			},
			GrantID:                 args.GrantID,
			LinkedAccountTitle:      args.LinkedAccountTitle,
			DestinationIdentity:     args.DestinationIdentity,
			DestinationIdentityType: args.DestinationIdentityType,
			Reference:               args.Description,
		})
		if err != nil {
			return fmt.Errorf("%w %s", openpayments.ErrInternal, err)
		}

		trxID = id

		return nil
	})
	if err != nil {
		return "", "", err
	}

	return fmt.Sprintf("%s/outgoing/%s", env.OpenPaymentsURL(), id), trxID, nil
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

func GetOutgoingPaymentByQuote(ctx context.Context, b Backends, qid string) (*openpayments.OutgoingPayment, error) {
	// Our friends may have provided the full ID with the payment pointer and the `incoming-payments` prefix.
	idxSlash := strings.LastIndex(qid, "/")
	if idxSlash > 0 {
		qid = qid[idxSlash+1:]
	}

	var op dbOutgoingPayments
	err := b.DB().GetContext(ctx, &op,
		fmt.Sprintf("SELECT %s FROM openpayments_outgoing_payment WHERE quote_id=$1", outgoingPaymentCols),
		qid)
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

	q := transformQuote(*dbq)

	return &openpayments.OutgoingPayment{
		ID:                fmt.Sprintf("%s/outgoing/%s", env.OpenPaymentsURL(), op.ID),
		PaymentPointer:    dbq.SenderWalletAddress.String,
		FromLinkedAccount: q.FromLinkedAccount,
		ToPaymentPointer:  dbq.ReceiverWalletAddress.String,
		Failed:            op.Failed,
		Receiver:          q.IncomingPayment,
		SendAmount:        q.SendAmount,
		ReceiveAmount:     q.ReceiveAmount,
		SentAmount: currency.Amount{
			Value:    op.SentAmount,
			Currency: currency.ParseCurrency(op.AssetCode),
			Scale:    op.AssetScale,
		},
		Description: op.Description,
		CreatedAt:   op.CreatedAt,
		UpdatedAt:   op.UpdatedAt,
		CreatedBy:   op.CreatedBy.String,
	}, nil
}

func FailOutgoingPayment(ctx context.Context, b Backends, id string) error {
	// Our friends may have provided the full ID with the payment pointer and the `incoming-payments` prefix.
	idxSlash := strings.LastIndex(id, "/")
	if idxSlash > 0 {
		id = id[idxSlash+1:]
	}

	op, err := GetOutgoingPayment(ctx, b, id)
	if err != nil {
		return err
	}

	wallet, err := b.Wallets().GetFromAddress(ctx, op.PaymentPointer)
	if err != nil {
		return err
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

		trx, err := b.Transactions().GetTransactionByForeignID(ctx, wallet.ID, id)
		if err != nil {
			return fmt.Errorf("%w %s", openpayments.ErrInternal, err)
		}

		return b.Transactions().SetTransactionState(ctx, trx.ID, transactions.StateFailed)
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

	receiverWallet, err := b.Wallets().GetFromAddress(ctx, op.ToPaymentPointer)
	if err != nil {
		return err
	}

	senderWallet, err := b.Wallets().GetFromAddress(ctx, op.PaymentPointer)
	if err != nil {
		return err
	}

	fromTrx, err := b.Transactions().GetTransactionByForeignID(ctx, senderWallet.ID, opID)
	if err != nil {
		return fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	var toTrx *transactions.Transaction
	if args.IncomingSuccess {
		toTrx, err = b.Transactions().GetTransactionByForeignID(ctx, receiverWallet.ID, ipID)
		if err != nil {
			return fmt.Errorf("%w %s", openpayments.ErrInternal, err)
		}
	}

	return crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		res, err := tx.ExecContext(ctx, "UPDATE openpayments_outgoing_payment SET updated_at=now(), completed=true, sent_amount=$1, sent_asset=$2, sent_scale=$3 WHERE id=$4 AND failed=false",
			args.SentAmount.Value, args.SentAmount.Currency, args.SentAmount.Scale, opID)
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

		err = b.Transactions().SetTransactionStateTx(ctx, tx, fromTrx.ID, transactions.StateCompleted)
		if err != nil {
			return err
		}
		err = b.Transactions().SetTransactionAmountTx(ctx, tx, fromTrx.ID, args.SentAmount)
		if err != nil {
			return err
		}

		if args.IncomingSuccess {
			res, err = tx.ExecContext(ctx, "UPDATE openpayments_incoming_payment SET updated_at=now(), received_amount=$1, asset_code=$2, asset_scale=$3, completed=true WHERE id=$4",
				args.SentAmount.Value, args.SentAmount.Currency, args.SentAmount.Scale, ipID)
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

			err = b.Transactions().SetTransactionStateTx(ctx, tx, toTrx.ID, transactions.StateCompleted)
			if err != nil {
				return err
			}
			err = b.Transactions().SetTransactionAmountTx(ctx, tx, toTrx.ID, args.SentAmount)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
