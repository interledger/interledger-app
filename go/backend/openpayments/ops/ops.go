package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/openpayments"
)

func CreatePaymentPointer(ctx context.Context, b Backends, pointer openpayments.PaymentPointer) error {
	err := b.Validator().Struct(pointer)
	if err != nil {
		return fmt.Errorf("%w %s", openpayments.ErrInvalidArgument, err)
	}

	_, err = b.DB().ExecContext(ctx, "INSERT INTO payment_pointers (wallet_id, url, alias, asset, scale) VALUES ($1,$2,$3,$4,$5)",
		pointer.WalletID, pointer.URL, pointer.Alias, pointer.Asset, pointer.AssetScale)

	if db.IsErrorCode(err, db.UniqueViolationError) {
		return fmt.Errorf("%w payment pointer url exists already (%s)", openpayments.ErrPaymentPointerExists, pointer.URL)
	}
	if err != nil {
		return fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	return nil
}

func GetPaymentPointer(ctx context.Context, b Backends, pointerURLRaw string) (*openpayments.PaymentPointer, error) {

	pointerURL, err := url.ParseRequestURI(pointerURLRaw)
	if err != nil {
		return nil, openpayments.ErrInvalidPointerURL
	}

	var pp openpayments.PaymentPointer
	err = b.DB().GetContext(ctx, &pp, "SELECT wallet_id, url, alias, asset, scale FROM payment_pointers WHERE url=$1", pointerURL.String())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w unkown payment pointer url(%s)", openpayments.ErrPaymentPointerNotFound, pointerURL.String())
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	return &pp, nil
}
