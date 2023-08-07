package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/wallets"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/env"
)

const incomingPaymentsCols = `id, description, asset_code, asset_scale, incoming_amount, received_amount, completed, expires_at, external_ref, ilp_stream_id, ilp_address, ilp_shared_secret, created_at, updated_at, created_by, sender_wallet_address, receiver_wallet_address`

type dbIncomingPayment struct {
	ID                    string         `db:"id"`
	Description           sql.NullString `db:"description"`
	AssetCode             sql.NullString `db:"asset_code"`
	AssetScale            sql.NullInt32  `db:"asset_scale"`
	IncomingAmount        uint64         `db:"incoming_amount"`
	ReceivedAmount        uint64         `db:"received_amount"`
	Completed             bool           `db:"completed"`
	ExternalRef           sql.NullString `db:"external_ref"`
	ILPStream             sql.NullString `db:"ilp_stream_id"`
	ILPAddress            sql.NullString `db:"ilp_address"`
	ILPSecret             sql.NullString `db:"ilp_shared_secret"`
	ExpiresAt             sql.NullTime   `db:"expires_at"`
	CreatedAt             time.Time      `db:"created_at"`
	UpdatedAt             time.Time      `db:"updated_at"`
	CreatedBy             sql.NullString `db:"created_by"`
	SenderWalletAddress   sql.NullString `db:"sender_wallet_address"`
	ReceiverWalletAddress sql.NullString `db:"receiver_wallet_address"`
}

func CreateIncomingPayment(ctx context.Context, b Backends, payment openpayments.CreateIncomingPaymentArgs) (*openpayments.IncomingPayment, error) {
	err := b.Validator().Struct(payment)
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInvalidArgument, err)
	}

	if !payment.ExpiresAt.IsZero() && payment.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("%w invalid expiry time", openpayments.ErrInvalidArgument)
	}

	_, err = b.Wallets().GetFromAddress(ctx, payment.PaymentPointer)
	if err != nil {
		return nil, err
	}

	_, err = b.Wallets().GetFromAddress(ctx, payment.FromPaymentPointer)
	if err != nil {
		return nil, err
	}

	if payment.IncomingAmount != nil && currency.USD != payment.IncomingAmount.Currency {
		return nil, fmt.Errorf("%w incompatible payment pointer assets", openpayments.ErrInvalidArgument)
	}

	senderWa, err := wallets.ParseAddress(payment.FromPaymentPointer)
	if err != nil {
		return nil, err
	}

	receiverWa, err := wallets.ParseAddress(payment.PaymentPointer)
	if err != nil {
		return nil, err
	}

	id := uuid.NewString()
	ib := db.NewInsert("openpayments_incoming_payment").
		Value("id", id).
		Value("received_amount", 0).
		Value("created_by", sql.NullString{
			String: payment.CreatedBy,
			Valid:  payment.CreatedBy != "",
		}).
		Value("sender_wallet_address", sql.NullString{
			String: senderWa.String(),
			Valid:  senderWa.String() != "",
		}).
		Value("receiver_wallet_address", sql.NullString{
			String: receiverWa.String(),
			Valid:  receiverWa.String() != "",
		})
	if payment.IncomingAmount != nil {
		ib.Value("asset_code", payment.IncomingAmount.Currency).
			Value("asset_scale", payment.IncomingAmount.Scale).
			Value("incoming_amount", payment.IncomingAmount.Value)
	} else {
		ib.Value("incoming_amount", 0)
	}
	if !payment.ExpiresAt.IsZero() {
		ib.Value("expires_at", payment.ExpiresAt)
	}
	if payment.ExternalRef != "" {
		ib.Value("external_ref", payment.ExternalRef)
	}
	if payment.Description != "" {
		ib.Value("description", payment.Description)
	}

	stmt, args, err := ib.GetStatement()
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	_, err = b.DB().ExecContext(ctx, stmt, args...)
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	return GetIncomingPayment(ctx, b, id)
}

func GetIncomingPayment(ctx context.Context, b Backends, id string) (*openpayments.IncomingPayment, error) {
	// Our friends may have provided the full ID with the payment pointer and the `incoming-payments` prefix.
	idxSlash := strings.LastIndex(id, "/")
	if idxSlash > 0 {
		id = id[idxSlash+1:]
	}

	var payment dbIncomingPayment
	err := b.DB().GetContext(ctx, &payment,
		fmt.Sprintf("SELECT %s FROM openpayments_incoming_payment WHERE id=$1", incomingPaymentsCols),
		id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", openpayments.ErrNotFound, err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	return transformIncomingPayment(payment), nil
}

func transformIncomingPayment(payment dbIncomingPayment) *openpayments.IncomingPayment {

	resp := &openpayments.IncomingPayment{
		ID:                 fmt.Sprintf("%s/incoming/%s", env.OpenPaymentsURL(), payment.ID),
		PaymentPointer:     payment.ReceiverWalletAddress.String,
		FromPaymentPointer: payment.SenderWalletAddress.String,
		Completed:          payment.Completed,
		ExternalRef:        payment.ExternalRef.String,
		ExpiresAt:          payment.ExpiresAt.Time,
		CreatedAt:          payment.CreatedAt,
		UpdatedAt:          payment.UpdatedAt,
		Description:        payment.Description.String,
		CreatedBy:          payment.CreatedBy.String,
	}
	if payment.IncomingAmount > 0 {
		resp.IncomingAmount = &currency.Amount{
			Value:    payment.IncomingAmount,
			Currency: currency.Currency(payment.AssetCode.String),
			Scale:    int(payment.AssetScale.Int32),
		}
	}
	if payment.ReceivedAmount > 0 {
		resp.ReceivedAmount = &currency.Amount{
			Value:    payment.ReceivedAmount,
			Currency: currency.Currency(payment.AssetCode.String),
			Scale:    int(payment.AssetScale.Int32),
		}
	}

	return resp
}
