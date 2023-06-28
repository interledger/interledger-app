package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/wallets"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/linkedaccounts"
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

	_, err = b.Wallets().AddAddress(ctx, pointer.WalletID, ppURL)
	if err != nil {
		return fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	b.Analytics().TrackWalletPaymentPointerCreated(pointer.WalletID)
	return nil
}

func PaymentPointerExists(ctx context.Context, b Backends, pointerURLRaw string) (bool, error) {
	// Validate that this is a valid payment pointer
	ppURL, err := validatePaymentPointer(pointerURLRaw)
	if err != nil {
		return false, err
	}

	pp, err := GetPaymentPointer(ctx, b, ppURL)
	if !errors.Is(err, openpayments.ErrPaymentPointerNotFound) {
		return false, err
	}
	if pp != nil {
		return true, nil
	}

	wa, err := b.Wallets().GetFromAddress(ctx, ppURL)
	if !errors.Is(err, wallets.ErrNoWalletFound) {
		return false, err
	}

	return pp != nil || wa != nil, err
}

func getPaymentPointerByID(ctx context.Context, b Backends, id string) (*openpayments.PaymentPointer, error) {
	var pp openpayments.PaymentPointer
	err := b.DB().GetContext(ctx, &pp, "SELECT id, wallet_id, url, alias, asset, scale FROM payment_pointers WHERE id = $1",
		id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w unkown payment pointer id(%s)", openpayments.ErrPaymentPointerNotFound, id)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	return &pp, nil
}

func GetPaymentPointer(ctx context.Context, b Backends, pointerURLRaw string) (*openpayments.PaymentPointer, error) {

	ppURL, err := sanitizePaymentPointer(pointerURLRaw)
	if err != nil {
		return nil, err
	}

	var pp openpayments.PaymentPointer
	err = b.DB().GetContext(ctx, &pp, "SELECT id, wallet_id, url, alias, asset, scale FROM payment_pointers WHERE lower(url) = lower($1)",
		ppURL)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w unkown payment pointer url(%s)", openpayments.ErrPaymentPointerNotFound, ppURL)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	return &pp, nil
}

var reservedURLParts = []string{"outgoing", "incoming", "quotes", "jwks.json", "identities"}

// sanitizePaymentPointer takes a full URL and checks for any reserved words and invalid formatting
func sanitizePaymentPointer(rawURL string) (string, error) {
	rawURL = StandardisePaymentPointer(rawURL)
	pointerURL, err := url.Parse(rawURL)
	if err != nil {
		return "", openpayments.ErrInvalidPointerURL
	}

	if pointerURL.Scheme == "" || pointerURL.Host == "" {
		return "", openpayments.ErrInvalidPointerURL
	}

	if len(pointerURL.Path) == 0 {
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
var pointerPrefixRegex = regexp.MustCompile(`^[A-Za-z]{3}$`)

// validatePaymentPointer returns the sanitized url or an error if the payment pointer is not in the format https://{base}/{variable}
// {variable} has the following conditions:
// - Between 4 and 42 characters
// - Only AlphaNumeric characters and underscore
// - The first 4 characters can only be alpha
// Assumption: {base} does not contain any slashes
func validatePaymentPointer(rawURL string) (string, error) {
	rawURL = StandardisePaymentPointer(rawURL)

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

	if len(path) < 3 {
		return "", fmt.Errorf("%w %s", openpayments.ErrInvalidPointerPath, "Your payment pointer must be longer than 3 characters")
	}
	if len(path) > 30 {
		return "", fmt.Errorf("%w %s", openpayments.ErrInvalidPointerPath, "Your payment pointer must be shorter than 30 characters")
	}

	if !pointerPrefixRegex.MatchString(path[:3]) {
		return "", fmt.Errorf("%w %s", openpayments.ErrInvalidPointerPath, "Your first 3 characters must be letters")
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

	formatted := path.Join(parsedUrl.Host, parsedUrl.Path)

	return formatted, nil
}

// StandardisePaymentPointer takes in a payment pointer in either the forms:
// - https://fynbos.me/alice
// - fynbos.me/alice
// - $fynbos.me/alice
// Returns the standard format of : https:///fynbos.me/alice
func StandardisePaymentPointer(pp string) string {
	if strings.HasPrefix(pp, "https://") {
		return pp
	}

	// Replace the $ with https://
	if strings.HasPrefix(pp, "$") {
		return strings.Replace(pp, "$", "https://", 1)
	}

	// We use https here
	if strings.HasPrefix(pp, "http://") {
		return strings.Replace(pp, "http://", "https://", 1)
	}

	// The payment pointer has no prefix assume we need to add https://
	return "https://" + pp
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
	err := b.DB().SelectContext(ctx, &pp, "SELECT id, wallet_id, url, alias, asset, scale FROM payment_pointers WHERE wallet_id=$1", walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w payment pointers fround for wallet(%s)", openpayments.ErrPaymentPointerNotFound, walletID)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	return pp, nil
}

func CheckWalletsCanSendRecv(ctx context.Context, b Backends, fromWalletID, fromLinkedAccID, toWalletID string) error {
	sendFeat, err := b.Features().Features(ctx, fromWalletID)
	if err != nil {
		return err
	}
	if !sendFeat.SendEnabled {
		return fmt.Errorf("%w walletID (%s)", openpayments.ErrPaymentPointerCannotSend, fromWalletID)
	}

	sendLa, err := b.LinkedAccounts().ListByWalletId(ctx, fromWalletID)
	if err != nil {
		return err
	}
	var canSend bool
	var foundLa linkedaccounts.LinkedAccount
	for _, sla := range sendLa {
		if sla.ID == fromLinkedAccID {
			foundLa = sla
		}
		if fromLinkedAccID != "" && sla.ID != fromLinkedAccID {
			continue
		}
		if sla.CanSend && sla.State == linkedaccounts.Verified {
			canSend = true
			break
		}

	}
	if !canSend {
		if foundLa.ID != "" {
			return fmt.Errorf("%w walletID (%s) accID (%s) acc state (%s) acc can send (%t)",
				openpayments.ErrPaymentPointerCannotSend, fromWalletID, foundLa.ID, foundLa.State, foundLa.CanSend)
		}
		return fmt.Errorf("%w walletID (%s)", openpayments.ErrPaymentPointerCannotSend, fromWalletID)
	}

	recvFeat, err := b.Features().Features(ctx, toWalletID)
	if err != nil {
		return err
	}
	if !recvFeat.ReceiveEnabled {
		return fmt.Errorf("%w walletID (%s)", openpayments.ErrPaymentPointerCannotRecv, toWalletID)
	}

	recvLa, err := b.LinkedAccounts().ListByWalletId(ctx, toWalletID)
	if err != nil {
		return err
	}
	var canRecv bool
	for _, rla := range recvLa {
		if rla.CanReceive && rla.State == linkedaccounts.Verified {
			canRecv = true
			break
		}
	}
	if !canRecv {
		return fmt.Errorf("%w walletID (%s)", openpayments.ErrPaymentPointerCannotRecv, toWalletID)
	}

	return nil
}

func CreateQuote(ctx context.Context, b Backends, args openpayments.CreateQuoteArgs) (*openpayments.Quote, error) {
	err := b.Validator().Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInvalidArgument, err)
	}

	if args.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("%w invalid expiry time", openpayments.ErrInvalidArgument)
	}

	recvPP, err := GetPaymentPointer(ctx, b, args.ReceivePaymentPointer)
	if err != nil {
		return nil, err
	}

	sendPP, err := GetPaymentPointer(ctx, b, args.SendPaymentPointer)
	if err != nil {
		return nil, err
	}

	if sendPP.Asset != recvPP.Asset || recvPP.Asset != args.SendAmount.Currency {
		return nil, fmt.Errorf("%w incompatible payment pointer assets", openpayments.ErrInvalidArgument)
	}

	// Tying to send money to yourself.
	if recvPP.ID == sendPP.ID {
		return nil, fmt.Errorf("%w cannot send money to the same payment pointer", openpayments.ErrInvalidArgument)
	}

	if args.LinkedAccID != "" {
		la, err := b.LinkedAccounts().Get(ctx, args.LinkedAccID)
		if err != nil {
			return nil, err
		}

		if la.WalletID != sendPP.WalletID {
			return nil, fmt.Errorf("%w specified linked account not associated with the send payment pointer", openpayments.ErrInvalidArgument)
		}
	}

	err = CheckWalletsCanSendRecv(ctx, b, sendPP.WalletID, args.LinkedAccID, recvPP.WalletID)
	if err != nil {
		return nil, err
	}

	// Create Incoming Payment
	ip, err := CreateIncomingPayment(ctx, b, openpayments.CreateIncomingPaymentArgs{
		PaymentPointer:     recvPP.URL,
		FromPaymentPointer: sendPP.URL,
		ExternalRef:        args.Reference,
		Description:        args.Description,
		IncomingAmount:     &args.SendAmount,
		CreatedBy:          args.CreatedBy,
	})
	if err != nil {
		return nil, err
	}

	hasTx, err := b.Transactions().GetHasTransacted(ctx, sendPP.WalletID, recvPP.URL)
	if err != nil {
		return nil, err
	}

	// TODO: Calculate the from/to conversion one day

	id := uuid.NewString()
	query, vals, err := db.NewInsert("openpayments_quotes").
		Value("id", id).
		Value("send_payment_pointer_id", sendPP.ID).
		Value("recv_payment_pointer_id", recvPP.ID).
		Value("incoming_payment_id", ip.ID[strings.LastIndex(ip.ID, "/")+1:]).
		Value("send_amount", args.SendAmount.Value).
		Value("send_asset", args.SendAmount.Currency).
		Value("send_scale", args.SendAmount.Scale).
		Value("recv_amount", args.SendAmount.Value).
		Value("recv_asset", args.SendAmount.Currency).
		Value("recv_scale", args.SendAmount.Scale).
		Value("expires_at", args.ExpiresAt).
		Value("send_linked_acc_id", sql.NullString{
			String: args.LinkedAccID,
			Valid:  args.LinkedAccID != "",
		}).
		Value("created_by", sql.NullString{
			String: args.CreatedBy,
			Valid:  args.CreatedBy != "",
		}).
		Value("recv_identity_type", sql.NullString{
			String: args.DestinationIdentityType,
			Valid:  args.DestinationIdentityType != "",
		}).
		Value("recv_identity", sql.NullString{
			String: args.DestinationIdentity,
			Valid:  args.DestinationIdentity != "",
		}).
		Value("otp_required", sql.NullBool{
			Bool:  !hasTx,
			Valid: true,
		}).GetStatement()
	if err != nil {
		return nil, fmt.Errorf("%w insert sql create failed (%s)", openpayments.ErrInternal, err)
	}
	_, err = b.DB().ExecContext(ctx, query, vals...)
	if err != nil {
		return nil, fmt.Errorf("%w insert failed (%s)", openpayments.ErrInternal, err)
	}

	return GetPaymentPointerQuote(ctx, b, sendPP.ID, id)
}

type dbQuote struct {
	ID                    string         `db:"id"`
	SendPaymentPointer    string         `db:"send_payment_pointer_id"`
	ReceivePaymentPointer string         `db:"recv_payment_pointer_id"`
	IncomingPaymentID     string         `db:"incoming_payment_id"`
	SendAmount            uint64         `db:"send_amount"`
	SendAsset             string         `db:"send_asset"`
	SendAssetScale        int            `db:"send_scale"`
	RecvAmount            uint64         `db:"recv_amount"`
	RecvAsset             string         `db:"recv_asset"`
	RecvAssetScale        int            `db:"recv_scale"`
	ExpiresAt             time.Time      `db:"expires_at"`
	CreatedAt             time.Time      `db:"created_at"`
	UpdatedAt             time.Time      `db:"updated_at"`
	SendLinkedAccountID   sql.NullString `db:"send_linked_acc_id"`
	CreatedBy             sql.NullString `db:"created_by"`
	RecvIdentity          sql.NullString `db:"recv_identity"`
	RecvIdentityType      sql.NullString `db:"recv_identity_type"`
	OTPRequired           sql.NullBool   `db:"otp_required"`
	OTPValidated          sql.NullBool   `db:"otp_validated"`
}

// getDBQuote returns a single quote in it's raw form from the DB without formatting.
// `where` is the where clause used in the SQL query with `args` used to fill the placeholders e.g. `id=$1`
func getDBQuote(ctx context.Context, b Backends, where string, args ...interface{}) (*dbQuote, error) {
	var dbq dbQuote
	err := b.DB().GetContext(ctx, &dbq,
		"SELECT id, send_linked_acc_id, send_payment_pointer_id, recv_payment_pointer_id, incoming_payment_id, send_amount, send_asset, send_scale, recv_amount, recv_asset, recv_scale, expires_at, created_at, updated_at, created_by, recv_identity, recv_identity_type, otp_required, otp_validated FROM openpayments_quotes WHERE "+where, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, openpayments.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	return &dbq, nil
}

func transformQuote(ctx context.Context, b Backends, dbq dbQuote) (*openpayments.Quote, error) {
	recvPP, err := getPaymentPointerByID(ctx, b, dbq.ReceivePaymentPointer)
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	sendPP, err := getPaymentPointerByID(ctx, b, dbq.SendPaymentPointer)
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	amount := currency.Amount{
		Value:    dbq.SendAmount,
		Currency: currency.ParseCurrency(dbq.SendAsset),
		Scale:    dbq.SendAssetScale,
	}
	return &openpayments.Quote{
		ID:                      fmt.Sprintf("%s/quotes/%s", sendPP.URL, dbq.ID),
		PaymentPointer:          sendPP.URL,
		IncomingPayment:         fmt.Sprintf("%s/incoming-payments/%s", recvPP.URL, dbq.IncomingPaymentID),
		ReceiveAmount:           amount,
		SendAmount:              amount,
		ExpiresAt:               dbq.ExpiresAt,
		CreatedAt:               dbq.CreatedAt,
		FromLinkedAccount:       dbq.SendLinkedAccountID.String,
		CreatedBy:               dbq.CreatedBy.String,
		DestinationIdentity:     dbq.RecvIdentity.String,
		DestinationIdentityType: dbq.RecvIdentityType.String,
		RequiresOTP:             dbq.OTPRequired.Bool,
		OTPValidated:            dbq.OTPValidated.Bool,
	}, nil
}

// GetQuote returns a quote for the given ID. No validation is done on if the caller/user/paymentpointer can access the quote.
func GetQuote(ctx context.Context, b Backends, id string) (*openpayments.Quote, error) {
	// Our friends may have provided the full ID with the payment pointer and the `quotes` prefix.
	idxSlash := strings.LastIndex(id, "/")
	if idxSlash > 0 {
		id = id[idxSlash+1:]
	}

	dbq, err := getDBQuote(ctx, b, "id=$1", id)
	if err != nil {
		return nil, err
	}

	return transformQuote(ctx, b, *dbq)
}

func GetWalletQuote(ctx context.Context, b Backends, walletID, id string) (*openpayments.Quote, error) {
	pp, err := ListWalletPaymentPointers(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	if len(pp) != 1 {
		return nil, fmt.Errorf("%w wallet has (%d) payment pointers", openpayments.ErrInternal, len(pp))
	}

	return GetPaymentPointerQuote(ctx, b, pp[0].ID, id)
}

func SendQuoteOTP(ctx context.Context, b Backends, qid string) error {
	idxSlash := strings.LastIndex(qid, "/")
	if idxSlash > 0 {
		qid = qid[idxSlash+1:]
	}

	q, err := GetQuote(ctx, b, qid)
	if err != nil {
		return err
	}

	if !q.RequiresOTP {
		return nil
	}

	sendPP, err := getPaymentPointerByID(ctx, b, q.PaymentPointer)
	if err != nil {
		return err
	}

	ul, err := b.Users().ListUsers(ctx, sendPP.WalletID)
	if err != nil || len(ul) == 0 {
		return fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	_, err = b.Twilio().SendVerificationCode(ctx, ul[0].PhoneNumber)
	if err != nil {
		return fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	return nil
}

func SetQuoteOTPValidated(ctx context.Context, b Backends, qid string) (*openpayments.Quote, error) {
	_, err := b.DB().ExecContext(ctx, "UPDATE openpayments_quotes SET otp_validated  = TRUE WHERE id=$1", qid)
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	return GetQuote(ctx, b, qid)
}

func GetPaymentPointerQuote(ctx context.Context, b Backends, paymentPointerID, id string) (*openpayments.Quote, error) {
	// Our friends may have provided the full ID with the payment pointer and the `quotes` prefix.
	idxSlash := strings.LastIndex(id, "/")
	if idxSlash > 0 {
		id = id[idxSlash+1:]
	}

	dbq, err := getDBQuote(ctx, b, "id=$1 AND send_payment_pointer_id=$2", id, paymentPointerID)
	if err != nil {
		return nil, err
	}

	return transformQuote(ctx, b, *dbq)
}

func ValidateCanSend(ctx context.Context, b Backends, walletID, ppString string) (bool, error) {
	pp, err := GetPaymentPointer(ctx, b, StandardisePaymentPointer(ppString))
	if errors.Is(err, openpayments.ErrPaymentPointerNotFound) {
		// Payment pointer doesn't exists, don't error just return false
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if walletID == "" {
		// Target Payment Pointer exists, Machnet wallet exists, unauthenticated request, so don't check that it's not sending to itself
		return true, nil
	}

	ppl, err := ListWalletPaymentPointers(ctx, b, walletID)
	if err != nil {
		return false, err
	}

	for _, own := range ppl {
		if own.ID == pp.ID {
			return false, nil
		}
	}

	// check recv pp has a linked account that is verified and receive enabled.
	receiveLAs, err := b.LinkedAccounts().ListByWalletId(ctx, pp.WalletID)
	if err != nil {
		return false, err
	}
	var canReceive bool
	for _, la := range receiveLAs {
		if la.CanReceive && la.State == linkedaccounts.Verified {
			canReceive = true
			break
		}
	}

	// check sending wallet has a linked account that is verified and send enabled.
	sendLAs, err := b.LinkedAccounts().ListByWalletId(ctx, walletID)
	if err != nil {
		return false, err
	}
	var canSend bool
	for _, la := range sendLAs {
		if la.CanSend && la.State == linkedaccounts.Verified {
			canSend = true
			break
		}
	}

	// Target Payment Pointer exists and can receive, sending wallet can send, it's not sending to itself, authenticated request
	return canReceive && canSend, nil
}

func GetWalletPaymentPointer(ctx context.Context, b Backends, walletID string) (*openpayments.PaymentPointer, error) {
	var pp openpayments.PaymentPointer
	err := b.DB().GetContext(ctx, &pp, "SELECT id, wallet_id, url, alias, asset, scale FROM payment_pointers WHERE wallet_id = $1 LIMIT 1;", walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w unkown payment pointer for walletID(%s)", openpayments.ErrPaymentPointerNotFound, walletID)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	return &pp, nil
}
