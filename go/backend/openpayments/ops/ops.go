package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/openpayments"
)

func CreatePaymentPointer(ctx context.Context, b Backends, pointer openpayments.PaymentPointer) error {
	err := b.Validator().Struct(pointer)
	if err != nil {
		return fmt.Errorf("%w %s", openpayments.ErrInvalidArgument, err)
	}

	ppURL, err := sanitizePaymentPointer(pointer.URL)
	if err != nil {
		return err
	}

	_, err = b.DB().ExecContext(ctx, "INSERT INTO payment_pointers (wallet_id, url, alias, asset, scale) VALUES ($1,$2,$3,$4,$5)",
		pointer.WalletID, ppURL, pointer.Alias, pointer.Asset, pointer.AssetScale)

	if db.IsErrorCode(err, db.UniqueViolationError) {
		return fmt.Errorf("%w payment pointer url exists already (%s)", openpayments.ErrPaymentPointerExists, pointer.URL)
	}
	if err != nil {
		return fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	return nil
}

func GetPaymentPointer(ctx context.Context, b Backends, pointerURLRaw string) (*openpayments.PaymentPointer, error) {

	ppURL, err := sanitizePaymentPointer(pointerURLRaw)
	if err != nil {
		return nil, err
	}

	var pp openpayments.PaymentPointer
	err = b.DB().GetContext(ctx, &pp, "SELECT wallet_id, url, alias, asset, scale FROM payment_pointers WHERE url=$1", ppURL)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w unkown payment pointer url(%s)", openpayments.ErrPaymentPointerNotFound, ppURL)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	return &pp, nil
}

var reservedURLParts = []string{"outgoing-payments", "incoming-payments", "quotes"}

// sanitizePaymentPointer takes a full URL and checks for any reserved words and invalid formatting
func sanitizePaymentPointer(rawURL string) (string, error) {
	pointerURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return "", openpayments.ErrInvalidPointerURL
	}

	pathParts := strings.Split(pointerURL.Path, "/")
	for _, pp := range pathParts {
		for _, res := range reservedURLParts {
			if strings.EqualFold(pp, res) {
				return "", openpayments.ErrInvalidPointerURL
			}
		}
	}

	return strings.TrimSuffix(pointerURL.String(), "/"), nil
}

// ExtractPaymentPointer takes a full URL and removes the known suffix and what is left is the original Payment pointer
// returns the payment pointer as well as the matching suffix
func ExtractPaymentPointer(rawURL string) (string, string, error) {
	var res string
	for _, res = range reservedURLParts {
		if strings.Contains(rawURL, res) {
			sanitized, err := sanitizePaymentPointer(rawURL[:strings.LastIndex(rawURL, res)])
			if err != nil {
				return "", "", err
			}
			return sanitized, res, nil
		}
	}

	// No suffix found, return the original sanitized
	sanitized, err := sanitizePaymentPointer(rawURL)
	return sanitized, "", err
}
