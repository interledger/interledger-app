package ops

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/slack"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/wallets"
	"go.temporal.io/api/enums/v1"
	temporal_client "go.temporal.io/sdk/client"
)

const cols = `id, public_id, state, sender_id, sender_id_type, sender_amount, sender_currency, sender_account, receiver_id, receiver_id_type, receiver_amount, receiver_currency, receiver_account, send_transaction_id, receive_transaction_id, action_three_ds_required, action_three_ds_id, action_otp_required, action_otp, note, ip_address, type, fx_rate, fx_fee_percentage, payment_protection_fee_percentage, created_at, updated_at`

type dbPayment struct {
	ID                             string                `db:"id"`
	PublicID                       string                `db:"public_id"`
	State                          payments.State        `db:"state"`
	ThreeDSRequired                bool                  `db:"action_three_ds_required"`
	ThreeDSID                      sql.NullString        `db:"action_three_ds_id"`
	SenderID                       string                `db:"sender_id"`
	SenderIDType                   payments.IdentityType `db:"sender_id_type"`
	SenderAmount                   uint64                `db:"sender_amount"`
	SenderCurrency                 string                `db:"sender_currency"`
	SenderAccount                  sql.NullString        `db:"sender_account"`
	ReceiverID                     string                `db:"receiver_id"`
	ReceiverIDType                 payments.IdentityType `db:"receiver_id_type"`
	ReceiverAmount                 uint64                `db:"receiver_amount"`
	ReceiverCurrency               string                `db:"receiver_currency"`
	ReceiverAccount                sql.NullString        `db:"receiver_account"`
	SendTransactionID              sql.NullString        `db:"send_transaction_id"`
	ReceiveTransactionID           sql.NullString        `db:"receive_transaction_id"`
	Note                           sql.NullString        `db:"note"`
	OTPRequired                    bool                  `db:"action_otp_required"`
	OTP                            sql.NullString        `db:"action_otp"`
	CreatedAt                      time.Time             `db:"created_at"`
	UpdatedAt                      time.Time             `db:"updated_at"`
	IPAddress                      sql.NullString        `db:"ip_address"`
	Type                           payments.Type         `db:"type"`
	FXRate                         sql.NullFloat64       `db:"fx_rate"`
	FXFeePercentage                sql.NullFloat64       `db:"fx_fee_percentage"`
	PaymentProtectionFeePercentage float64               `db:"payment_protection_fee_percentage"`
}

func lookupWallet(ctx context.Context, b Backends, identity payments.Identity) (*wallets.Wallet, error) {
	var resp *wallets.Wallet
	var err error
	switch identity.Type {
	case payments.IdentityTypeWalletID:
		resp, err = b.Wallets().Get(ctx, identity.Identifier)
	case payments.IdentityTypeWalletURL:
		resp, err = b.Wallets().GetFromAddress(ctx, identity.Identifier)
	case payments.IdentityTypeTwitter:
		var id *identities.Identity
		id, err = b.Identities().GetByIdentifier(ctx, identity.Identifier)
		if err != nil {
			return nil, err
		}
		if !strings.EqualFold(string(id.Platform), string(identities.PlatformTwitter)) {
			return nil, fmt.Errorf("identifier (%s) type mismatch expected (%s) got (%s)", identity.Identifier, identities.PlatformTwitter, identity.Type)
		}
		resp, err = b.Wallets().Get(ctx, id.WalletID)
	case payments.IdentityTypeSlack:
		var id *identities.Identity
		id, err = b.Identities().GetByIdentifier(ctx, identity.Identifier)
		if err != nil {
			return nil, err
		}
		if !strings.EqualFold(string(id.Platform), string(identities.PlatformSlack)) {
			return nil, fmt.Errorf("identifier (%s) type mismatch expected (%s) got (%s)", identity.Identifier, identities.PlatformSlack, identity.Type)
		}
		resp, err = b.Wallets().Get(ctx, id.WalletID)
	case payments.IdentityTypeDiscord:
		var id *identities.Identity
		id, err = b.Identities().GetByIdentifier(ctx, identity.Identifier)
		if err != nil {
			return nil, err
		}
		if !strings.EqualFold(string(id.Platform), string(identities.PlatformDiscord)) {
			return nil, fmt.Errorf("identifier (%s) type mismatch expected (%s) got (%s)", identity.Identifier, identities.PlatformDiscord, identity.Type)
		}
		resp, err = b.Wallets().Get(ctx, id.WalletID)
	default:
		return nil, fmt.Errorf("unknown identity type %s", identity.Type)
	}
	return resp, err
}

