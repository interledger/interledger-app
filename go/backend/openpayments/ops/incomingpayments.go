package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/openpayments"
)

type dbIncomingPayment struct {
	ID               string         `db:"id"`
	PaymentPointerID string         `db:"payment_pointer_id"`
	AssetCode        sql.NullString `db:"asset_code"`
	AssetScale       sql.NullInt32  `db:"asset_scale"`
	IncomingAmount   uint64         `db:"incoming_amount"`
	ReceivedAmount   uint64         `db:"received_amount"`
	Completed        bool           `db:"completed"`
	ExternalRef      sql.NullString `db:"external_ref"`
	ILPStream        sql.NullString `db:"ilp_stream_id"`
	ILPAddress       sql.NullString `db:"ilp_address"`
	ILPSecret        sql.NullString `db:"ilp_shared_secret"`
	ExpiresAt        sql.NullTime   `db:"expires_at"`
	CreatedAt        time.Time      `db:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at"`
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

	return transformIncomingPayment(ctx, b, dbIncomingPayment{})
}

func transformIncomingPayment(ctx context.Context, b Backends, payment dbIncomingPayment) (*openpayments.IncomingPayment, error) {
	var pp string
	err := b.DB().GetContext(ctx, &pp, "SELECT url FROM payment_pointers WHERE id=$1", payment.PaymentPointerID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	resp := &openpayments.IncomingPayment{
		ID:             fmt.Sprintf("%s/incoming-payments/%s", pp, payment.ID),
		PaymentPointer: pp,
		Completed:      payment.Completed,
		ExternalRef:    payment.ExternalRef.String,
		ExpiresAt:      payment.ExpiresAt.Time,
		CreatedAt:      payment.CreatedAt,
		UpdatedAt:      payment.UpdatedAt,
	}
	if payment.IncomingAmount > 0 {
		resp.IncomingAmount = &openpayments.Amount{
			Value:      payment.IncomingAmount,
			Asset:      payment.AssetCode.String,
			AssetScale: int(payment.AssetScale.Int32),
		}
	}
	if payment.ReceivedAmount > 0 {
		resp.ReceivedAmount = &openpayments.Amount{
			Value:      payment.ReceivedAmount,
			Asset:      payment.AssetCode.String,
			AssetScale: int(payment.AssetScale.Int32),
		}
	}

	return resp, nil
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

	if payment.IncomingAmount != nil && pp.Asset != payment.IncomingAmount.Asset {
		return nil, fmt.Errorf("%w incompatible payment pointer assets", openpayments.ErrInvalidArgument)
	}

	id := uuid.NewString()
	ib := db.NewInsert("openpayments_incoming_payment").
		Value("id", id).
		Value("payment_pointer_id", pp.ID).
		Value("received_amount", 0)
	if payment.IncomingAmount != nil {
		ib.Value("asset_code", payment.IncomingAmount.Asset).
			Value("asset_scale", payment.IncomingAmount.AssetScale).
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

const incomingPaymentsCols = `id, payment_pointer_id, asset_code, asset_scale, incoming_amount, received_amount, completed, expires_at, external_ref, ilp_stream_id, ilp_address, ilp_shared_secret, created_at, updated_at`

func ListIncomingPayments(ctx context.Context, b Backends, walletID string) ([]openpayments.IncomingPayment, error) {
	pp, err := ListWalletPaymentPointers(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	if len(pp) != 1 {
		return nil, fmt.Errorf("%w wallet has (%s) payment pointer", openpayments.ErrInternal, len(pp))
	}

	var dbList []dbIncomingPayment
	// TODO pagination
	err = b.DB().SelectContext(ctx, &dbList,
		fmt.Sprintf("SELECT %s FROM openpayments_incoming_payment WHERE completed=true AND payment_pointer_id=$1", incomingPaymentsCols),
		pp[0].ID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	resp := make([]openpayments.IncomingPayment, len(dbList))
	for i, ip := range dbList {
		transform, err := transformIncomingPayment(ctx, b, ip)
		if err != nil {
			return nil, err
		}
		resp[i] = *transform
	}

	return resp, nil
}
