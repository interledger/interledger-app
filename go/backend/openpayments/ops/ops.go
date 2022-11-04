package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/jmoiron/sqlx"

	"github.com/google/uuid"

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

	if sendPP.Asset != recvPP.Asset || recvPP.Asset != args.SendAmount.Asset {
		return nil, fmt.Errorf("%w incompatible payment pointer assets", openpayments.ErrInvalidArgument)
	}

	// Create Incoming Payment
	ip, err := CreateIncomingPayment(ctx, b, openpayments.CreateIncomingPaymentArgs{
		PaymentPointer:     recvPP.URL,
		FromPaymentPointer: sendPP.URL,
		ExternalRef:        args.Reference,
		Description:        args.Description,
	})
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
		Value("send_asset", args.SendAmount.Asset).
		Value("send_scale", args.SendAmount.AssetScale).
		Value("recv_amount", args.SendAmount.Value).
		Value("recv_asset", args.SendAmount.Asset).
		Value("recv_scale", args.SendAmount.AssetScale).
		Value("expires_at", args.ExpiresAt).GetStatement()
	if err != nil {
		return nil, fmt.Errorf("%w insert sql create failed (%s)", openpayments.ErrInternal, err)
	}
	_, err = b.DB().ExecContext(ctx, query, vals...)
	if err != nil {
		return nil, fmt.Errorf("%w insert failed (%s)", openpayments.ErrInternal, err)
	}

	return GetQuote(ctx, b, id)
}

type dbQuote struct {
	ID                    string    `db:"id"`
	SendPaymentPointer    string    `db:"send_payment_pointer_id"`
	ReceivePaymentPointer string    `db:"recv_payment_pointer_id"`
	IncomingPaymentID     string    `db:"incoming_payment_id"`
	SendAmount            uint64    `db:"send_amount"`
	SendAsset             string    `db:"send_asset"`
	SendAssetScale        int       `db:"send_scale"`
	RecvAmount            uint64    `db:"recv_amount"`
	RecvAsset             string    `db:"recv_asset"`
	RecvAssetScale        int       `db:"recv_scale"`
	ExpiresAt             time.Time `db:"expires_at"`
	CreatedAt             time.Time `db:"created_at"`
	UpdatedAt             time.Time `db:"updated_at"`
}