func transformPayment(ctx context.Context, b Backends, db dbPayment) (*payments.Payment, error) {
	var senderWalletID, receiverWalletID string
	if db.SenderIDType == payments.IdentityTypeWalletID {
		senderWalletID = db.SenderID
	} else {
		senderWallet, err := lookupWallet(ctx, b, payments.Identity{
			Type:       db.SenderIDType,
			Identifier: db.SenderID,
		})
		if err != nil {
			return nil, err
		}
		senderWalletID = senderWallet.ID
	}

	if db.ReceiverIDType == payments.IdentityTypeWalletID {
		receiverWalletID = db.ReceiverID
	} else if db.ReceiverIDType.Valid() {
		receiverWallet, err := lookupWallet(ctx, b, payments.Identity{
			Type:       db.ReceiverIDType,
			Identifier: db.ReceiverID,
		})
		if err != nil && !errors.Is(err, identities.ErrNotFound) {
			return nil, err
		}
		if receiverWallet != nil {
			receiverWalletID = receiverWallet.ID
		}
	}

	return &payments.Payment{
		ID:       db.ID,
		PublicID: db.PublicID,
		State:    db.State,
		Sender: payments.Identity{
			Type:       db.SenderIDType,
			Identifier: db.SenderID,
			WalletID:   senderWalletID,
		},
		Receiver: payments.Identity{
			Type:       db.ReceiverIDType,
			Identifier: db.ReceiverID,
			WalletID:   receiverWalletID,
		},
		SenderAmount:                   currency.FromUInt64(db.SenderAmount, currency.ParseCurrency(db.SenderCurrency)),
		ReceiverAmount:                 currency.FromUInt64(db.ReceiverAmount, currency.ParseCurrency(db.ReceiverCurrency)),
		SenderAccount:                  db.SenderAccount.String,
		ReceiverAccount:                db.ReceiverAccount.String,
		SendTransactionID:              db.SendTransactionID.String,
		ReceiveTransactionID:           db.ReceiveTransactionID.String,
		RequiredActions:                getRequiredActions(&db),
		Note:                           db.Note.String,
		IPAddress:                      db.IPAddress.String,
		UpdatedAt:                      db.UpdatedAt,
		Type:                           db.Type,
		FXRate:                         db.FXRate.Float64,
		FXFeePercentage:                db.FXFeePercentage.Float64,
		PaymentProtectionFeePercentage: db.PaymentProtectionFeePercentage,
	}, nil
}

func getPayment(ctx context.Context, b Backends, id string) (*dbPayment, error) {
	dbID, err := uuid.Parse(id)
	if err != nil {
		return getPaymentByPublicID(ctx, b, id)
	}

	var res dbPayment
	err = b.DB().GetContext(ctx, &res, fmt.Sprintf("SELECT %s FROM payments WHERE id=$1;", cols), dbID.String())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, payments.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	return &res, nil
}

