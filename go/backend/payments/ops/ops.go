package ops

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"gitlab.com/fynbos/backend/providers/chimoney"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/pti"
	xago_external "gitlab.com/fynbos/backend/providers/xago/external"
	"gitlab.com/fynbos/backend/rafiki"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"

	"gitlab.com/fynbos/backend/providers/xago"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/wallets"
	"go.temporal.io/api/enums/v1"
	temporal_client "go.temporal.io/sdk/client"
)

const cols = `id, public_id, state, sender_id, sender_id_type, sender_amount, sender_currency, sender_account, receiver_id, receiver_id_type, receiver_amount, receiver_currency, receiver_account, send_transaction_id, receive_transaction_id, action_three_ds_required, action_three_ds_id, action_otp_required, action_otp, note, ip_address, type, fx_rate, fx_fee_percentage, revision, created_at, updated_at`

type dbPayment struct {
	ID                   string                `db:"id"`
	PublicID             string                `db:"public_id"`
	State                payments.State        `db:"state"`
	ThreeDSRequired      bool                  `db:"action_three_ds_required"`
	ThreeDSID            sql.NullString        `db:"action_three_ds_id"`
	SenderID             string                `db:"sender_id"`
	SenderIDType         payments.IdentityType `db:"sender_id_type"`
	SenderAmount         int64                 `db:"sender_amount"`
	SenderCurrency       string                `db:"sender_currency"`
	SenderAccount        sql.NullString        `db:"sender_account"`
	ReceiverID           string                `db:"receiver_id"`
	ReceiverIDType       payments.IdentityType `db:"receiver_id_type"`
	ReceiverAmount       int64                 `db:"receiver_amount"`
	ReceiverCurrency     string                `db:"receiver_currency"`
	ReceiverAccount      sql.NullString        `db:"receiver_account"`
	SendTransactionID    sql.NullString        `db:"send_transaction_id"`
	ReceiveTransactionID sql.NullString        `db:"receive_transaction_id"`
	Note                 sql.NullString        `db:"note"`
	OTPRequired          bool                  `db:"action_otp_required"`
	OTP                  sql.NullString        `db:"action_otp"`
	CreatedAt            time.Time             `db:"created_at"`
	UpdatedAt            time.Time             `db:"updated_at"`
	IPAddress            sql.NullString        `db:"ip_address"`
	Type                 payments.Type         `db:"type"`
	FXRate               sql.NullFloat64       `db:"fx_rate"`
	FXFeePercentage      sql.NullFloat64       `db:"fx_fee_percentage"`
	Revision             int                   `db:"revision"`
}

func lookupWallet(ctx context.Context, b Backends, identity payments.Identity) (*wallets.Wallet, error) {
	var resp *wallets.Wallet
	var err error
	switch identity.Type {
	case payments.IdentityTypeExternalWalletURL:
		return nil, fmt.Errorf("%w external wallet URL %s", identities.ErrNotFound, identity.Identifier)
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
	default:
		return nil, fmt.Errorf("%w unknown identity type %s", identities.ErrNotFound, identity.Type)
	}
	return resp, err
}

