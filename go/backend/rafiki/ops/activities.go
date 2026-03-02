package ops

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.temporal.io/sdk/temporal"

	"gitlab.com/fynbos/log"
	"go.uber.org/zap"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/gatehub"
	ghExternal "gitlab.com/fynbos/backend/providers/gatehub/external"
	"gitlab.com/fynbos/backend/rafiki"
	rafikiExternal "gitlab.com/fynbos/backend/rafiki/external"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/pacioli"
)

type Activity struct {
	b              ActivityBackends
	rafikiExternal rafikiExternal.Client
}

func NewActivity(b ActivityBackends, rafikiExt rafikiExternal.Client) *Activity {
	return &Activity{
		b:              b,
		rafikiExternal: rafikiExt,
	}
}

type ValidationResult struct {
	Valid  bool
	Reason string
}

type dbPayment struct {
	ID           string    `db:"id"`
	FromWalletID string    `db:"from_wallet"`
	ToWalletID   string    `db:"to_wallet"`
	Amount       uint64    `db:"amount"`
	Asset        string    `db:"amount_asset"`
	Timestamp    time.Time `db:"created_at"`
}

type RafikiPayment struct {
	IDs          pq.StringArray `db:"ids"`
	FromWalletID string         `db:"from_wallet"`
	ToWalletID   string         `db:"to_wallet"`
	Amount       uint64         `db:"amount"`
	Asset        string         `db:"amount_asset"`
}

func parseAmountValue(val string) (uint64, error) {
	return strconv.ParseUint(val, 10, 64)
}

func (a *Activity) ListPaymentsToMake(ctx context.Context) ([]RafikiPayment, error) {
	var dbPayments []RafikiPayment
	err := a.b.DB().SelectContext(ctx, &dbPayments, `SELECT ARRAY_AGG(id::text) AS ids, from_wallet, to_wallet, amount_asset, SUM(amount) AS amount
																FROM 
																	rafiki_outgoing_payments
																WHERE 
																	payment_id IS NULL
																GROUP BY 
																	from_wallet, 
																	to_wallet, 
																	amount_asset;`)
	if err != nil {
		return nil, err
	}
	return dbPayments, nil
}

func (a *Activity) CreateWebMonetizationPayment(ctx context.Context, payment RafikiPayment) (string, error) {
	senderBalances, err := a.b.LinkedAccounts().ListBalances(ctx, payment.FromWalletID)
	if err != nil {
		return "", err
	}

	var senderAcc *linkedaccounts.LinkedAccount
	for _, bal := range senderBalances {
		if bal.SendCurrency == currency.Currency(payment.Asset) {
			senderAcc = &bal
			break
		}
	}
	if senderAcc == nil {
		err = fmt.Errorf("%w failed to find sender account for currency=%s", rafiki.ErrNotFound, payment.Asset)
		return "", temporal.NewNonRetryableApplicationError("web monetization payment cron: no sending account found", "ErrInternal", err)
	}

	receiverAccs, err := a.b.LinkedAccounts().ListBalances(ctx, payment.ToWalletID)
	if err != nil {
		log.Error("failed to lookup balance accounts for receiver", zap.Error(err))
		return "", err
	}

	var receiverAcc *linkedaccounts.LinkedAccount
	for _, la := range receiverAccs {
		if currency.Currency(payment.Asset) == la.ReceiveCurrency {
			receiverAcc = &la
			break
		}
	}
	if receiverAcc == nil {
		err = fmt.Errorf("%w failed to find receiver account for currency=%s", rafiki.ErrNotFound, payment.Asset)
		return "", temporal.NewNonRetryableApplicationError("web monetization payment cron: no receiving account found", "ErrInternal", err)
	}

	p, err := a.b.Payments().Create(ctx, payments.CreateArgs{
		Sender: payments.Identity{
			Type:       payments.IdentityTypeWalletID,
			Identifier: payment.FromWalletID,
		},
		Receiver: payments.Identity{
			Type:       payments.IdentityTypeWalletID,
			Identifier: payment.ToWalletID,
		},
		SenderAccount:   senderAcc.ID,
		ReceiverAccount: receiverAcc.ID,
		Type:            payments.TypeWebMonetization,
		SenderAmount:    currency.FromUInt64(payment.Amount, currency.ParseCurrency(payment.Asset)),
		ReceiverAmount:  currency.FromUInt64(payment.Amount, currency.ParseCurrency(payment.Asset)),
		IPAddress:       "198.0.0.2", // TODO: Add a our static IP Address
	})

	if err != nil {
		return "", err
	}

	return p.ID, nil
}

