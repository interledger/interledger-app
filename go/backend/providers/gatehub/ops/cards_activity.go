package ops

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/notify"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/gatehub/external"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/pacioli"
	"go.temporal.io/sdk/temporal"
)

func (a *Activity) CreateNewDeliveryAddress(ctx context.Context, userID, customerID string, args external.CreateCustomerDeliveryAddressArgs) (string, error) {
	id, err := a.external.CreateCustomerDeliveryAddress(ctx, userID, customerID, args)
	if err != nil {
		return "", fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return id, nil
}

func (a *Activity) CreateCard(ctx context.Context, userID, accountID string, args external.OrderCardArgs) (*external.Card, error) {
	card, err := a.external.OrderCard(ctx, userID, accountID, args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return card, nil
}

func (a *Activity) StoreCustomerIDs(ctx context.Context, userID, customerID, accountID string) (bool, error) {
	shouldOrderPlastic, err := updateCustomerIDsReturningPlasticFlag(ctx, a.b, userID, customerID, accountID)
	if err != nil {
		return false, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	if !shouldOrderPlastic.Valid {
		return false, fmt.Errorf("%w unknown card type - flag was not set when creating customer", gatehub.ErrInternal)
	}

	return shouldOrderPlastic.Bool, nil
}

func (a *Activity) CreatePlasticForCard(ctx context.Context, userID, cardID string) error {
	return a.external.CreatePlasticForCard(ctx, userID, cardID)
}

func (a *Activity) MarkFirstCardAsProcessed(ctx context.Context, userID string) error {
	err := updateFirstCardProcessTime(ctx, a.b, userID)
	if err != nil {
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return nil
}

func (a *Activity) IsCustomerCreated(ctx context.Context, userID string) (bool, error) {
	externalIDs, err := getExternalIDsByUserID(ctx, a.b, userID)
	if err != nil {
		return false, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return externalIDs.IsCustomerCreated(), nil
}

func (a *Activity) SendCardReadyEmail(ctx context.Context, walletID, cardID string) error {
	a.b.Email().SendCardCreatedEmail(ctx, walletID, cardID)
	return nil
}

func (a *Activity) Notify(ctx context.Context, walletID string, notificationType notify.NotificationType) error {
	a.b.Notify().NotifyWallet(ctx, walletID, notificationType)
	return nil
}

func (a *Activity) GetCardDetails(ctx context.Context, userID, cardID string) (*external.Card, error) {
	return a.external.GetCardDetails(ctx, userID, cardID)
}

func (a *Activity) GetCardTransaction(ctx context.Context, userID, txID string) (*external.CardTransaction, error) {
	return a.external.GetCardTransaction(ctx, userID, txID)
}

func (a *Activity) SaveGatehubCardTransaction(ctx context.Context, userID, cardID, cardMaskedPan string, tx external.CardTransaction) error {
	raw, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	maskedPan := formatMaskedPan(cardMaskedPan)

	_, err = a.b.DB().ExecContext(ctx,
		"INSERT INTO gatehub_card_transactions (id, user_id, card_id, card_masked_pan, gatehub_response_code, gatehub_response_description, type, status, mcc, transaction_amount, transaction_currency, raw_transaction) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) ON CONFLICT DO NOTHING;",
		tx.TransactionID,
		userID,
		cardID,
		maskedPan,
		tx.GHResponseCode,
		tx.GHResponseDescription,
		tx.Type,
		tx.TxStatus,
		tx.Mcc,
		tx.TransactionAmount,
		tx.TransactionCurrency,
		raw,
	)
	if err != nil {
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return nil
}

func (a *Activity) FinalizeGatehubCardTransaction(ctx context.Context, cardTxID, internalTxID string) error {
	err := FinaliseReserve(ctx, a.b, internalTxID)
	if err != nil {
		return err
	}

	// This can be removed - no more storing transfers in wallet_backend for card transactions
	transfers, err := a.b.Transactions().ListTransfers(ctx, internalTxID)
	if err != nil {
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	for _, t := range transfers {
		err = a.b.Transactions().SetTransferState(ctx, t.ID, transactions.StateCompleted)
		if err != nil {
			return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
		}
	}

	err = a.b.Transactions().SetTransactionState(ctx, internalTxID, transactions.StateCompleted)
	if err != nil {
		return err
	}

	err = updateCardTransactionStatus(ctx, a.b, cardTxID, external.CardTransactionStatusCompleted)
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) RollbackGatehubCardTransaction(ctx context.Context, cardTxID, internalTxID string) error {
	err := RollbackReserve(ctx, a.b, internalTxID)
	if err != nil {
		return err
	}

	err = a.b.Transactions().SetTransactionState(ctx, internalTxID, transactions.StateFailed)
	if err != nil {
		return err
	}

	err = updateCardTransactionStatus(ctx, a.b, cardTxID, external.CardTransactionStatusFailed)
	if err != nil {
		return err
	}

	return nil
}

type CardTransactionMeta struct {
	WalletID        string
	WalletAddress   string
	LinkedAccountID string
	MerchantName    string
	BillingAmount   currency.Amount
}

func (a *Activity) ComputeCardTransactionMeta(ctx context.Context, userID string, tx external.CardTransaction) (CardTransactionMeta, error) {
	if tx.BillingCurrency == nil || tx.BillingAmount == nil {
		// Replace with `new("EUR")` when moving to golang 1.26
		c := "EUR"
		amount := "0"
		tx.BillingCurrency = &c
		tx.BillingAmount = &amount
	}

	// Sanity check
	if *tx.BillingCurrency != currency.EUR.String() {
		return CardTransactionMeta{}, temporal.NewNonRetryableApplicationError("Invalid currency", "ErrInternal", fmt.Errorf("%w invalid currency", gatehub.ErrInternal))
	}

	walletID, err := getWalletID(ctx, a.b, userID)
	if err != nil {
		return CardTransactionMeta{}, err
	}

	wallet, err := a.b.Wallets().Get(ctx, walletID)
	if errors.Is(err, gatehub.ErrNotFound) {
		return CardTransactionMeta{}, temporal.NewNonRetryableApplicationError("Wallet not found", "ErrNotFound", fmt.Errorf("%w No wallet found for gatehub user", gatehub.ErrNotFound))
	}
	if err != nil {
		return CardTransactionMeta{}, err
	}

	las, err := a.b.LinkedAccounts().ListBalances(ctx, walletID)
	if err != nil {
		return CardTransactionMeta{}, err
	}

	var linkedAccount *linkedaccounts.LinkedAccount
	for _, la := range las {
		if la.Provider == gatehub.ProviderName && la.Type == gatehub.AccTypeBalance {
			linkedAccount = &la
			break
		}
	}

	if linkedAccount == nil {
		return CardTransactionMeta{}, temporal.NewNonRetryableApplicationError("Gatehub EUR balance account not found", "ErrInternal", fmt.Errorf("%w Gatehub EUR balance account not found", gatehub.ErrNotFound))
	}

	val, err := StringToScaledUInt(*tx.BillingAmount)
	if err != nil {
		return CardTransactionMeta{}, temporal.NewNonRetryableApplicationError("Invalid billing amount", "ErrInternal", fmt.Errorf("%w invalid billing amount: %s", gatehub.ErrInternal, *tx.BillingAmount))
	}

	return CardTransactionMeta{
		WalletID:        walletID,
		WalletAddress:   wallet.AddressString(),
		LinkedAccountID: linkedAccount.ID,
		MerchantName:    getMerchantName(tx),
		BillingAmount: currency.Amount{
			Value:    val,
			Currency: currency.EUR,
			Scale:    2,
		},
	}, nil
}

type RecordGatehubCardTxData struct {
	WalletID        string
	WalletAddress   string
	LinkedAccountID string
	MerchantName    string
	BillingAmount   currency.Amount
	Note            string
}
type RecordGatehubCardFXData struct {
	ExchangeRateApplied   string
	ExchangeRateReference string
	ExchangeRateSurcharge string
	TargetAmount          *currency.Amount
}

type RecordGatehubCardWithdrawalArgs struct {
	RecordGatehubCardTxData
	RecordGatehubCardFXData
}

func (a *Activity) RecordGatehubCardWithdrawal(ctx context.Context, txID string, tx external.CardTransaction, args RecordGatehubCardWithdrawalArgs) error {
	if args.BillingAmount.Value > 0 {
		if err := createCardTransactionLedgerTransfer(ctx, a.b.Pacioli(), createCardTransactionLedgerTransferArgs{
			txID:      txID,
			amount:    args.BillingAmount,
			debitID:   args.LinkedAccountID,
			creditID:  gatehub.EUROpsAccount,
			isPending: true,
		}); err != nil {
			return err
		}
	}

	createArgs := transactions.CreateTransactionArgs{
		ID:                    txID,
		WalletID:              args.WalletID,
		Provider:              gatehub.ProviderName,
		State:                 transactions.StatePending,
		ForeignID:             tx.TransactionID,
		ForeignType:           transactions.TransactionTypeCardTransaction,
		Source:                args.WalletAddress,
		LinkedAccountTitle:    "EUR Balance",
		Title:                 args.MerchantName,
		Note:                  args.Note,
		Reference:             args.Note,
		Amount:                args.BillingAmount,
		ExchangeRateApplied:   args.ExchangeRateApplied,
		ExchangeRateReference: args.ExchangeRateReference,
		ExchangeRateSurcharge: args.ExchangeRateSurcharge,
		TargetAmount:          args.TargetAmount,
	}

	if _, err := a.b.Transactions().CreateTransaction(ctx, createArgs); err != nil {
		return err
	}

	return nil
}

type RecordGatehubCardDepositArgs struct {
	RecordGatehubCardTxData
}

func (a *Activity) RecordGatehubCardDeposit(ctx context.Context, txID string, tx external.CardTransaction, args RecordGatehubCardDepositArgs) error {
	if args.BillingAmount.Value > 0 {
		if err := createCardTransactionLedgerTransfer(ctx, a.b.Pacioli(), createCardTransactionLedgerTransferArgs{
			txID:      txID,
			amount:    args.BillingAmount,
			debitID:   gatehub.EUROpsAccount,
			creditID:  args.LinkedAccountID,
			isPending: true,
		}); err != nil {
			return err
		}
	}

	createArgs := transactions.CreateTransactionArgs{
		ID:                 txID,
		WalletID:           args.WalletID,
		Provider:           gatehub.ProviderName,
		State:              transactions.StatePending,
		ForeignID:          tx.TransactionID,
		ForeignType:        transactions.TransactionTypeCardTransaction,
		Source:             args.WalletAddress,
		LinkedAccountTitle: "EUR Balance",
		Title:              args.MerchantName,
		Note:               args.Note,
		Reference:          args.Note,
		Amount:             args.BillingAmount,
	}

	if _, err := a.b.Transactions().CreateTransaction(ctx, createArgs); err != nil {
		return err
	}

	return nil
}

type RecordGatehubCardInformationalArgs struct {
	RecordGatehubCardTxData
	RecordGatehubCardFXData
	State transactions.State
}

func (a *Activity) RecordGatehubCardInformational(ctx context.Context, txID string, tx external.CardTransaction, args RecordGatehubCardInformationalArgs) error {
	createArgs := transactions.CreateTransactionArgs{
		ID:                    txID,
		WalletID:              args.WalletID,
		Provider:              gatehub.ProviderName,
		State:                 args.State,
		ForeignID:             tx.TransactionID,
		ForeignType:           transactions.TransactionTypeCardTransaction,
		Source:                args.WalletAddress,
		LinkedAccountTitle:    "EUR Balance",
		Title:                 args.MerchantName,
		Note:                  args.Note,
		Reference:             args.Note,
		Amount:                args.BillingAmount,
		ExchangeRateApplied:   args.ExchangeRateApplied,
		ExchangeRateReference: args.ExchangeRateReference,
		ExchangeRateSurcharge: args.ExchangeRateSurcharge,
		TargetAmount:          args.TargetAmount,
	}

	if _, err := a.b.Transactions().CreateTransaction(ctx, createArgs); err != nil {
		return err
	}

	switch args.State {
	case transactions.StateCompleted:
		return updateCardTransactionStatus(ctx, a.b, tx.TransactionID, external.CardTransactionStatusCompleted)
	case transactions.StateFailed:
		return updateCardTransactionStatus(ctx, a.b, tx.TransactionID, external.CardTransactionStatusFailed)
	}

	return nil
}

func (a *Activity) GetInternalTransactionByForeignID(ctx context.Context, walletID, foreignID string) (string, error) {
	tx, err := a.b.Transactions().GetTransactionByForeignID(ctx, walletID, foreignID)
	if err != nil {
		return "", fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	return tx.ID, nil
}

type SendCardTransactionFXEmailArgs struct {
	WalletID          string
	MaskedPAN         string
	MerchantName      string
	Date              string
	TransactionAmount currency.Amount
	BillingAmount     currency.Amount
	Surcharge         string
}

func (a *Activity) SendCardTransactionFXEmail(ctx context.Context, args SendCardTransactionFXEmailArgs) error {
	a.b.Email().SendCardTransactionFXEmail(ctx, args.WalletID, args.MaskedPAN, args.MerchantName, args.Date, args.Surcharge, args.TransactionAmount.Format(), args.BillingAmount.Format())
	return nil
}

type createCardTransactionLedgerTransferArgs struct {
	txID      string
	amount    currency.Amount
	debitID   string
	creditID  string
	isPending bool
}

func createCardTransactionLedgerTransfer(ctx context.Context, client pacioli.Client, args createCardTransactionLedgerTransferArgs) error {
	createArgs := []pacioli.CreateTransferArgs{{
		ID:              args.txID,
		Amount:          args.amount.Value,
		DebitAccountID:  args.debitID,
		CreditAccountID: args.creditID,
		Pending:         args.isPending,
		Code:            1,
		Ledger:          gatehub.LedgerIDEUR,
	}}
	if args.isPending {
		createArgs[0].Timeout = uint64(time.Hour * 24 * 365)
	}

	transfers, err := client.CreateTransfers(ctx, createArgs)
	if err != nil {
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	if len(transfers) > 0 && transfers[0].Code != 0 {
		switch transfers[0].Code {
		case pacioli.TransferExceedsCredits, pacioli.TransferExceedsDebits, pacioli.TransferExceedsPendingTransferAmount:
			return fmt.Errorf("%w insufficient balance (%s)", gatehub.ErrInsufficientBalance, transfers[0].Code.String())
		case pacioli.TransferExists:
			// ignore for idempotency
		default:
			return fmt.Errorf("%w non success code (%s)", gatehub.ErrInternal, transfers[0].Code.String())
		}
	}
	return nil
}

func buildFX(mcConversion *external.MastercardConversion, transactionCurrency *string, transactionAmount *string) (*RecordGatehubCardFXData, error) {
	if mcConversion == nil {
		return nil, fmt.Errorf("missing Mastercard conversion data")
	}
	if transactionCurrency == nil || transactionAmount == nil {
		return nil, fmt.Errorf("missing transaction amount or currency")
	}

	targetVal, err := StringToScaledUInt(*transactionAmount)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction amount: %s", *transactionAmount)
	}

	fx := &RecordGatehubCardFXData{
		TargetAmount: &currency.Amount{
			Value:    targetVal,
			Currency: currency.ParseCurrency(*transactionCurrency),
			Scale:    2,
		},
	}
	if mcConversion.ConvRate != nil {
		fx.ExchangeRateApplied = *mcConversion.ConvRate
	}
	if mcConversion.RefConfRate != nil {
		fx.ExchangeRateReference = *mcConversion.RefConfRate
	}
	if mcConversion.RefConfRateDiff != nil {
		fx.ExchangeRateSurcharge = *mcConversion.RefConfRateDiff
	}

	return fx, nil
}

func getNoteForWithdrawals(txType int) string {
	switch txType {
	case external.CardTransactionTypeATMWithdrawal:
		return "ATM Withdrawal"
	case external.CardTransactionTypeCashAdvance:
		return "Cash Advance"
	case external.CardTransactionTypeTransferFromAccount:
		return "Mastercard Send"
	case external.CardTransactionTypePreauthorization,
		external.CardTransactionTypePreauthorizationIncremental,
		external.CardTransactionTypePreauthorizationCompletion:
		return "Preauthorization"
	default:
		return ""
	}
}

func getNoteForDeposits(txType int, classification string) string {
	if classification == external.CardTransactionClassificationReversal {
		return "Refund"
	}

	switch txType {
	case external.CardTransactionTypeTransferToAccount:
		return "Mastercard Send"
	default:
		return ""
	}
}

func getNoteForInformationals(txType int, ghResponseCode string, responseCode string) string {
	if ghResponseCode == external.CardTransactionGHResponseCodeCRGUI {
		return "Insufficient Balance"
	}

	if txType == external.CardTransactionTypeCardVerificationInquiry &&
		ghResponseCode == external.CardTransactionGHResponseCodeOK &&
		responseCode == external.CardTransactionResponseCodeOK {
		return "Card Verification"
	}

	switch responseCode {
	case external.CardTransactionResponseCodeWCVV1,
		external.CardTransactionResponseCodeWCVV2:
		return "Invalid CVV"
	case external.CardTransactionResponseCodeWPIN:
		return "Invalid PIN"
	case external.CardTransactionResponseCodeCAEDM:
		return "Wrong Expiration Date"
	case external.CardTransactionResponseCodeCAEDI:
		return "Invalid Card Expiration Date"
	case external.CardTransactionResponseCodeCAEXP:
		return "Card Expired"
	case external.CardTransactionResponseCodeCASUS:
		return "Frozen Card"
	case external.CardTransactionResponseCodeCALOS:
		return "Lost Card"
	case external.CardTransactionResponseCodeCASTO:
		return "Stolen Card"
	case external.CardTransactionResponseCodeCAUSR:
		return "Card Suspended by User Request"
	default:
		return ""
	}
}

func getMerchantName(tx external.CardTransaction) string {
	if tx.MerchantName != nil && *tx.MerchantName != "" {
		merchant := *tx.MerchantName

		if tx.MerchantCountry != nil && *tx.MerchantCountry != "" {
			merchant = fmt.Sprintf("%s, %s", merchant, *tx.MerchantCountry)
		}

		return merchant
	}

	switch tx.Type {
	case external.CardTransactionTypePurchase:
		return "Purchase"
	case external.CardTransactionTypeATMWithdrawal:
		return "ATM Withdrawal"
	case external.CardTransactionTypePreauthorization,
		external.CardTransactionTypePreauthorizationIncremental,
		external.CardTransactionTypePreauthorizationCompletion:
		return "Preauthorization"
	case external.CardTransactionTypeCashAdvance:
		return "Cash Advance"
	case external.CardTransactionTypeTransferToAccount,
		external.CardTransactionTypeTransferFromAccount:
		return "Mastercard Send"
	case external.CardTransactionTypeCardVerificationInquiry:
		return "Card Verification"
	default:
		return "Card Transaction"
	}
}
