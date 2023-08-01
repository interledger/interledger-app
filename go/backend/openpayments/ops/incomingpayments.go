package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/env"
)

const incomingPaymentsCols = `id, payment_pointer_id, from_payment_pointer_id, description, asset_code, asset_scale, incoming_amount, received_amount, completed, expires_at, external_ref, ilp_stream_id, ilp_address, ilp_shared_secret, created_at, updated_at, created_by, sender_wallet_address, receiver_wallet_address`

type dbIncomingPayment struct {
	ID                    string         `db:"id"`
	PaymentPointerID      string         `db:"payment_pointer_id"`
	FromPaymentPointerID  sql.NullString `db:"from_payment_pointer_id"`
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

	pp, err := GetPaymentPointer(ctx, b, payment.PaymentPointer)
	if err != nil {
		return nil, err
	}

	fromPP, err := GetPaymentPointer(ctx, b, payment.FromPaymentPointer)
	if err != nil {
		return nil, err
	}

	if payment.IncomingAmount != nil && pp.Asset != payment.IncomingAmount.Currency {
		return nil, fmt.Errorf("%w incompatible payment pointer assets", openpayments.ErrInvalidArgument)
	}

	id := uuid.NewString()
	ib := db.NewInsert("openpayments_incoming_payment").
		Value("id", id).
		Value("payment_pointer_id", pp.ID).
		Value("received_amount", 0).
		Value("from_payment_pointer_id", fromPP.ID).
		Value("created_by", sql.NullString{
			String: payment.CreatedBy,
			Valid:  payment.CreatedBy != "",
		}).
		Value("sender_wallet_address", sql.NullString{
			String: fromPP.URL,
			Valid:  fromPP.URL != "",
		}).
		Value("receiver_wallet_address", sql.NullString{
			String: pp.URL,
			Valid:  pp.URL != "",
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

	return transformIncomingPayment(ctx, b, payment)
}

func transformIncomingPayment(ctx context.Context, b Backends, payment dbIncomingPayment) (*openpayments.IncomingPayment, error) {
	fromPP, err := getPaymentPointerByID(ctx, b, payment.FromPaymentPointerID.String)
	if err != nil {
		return nil, err
	}

	toPP, err := getPaymentPointerByID(ctx, b, payment.PaymentPointerID)
	if err != nil {
		return nil, err
	}

	resp := &openpayments.IncomingPayment{
		ID:                 fmt.Sprintf("%s/incoming/%s", env.OpenPaymentsURL(), payment.ID),
		PaymentPointer:     toPP.URL,
		FromPaymentPointer: fromPP.URL,
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

	return resp, nil
}
