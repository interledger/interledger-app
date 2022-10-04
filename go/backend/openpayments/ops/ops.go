package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/openpayments"
)

func CreatePaymentPointer(ctx context.Context, b Backends, pointer openpayments.PaymentPointer) error {
	err := b.Validator().Struct(pointer)
	if err != nil {
		return fmt.Errorf("%w %s", openpayments.ErrInvalidArgument, err)
	}

	ppURL, err := validatePaymentPointer(pointer.URL)
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

func PaymentPointerExists(ctx context.Context, b Backends, pointerURLRaw string) (bool, error) {
	// Validate that this is a valid payment pointer
	ppURL, err := validatePaymentPointer(pointerURLRaw)
	if err != nil {
		return false, err
	}

	_, err = GetPaymentPointer(ctx, b, ppURL)
	if errors.Is(err, openpayments.ErrPaymentPointerNotFound) {
		return false, nil
	}
	return true, err
}

func GetPaymentPointer(ctx context.Context, b Backends, pointerURLRaw string) (*openpayments.PaymentPointer, error) {

	ppURL, err := sanitizePaymentPointer(pointerURLRaw)
	if err != nil {
		return nil, err
	}

	var pp openpayments.PaymentPointer
	err = b.DB().GetContext(ctx, &pp, "SELECT wallet_id, url, alias, asset, scale FROM payment_pointers WHERE lower(url) = lower($1)",
		ppURL)
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
	pointerURL, err := url.Parse(rawURL)
	if err != nil {
		return "", openpayments.ErrInvalidPointerURL
	}

	if pointerURL.Scheme == "" || pointerURL.Host == "" {
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

var pointerRegex = regexp.MustCompile(`^[A-Za-z]{4}[a-zA-z0\d_]{0,26}$`)
var pointerPrefixRegex = regexp.MustCompile(`^[A-Za-z]{4}$`)

// validatePaymentPointer returns the sanitized url or an error if the payment pointer is not in the format https://{base}/{variable}
// {variable} has the following conditions:
//- Between 4 and 42 characters
//- Only AlphaNumeric characters and underscore
//- The first 4 characters can only be alpha
// Assumption: {base} does not contain any slashes
func validatePaymentPointer(rawURL string) (string, error) {

	unescaped, err := url.PathUnescape(rawURL)
	if err != nil || unescaped != rawURL {
		// Some URL escapes where added or invalid URL escapes are present
		return "", fmt.Errorf("%w %s", openpayments.ErrInvalidPointerPath, "Your payment pointer can only contain letters, numbers and '_'")
	}

	pointerURL, err := url.Parse(rawURL)
	if err != nil {
		return "", openpayments.ErrInvalidPointerURL
	}

	if pointerURL.Scheme == "" || pointerURL.Host == "" {
		return "", fmt.Errorf("%w %s", openpayments.ErrInvalidPointerPath, "Your payment pointer needs to contain a host and a http scheme")
	}

	// Fragments are after a '#' character in the url.
	// Payment pointers do not contain queries.
	if pointerURL.Fragment != "" || pointerURL.RawQuery != "" {
		return "", fmt.Errorf("%w %s", openpayments.ErrInvalidPointerPath, "Your payment pointer can only contain letters, numbers and '_'")
	}

	path := strings.TrimPrefix(pointerURL.Path, "/")

	if len(path) < 4 {
		return "", fmt.Errorf("%w %s", openpayments.ErrInvalidPointerPath, "Your payment pointer must be longer than 4 characters")
	}
	if len(path) > 30 {
		return "", fmt.Errorf("%w %s", openpayments.ErrInvalidPointerPath, "Your payment pointer must be shorter than 30 characters")
	}

	if !pointerPrefixRegex.MatchString(path[:4]) {
		return "", fmt.Errorf("%w %s", openpayments.ErrInvalidPointerPath, "Your first 4 characters must be letters")
	}

	if !pointerRegex.MatchString(path) {
		return "", fmt.Errorf("%w %s", openpayments.ErrInvalidPointerPath, "Your payment pointer can only contain letters, numbers and '_'")
	}

	ppURL, err := sanitizePaymentPointer(rawURL)
	if err != nil {
		return "", err
	}

	// Some characters where removed from the URL, error
	if !strings.EqualFold(ppURL, rawURL) {
		return "", fmt.Errorf("%w %s", openpayments.ErrInvalidPointerPath, "Your payment pointer can only contain letters, numbers and '_'")
	}

	return ppURL, nil
}

func FormattedPaymentPointer(rawURL string) (string, error) {
	parsedUrl, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	formatted := fmt.Sprintf("$%s%s", parsedUrl.Host, parsedUrl.Path)

	return formatted, nil
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

func ListWalletPaymentPointers(ctx context.Context, b Backends, walletID string) ([]openpayments.PaymentPointer, error) {
	var pp []openpayments.PaymentPointer
	err := b.DB().SelectContext(ctx, &pp, "SELECT wallet_id, url, alias, asset, scale FROM payment_pointers WHERE wallet_id=$1", walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w payment pointers fround for wallet(%s)", openpayments.ErrPaymentPointerNotFound, walletID)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	return pp, nil
}
