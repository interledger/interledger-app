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
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/openpayments"
)

func PaymentPointerExists(ctx context.Context, b Backends, pointerURLRaw string) (bool, error) {
	// Validate that this is a valid payment pointer
	ppURL, err := wallets.ParseAddress(pointerURLRaw)
	if err != nil {
		return false, err
	}

	wa, err := b.Wallets().GetFromAddress(ctx, ppURL.String())
	if err != nil && !errors.Is(err, wallets.ErrNoWalletFound) {
		return false, err
	}

	return wa != nil, nil
}

// ExtractPaymentPointer takes a full URL and removes the known suffix and what is left is the original Payment pointer
// returns the payment pointer as well as the matching suffix
func ExtractPaymentPointer(rawURL string) (string, string, error) {
	var res string
	for _, res = range wallets.ReservedURLParts {
		if strings.Contains(rawURL, res) {
			waRaw := rawURL[:strings.LastIndex(rawURL, res)]

			wa, err := wallets.ParseAddress(waRaw)
			if err != nil {
				return "", "", err
			}

			return wa.String(), res, nil
		}
	}

	// No suffix found, return the original sanitized
	wa, err := wallets.ParseAddress(rawURL)
	if err != nil {
		return "", "", err
	}
	return wa.String(), "", err
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

	recvAddress, err := wallets.ParseAddress(args.ReceivePaymentPointer)
	if err != nil {
		return nil, err
	}

	senderAddress, err := wallets.ParseAddress(args.SendPaymentPointer)
	if err != nil {
		return nil, err
	}

	// Tying to send money to yourself.
	if recvAddress.String() == senderAddress.String() {
		return nil, fmt.Errorf("%w cannot send money to the same payment pointer", openpayments.ErrInvalidArgument)
	}

	senderWallet, err := b.Wallets().GetFromAddress(ctx, senderAddress.String())
	if err != nil {
		return nil, err
	}
	if args.LinkedAccID != "" {
		la, err := b.LinkedAccounts().Get(ctx, args.LinkedAccID)
		if err != nil {
			return nil, err
		}

		if la.WalletID != senderWallet.ID {
			return nil, fmt.Errorf("%w specified linked account not associated with the send payment pointer", openpayments.ErrInvalidArgument)
		}
	}

	recvWallet, err := b.Wallets().GetFromAddress(ctx, recvAddress.String())
	if err != nil {
		return nil, err
	}

	err = CheckWalletsCanSendRecv(ctx, b, senderWallet.ID, args.LinkedAccID, recvWallet.ID)
	if err != nil {
		return nil, err
	}

	// Create Incoming Payment
	ip, err := CreateIncomingPayment(ctx, b, openpayments.CreateIncomingPaymentArgs{
		PaymentPointer:     recvAddress.String(),
		FromPaymentPointer: senderAddress.String(),
		ExternalRef:        args.Reference,
		Description:        args.Description,
		IncomingAmount:     &args.SendAmount,
		CreatedBy:          args.CreatedBy,
	})
	if err != nil {
		return nil, err
	}

	hasTx, err := b.Transactions().GetHasTransacted(ctx, senderWallet.ID, recvAddress.String())
	if err != nil {
		return nil, err
	}

	// TODO: Calculate the from/to conversion one day

	id := uuid.NewString()
	query, vals, err := db.NewInsert("openpayments_quotes").
		Value("id", id).
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
		}).
		Value("sender_wallet_address", sql.NullString{
			String: senderAddress.String(),
			Valid:  true,
		}).
		Value("receiver_wallet_address", sql.NullString{
			String: recvAddress.String(),
			Valid:  true,
		}).GetStatement()
	if err != nil {
		return nil, fmt.Errorf("%w insert sql create failed (%s)", openpayments.ErrInternal, err)
	}
	_, err = b.DB().ExecContext(ctx, query, vals...)
	if err != nil {
		return nil, fmt.Errorf("%w insert failed (%s)", openpayments.ErrInternal, err)
	}

	return GetWalletAddressQuote(ctx, b, senderAddress.String(), id)
}