func getPaymentByPublicID(ctx context.Context, b Backends, publicID string) (*dbPayment, error) {
	var res dbPayment
	err := b.DB().GetContext(ctx, &res, fmt.Sprintf("SELECT %s FROM payments WHERE public_id=$1", cols), publicID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, payments.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	return &res, nil
}

func getPaymentByPublicIDTx(ctx context.Context, tx *sqlx.Tx, publicID string) (*dbPayment, error) {
	var res dbPayment
	err := tx.GetContext(ctx, &res, fmt.Sprintf("SELECT %s FROM payments WHERE public_id=$1", cols), publicID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, payments.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	return &res, nil
}

func getPaymentTX(ctx context.Context, tx *sqlx.Tx, id string) (*dbPayment, error) {
	dbID, err := uuid.Parse(id)
	if err != nil {
		return getPaymentByPublicIDTx(ctx, tx, id)
	}

	var res dbPayment
	err = tx.GetContext(ctx, &res, fmt.Sprintf("SELECT %s FROM payments WHERE id=$1;", cols), dbID.String())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, payments.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	return &res, nil
}

func Lookup(ctx context.Context, b Backends, id string) (*payments.Payment, error) {
	dbp, err := getPayment(ctx, b, id)
	if err != nil {
		return nil, err
	}

	return transformPayment(ctx, b, *dbp)
}

func accountCanSend(ctx context.Context, b Backends, id payments.Identity, accountID string) (bool, error) {
	if accountID == "" || id.IsEmpty() {
		return false, nil
	}
	acc, err := b.LinkedAccounts().Get(ctx, accountID)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	w, err := lookupWallet(ctx, b, id)
	if err != nil {
		return false, err
	}

	if acc.WalletID != w.ID {
		return false, nil
	}

	return acc.CanSend, nil
}

func accountCanReceive(ctx context.Context, b Backends, id payments.Identity, accountID string) (bool, error) {
	if accountID == "" || id.IsEmpty() {
		return false, nil
	}
	acc, err := b.LinkedAccounts().Get(ctx, accountID)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	w, err := lookupWallet(ctx, b, id)
	if errors.Is(err, identities.ErrNotFound) {
		return false, nil
	}
	if err != nil && !errors.Is(err, identities.ErrNotFound) {
		return false, err
	}

	if w.ID != acc.WalletID {
		return false, nil
	}

	return acc.CanReceive, nil
}

func defaultReceiveAccount(ctx context.Context, b Backends, id payments.Identity) (string, error) {
	if id.IsEmpty() || !id.Type.Valid() {
		return "", nil
	}

	wallet, err := lookupWallet(ctx, b, id)
	if err != nil {
		return "", err
	}

	la, err := b.LinkedAccounts().GetDefaultReceive(ctx, wallet.ID)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	return la.ID, nil
}

func requiresOTP(ctx context.Context, b Backends, typ payments.Type, sender, receiver payments.Identity) (bool, error) {

	// Web monetization payouts don't need an OTP
	if typ == payments.TypeWebMonetization || typ == payments.TypeReferral {
		return false, nil
	}

	// Default to requiring
	if receiver.IsEmpty() || !receiver.Type.Valid() {
		return true, nil
	}

	senderWallet, err := lookupWallet(ctx, b, sender)
	if err != nil {
		return false, err
	}

	receiverWallet, err := lookupWallet(ctx, b, receiver)
	if errors.Is(err, identities.ErrNotFound) {
		// Identity not yet linked, require OTP
		return true, nil
	}
	if err != nil {
		return false, err
	}

	hasTx, err := b.Transactions().GetHasTransacted(ctx, senderWallet.ID, receiverWallet.AddressString())
	if err != nil {
		return false, err
	}

	return !hasTx, nil
}

func requires3DS(sender payments.Identity) bool {
	return sender.Identifier != wallets.WebMonetizationWalletID && sender.Identifier != wallets.ReferralsWalletID
}

// Create The `Sender` is the minimum required information to create a payment. If the specified identity
// is not of type WalletID, then the walletID will be looked up and used as the `Sender`.
func Create(ctx context.Context, b Backends, p payments.CreateArgs) (*payments.Payment, error) {
	// convert sender identity to walletID
	if p.Sender.IsEmpty() || !p.Sender.Type.Valid() {
		return nil, fmt.Errorf("%w Sender is invalid.", payments.ErrInvalidIdentifier)
	}

	// Default to Peer to Peer
	if p.Type == payments.TypeUnknown {
		p.Type = payments.TypePeer2Peer
	}

	senderWallet, err := lookupWallet(ctx, b, p.Sender)
	if err != nil {
		return nil, err
	}

	canSend, err := accountCanSend(ctx, b, p.Sender, p.SenderAccount)
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}
	if !canSend {
		p.SenderAccount = ""
	}

	canReceive, err := accountCanReceive(ctx, b, p.Receiver, p.ReceiverAccount)
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	if p.ReceiverAccount == "" || !canReceive {
		p.ReceiverAccount, _ = defaultReceiveAccount(ctx, b, p.Receiver)
	}

	requireOTP, err := requiresOTP(ctx, b, p.Type, p.Sender, p.Receiver)
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	require3DS := requires3DS(p.Sender)

	publicID, err := NewSoftDescriptor(time.Now())
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	p, fxRate, fxFeePerc, err := applyFXCreate(ctx, b, p)
	if err != nil {
		return nil, err
	}

	var paymentProtectionFeePercentage float64
	if p.AddPaymentProtection.Valid && p.AddPaymentProtection.Bool {
		amount, feePercentage, err := applyPaymentProtection(ctx, b, p.Sender, p.SenderAmount)
		if err != nil {
			return nil, err
		}
		paymentProtectionFeePercentage = feePercentage
		p.SenderAmount.Value = amount.Value
	}

	// TODO Calculate more actions required
	id := uuid.NewString()
	stmt, args, err := db.NewInsert("payments").
		Value("id", id).
		Value("public_id", publicID).
		Value("state", payments.StateCreated).
		Value("sender_id", senderWallet.ID).
		Value("sender_id_type", payments.IdentityTypeWalletID).
		Value("sender_amount", p.SenderAmount.Value).
		Value("sender_currency", p.SenderAmount.Currency).
		Value("sender_account", sql.NullString{String: p.SenderAccount, Valid: p.SenderAccount != "" && canSend}).
		Value("receiver_id", p.Receiver.Identifier).
		Value("receiver_id_type", p.Receiver.Type).
		Value("receiver_amount", p.ReceiverAmount.Value).
		Value("receiver_currency", p.ReceiverAmount.Currency).
		Value("receiver_account", sql.NullString{String: p.ReceiverAccount, Valid: p.ReceiverAccount != ""}).
		Value("action_three_ds_required", require3DS).
		Value("note", sql.NullString{String: p.Note, Valid: p.Note != ""}).
		Value("action_otp_required", requireOTP).
		Value("ip_address", sql.NullString{String: p.IPAddress, Valid: b.Validator().Var(p.IPAddress, "ip_addr") == nil}).
		Value("type", p.Type).
		Value("fx_rate", sql.NullFloat64{Float64: fxRate, Valid: fxRate > 0}).
		Value("fx_fee_percentage", sql.NullFloat64{Float64: fxFeePerc, Valid: fxFeePerc > 0}).
		Value("payment_protection_fee_percentage", paymentProtectionFeePercentage).
		GetStatement()
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	_, err = b.DB().ExecContext(ctx, stmt, args...)
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	return Lookup(ctx, b, id)
}

// Hard coded to be calculated as 3%.
// TODO: Make it dynamic by looking at senders payment history
func applyPaymentProtection(ctx context.Context, b Backends, sender payments.Identity, amount currency.Amount) (currency.Amount, float64, error) {
	rate := float64(0.03)
	smallest := currency.FromUInt64(1, amount.Currency)
	fee := amount.Float64() * rate
	if fee < smallest.Float64() {
		fee = 0
	}

	feeAmount := currency.FromFloat64(fee, amount.Currency)

	return currency.FromUInt64(feeAmount.Value+amount.Value, amount.Currency), rate, nil
}

func applyFXCreate(ctx context.Context, b Backends, args payments.CreateArgs) (payments.CreateArgs, float64, float64, error) {
	if args.SenderAccount == "" || args.ReceiverAccount == "" {
		return args, 0, 0, nil
	}

	senderAcc, err := b.LinkedAccounts().Get(ctx, args.SenderAccount)
	if err != nil {
		return args, 0, 0, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}
	if senderAcc.SendCurrency != currency.USD {
		return args, 0, 0, fmt.Errorf("%w currently only from USD to other currencies are supported", payments.ErrInternal)
	}

	receiverAcc, err := b.LinkedAccounts().Get(ctx, args.ReceiverAccount)
	if err != nil {
		return args, 0, 0, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	if receiverAcc.ReceiveCurrency == senderAcc.SendCurrency {
		if !args.SenderAmount.IsEqual(args.ReceiverAmount) {
			// Equalize sender and receiver amounts until we add fees.
			if args.SenderAmount.Value > args.ReceiverAmount.Value {
				args.ReceiverAmount = args.SenderAmount
			} else {
				args.SenderAmount = args.ReceiverAmount
			}
		}
		return args, 0, 0, nil
	}

	if receiverAcc.ReceiveCurrency == currency.USD {
		return args, 0, 0, fmt.Errorf("%w currently only from USD to other currencies are supported", payments.ErrInternal)
	}

	fx, err := b.Tabapay().GetFXRate(ctx, receiverAcc.ReceiveCurrency)
	if err != nil {
		return args, 0, 0, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	rate := fx.MatercardRate
	if strings.Contains(strings.ToLower(receiverAcc.ReceiveNetwork), "visa") {
		rate = fx.VisaRate
	}

	if args.SenderAmount.Value > 0 {
		args.ReceiverAmount = currency.FromFloat64(rate.FromUSD(args.SenderAmount.Float64()), receiverAcc.ReceiveCurrency)
		return args, rate.BuyRateInv, 0, nil
	}
	args.SenderAmount = currency.FromFloat64(rate.ToUSD(args.ReceiverAmount.Float64()), senderAcc.SendCurrency)
	return args, rate.SellRate, 0, nil
}

func SetState(ctx context.Context, b Backends, id string, state payments.State) error {
	p, err := Lookup(ctx, b, id)
	if err != nil {
		return err
	}
	if !p.State.CanTransitionTo(state) {
		return fmt.Errorf("%w id=%s current state=%s, proposed state=%s", payments.ErrInvalidStateTransition, id, p.State, state)
	}

	result, err := b.DB().ExecContext(ctx, "UPDATE payments SET state=$1 WHERE id=$2 AND state=$3;", state, p.ID, p.State)
	if err != nil {
		return fmt.Errorf("%w %s", payments.ErrInternal, err)
	}
	if rows, _ := result.RowsAffected(); rows < 1 {
		return fmt.Errorf("%w Failed to update state. id=%s, proposed state=%s", payments.ErrInternal, id, state)
	}

	return nil
}

func setStateTX(ctx context.Context, tx *sqlx.Tx, id string, state payments.State) error {
	p, err := getPaymentTX(ctx, tx, id)
	if err != nil {
		return err
	}
	if !p.State.CanTransitionTo(state) {
		return fmt.Errorf("%w id=%s current state=%s, proposed state=%s", payments.ErrInvalidStateTransition, id, p.State, state)
	}

	result, err := tx.ExecContext(ctx, "UPDATE payments SET state=$1 WHERE id=$2 AND state=$3;", state, p.ID, p.State)
	if err != nil {
		return fmt.Errorf("%w %s", payments.ErrInternal, err)
	}
	if rows, _ := result.RowsAffected(); rows < 1 {
		return fmt.Errorf("%w Failed to update state. id=%s, proposed state=%s", payments.ErrInternal, id, state)
	}

	return nil
}

func setReceiveTransactionID(ctx context.Context, b Backends, paymentID, txID string) error {
	_, err := b.DB().ExecContext(ctx, "UPDATE payments SET receive_transaction_id=$1 WHERE id=$2", txID, paymentID)
	if err != nil {
		return fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	return nil
}

/*
GetRequiredActions checks that the payment has the following:

1) Sender and send account
2) Receiver identifier
3) Send amount
4) Receive amount
5) 3DSID
4) OTP if required
*/
func GetRequiredActions(ctx context.Context, b Backends, id string) ([]payments.RequiredActionType, error) {
	payment, err := getPayment(ctx, b, id)
	if err != nil {
		return nil, err
	}

	requiredActions := getRequiredActions(payment)

	return requiredActions, nil
}

func getRequiredActions(payment *dbPayment) []payments.RequiredActionType {
	var requiredActions []payments.RequiredActionType
	if payment.SenderID == "" {
		requiredActions = append(requiredActions, payments.RequiredActionTypeSenderIdentifier)
	}

	if payment.SenderAccount.String == "" {
		requiredActions = append(requiredActions, payments.RequiredActionTypeSenderAccount)
	}

	if payment.ReceiverID == "" || payment.ReceiverIDType == payments.IdentityTypeUnknown {
		requiredActions = append(requiredActions, payments.RequiredActionTypeReceiverIdentifier)
	}

	if payment.SenderAmount < 1 || !currency.ParseCurrency(payment.SenderCurrency).Valid() {
		requiredActions = append(requiredActions, payments.RequiredActionTypeSenderAmount)
	}

	if payment.ReceiverAmount < 1 || !currency.ParseCurrency(payment.ReceiverCurrency).Valid() {
		requiredActions = append(requiredActions, payments.RequiredActionTypeReceiverAmount)
	}

	if payment.OTPRequired && payment.OTP.String == "" {
		requiredActions = append(requiredActions, payments.RequiredActionTypeOTP)
	}

	if payment.ThreeDSRequired && payment.ThreeDSID.String == "" {
		requiredActions = append(requiredActions, payments.RequiredActionTypeThreeDS)
	}

	if payment.IPAddress.String == "" {
		requiredActions = append(requiredActions, payments.RequiredActionTypeIPAddress)
	}

	return requiredActions
}

func Confirm(ctx context.Context, b Backends, id string) (*payments.Payment, []payments.RequiredActionType, error) {
	requiredActions, err := GetRequiredActions(ctx, b, id)
	if err != nil {
		return nil, nil, err
	}
	if len(requiredActions) > 0 {
		return nil, requiredActions, payments.ErrRequiredActions
	}

	// Lookup in case they used the public ID
	dbp, err := Lookup(ctx, b, id)
	if err != nil {
		return nil, nil, err
	}

	destination := dbp.Receiver.Identifier
	receiverWallet, err := lookupWallet(ctx, b, dbp.Receiver)
	if err != nil && !errors.Is(err, identities.ErrNotFound) {
		return nil, nil, err
	}
	if receiverWallet != nil {
		destination = receiverWallet.AddressString()
	}

	senderWallet, err := lookupWallet(ctx, b, dbp.Sender)
	if err != nil {
		return nil, nil, err
	}

	la, err := b.LinkedAccounts().Get(ctx, dbp.SenderAccount)
	if err != nil {
		return nil, nil, err
	}

	// Do all precursor operations in single TX so we don't get inconsistent state.
	err = crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		err := setStateTX(ctx, tx, id, payments.StateConfirmed)
		if err != nil {
			return err
		}

		txID, err := b.Transactions().CreateTransactionTx(ctx, tx, transactions.CreateTransactionArgs{
			WalletID:                       senderWallet.ID,
			ForeignID:                      dbp.ID,
			ForeignType:                    transactions.TransactionTypeOutgoing,
			Provider:                       transactions.ProviderPaymentsEngine,
			State:                          transactions.StatePending,
			Note:                           dbp.Note,
			Source:                         senderWallet.AddressString(),
			Destination:                    destination,
			Amount:                         dbp.SenderAmount,
			LinkedAccountTitle:             la.Title(),
			DestinationIdentity:            dbp.Receiver.Identifier,
			DestinationIdentityType:        dbp.Receiver.Type.String(),
			Reference:                      dbp.Note,
			PaymentProtectionFeePercentage: dbp.PaymentProtectionFeePercentage,
		})
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, "UPDATE payments SET send_transaction_id=$1 WHERE id=$2", txID, dbp.ID)
		if err != nil {
			return fmt.Errorf("%w %s", payments.ErrInternal, err)
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	workflowOptions := temporal_client.StartWorkflowOptions{
		ID:                       "payments_" + dbp.ID,
		TaskQueue:                "backend",
		WorkflowExecutionTimeout: time.Hour * 24 * 8, // Workflow has 8 days to complete
		WorkflowIDReusePolicy:    enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}

	_, err = b.Temporal().ExecuteWorkflow(ctx, workflowOptions, PaymentWorkflow, dbp.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	if dbp.Type == payments.TypeReferral {
		slack.SendToChannel(ctx, slack.ChannelNotifyEvents, "Fynbot", fmt.Sprintf(":money_with_wings: New Referral Payment Created\nID: %s\nReceiver Wallet ID: %s\nAmount:%s\nTemporal WorkflowID: %s", dbp.ID, dbp.Receiver.WalletID, dbp.SenderAmount.Format(), "payments_"+dbp.ID))
	}

	payment, err := Lookup(ctx, b, id)
	if err != nil {
		return nil, nil, err
	}

	return payment, nil, nil
}

// Update does validation on all fields that can be set from outside of the payments engine and the calls internal update functionality.
func Update(ctx context.Context, b Backends, args payments.UpdateArgs) (*payments.Payment, error) {
	payment, err := getPayment(ctx, b, args.ID)
	if err != nil {
		return nil, err
	}

	if payment.State != payments.StateCreated {
		return nil, fmt.Errorf("%w Cannot update payment in state (%s)", payments.ErrInvalidState, payment.State)
	}

	return update(ctx, b, args)
}

// update performs minimal validation and updates a payment. This is only available internally to the payments engine
// where updates can be made to the payment regardless of the state of the payment.
func update(ctx context.Context, b Backends, args payments.UpdateArgs) (*payments.Payment, error) {
	payment, err := getPayment(ctx, b, args.ID)
	if err != nil {
		return nil, err
	}
	if !args.SenderAmount.IsEmpty() && !args.SenderAmount.Currency.Valid() {
		return nil, fmt.Errorf("%w Sender amount currency is invalid", payments.ErrInvalidAmount)
	}
	if !args.ReceiverAmount.IsEmpty() && !args.ReceiverAmount.Currency.Valid() {
		return nil, fmt.Errorf("%w Receiver amount currency is invalid", payments.ErrInvalidAmount)
	}
	if !args.Receiver.IsEmpty() && !args.Receiver.Type.Valid() {
		return nil, fmt.Errorf("%w Receiver is invalid", payments.ErrInvalidIdentifier)
	}

	var receiverAmtUpdated, senderAmtUpdated bool
	noop := true
	receiver := payments.Identity{Identifier: payment.ReceiverID, Type: payment.ReceiverIDType}
	if !args.Receiver.IsEmpty() && !args.Receiver.IsEqual(receiver) {
		payment.ReceiverID = args.Receiver.Identifier
		payment.ReceiverIDType = args.Receiver.Type
		noop = false
	}

	if args.ReceiverAccount != "" && args.ReceiverAccount != payment.ReceiverAccount.String {
		canReceive, err := accountCanReceive(ctx, b, payments.Identity{Identifier: payment.ReceiverID, Type: payment.ReceiverIDType}, args.ReceiverAccount)
		if err != nil {
			return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
		}
		payment.ReceiverAccount.String = args.ReceiverAccount
		payment.ReceiverAccount.Valid = canReceive
		noop = false
	}
	if args.SenderAccount != "" && args.SenderAccount != payment.SenderAccount.String {
		canSend, err := accountCanSend(ctx, b, payments.Identity{Identifier: payment.SenderID, Type: payment.SenderIDType}, args.SenderAccount)
		if err != nil {
			return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
		}
		payment.SenderAccount.String = args.SenderAccount
		payment.SenderAccount.Valid = canSend
		noop = false
	}

	receiveAmount := currency.FromUInt64(payment.ReceiverAmount, currency.Currency(payment.ReceiverCurrency))
	if !args.ReceiverAmount.IsEmpty() && !args.ReceiverAmount.IsEqual(receiveAmount) {
		payment.ReceiverAmount = args.ReceiverAmount.Value
		payment.ReceiverCurrency = args.ReceiverAmount.Currency.String()
		noop = false
		receiverAmtUpdated = true
	}
	sendAmount := currency.FromUInt64(payment.SenderAmount, currency.Currency(payment.SenderCurrency))
	if !args.SenderAmount.IsEmpty() && !args.SenderAmount.IsEqual(sendAmount) {
		payment.SenderAmount = args.SenderAmount.Value
		payment.SenderCurrency = args.SenderAmount.Currency.String()
		noop = false
		senderAmtUpdated = true
	}
	if args.ThreeDSID != "" && args.ThreeDSID != payment.ThreeDSID.String {
		payment.ThreeDSID.String = args.ThreeDSID
		payment.ThreeDSID.Valid = true
		noop = false
	}
	if args.OTP != "" && args.OTP != payment.OTP.String {
		payment.OTP.String = args.OTP
		payment.OTP.Valid = true
		noop = false
	}
	if args.Note != "" && args.Note != payment.Note.String {
		payment.Note = sql.NullString{
			String: args.Note,
			Valid:  true,
		}
		noop = false
	}
	if args.IPAddress != "" && args.IPAddress != payment.IPAddress.String && b.Validator().Var(args.IPAddress, "ip_addr") == nil {
		noop = false
		payment.IPAddress = sql.NullString{
			String: args.IPAddress,
			Valid:  true,
		}
	}

	if noop {
		return transformPayment(ctx, b, *payment)
	}

	// Something changed, update the FX calculations
	if receiverAmtUpdated || senderAmtUpdated {
		payment, err = applyFXUpdate(ctx, b, payment, receiverAmtUpdated)
		if err != nil {
			return nil, err
		}
	}

	hasPaymentProtection := payment.PaymentProtectionFeePercentage != 0
	if (receiverAmtUpdated || senderAmtUpdated) && hasPaymentProtection {
		amount, feePercentage, err := applyPaymentProtection(
			ctx,
			b,
			payments.Identity{
				Type:       payment.SenderIDType,
				Identifier: payment.SenderID,
			},
			currency.FromUInt64(payment.SenderAmount, currency.ParseCurrency(payment.SenderCurrency)),
		)
		if err != nil {
			return nil, err
		}

		payment.SenderAmount = amount.Value
		payment.PaymentProtectionFeePercentage = feePercentage
	}

	payment.UpdatedAt = time.Now()
	stmt, values, err := db.NewUpdate("payments").ID(payment.ID).
		Value("sender_amount", payment.SenderAmount).
		Value("sender_currency", payment.SenderCurrency).
		Value("sender_account", payment.SenderAccount).
		Value("receiver_id", payment.ReceiverID).
		Value("receiver_id_type", payment.ReceiverIDType).
		Value("receiver_amount", payment.ReceiverAmount).
		Value("receiver_currency", payment.ReceiverCurrency).
		Value("receiver_account", payment.ReceiverAccount).
		Value("updated_at", payment.UpdatedAt).
		Value("note", payment.Note).
		Value("action_three_ds_id", payment.ThreeDSID).
		Value("ip_address", payment.IPAddress).
		Value("action_otp", payment.OTP).
		Value("fx_rate", payment.FXRate).
		Value("fx_fee_percentage", payment.FXFeePercentage).
		Returning(cols).GetStatement()
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	var ret dbPayment
	err = b.DB().GetContext(ctx, &ret, stmt, values...)
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	return transformPayment(ctx, b, ret)
}

func AddPaymentProtection(ctx context.Context, b Backends, id string, add bool) (*payments.Payment, error) {
	p, err := Lookup(ctx, b, id)
	if err != nil {
		return nil, err
	}

	// remove old payment percentage. This is defaulted to 0.00. Then check if we must add it back.
	protectionFee := p.PaymentProtectionAmount()
	newSenderAmount := currency.Amount{
		Value:    p.SenderAmount.Value - protectionFee.Value,
		Currency: p.SenderAmount.Currency,
	}
	newPaymentProtectionFeePercentage := float64(0)
	if add {
		amount, feePercentage, err := applyPaymentProtection(
			ctx,
			b,
			p.Sender,
			newSenderAmount,
		)
		if err != nil {
			return nil, err
		}

		newSenderAmount = amount
		newPaymentProtectionFeePercentage = feePercentage
	}

	stmt, values, err := db.NewUpdate("payments").ID(p.ID).
		Value("sender_amount", newSenderAmount.Value).
		Value("payment_protection_fee_percentage", newPaymentProtectionFeePercentage).
		Returning(cols).GetStatement()

	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	var ret dbPayment
	err = b.DB().GetContext(ctx, &ret, stmt, values...)
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	return transformPayment(ctx, b, ret)
}

func applyFXUpdate(ctx context.Context, b Backends, existing *dbPayment, receiverAmtUpdated bool) (*dbPayment, error) {
	if !existing.SenderAccount.Valid || !existing.ReceiverAccount.Valid {
		return existing, nil
	}

	senderAcc, err := b.LinkedAccounts().Get(ctx, existing.SenderAccount.String)
	if err != nil {
		return existing, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}
	if senderAcc.SendCurrency != currency.USD {
		return existing, fmt.Errorf("%w currently only from USD to other currencies are supported", payments.ErrInternal)
	}

	receiverAcc, err := b.LinkedAccounts().Get(ctx, existing.ReceiverAccount.String)
	if err != nil {
		return existing, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	if receiverAcc.ReceiveCurrency == senderAcc.SendCurrency {
		// Equalize sender and receiver amounts until we add fees.
		if receiverAmtUpdated {
			existing.SenderAmount = existing.ReceiverAmount
			return existing, nil
		}

		existing.ReceiverAmount = existing.SenderAmount
		return existing, nil
	}

	if receiverAcc.ReceiveCurrency == currency.USD {
		return existing, fmt.Errorf("%w currently only from USD to other currencies are supported", payments.ErrInternal)
	}

	fx, err := b.Tabapay().GetFXRate(ctx, receiverAcc.ReceiveCurrency)
	if err != nil {
		return existing, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	rate := fx.MatercardRate
	if strings.Contains(strings.ToLower(receiverAcc.ReceiveNetwork), "visa") {
		rate = fx.VisaRate
	}

	if receiverAmtUpdated {
		receiveAmt := currency.FromUInt64(existing.ReceiverAmount, receiverAcc.ReceiveCurrency)
		existing.SenderAmount = currency.FromFloat64(rate.ToUSD(receiveAmt.Float64()), senderAcc.SendCurrency).Value
		existing.SenderCurrency = senderAcc.SendCurrency.String()
		existing.FXRate = sql.NullFloat64{Float64: rate.SellRate, Valid: true}
		return existing, nil
	}
	sendAmt := currency.FromUInt64(existing.SenderAmount, senderAcc.SendCurrency)
	existing.ReceiverAmount = currency.FromFloat64(rate.FromUSD(sendAmt.Float64()), receiverAcc.ReceiveCurrency).Value
	existing.ReceiverCurrency = receiverAcc.ReceiveCurrency.String()
	existing.FXRate = sql.NullFloat64{Float64: rate.BuyRateInv, Valid: true}
	return existing, nil
}

func UpdateReceiver(ctx context.Context, b Backends, id string, identity payments.Identity) error {
	_, err := update(ctx, b, payments.UpdateArgs{ID: id, Receiver: identity})
	if err != nil {
		return err
	}

	_, err = b.DB().ExecContext(ctx, "UPDATE payments_workflow_refs SET identifier=$1 WHERE payment_id=$2", identity.Identifier, id)
	if err != nil {
		return fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	return nil
}

type workflowRef struct {
	PaymentID  string `db:"payment_id"`
	Identifier string `db:"identifier"`
	WalletID   string `db:"wallet_id"`
	WorkflowID string `db:"workflow_id"`
	RunID      string `db:"workflow_run_id"`
}

func SignalIdentityCreated(ctx context.Context, b Backends, identifier string) error {
	var refs []workflowRef
	err := b.DB().SelectContext(ctx, &refs, "SELECT payment_id, identifier, wallet_id, workflow_id, workflow_run_id FROM payments_workflow_refs WHERE identifier=$1", identifier)
	if err != nil {
		return fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	for _, ref := range refs {
		err = b.Temporal().SignalWorkflow(ctx, ref.WorkflowID, ref.RunID, identityChanName, nil)
		if err != nil {
			return fmt.Errorf("%w %s", payments.ErrInternal, err)
		}
	}

	return nil
}

func SignalAccountLinked(ctx context.Context, b Backends, walletID string) error {
	var refs []workflowRef
	err := b.DB().SelectContext(ctx, &refs, "SELECT payment_id, identifier, wallet_id, workflow_id, workflow_run_id FROM payments_workflow_refs WHERE wallet_id=$1", walletID)
	if err != nil {
		return fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	for _, ref := range refs {
		err = b.Temporal().SignalWorkflow(ctx, ref.WorkflowID, ref.RunID, identityChanName, nil)
		if err != nil {
			return fmt.Errorf("%w %s", payments.ErrInternal, err)
		}
	}

	return nil
}

func ListAwaitingSignal(ctx context.Context, b Backends) ([]payments.Payment, error) {
	var dbRes []dbPayment
	err := b.DB().GetContext(ctx, &dbRes, fmt.Sprintf("SELECT %s FROM payments WHERE id IN (select payment_id from payments_workflow_refs where completed=false)", cols))
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	res := make([]payments.Payment, len(dbRes))
	for i, p := range dbRes {
		temp, err := transformPayment(ctx, b, p)
		if err != nil {
			return nil, err
		}
		res[i] = *temp
	}

	return res, nil
}

const (
	geohashBase32Alphabet = "0123456789bcdefghjkmnpqrstuvwxyz"
	hoursInUnixTimestamp  = 60 * 60
)

// Generates a soft descriptor to show on the user's card statement.
// See https://www.notion.so/fynbos/Soft-Descriptors-08b6693f96194f54ba0d62e21efd22d6
func NewSoftDescriptor(date time.Time) (string, error) {
	// Generate a random uint64.
	buffer := make([]byte, 8)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}
	// Put 5 bytes of randomness in first 40 bits of uint64
	randomNum := binary.BigEndian.Uint64(buffer) << 24

	// Put 20 bits of date in bits 41 - 60 of unit64
	dateValue := uint64(date.Unix()/hoursInUnixTimestamp) << 4

	// XOR the two values together and convert to a byte slice.
	binary.BigEndian.PutUint64(buffer, randomNum^dateValue)

	// Create a Geohash Base32 encoder.
	geohashEncoder := base32.NewEncoding(geohashBase32Alphabet).WithPadding(base32.NoPadding)

	// Encode the buffer as a Base32 string.
	key := geohashEncoder.EncodeToString(buffer)

	// Trim key to 12 characters, we don't care about the last 4 bits
	return key[:12], nil
}