// getDBQuote returns a single quote in it's raw form from the DB without formatting.
// `where` is the where clause used in the SQL query with `args` used to fill the placeholders e.g. `id=$1`
func getDBQuote(ctx context.Context, b Backends, where string, args ...interface{}) (*dbQuote, error) {
	var dbq dbQuote
	err := b.DB().GetContext(ctx, &dbq,
		"SELECT id, send_payment_pointer_id, recv_payment_pointer_id, incoming_payment_id, send_amount, send_asset, send_scale, recv_amount, recv_asset, recv_scale, expires_at, created_at, updated_at FROM openpayments_quotes WHERE "+where, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, openpayments.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	return &dbq, nil
}

func GetQuote(ctx context.Context, b Backends, id string) (*openpayments.Quote, error) {
	// Our friends may have provided the full ID with the payment pointer and the `incoming-payments` prefix.
	idxSlash := strings.LastIndex(id, "/")
	if idxSlash > 0 {
		id = id[idxSlash+1:]
	}

	dbq, err := getDBQuote(ctx, b, "id=$1", id)
	if err != nil {
		return nil, err
	}

	recvPP, err := getPaymentPointerByID(ctx, b, dbq.ReceivePaymentPointer)
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	sendPP, err := getPaymentPointerByID(ctx, b, dbq.SendPaymentPointer)
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	amount := openpayments.Amount{
		Value:      dbq.SendAmount,
		Asset:      dbq.SendAsset,
		AssetScale: dbq.SendAssetScale,
	}
	return &openpayments.Quote{
		ID:              fmt.Sprintf("%s/quotes/%s", sendPP.URL, dbq.ID),
		PaymentPointer:  sendPP.URL,
		IncomingPayment: fmt.Sprintf("%s/incoming-payments/%s", recvPP.URL, dbq.IncomingPaymentID),
		ReceiveAmount:   amount,
		SendAmount:      amount,
		ExpiresAt:       dbq.ExpiresAt,
		CreatedAt:       dbq.CreatedAt,
	}, nil

}

type dbIncomingPayment struct {
	ID                   string         `db:"id"`
	PaymentPointerID     string         `db:"payment_pointer_id"`
	FromPaymentPointerID sql.NullString `db:"from_payment_pointer_id"`
	Description          sql.NullString `db:"description"`
	AssetCode            sql.NullString `db:"asset_code"`
	AssetScale           sql.NullInt32  `db:"asset_scale"`
	IncomingAmount       uint64         `db:"incoming_amount"`
	ReceivedAmount       uint64         `db:"received_amount"`
	Completed            bool           `db:"completed"`
	ExternalRef          sql.NullString `db:"external_ref"`
	ILPStream            sql.NullString `db:"ilp_stream_id"`
	ILPAddress           sql.NullString `db:"ilp_address"`
	ILPSecret            sql.NullString `db:"ilp_shared_secret"`
	ExpiresAt            sql.NullTime   `db:"expires_at"`
	CreatedAt            time.Time      `db:"created_at"`
	UpdatedAt            time.Time      `db:"updated_at"`
}

func GetIncomingPayment(ctx context.Context, b Backends, id string) (*openpayments.IncomingPayment, error) {
	// Our friends may have provided the full ID with the payment pointer and the `incoming-payments` prefix.
	idxSlash := strings.LastIndex(id, "/")
	if idxSlash > 0 {
		id = id[idxSlash+1:]
	}

	var payment dbIncomingPayment
	err := b.DB().GetContext(ctx, &payment,
		"SELECT id, payment_pointer_id, from_payment_pointer_id, description, asset_code, asset_scale, incoming_amount, received_amount, completed, expires_at, external_ref, ilp_stream_id, ilp_address, ilp_shared_secret, created_at, updated_at FROM openpayments_incoming_payment WHERE id=$1",
		id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", openpayments.ErrNotFound, err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	fromPP, err := getPaymentPointerByID(ctx, b, payment.FromPaymentPointerID.String)
	if err != nil {
		return nil, err
	}

	toPP, err := getPaymentPointerByID(ctx, b, payment.PaymentPointerID)
	if err != nil {
		return nil, err
	}

	op := &openpayments.IncomingPayment{
		ID:                 fmt.Sprintf("%s/incoming-payments/%s", toPP.URL, payment.ID),
		PaymentPointer:     toPP.URL,
		FromPaymentPointer: fromPP.URL,
		Completed:          payment.Completed,
		ExternalRef:        payment.ExternalRef.String,
		ExpiresAt:          payment.ExpiresAt.Time,
		CreatedAt:          payment.CreatedAt,
		UpdatedAt:          payment.UpdatedAt,
		Description:        payment.Description.String,
	}
	if payment.IncomingAmount > 0 {
		op.IncomingAmount = &openpayments.Amount{
			Value:      payment.IncomingAmount,
			Asset:      payment.AssetCode.String,
			AssetScale: int(payment.AssetScale.Int32),
		}
	}
	if payment.ReceivedAmount > 0 {
		op.ReceivedAmount = &openpayments.Amount{
			Value:      payment.ReceivedAmount,
			Asset:      payment.AssetCode.String,
			AssetScale: int(payment.AssetScale.Int32),
		}
	}

	return op, nil
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

	if payment.IncomingAmount != nil && pp.Asset != payment.IncomingAmount.Asset {
		return nil, fmt.Errorf("%w incompatible payment pointer assets", openpayments.ErrInvalidArgument)
	}

	id := uuid.NewString()
	ib := db.NewInsert("openpayments_incoming_payment").
		Value("id", id).
		Value("payment_pointer_id", pp.ID).
		Value("received_amount", 0).
		Value("from_payment_pointer_id", fromPP.ID)
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

		return nil
	})
	if err != nil {
		return "", err
	}

	toPP, err := getPaymentPointerByID(ctx, b, q.SendPaymentPointer)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/outgoing-payments/%s", toPP.URL, id), nil
}

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

func GetOutgoingPayment(ctx context.Context, b Backends, id string) (*openpayments.OutgoingPayment, error) {
	// Our friends may have provided the full ID with the payment pointer and the `incoming-payments` prefix.
	idxSlash := strings.LastIndex(id, "/")
	if idxSlash > 0 {
		id = id[idxSlash+1:]
	}

	var op dbOutgoingPayments
	err := b.DB().GetContext(ctx, &op,
		"SELECT id, to_payment_pointer_id, quote_id, failed, description, sent_amount, sent_asset, sent_scale, created_at, updated_at FROM openpayments_outgoing_payment WHERE id=$1",
		id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", openpayments.ErrNotFound, err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	q, err := GetQuote(ctx, b, op.QuoteID)
	if err != nil {
		return nil, err
	}

	dbq, err := getDBQuote(ctx, b, "id=$1", op.QuoteID)
	if err != nil {
		return nil, err
	}

	toPP, err := getPaymentPointerByID(ctx, b, dbq.ReceivePaymentPointer)
	if err != nil {
		return nil, err
	}

	return &openpayments.OutgoingPayment{
		ID:               fmt.Sprintf("%s/outgoing-payments/%s", q.PaymentPointer, op.ID),
		PaymentPointer:   q.PaymentPointer,
		ToPaymentPointer: toPP.URL,
		Failed:           op.Failed,
		Receiver:         q.IncomingPayment,
		SendAmount:       q.SendAmount,
		ReceiveAmount:    q.ReceiveAmount,
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

	res, err := b.DB().ExecContext(ctx, "UPDATE openpayments_outgoing_payment SET failed=true, updated_at=now() WHERE id=$1", id)
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

	return nil
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

	return crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		res, err := tx.ExecContext(ctx, "UPDATE openpayments_outgoing_payment SET updated_at=now(), sent_amount=$1, sent_asset=$2, sent_scale=$3 WHERE id=$4 AND failed=false",
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

		res, err = tx.ExecContext(ctx, "UPDATE openpayments_incoming_payment SET received_amount=$1 WHERE id=$2 AND received_amount<=$1", args.SentAmount.Value, ipID)
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

		return nil
	})
}