type dbQuote struct {
	ID                    string         `db:"id"`
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
	SenderWalletAddress   sql.NullString `db:"sender_wallet_address"`
	ReceiverWalletAddress sql.NullString `db:"receiver_wallet_address"`
}

// getDBQuote returns a single quote in it's raw form from the DB without formatting.
// `where` is the where clause used in the SQL query with `args` used to fill the placeholders e.g. `id=$1`
func getDBQuote(ctx context.Context, b Backends, where string, args ...interface{}) (*dbQuote, error) {
	var dbq dbQuote
	err := b.DB().GetContext(ctx, &dbq,
		"SELECT id, send_linked_acc_id, incoming_payment_id, send_amount, send_asset, send_scale, recv_amount, recv_asset, recv_scale, expires_at, created_at, updated_at, created_by, recv_identity, recv_identity_type, otp_required, otp_validated, sender_wallet_address, receiver_wallet_address FROM openpayments_quotes WHERE "+where, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, openpayments.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	return &dbq, nil
}

func transformQuote(dbq dbQuote) *openpayments.Quote {

	amount := currency.Amount{
		Value:    dbq.SendAmount,
		Currency: currency.ParseCurrency(dbq.SendAsset),
		Scale:    dbq.SendAssetScale,
	}
	return &openpayments.Quote{
		ID:                      fmt.Sprintf("%s/quotes/%s", dbq.SenderWalletAddress.String, dbq.ID),
		PaymentPointer:          dbq.SenderWalletAddress.String,
		IncomingPayment:         fmt.Sprintf("%s/incoming-payments/%s", dbq.ReceiverWalletAddress.String, dbq.IncomingPaymentID),
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
	}
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

	return transformQuote(*dbq), nil
}

func GetWalletQuote(ctx context.Context, b Backends, walletID, id string) (*openpayments.Quote, error) {
	wallet, err := b.Wallets().Get(ctx, walletID)
	if err != nil {
		return nil, err
	}

	if len(wallet.Addresses) != 1 {
		return nil, fmt.Errorf("%w wallet has (%d) payment pointers", openpayments.ErrInternal, len(wallet.Addresses))
	}

	return GetWalletAddressQuote(ctx, b, wallet.AddressString(), id)
}

func SetQuoteOTPValidated(ctx context.Context, b Backends, qid string) (*openpayments.Quote, error) {
	_, err := b.DB().ExecContext(ctx, "UPDATE openpayments_quotes SET otp_validated  = TRUE WHERE id=$1", qid)
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	return GetQuote(ctx, b, qid)
}

func GetWalletAddressQuote(ctx context.Context, b Backends, senderAddress, id string) (*openpayments.Quote, error) {
	// Our friends may have provided the full ID with the payment pointer and the `quotes` prefix.
	idxSlash := strings.LastIndex(id, "/")
	if idxSlash > 0 {
		id = id[idxSlash+1:]
	}

	dbq, err := getDBQuote(ctx, b, "id=$1 AND sender_wallet_address=$2", id, senderAddress)
	if err != nil {
		return nil, err
	}

	return transformQuote(*dbq), nil
}

func ValidateCanSend(ctx context.Context, b Backends, walletID, walletAddress string) (bool, error) {
	addressWallet, err := b.Wallets().GetFromAddress(ctx, walletAddress)
	if errors.Is(err, wallets.ErrNoWalletFound) {
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

	wallet, err := b.Wallets().Get(ctx, walletID)
	if err != nil {
		return false, err
	}

	if wallet.ID == addressWallet.ID {
		return false, nil
	}
	for _, wa := range wallet.Addresses {
		for _, recvWA := range addressWallet.Addresses {
			if wa.String() == recvWA.String() {
				return false, nil
			}
		}

	}

	// check recv pp has a linked account that is verified and receive enabled.
	receiveLAs, err := b.LinkedAccounts().ListByWalletId(ctx, addressWallet.ID)
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