func (a *Activity) ConfirmPayment(ctx context.Context, id string) error {
	_, ra, err := a.b.Payments().Confirm(ctx, id)
	if errors.Is(err, payments.ErrRequiredActions) {
		log.Error("required actions outstanding for web monitization payout", zap.String("actions", fmt.Sprintf("%+v", ra)))
	}
	return err
}

func (a *Activity) AddWebMonetizationPayment(ctx context.Context, payout RafikiPayment, paymentID string) error {
	query, args, err := sqlx.In("UPDATE rafiki_outgoing_payments SET payment_id=?,completed=? WHERE id in (?)", paymentID, true, []string(payout.IDs))
	if err != nil {
		return err
	}
	_, err = a.b.DB().ExecContext(ctx, a.b.DB().Rebind(query), args...)
	return err
}

func (a *Activity) getGatehubLinkedAccount(ctx context.Context, walletAddressID string) (*linkedaccounts.LinkedAccount, string, error) {
	walletID, err := lookupWalletIDFromActivity(ctx, a.b, walletAddressID)
	if err != nil {
		return nil, "", err
	}

	accs, err := a.b.LinkedAccounts().ListBalances(ctx, walletID)
	if err != nil {
		return nil, walletID, fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	for _, la := range accs {
		if la.Provider == gatehub.ProviderName && la.Type == gatehub.AccTypeBalance {
			return &la, walletID, nil
		}
	}

	return nil, walletID, fmt.Errorf("%w no gatehub balance account found for wallet %s", rafiki.ErrNotFound, walletID)
}

func (a *Activity) getGatehubExternalUserID(ctx context.Context, walletID string) (string, error) {
	var externalID string
	err := a.b.DB().GetContext(ctx, &externalID, "SELECT external_id FROM gatehub_users WHERE wallet_id=$1;", walletID)
	if err != nil {
		return "", fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}
	return externalID, nil
}

func lookupWalletIDFromActivity(ctx context.Context, b ActivityBackends, paymentPointerID string) (string, error) {
	var wid string
	err := b.DB().GetContext(ctx, &wid, "SELECT wallet_id FROM rafiki_payment_pointers WHERE payment_pointer_id=$1", paymentPointerID)
	if err != nil {
		return "", fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}
	return wid, nil
}

// Creates a GateHub hosted transfer from the system intermediary
// account to the user's GateHub wallet. Returns the GateHub transaction ID.
func (a *Activity) TransferFromIntermediaryToUser(ctx context.Context, walletAddressID string, amt amount) (string, error) {
	la, walletID, err := a.getGatehubLinkedAccount(ctx, walletAddressID)
	if err != nil {
		return "", temporal.NewNonRetryableApplicationError("failed to get gatehub linked account", "ErrNotFound", err)
	}

	parsedAmt, err := parseAmountValue(amt.Value)
	if err != nil {
		return "", temporal.NewNonRetryableApplicationError("invalid amount value", "ErrInternal", err)
	}

	sendingAddress := os.Getenv("GATEHUB_MANAGED_USER_WALLET")
	if sendingAddress == "" {
		return "", temporal.NewNonRetryableApplicationError("intermediary gatehub credentials not configured", "ErrInternal",
			fmt.Errorf("missing GATEHUB_MANAGED_USER_WALLET"))
	}

	cc := currency.ParseCurrency(amt.AssetCode)
	currencyAmt := currency.FromUInt64(parsedAmt, cc)
	floatAmt := currencyAmt.Float64()

	wallet, err := a.b.Wallets().Get(ctx, walletID)
	if err != nil {
		return "", fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	ghClient := a.b.Gatehub().ExternalClient()
	tx, err := ghClient.CreateTransaction(ctx, ghExternal.CreateTransactionRequest{
		SendingAddress:   sendingAddress,
		ReceivingAddress: la.ProviderID,
		Amount:           floatAmt,
		Message:          fmt.Sprintf("Rafiki incoming payment to %s", wallet.Name),
		Type:             ghExternal.TransactionTypeHosted,
		VaultID:          ghClient.GetVaultID(),
	})
	if err != nil {
		return "", fmt.Errorf("gatehub transfer from intermediary to user failed: %w", err)
	}

	return tx.ID, nil
}

// Creates a GateHub hosted transfer from the user's
// GateHub wallet to the system intermediary account. Returns the GateHub transaction ID.
func (a *Activity) TransferFromUserToIntermediary(ctx context.Context, walletAddressID string, amt amount) (string, error) {
	la, walletID, err := a.getGatehubLinkedAccount(ctx, walletAddressID)
	if err != nil {
		return "", temporal.NewNonRetryableApplicationError("failed to get gatehub linked account", "ErrNotFound", err)
	}

	parsedAmt, err := parseAmountValue(amt.Value)
	if err != nil {
		return "", temporal.NewNonRetryableApplicationError("invalid amount value", "ErrInternal", err)
	}

	sendingAddress := os.Getenv("GATEHUB_MANAGED_USER_WALLET")
	if sendingAddress == "" {
		return "", temporal.NewNonRetryableApplicationError("intermediary gatehub credentials not configured", "ErrInternal",
			fmt.Errorf("missing GATEHUB_MANAGED_USER_WALLET"))
	}

	externalUserID, err := a.getGatehubExternalUserID(ctx, walletID)
	if err != nil {
		return "", temporal.NewNonRetryableApplicationError("failed to get gatehub external user", "ErrNotFound", err)
	}

	cc := currency.ParseCurrency(amt.AssetCode)
	currencyAmt := currency.FromUInt64(parsedAmt, cc)
	floatAmt := currencyAmt.Float64()

	ghClient := a.b.Gatehub().ExternalClient()
	tx, err := ghClient.CreateTransaction(ctx, ghExternal.CreateTransactionRequest{
		SendingUserID:    externalUserID,
		SendingAddress:   la.ProviderID,
		ReceivingAddress: sendingAddress,
		Amount:           floatAmt,
		Message:          "Rafiki outgoing payment",
		Type:             ghExternal.TransactionTypeHosted,
		VaultID:          ghClient.GetVaultID(),
	})
	if err != nil {
		return "", fmt.Errorf("gatehub transfer from user to intermediary failed: %w", err)
	}

	return tx.ID, nil
}

// Stores the mapping between a GateHub transaction ID
// and a Temporal workflow ID so the GateHub webhook can signal the correct workflow.
func (a *Activity) StoreGatehubTransferMapping(ctx context.Context, gatehubTxID, workflowID string) error {
	_, err := a.b.DB().ExecContext(ctx,
		`INSERT INTO rafiki_gatehub_transfers (gatehub_tx_id, workflow_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		gatehubTxID, workflowID)
	return err
}

// Checks KYC status and validates that the outgoing payment can proceed.
// For local receivers it checks that sender and receiver have the same currency.
func (a *Activity) ValidateOutgoingPayment(ctx context.Context, op outgoingPaymentData) (*ValidationResult, error) {
	senderWalletID, err := lookupWalletIDFromActivity(ctx, a.b, op.WalletAddressID)
	if err != nil {
		return nil, err
	}

	approved, err := a.b.KYC().IsKYCApproved(ctx, senderWalletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}
	if !approved {
		return &ValidationResult{Valid: false, Reason: "user KYC not approved"}, nil
	}

	return a.validateLocalReceiver(ctx, op, senderWalletID)
}

func (a *Activity) validateLocalReceiver(ctx context.Context, op outgoingPaymentData, senderWalletID string) (*ValidationResult, error) {
	receiverID := extractIncomingPaymentID(op.Receiver)

	receiverIP, err := a.rafikiExternal.GetIncomingPayment(ctx, receiverID)
	if err != nil {
		return &ValidationResult{Valid: false, Reason: "could not resolve receiver"}, nil
	}

	var receiverWalletID string
	err = a.b.DB().GetContext(ctx, &receiverWalletID,
		"SELECT wallet_id FROM rafiki_payment_pointers WHERE payment_pointer_id=$1", receiverIP.WalletAddressId)
	if err != nil {
		return &ValidationResult{Valid: true}, nil
	}

	if senderWalletID == receiverWalletID {
		return &ValidationResult{Valid: false, Reason: "sending wallet cannot be the same as receiving wallet"}, nil
	}

	return a.validateCurrencyMatch(ctx, op, senderWalletID, receiverWalletID)
}

func extractIncomingPaymentID(receiverURL string) string {
	const urlPart = "incoming-payments"
	if strings.Contains(receiverURL, urlPart) {
		parts := strings.Split(receiverURL, "/")
		return parts[len(parts)-1]
	}
	return receiverURL
}

func (a *Activity) validateCurrencyMatch(ctx context.Context, op outgoingPaymentData, senderWalletID, receiverWalletID string) (*ValidationResult, error) {
	senderAccs, err := a.b.LinkedAccounts().ListBalances(ctx, senderWalletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}
	receiverAccs, err := a.b.LinkedAccounts().ListBalances(ctx, receiverWalletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	senderCurrency := findProviderCurrency(senderAccs, op.DebitAmount.AssetCode, true)
	receiverCurrency := findProviderCurrency(receiverAccs, op.DebitAmount.AssetCode, false)

	if senderCurrency != "" && receiverCurrency != "" && senderCurrency != receiverCurrency {
		return &ValidationResult{Valid: false, Reason: "cross-jurisdiction transfers not supported for local receivers"}, nil
	}

	return &ValidationResult{Valid: true}, nil
}

func findProviderCurrency(accs []linkedaccounts.LinkedAccount, assetCode string, isSender bool) string {
	for _, la := range accs {
		if la.Provider != gatehub.ProviderName {
			continue
		}
		if isSender && la.SendCurrency.String() == assetCode {
			return la.SendCurrency.String()
		}
		if !isSender && la.ReceiveCurrency.String() == assetCode {
			return la.ReceiveCurrency.String()
		}
	}
	return ""
}

func (a *Activity) CancelOutgoingPayment(ctx context.Context, paymentID, reason string) error {
	return a.rafikiExternal.CancelOutgoingPayment(ctx, paymentID, reason)
}

func (a *Activity) CreateIncomingPaymentTransaction(ctx context.Context, ip incomingPaymentData) error {
	walletID, err := lookupWalletIDFromActivity(ctx, a.b, ip.WalletAddressID)
	if err != nil {
		return err
	}

	la, _, err := a.getGatehubLinkedAccount(ctx, ip.WalletAddressID)
	if err != nil {
		return err
	}

	receivedAmt, err := parseAmountValue(ip.ReceivedAmount.Value)
	if err != nil {
		return temporal.NewNonRetryableApplicationError("invalid received amount", "ErrInternal", err)
	}

	cc := currency.ParseCurrency(ip.ReceivedAmount.AssetCode)
	amt := currency.FromUInt64(receivedAmt, cc)

	wallet, err := a.b.Wallets().Get(ctx, walletID)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	_, err = a.b.Transactions().CreateTransaction(ctx, transactions.CreateTransactionArgs{
		WalletID:                walletID,
		ForeignID:               ip.ID,
		ForeignType:             transactions.TransactionTypeOpenPaymentIncoming,
		Provider:                gatehub.ProviderName,
		State:                   transactions.StateCompleted,
		Source:                  wallet.AddressString(),
		Destination:             wallet.AddressString(),
		Title:                   "Incoming Payment",
		DestinationIdentity:     walletID,
		DestinationIdentityType: payments.IdentityTypeWalletID.String(),
		Amount:                  amt,
		LinkedAccountTitle:      "EUR Balance",
		Transfers: []transactions.TransferArgs{
			{
				LinkedAccountID: la.ID,
				ForeignID:       ip.ID,
				Amount:          amt,
				Type:            transactions.TransferTypeCreditBalance,
				State:           transactions.StateCompleted,
			},
		},
	})
	return err
}

// Creates a non-pending Pacioli ledger transfer (debit ops account, credit user account) and posts it immediately.
func (a *Activity) CreateAndPostLedgerTransferForIncoming(ctx context.Context, ip incomingPaymentData) error {
	la, _, err := a.getGatehubLinkedAccount(ctx, ip.WalletAddressID)
	if err != nil {
		return err
	}

	receivedAmt, err := parseAmountValue(ip.ReceivedAmount.Value)
	if err != nil {
		return temporal.NewNonRetryableApplicationError("invalid received amount", "ErrInternal", err)
	}

	opsAcc := gatehub.EUROpsAccount
	ledger := gatehub.LedgerIDEUR

	results, err := a.b.Pacioli().CreateTransfers(ctx, []pacioli.CreateTransferArgs{
		{
			ID:              ip.ID,
			Amount:          receivedAmt,
			CreditAccountID: la.ID,
			DebitAccountID:  opsAcc,
			Pending:         false,
			Code:            1,
			Ledger:          ledger,
		},
	})
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}
	if len(results) > 0 && results[0].Code != 0 {
		return fmt.Errorf("pacioli transfer failed with code %s", results[0].Code.String())
	}

	return nil
}

func (a *Activity) WithdrawIncomingPaymentLiquidity(ctx context.Context, incomingPaymentID string) error {
	err := a.rafikiExternal.WithdrawIncomingPaymentLiquidity(ctx, incomingPaymentID, 0)
	if err != nil {
		log.Error("failed to withdraw incoming payment liquidity",
			zap.String("incomingPaymentId", incomingPaymentID),
			zap.Error(err))
	}
	return nil
}

func (a *Activity) CreateOutgoingPaymentTransaction(ctx context.Context, op outgoingPaymentData) error {
	walletID, err := lookupWalletIDFromActivity(ctx, a.b, op.WalletAddressID)
	if err != nil {
		return err
	}

	la, _, err := a.getGatehubLinkedAccount(ctx, op.WalletAddressID)
	if err != nil {
		return err
	}

	debitAmt, err := parseAmountValue(op.DebitAmount.Value)
	if err != nil {
		return temporal.NewNonRetryableApplicationError("invalid debit amount", "ErrInternal", err)
	}

	cc := currency.ParseCurrency(op.DebitAmount.AssetCode)
	amt := currency.FromUInt64(debitAmt, cc)

	wallet, err := a.b.Wallets().Get(ctx, walletID)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	_, err = a.b.Transactions().CreateTransaction(ctx, transactions.CreateTransactionArgs{
		WalletID:                walletID,
		ForeignID:               op.ID,
		ForeignType:             transactions.TransactionTypeOpenOutgoingPayment,
		Provider:                gatehub.ProviderName,
		State:                   transactions.StatePending,
		Source:                  wallet.AddressString(),
		Destination:             op.Receiver,
		Title:                   "Outgoing Payment",
		DestinationIdentity:     op.Receiver,
		DestinationIdentityType: payments.IdentityTypeExternalWalletURL.String(),
		Amount:                  amt,
		LinkedAccountTitle:      "EUR Balance",
		Transfers: []transactions.TransferArgs{
			{
				LinkedAccountID: la.ID,
				ForeignID:       op.ID,
				Amount:          amt,
				Type:            transactions.TransferTypeDebitBalance,
				State:           transactions.StatePending,
			},
		},
	})
	return err
}

// Creates a pending Pacioli ledger transfer (debit user account, credit ops account) to reserve the outgoing payment amount.
func (a *Activity) ReserveBalanceForOutgoing(ctx context.Context, op outgoingPaymentData) error {
	la, _, err := a.getGatehubLinkedAccount(ctx, op.WalletAddressID)
	if err != nil {
		return err
	}

	debitAmt, err := parseAmountValue(op.DebitAmount.Value)
	if err != nil {
		return temporal.NewNonRetryableApplicationError("invalid debit amount", "ErrInternal", err)
	}

	opsAcc := gatehub.EUROpsAccount
	ledger := gatehub.LedgerIDEUR
	timeout := time.Hour * 24 * 365

	results, err := a.b.Pacioli().CreateTransfers(ctx, []pacioli.CreateTransferArgs{
		{
			ID:              op.ID,
			Amount:          debitAmt,
			DebitAccountID:  la.ID,
			CreditAccountID: opsAcc,
			Pending:         true,
			Code:            1,
			Timeout:         uint64(timeout),
			Ledger:          ledger,
		},
	})
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}
	if len(results) > 0 && results[0].Code != 0 {
		return fmt.Errorf("pacioli reserve failed with code %s", results[0].Code.String())
	}

	return nil
}

func (a *Activity) DepositOutgoingPaymentLiquidity(ctx context.Context, outgoingPaymentID string) error {
	err := a.rafikiExternal.FundOutgoingPayment(ctx, outgoingPaymentID)
	if err != nil {
		if strings.Contains(err.Error(), "wrong state") {
			log.Info("rafiki outgoing payment already funded", zap.String("paymentId", outgoingPaymentID))
			return nil
		}
		return err
	}
	return nil
}

func (a *Activity) UpdateOutgoingPaymentTransactionState(ctx context.Context, op outgoingPaymentData, state string) error {
	walletID, err := lookupWalletIDFromActivity(ctx, a.b, op.WalletAddressID)
	if err != nil {
		return err
	}

	txState := transactions.State(state)

	trx, err := a.b.Transactions().GetTransactionByForeignID(ctx, walletID, op.ID)
	if err != nil {
		return fmt.Errorf("transaction not found for outgoing payment %s: %w", op.ID, err)
	}

	transfers, err := a.b.Transactions().ListTransfers(ctx, trx.ID)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}
	for _, t := range transfers {
		if setErr := a.b.Transactions().SetTransferState(ctx, t.ID, txState); setErr != nil {
			return fmt.Errorf("%w %s", rafiki.ErrInternal, setErr)
		}
	}

	return a.b.Transactions().SetTransactionState(ctx, trx.ID, txState)
}

// Posts (finalizes) the pending Pacioli ledger transfer that was created during outgoing_payment.created.
func (a *Activity) PostLedgerTransferForOutgoing(ctx context.Context, op outgoingPaymentData) error {
	results, err := a.b.Pacioli().PostTransfers(ctx, []string{op.ID})
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}
	if len(results) > 0 &&
		results[0].Code != 0 &&
		results[0].Code != pacioli.TransferPendingTransferAlreadyPosted {
		return fmt.Errorf("pacioli post transfer failed with code %s", results[0].Code.String())
	}
	return nil
}

// Voids the pending Pacioli ledger transfer that was created during outgoing_payment.created (used on failure).
func (a *Activity) VoidLedgerTransferForOutgoing(ctx context.Context, op outgoingPaymentData) error {
	results, err := a.b.Pacioli().VoidTransfers(ctx, []string{op.ID})
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}
	if len(results) > 0 &&
		results[0].Code != 0 &&
		results[0].Code != pacioli.TransferPendingTransferAlreadyVoided &&
		results[0].Code != pacioli.TransferPendingTransferNotFound &&
		results[0].Code != pacioli.TransferPendingTransferExpired {
		return fmt.Errorf("pacioli void transfer failed with code %s", results[0].Code.String())
	}
	return nil
}

func (a *Activity) WithdrawOutgoingPaymentLiquidity(ctx context.Context, outgoingPaymentID string) error {
	err := a.rafikiExternal.WithdrawOutgoingPaymentLiquidity(ctx, outgoingPaymentID, 0)
	if err != nil {
		log.Error("failed to withdraw outgoing payment liquidity",
			zap.String("paymentId", outgoingPaymentID),
			zap.Error(err))
	}
	return nil
}