func transformPayment(ctx context.Context, b Backends, db dbPayment, senderWallet, receiverWallet *wallets.Wallet) (*payments.Payment, error) {
	var senderWalletID, receiverWalletID string
	var err error
	if db.SenderIDType == payments.IdentityTypeWalletID {
		senderWalletID = db.SenderID
	} else if senderWallet != nil {
		senderWalletID = senderWallet.ID
	} else {
		senderWallet, err = lookupWallet(ctx, b, payments.Identity{
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
	} else if receiverWallet != nil {
		receiverWalletID = receiverWallet.ID
	} else if db.ReceiverIDType.Valid() {
		receiverWallet, err = lookupWallet(ctx, b, payments.Identity{
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
		SenderAmount:         currency.FromUInt64(int64(db.SenderAmount), currency.ParseCurrency(db.SenderCurrency)),
		ReceiverAmount:       currency.FromUInt64(int64(db.ReceiverAmount), currency.ParseCurrency(db.ReceiverCurrency)),
		SenderAccount:        db.SenderAccount.String,
		ReceiverAccount:      db.ReceiverAccount.String,
		SendTransactionID:    db.SendTransactionID.String,
		ReceiveTransactionID: db.ReceiveTransactionID.String,
		RequiredActions:      getRequiredActions(&db),
		Note:                 db.Note.String,
		IPAddress:            db.IPAddress.String,
		UpdatedAt:            db.UpdatedAt,
		Type:                 db.Type,
		FXRate:               db.FXRate.Float64,
		FXFeePercentage:      db.FXFeePercentage.Float64,
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

	return transformPayment(ctx, b, *dbp, nil, nil)
}

func accountCanSend(ctx context.Context, b Backends, w *wallets.Wallet, accountID string, typ payments.Type) (bool, currency.Currency, error) {
	if accountID == "" || w == nil {
		return false, currency.USD, nil
	}
	acc, err := b.LinkedAccounts().Get(ctx, accountID)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return false, currency.USD, nil
	}
	if err != nil {
		return false, currency.USD, err
	}

	// PTI card accounts can't send unless it's a deposit
	if typ == payments.TypeDeposit && acc.Provider == pti.ProviderName && acc.Type == pti.TypeCard {
		return w.ID == acc.WalletID, acc.SendCurrency, nil
	}

	return acc.CanSend && w.ID == acc.WalletID, acc.SendCurrency, nil
}

func accountCanReceive(ctx context.Context, b Backends, w *wallets.Wallet, accountID string, typ payments.Type) (bool, currency.Currency, error) {
	if accountID == "" || w == nil {
		return false, "", nil
	}
	acc, err := b.LinkedAccounts().Get(ctx, accountID)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return false, currency.USD, nil
	}
	if err != nil {
		return false, currency.USD, err
	}

	// PTI card accounts can't receive unless it's a withdrawal
	if typ == payments.TypeWithdrawal && acc.Provider == pti.ProviderName && acc.Type == pti.TypeCard {
		return w.ID == acc.WalletID, acc.SendCurrency, nil
	}

	return acc.CanReceive && w.ID == acc.WalletID, acc.ReceiveCurrency, nil
}

func defaultReceiveAccount(ctx context.Context, b Backends, w *wallets.Wallet, cc currency.Currency) (*linkedaccounts.LinkedAccount, error) {
	if w == nil {
		return nil, nil
	}

	la, err := b.LinkedAccounts().GetDefaultReceive(ctx, w.ID, cc)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	return la, nil
}

func requiresOTP(ctx context.Context, b Backends, typ payments.Type, sender, receiver *wallets.Wallet) (bool, error) {

	// Web monetization payouts don't need an OTP
	if typ == payments.TypeWebMonetization || typ == payments.TypeRafikiPeer2Peer || typ == payments.TypeRafiki2External || typ == payments.TypeWithdrawal || typ == payments.TypeDeposit {
		return false, nil
	}

	// Default to requiring
	if receiver == nil || sender == nil {
		return true, nil
	}

	hasTx, err := b.Transactions().GetHasTransacted(ctx, sender.ID, receiver.AddressString())
	if err != nil {
		return false, err
	}

	return !hasTx, nil
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

	receiverWallet, err := lookupWallet(ctx, b, p.Receiver)
	if err != nil && !errors.Is(err, identities.ErrNotFound) {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	if p.SenderAccount == "" && p.ReceiverAccount == "" {
		sendBalances, err := b.LinkedAccounts().ListBalances(ctx, senderWallet.ID)
		if err != nil {
			return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
		}

		var recvBalances []linkedaccounts.LinkedAccount
		if receiverWallet != nil {
			recvBalances, err = b.LinkedAccounts().ListBalances(ctx, receiverWallet.ID)
			if err != nil {
				return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
			}
		}

		for _, sendLA := range sendBalances {
			for _, recvLA := range recvBalances {
				// TODO check if we can find a better way
				crossProvider := isXagoGatehubPair(sendLA.Provider, recvLA.Provider) &&
					recvLA.CanReceive && recvLA.State == linkedaccounts.Verified
				if sendLA.CanPay(recvLA) || crossProvider {
					p.SenderAccount = sendLA.ID
					p.ReceiverAccount = recvLA.ID
					p.SenderAmount.Currency = sendLA.SendCurrency
					p.ReceiverAmount.Currency = recvLA.ReceiveCurrency
					break
				}
			}

			if p.SenderAccount != "" && p.ReceiverAccount != "" {
				break
			}
		}
	}

	canSend, sendCurrency, err := accountCanSend(ctx, b, senderWallet, p.SenderAccount, p.Type)
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}
	if !canSend {
		p.SenderAccount = ""
	}
	if p.SenderAmount.Currency.String() == "" && canSend {
		p.SenderAmount.Currency = sendCurrency
	}

	canReceive, _, err := accountCanReceive(ctx, b, receiverWallet, p.ReceiverAccount, p.Type)
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	if p.ReceiverAccount == "" || !canReceive {
		defaultReceive, _ := defaultReceiveAccount(ctx, b, receiverWallet, sendCurrency)
		if defaultReceive != nil {
			p.ReceiverAccount = defaultReceive.ID
		}

		if defaultReceive != nil && p.ReceiverAmount.Currency.String() == "" {
			p.ReceiverAmount.Currency = defaultReceive.ReceiveCurrency
		}
	}

	requireOTP, err := requiresOTP(ctx, b, p.Type, senderWallet, receiverWallet)
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	publicID, err := NewSoftDescriptor(time.Now())
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	p, fxRate, fxFeePerc, err := applyFXCreate(ctx, b, p)
	if err != nil {
		return nil, err
	}

	err = validateWithdrawal(ctx, b, p.Type, p.SenderAccount, p.ReceiverAccount)
	if err != nil {
		return nil, err
	}

	err = validateDeposit(ctx, b, p.Type, p.SenderAccount, p.ReceiverAccount)
	if err != nil {
		return nil, err
	}

	err = validateSenderReceiver(ctx, b, p.Type, p.SenderAccount, p.ReceiverAccount)
	if err != nil {
		return nil, err
	}

	err = validateSendBalances(ctx, b, p.SenderAccount, p.SenderAmount)
	if err != nil {
		return nil, err
	}

	// TODO Calculate more actions required
	id := p.IdempotencyKey
	if id == "" {
		id = uuid.NewString()
	}
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
		Value("action_three_ds_required", false).
		Value("note", sql.NullString{String: p.Note, Valid: p.Note != ""}).
		Value("action_otp_required", requireOTP).
		Value("ip_address", sql.NullString{String: p.IPAddress, Valid: b.Validator().Var(p.IPAddress, "ip_addr") == nil}).
		Value("type", p.Type).
		Value("fx_rate", sql.NullFloat64{Float64: fxRate, Valid: fxRate > 0}).
		Value("fx_fee_percentage", sql.NullFloat64{Float64: fxFeePerc, Valid: fxFeePerc > 0}).
		Returning(cols).
		GetStatement()
	if err != nil {
		if db.IsErrorCode(err, db.UniqueViolationError) {
			return nil, fmt.Errorf("%w %s", payments.ErrIdempotencyViolation, err)
		}
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	var dbp dbPayment
	err = b.DB().GetContext(ctx, &dbp, stmt, args...)
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	return transformPayment(ctx, b, dbp, senderWallet, receiverWallet)
}

func validateSendBalances(ctx context.Context, b Backends, sendAccID string, amt currency.Amount) error {
	if sendAccID == "" || amt.IsEmpty() || sendAccID == rafiki.ZARBalanceAccount {
		return nil
	}

	sa, err := b.LinkedAccounts().Get(ctx, sendAccID)
	if err != nil {
		return err
	}

	// Only relevant for balance accounts
	if sa.Type != xago.AccTypeBalance && sa.Type != pti.AccTypeBalance {
		return nil
	}

	if sa.Provider == xago.ProviderName {
		bal, err := b.Xago().GetBalance(ctx, sa.ID)
		if err != nil {
			return err
		}

		if bal.Available.Value < amt.Value {
			return fmt.Errorf("%w insufficient balance", payments.ErrInsufficientFunds)
		}
	} else if sa.Provider == pti.ProviderName {
		w, err := b.PTI().GetBalance(ctx, sa.ID)
		if err != nil {
			return fmt.Errorf("%w failed to get pti wallet for balance validation", err)
		}

		if w.Available.Value < amt.Value {
			return fmt.Errorf("%w insufficient balance", payments.ErrInsufficientFunds)
		}
	}

	return nil
}

func validateDeposit(ctx context.Context, b Backends, typ payments.Type, senderAccID, receiverAccID string) error {
	if typ != payments.TypeDeposit {
		return nil
	}

	if senderAccID == "" || receiverAccID == "" {
		return payments.ErrInvalidDeposit
	}

	senderAcc, err := b.LinkedAccounts().Get(ctx, senderAccID)
	if err != nil {
		return fmt.Errorf("%w %s", payments.ErrInternal, err)
	}
	if !(senderAcc.Provider == pti.ProviderName && senderAcc.Type == pti.TypeCard) &&
		!(senderAcc.Provider == pti.ProviderName && senderAcc.Type == pti.TypeBank) {
		return payments.ErrInvalidDeposit
	}

	receiverAcc, err := b.LinkedAccounts().Get(ctx, receiverAccID)
	if err != nil {
		return fmt.Errorf("%w %s", payments.ErrInternal, err)
	}
	if receiverAcc.Provider != pti.ProviderName || receiverAcc.Type != pti.AccTypeBalance || senderAcc.WalletID != receiverAcc.WalletID {
		return payments.ErrInvalidDeposit
	}

	return nil
}

func validateWithdrawal(ctx context.Context, b Backends, typ payments.Type, senderAccID, receiverAccID string) error {
	if typ != payments.TypeWithdrawal {
		return nil
	}

	if senderAccID == "" || receiverAccID == "" {
		return fmt.Errorf("%w sender(%s) receiver(%s)", payments.ErrInvalidWithdrawal, senderAccID, receiverAccID)
	}

	senderAcc, err := b.LinkedAccounts().Get(ctx, senderAccID)
	if err != nil {
		return fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	// Only Xago, Chimoney and PTI (card) supports withdrawal
	if !(senderAcc.Provider == xago.ProviderName && senderAcc.Type == xago.AccTypeBalance) &&
		!(senderAcc.Provider == chimoney.ProviderName && senderAcc.Type == chimoney.AccTypeBalance) &&
		!(senderAcc.Provider == pti.ProviderName && senderAcc.Type == pti.AccTypeBalance) {
		return fmt.Errorf("%w sender provider(%s) type(%s)", payments.ErrInvalidWithdrawal, senderAcc.Provider, senderAcc.Type)
	}

	receiverAcc, err := b.LinkedAccounts().Get(ctx, receiverAccID)
	if err != nil {
		return fmt.Errorf("%w %s", payments.ErrInternal, err)
	}
	if (!(receiverAcc.Provider == xago.ProviderName && receiverAcc.Type == xago.AccTypeBank) &&
		!(receiverAcc.Provider == pti.ProviderName && receiverAcc.Type == pti.TypeBank) &&
		!(receiverAcc.Provider == chimoney.ProviderName && receiverAcc.Type == chimoney.AccTypeInterac)) ||
		senderAcc.WalletID != receiverAcc.WalletID {
		return fmt.Errorf("%w receiver provider(%s) type(%s)", payments.ErrInvalidWithdrawal, receiverAcc.Provider, receiverAcc.Type)
	}

	return nil
}

func isXagoGatehubPair(senderProvider, receiverProvider string) bool {
	isXagoGatehubPair := (senderProvider == gatehub.ProviderName && receiverProvider == xago.ProviderName) ||
		(senderProvider == xago.ProviderName && receiverProvider == gatehub.ProviderName)

	return isXagoGatehubPair
}

// Allow cross-provider transfers only for supported provider pair and only if the feature is enabled for both wallets
func validateCrossProviderSenderReceiver(ctx context.Context, b Backends, senderAcc, receiverAcc *linkedaccounts.LinkedAccount) error {
	// Supported provider pair only
	if !isXagoGatehubPair(senderAcc.Provider, receiverAcc.Provider) {
		return payments.ErrIncompatibleAccounts
	}

	// Check feature flag enabled for both wallets
	senderFeats, err := b.Features().Features(ctx, senderAcc.WalletID)
	if err != nil {
		return err
	}
	receiverFeats, err := b.Features().Features(ctx, receiverAcc.WalletID)
	if err != nil {
		return err
	}
	if !senderFeats.XagoGatehubPaymentsEnabled || !receiverFeats.XagoGatehubPaymentsEnabled {
		return payments.ErrIncompatibleAccounts
	}

	// Check account type
	if senderAcc.Provider == xago.ProviderName && senderAcc.Type != xago.AccTypeBalance {
		return payments.ErrIncompatibleAccounts
	}
	if receiverAcc.Provider == xago.ProviderName && receiverAcc.Type != xago.AccTypeBalance {
		return payments.ErrIncompatibleAccounts
	}

	return nil
}

func validateSenderReceiver(ctx context.Context, b Backends, typ payments.Type, senderAccID, receiverAccID string) error {
	if typ != payments.TypePeer2Peer || senderAccID == "" || receiverAccID == "" {
		return nil
	}

	senderAcc, err := b.LinkedAccounts().Get(ctx, senderAccID)
	if err != nil {
		return err
	}

	receiverAcc, err := b.LinkedAccounts().Get(ctx, receiverAccID)
	if err != nil {
		return err
	}

	if senderAcc.Provider != receiverAcc.Provider {
		return validateCrossProviderSenderReceiver(ctx, b, senderAcc, receiverAcc)
	}

	if senderAcc.Provider == xago.ProviderName && (senderAcc.Type != xago.AccTypeBalance || receiverAcc.Type != xago.AccTypeBalance) {
		return payments.ErrIncompatibleAccounts
	}

	return nil
}

func SellerRisk(sendTransactions int) float64 {
	const (
		k = float64(0.1198)
		A = float64(0.04)
		B = float64(0.01)
	)
	return A*math.Exp(-k*float64(sendTransactions)) + B
}

func applyFXCreate(ctx context.Context, b Backends, args payments.CreateArgs) (payments.CreateArgs, float64, float64, error) {
	if args.SenderAccount == "" || args.ReceiverAccount == "" || args.Type == payments.TypeWithdrawal {
		return args, 0, 0, nil
	}

	senderAcc, err := b.LinkedAccounts().Get(ctx, args.SenderAccount)
	if err != nil {
		return args, 0, 0, fmt.Errorf("%w %s", payments.ErrInternal, err)
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

	// TODO maybe extract this into a separate function, maybe reuse in both applyFXCreate and applyFXUpdate
	// Cross-provider GateHub <-> Xago FX
	if isXagoGatehubPair(senderAcc.Provider, receiverAcc.Provider) {
		var pair xago_external.ConvertCurrencyPairEnum
		var receiverCurrency currency.Currency
		if senderAcc.Provider == gatehub.ProviderName {
			pair = xago_external.EURtoZAR
			receiverCurrency = currency.ZAR
		} else {
			pair = xago_external.ZARtoEUR
			receiverCurrency = currency.EUR
		}

		estimate, err := b.Xago().EstimateConvertCurrency(ctx, pair, args.SenderAmount.Float64())
		if err != nil {
			return args, 0, 0, fmt.Errorf("%w failed to estimate FX rate: %s", payments.ErrInternal, err)
		}

		args.ReceiverAmount = currency.FromFloat64(estimate.ReceivedAmount, receiverCurrency)
		return args, estimate.EstimatedRate, 0, nil
	}

	if senderAcc.SendCurrency != currency.USD || receiverAcc.ReceiveCurrency == currency.USD {
		return args, 0, 0, fmt.Errorf("%w currently only from USD to other currencies are supported", payments.ErrInternal)
	}

	return args, 0, 0, fmt.Errorf("%w cross currency not supported", payments.ErrInternal)
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

	if (payment.ReceiverID == "" || payment.ReceiverIDType == payments.IdentityTypeUnknown) && payment.Type != payments.TypeWithdrawal {
		requiredActions = append(requiredActions, payments.RequiredActionTypeReceiverIdentifier)
	}

	if payment.SenderAmount < 1 || !currency.ParseCurrency(payment.SenderCurrency).Valid() {
		requiredActions = append(requiredActions, payments.RequiredActionTypeSenderAmount)
	}

	if payment.ReceiverAmount < 1 || !currency.ParseCurrency(payment.ReceiverCurrency).Valid() {
		requiredActions = append(requiredActions, payments.RequiredActionTypeReceiverAmount)
	}

	// if payment.OTPRequired && payment.OTP.String == "" {
	// 	requiredActions = append(requiredActions, payments.RequiredActionTypeOTP)
	// }

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

	fee := currency.FromFloat64(0, currency.USD)
	if dbp.Type == payments.TypeWithdrawal && la.Provider == chimoney.ProviderName && la.Type == chimoney.AccTypeBalance {
		fee, err = b.Chimoney().GetEstimatedFee(ctx, dbp.SenderAmount)
		if err != nil {
			return nil, nil, err
		}
	}

	// Do all precursor operations in single TX so we don't get inconsistent state.
	err = crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		err := setStateTX(ctx, tx, id, payments.StateConfirmed)
		if err != nil {
			return err
		}

		txType := transactions.TransactionTypeSent
		var title string
		if dbp.Type == payments.TypeWithdrawal {
			title = "Withdrawal"
			txType = transactions.TransactionTypeWithdrawal
		}
		if dbp.Type == payments.TypeWebMonetization {
			if receiverWallet != nil {
				title = receiverWallet.Name
			}
			txType = transactions.TransactionTypeWebMonetizationOutgoing
		}
		if dbp.Type == payments.TypeRafikiPeer2Peer && receiverWallet != nil {
			title = receiverWallet.Name
		}
		if dbp.Type == payments.TypeDeposit {
			title = "Deposit"
			txType = transactions.TransactionTypeDeposit
		}

		txID, err := b.Transactions().CreateTransactionTx(ctx, tx, transactions.CreateTransactionArgs{
			WalletID:                senderWallet.ID,
			ForeignID:               dbp.ID,
			ForeignType:             txType,
			Provider:                transactions.ProviderPaymentsEngine,
			State:                   transactions.StatePending,
			Note:                    dbp.Note,
			Source:                  senderWallet.AddressString(),
			Destination:             destination,
			Amount:                  dbp.SenderAmount,
			ProviderFee:             &fee,
			LinkedAccountTitle:      la.Title(),
			DestinationIdentity:     dbp.Receiver.Identifier,
			DestinationIdentityType: dbp.Receiver.Type.String(),
			Reference:               dbp.Note,
			Title:                   title,
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

	payment, err := Lookup(ctx, b, id)
	if err != nil {
		return nil, nil, err
	}

	if payment.Type == payments.TypeWithdrawal && la.Provider == chimoney.ProviderName && la.Type == chimoney.AccTypeBalance {
		err = b.Chimoney().ExecuteWithdraw(ctx, la.WalletID, payment.SendTransactionID)
		if err != nil {
			return nil, nil, err
		}

		return payment, nil, nil
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

	return update(ctx, b, args, payment)
}

// update performs minimal validation and updates a payment. This is only available internally to the payments engine
// where updates can be made to the payment regardless of the state of the payment.
func update(ctx context.Context, b Backends, args payments.UpdateArgs, payment *dbPayment) (*payments.Payment, error) {
	var err error
	if payment == nil {
		payment, err = getPayment(ctx, b, args.ID)
		if err != nil {
			return nil, err
		}
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

	var receiverAmtUpdated, senderAmtUpdated, accountsUpdated bool
	noop := true
	receiver := payments.Identity{Identifier: payment.ReceiverID, Type: payment.ReceiverIDType}
	if !args.Receiver.IsEmpty() && !args.Receiver.IsEqual(receiver) {
		payment.ReceiverID, receiver.Identifier = args.Receiver.Identifier, args.Receiver.Identifier
		payment.ReceiverIDType, receiver.Type = args.Receiver.Type, args.Receiver.Type
		noop = false
	}
	var wg sync.WaitGroup
	var pErr error
	var senderWallet, receiverWallet *wallets.Wallet

	wg.Add(1)
	go func() {
		defer wg.Done()
		var syncErr error
		senderWallet, syncErr = lookupWallet(ctx, b, payments.Identity{Identifier: payment.SenderID, Type: payment.SenderIDType})
		if err != nil {
			pErr = syncErr
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var syncErr error
		receiverWallet, syncErr = lookupWallet(ctx, b, receiver)
		if err != nil && !errors.Is(err, identities.ErrNotFound) {
			pErr = syncErr
		}
	}()

	wg.Wait()
	if pErr != nil {
		return nil, pErr
	}

	if args.ReceiverAccount != "" && args.ReceiverAccount != payment.ReceiverAccount.String {
		wg.Add(1)
		go func() {
			defer wg.Done()
			canReceive, recvCurrency, syncErr := accountCanReceive(ctx, b, receiverWallet, args.ReceiverAccount, payment.Type)
			if syncErr != nil {
				pErr = fmt.Errorf("%w %s", payments.ErrInternal, err)
				return
			}
			payment.ReceiverAccount.String = args.ReceiverAccount
			payment.ReceiverAccount.Valid = canReceive
			payment.ReceiverCurrency = recvCurrency.String()
			noop = false
			accountsUpdated = true
		}()
	}

	if args.SenderAccount != "" && args.SenderAccount != payment.SenderAccount.String {
		wg.Add(1)
		go func() {
			defer wg.Done()
			canSend, sendCurrency, err := accountCanSend(ctx, b, senderWallet, args.SenderAccount, payment.Type)
			if err != nil {
				pErr = fmt.Errorf("%w %s", payments.ErrInternal, err)
				return
			}
			payment.SenderAccount.String = args.SenderAccount
			payment.SenderAccount.Valid = canSend
			payment.SenderCurrency = sendCurrency.String()
			noop = false
			accountsUpdated = true
		}()
	}

	receiveAmount := currency.FromUInt64(int64(payment.ReceiverAmount), currency.Currency(payment.ReceiverCurrency))
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

	// Wait for parallel checks to complete
	wg.Wait()
	if pErr != nil {
		return nil, pErr
	}

	if noop {
		return transformPayment(ctx, b, *payment, senderWallet, receiverWallet)
	}

	if accountsUpdated {
		// look for default receive account if receive account is not explicitly being set
		if args.ReceiverAccount == "" {
			la, err := defaultReceiveAccount(ctx, b, receiverWallet, currency.ParseCurrency(payment.SenderCurrency))
			if err != nil {
				return nil, err
			}
			payment.ReceiverAccount.String = la.ID
			payment.ReceiverAccount.Valid = true
		}

		err := validateSenderReceiver(ctx, b, payment.Type, payment.SenderAccount.String, payment.ReceiverAccount.String)
		if err != nil {
			return nil, err
		}

		senderAmtUpdated = true
	}

	// Something changed, update the FX calculations
	if receiverAmtUpdated || senderAmtUpdated {
		payment, err = applyFXUpdate(ctx, b, payment, receiverAmtUpdated)
		if err != nil {
			return nil, err
		}

		err = validateSendBalances(ctx, b, payment.SenderAccount.String, currency.FromUInt64(payment.SenderAmount, currency.ParseCurrency(payment.SenderCurrency)))
		if err != nil {
			return nil, err
		}
	}

	err = validateWithdrawal(ctx, b, payment.Type, payment.SenderAccount.String, payment.ReceiverAccount.String)
	if err != nil {
		return nil, err
	}

	err = validateDeposit(ctx, b, payment.Type, payment.SenderAccount.String, payment.ReceiverAccount.String)
	if err != nil {
		return nil, err
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
		Value("action_three_ds_required", payment.ThreeDSRequired).
		Value("action_three_ds_id", payment.ThreeDSID).
		Value("ip_address", payment.IPAddress).
		Value("action_otp", payment.OTP).
		Value("fx_rate", payment.FXRate).
		Value("fx_fee_percentage", payment.FXFeePercentage).
		Value("revision", payment.Revision+1).
		Where("revision", payment.Revision).
		Returning(cols).GetStatement()
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	var ret dbPayment
	err = b.DB().GetContext(ctx, &ret, stmt, values...)
	if errors.Is(err, sql.ErrNoRows) {
		// Update lost the race condition for updating, just return the latest DB entry
		return Lookup(ctx, b, payment.ID)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	return transformPayment(ctx, b, ret, senderWallet, receiverWallet)
}

func applyFXUpdate(ctx context.Context, b Backends, existing *dbPayment, receiverAmtUpdated bool) (*dbPayment, error) {
	if !existing.SenderAccount.Valid || !existing.ReceiverAccount.Valid || existing.Type == payments.TypeWithdrawal {
		return existing, nil
	}

	senderAcc, err := b.LinkedAccounts().Get(ctx, existing.SenderAccount.String)
	if err != nil {
		return existing, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	receiverAcc, err := b.LinkedAccounts().Get(ctx, existing.ReceiverAccount.String)
	if err != nil {
		return existing, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	if receiverAcc.ReceiveCurrency == senderAcc.SendCurrency {
		// Equalize sender and receiver amounts until we add fees.
		if receiverAmtUpdated {
			existing.SenderAmount = existing.ReceiverAmount
			existing.SenderCurrency = senderAcc.ReceiveCurrency.String()
			return existing, nil
		}

		existing.ReceiverAmount = existing.SenderAmount
		existing.ReceiverCurrency = receiverAcc.ReceiveCurrency.String()
		return existing, nil
	}

	// TODO maybe extract this into a separate function, maybe reuse in both applyFXCreate and applyFXUpdate
	// Cross-provider GateHub <-> Xago FX
	if isXagoGatehubPair(senderAcc.Provider, receiverAcc.Provider) {
		var pair xago_external.ConvertCurrencyPairEnum
		var receiverCurrency currency.Currency
		if senderAcc.Provider == gatehub.ProviderName {
			pair = xago_external.EURtoZAR
			receiverCurrency = currency.ZAR
		} else {
			pair = xago_external.ZARtoEUR
			receiverCurrency = currency.EUR
		}

		senderAmount := currency.FromUInt64(existing.SenderAmount, currency.ParseCurrency(existing.SenderCurrency))
		estimate, err := b.Xago().EstimateConvertCurrency(ctx, pair, senderAmount.Float64())
		if err != nil {
			return existing, fmt.Errorf("%w failed to estimate FX rate: %s", payments.ErrInternal, err)
		}

		receiverAmt := currency.FromFloat64(estimate.ReceivedAmount, receiverCurrency)
		existing.ReceiverAmount = receiverAmt.Value
		existing.ReceiverCurrency = receiverCurrency.String()
		existing.FXRate = sql.NullFloat64{Float64: estimate.EstimatedRate, Valid: true}
		return existing, nil
	}

	if senderAcc.SendCurrency != currency.USD {
		return existing, fmt.Errorf("%w currently only from USD to other currencies are supported", payments.ErrInternal)
	}

	return existing, fmt.Errorf("%w cross currency currently not supported", payments.ErrInternal)
}

func UpdateReceiver(ctx context.Context, b Backends, id string, identity payments.Identity) error {
	_, err := update(ctx, b, payments.UpdateArgs{ID: id, Receiver: identity}, nil)
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
	err := b.DB().SelectContext(ctx, &refs, "SELECT payment_id, identifier, wallet_id, workflow_id, workflow_run_id FROM payments_workflow_refs WHERE identifier=$1 AND completed=false", identifier)
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
	err := b.DB().SelectContext(ctx, &refs, "SELECT payment_id, identifier, wallet_id, workflow_id, workflow_run_id FROM payments_workflow_refs WHERE wallet_id=$1 AND completed=false", walletID)
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

func SignalExternalPayoutComplete(ctx context.Context, b Backends, id string, success bool) error {
	dbp, err := getPayment(ctx, b, id)
	if err != nil {
		return err
	}

	// Nothing to signal if it's not a external payment
	if dbp.Type != payments.TypeRafiki2External {
		return nil
	}

	return b.Temporal().SignalWorkflow(ctx, fmt.Sprintf(payoutWorkflowFmt, id), "", signalChanName, PaySignal{
		PaymentID:     id,
		PayOutSuccess: success,
	})
}

func SignalGatehubTransferComplete(ctx context.Context, b Backends, externalTransactionID string) error {
	var paymentID string
	err := b.DB().GetContext(ctx, &paymentID, "SELECT payment_id FROM gatehub_transactions WHERE external_id=$1;", externalTransactionID)
	if errors.Is(err, sql.ErrNoRows) {
		log.Error("%w No payment found for gatehub transfer complete signal", zap.String("external_transaction_id", externalTransactionID))
		return nil
	} else if err != nil {
		return err
	}

	// TODO maybe ignore the serviceerror.NotFound error (workflow already completed)
	return b.Temporal().SignalWorkflow(ctx, fmt.Sprintf(payoutWorkflowFmt, paymentID), "", gatehubNotifyChanName, nil)

	// err = b.Temporal().SignalWorkflow(ctx, fmt.Sprintf(payoutWorkflowFmt, paymentID), "", gatehubNotifyChanName, nil)
	// if _, ok := err.(*serviceerror.NotFound); ok {
	// 	// Workflow already completed before the webhook arrived — signal is no longer needed.
	//  // TODO log that the error was swallowed and the reason
	// 	return nil
	// }
	// return err
}

func ListAwaitingSignal(ctx context.Context, b Backends) ([]payments.Payment, error) {
	var dbRes []dbPayment
	err := b.DB().SelectContext(ctx, &dbRes, fmt.Sprintf("SELECT %s FROM payments WHERE id IN (select payment_id::uuid from payments_workflow_refs where completed=false)", cols))
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	res := make([]payments.Payment, len(dbRes))
	for i, p := range dbRes {
		temp, err := transformPayment(ctx, b, p, nil, nil)
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
