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

func (a *Activity) CreateGatehubCardTransaction(ctx context.Context, userID, txID string, tx external.CardTransaction) error {
	if tx.BillingCurrency == nil || tx.BillingAmount == nil {
		return temporal.NewNonRetryableApplicationError("Invalid billing currency or amount", "ErrInternal", fmt.Errorf("%w invalid currency or amount", gatehub.ErrInternal))
	}

	if *tx.BillingCurrency != currency.EUR.String() {
		return temporal.NewNonRetryableApplicationError("Invalid currency", "ErrInternal", fmt.Errorf("%w invalid currency", gatehub.ErrInternal))
	}

	merchantName := getMerchantName(tx)

	walletID, err := getWalletID(ctx, a.b, userID)
	if err != nil {
		return err
	}

	wallet, err := a.b.Wallets().Get(ctx, walletID)
	if errors.Is(err, gatehub.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError("Wallet not found", "ErrNotFound", fmt.Errorf("%w No wallet found for gatehub user", gatehub.ErrNotFound))
	}
	if err != nil {
		return err
	}

	las, err := a.b.LinkedAccounts().ListBalances(ctx, walletID)
	if err != nil {
		return err
	}

	var eurBalance *linkedaccounts.LinkedAccount
	for _, la := range las {
		if la.Provider == gatehub.ProviderName && la.Type == gatehub.AccTypeBalance {
			eurBalance = &la
			break
		}
	}

	if eurBalance == nil {
		return temporal.NewNonRetryableApplicationError("Gatehub EUR balance account not found", "ErrInternal", fmt.Errorf("%w Gatehub EUR balance account not found", gatehub.ErrInternal))
	}

	transactionArgs := transactions.CreateTransactionArgs{
		ID:                 txID,
		WalletID:           walletID,
		Provider:           gatehub.ProviderName,
		State:              transactions.StatePending,
		ForeignID:          tx.TransactionID,
		ForeignType:        transactions.TransactionTypeCardTransaction,
		Source:             wallet.AddressString(),
		LinkedAccountTitle: "EUR Balance",
		Title:              merchantName,
		Destination:        merchantName,
	}

	val, err := StringToScaledUInt(*tx.BillingAmount)
	if err != nil {
		return temporal.NewNonRetryableApplicationError("Invalid billing amount", "ErrInternal", fmt.Errorf("%w invalid billing amount: %s", gatehub.ErrInternal, *tx.BillingAmount))
	}

	amount := currency.Amount{
		Value:    val,
		Currency: currency.EUR,
		Scale:    2,
	}

	transactionArgs.Amount = amount
	transactionArgs.Transfers = []transactions.TransferArgs{
		{
			LinkedAccountID: eurBalance.ID,
			ForeignID:       tx.TransactionID,
			Amount:          amount,
			Type:            transactions.TransferTypeDebitCard,
			State:           transactions.StatePending,
		},
	}

	_, err = a.b.Transactions().CreateTransaction(ctx, transactionArgs)
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) FinalizeGatehubCardTransaction(ctx context.Context, cardTxID, internalTxID string) error {
	err := FinaliseReserve(ctx, a.b, internalTxID)
	if err != nil {
		return err
	}

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

	err = updateCardTransactionStatus(ctx, a.b, cardTxID, external.CardTractionStatusCompleted)
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

	transfers, err := a.b.Transactions().ListTransfers(ctx, internalTxID)
	if err != nil {
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	for _, t := range transfers {
		err = a.b.Transactions().SetTransferState(ctx, t.ID, transactions.StateFailed)
		if err != nil {
			return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
		}
	}

	err = a.b.Transactions().SetTransactionState(ctx, internalTxID, transactions.StateFailed)
	if err != nil {
		return err
	}

	err = updateCardTransactionStatus(ctx, a.b, cardTxID, external.CardTractionStatusFailed)
	if err != nil {
		return err
	}

	return nil
}

type CardTransactionMeta struct {
	WalletID      string
	WalletAddress string
	EURBalanceID  string
	MerchantName  string
	Note          string
	BillingAmount currency.Amount
}

func (a *Activity) ComputeCardTransactionMeta(ctx context.Context, userID string, tx external.CardTransaction) (CardTransactionMeta, error) {
	if tx.BillingCurrency == nil || tx.BillingAmount == nil {
		return CardTransactionMeta{}, temporal.NewNonRetryableApplicationError("Invalid billing currency or amount", "ErrInternal", fmt.Errorf("%w invalid currency or amount", gatehub.ErrInternal))
	}

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

	var eurBalance *linkedaccounts.LinkedAccount
	for _, la := range las {
		if la.Provider == gatehub.ProviderName && la.Type == gatehub.AccTypeBalance {
			eurBalance = &la
			break
		}
	}

	if eurBalance == nil {
		return CardTransactionMeta{}, temporal.NewNonRetryableApplicationError("Gatehub EUR balance account not found", "ErrInternal", fmt.Errorf("%w Gatehub EUR balance account not found", gatehub.ErrInternal))
	}

	val, err := StringToScaledUInt(*tx.BillingAmount)
	if err != nil {
		return CardTransactionMeta{}, temporal.NewNonRetryableApplicationError("Invalid billing amount", "ErrInternal", fmt.Errorf("%w invalid billing amount: %s", gatehub.ErrInternal, *tx.BillingAmount))
	}

	classification := ""
	if tx.TransactionClassification != nil {
		classification = *tx.TransactionClassification
	}

	return CardTransactionMeta{
		WalletID:      walletID,
		WalletAddress: wallet.AddressString(),
		EURBalanceID:  eurBalance.ID,
		MerchantName:  getMerchantName(tx),
		Note:          getNoteForCardTransaction(tx.Type, classification),
		BillingAmount: currency.Amount{
			Value:    val,
			Currency: currency.EUR,
			Scale:    2,
		},
	}, nil
}

type RecordGatehubCardAuthorizationArgs struct {
	WalletID              string
	WalletAddress         string
	EURBalanceID          string
	MerchantName          string
	Note                  string
	BillingAmount         currency.Amount
	ExchangeRateApplied   string
	ExchangeRateReference string
	ExchangeRateSurcharge string
	TargetAmount          *currency.Amount
}

func (a *Activity) RecordGatehubCardAuthorization(ctx context.Context, txID string, tx external.CardTransaction, args RecordGatehubCardAuthorizationArgs) error {
	if args.BillingAmount.Value > 0 {
		const cardAuthTimeout = 30 * 24 * time.Hour
		if _, err := ReserveBalance(ctx, a.b, args.EURBalanceID, txID, args.BillingAmount, cardAuthTimeout); err != nil {
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
		Destination:           args.MerchantName,
		Note:                  args.Note,
		Amount:                args.BillingAmount,
		ExchangeRateApplied:   args.ExchangeRateApplied,
		ExchangeRateReference: args.ExchangeRateReference,
		ExchangeRateSurcharge: args.ExchangeRateSurcharge,
		TargetAmount:          args.TargetAmount,
		Transfers: []transactions.TransferArgs{{
			LinkedAccountID: args.EURBalanceID,
			ForeignID:       tx.TransactionID,
			Amount:          args.BillingAmount,
			Type:            transactions.TransferTypeDebitCard,
			State:           transactions.StatePending,
		}},
	}

	_, err := a.b.Transactions().CreateTransaction(ctx, createArgs)
	return err
}

func (a *Activity) VoidPreviousGatehubCardAuthorization(ctx context.Context, walletID, refTransactionID string) error {
	prev, err := a.b.Transactions().GetTransactionByForeignID(ctx, walletID, refTransactionID)
	if err != nil {
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	if err = RollbackReserve(ctx, a.b, prev.ID); err != nil {
		return err
	}

	transfers, err := a.b.Transactions().ListTransfers(ctx, prev.ID)
	if err != nil {
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	for _, t := range transfers {
		if err = a.b.Transactions().SetTransferState(ctx, t.ID, transactions.StateFailed); err != nil {
			return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
		}
	}

	if err = a.b.Transactions().SetTransactionState(ctx, prev.ID, transactions.StateFailed); err != nil {
		return err
	}

	return updateCardTransactionStatus(ctx, a.b, refTransactionID, external.CardTractionStatusFailed)
}

type cardFX struct {
	ExchangeRateApplied   string
	ExchangeRateReference string
	ExchangeRateSurcharge string
	TargetAmount          *currency.Amount
}

func buildFX(mcConversion *external.MastercardConversion, transactionAmount *string, transactionCurrency string) (*cardFX, error) {
	if mcConversion == nil {
		return nil, fmt.Errorf("missing Mastercard conversion data")
	}
	if transactionAmount == nil {
		return nil, fmt.Errorf("missing transaction amount")
	}
	targetVal, err := StringToScaledUInt(*transactionAmount)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction amount: %s", *transactionAmount)
	}
	fx := &cardFX{
		TargetAmount: &currency.Amount{
			Value:    targetVal,
			Currency: currency.ParseCurrency(transactionCurrency),
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

func getNoteForCardTransaction(txType int, classification string) string {
	switch classification {
	case external.CardTransactionClassificationReversal:
		switch txType {
		case external.CardTransactionTypePurchase:
			return "Purchase reversed"
		case external.CardTransactionTypeATMWithdrawal:
			return "ATM withdrawal reversed"
		case external.CardTransactionTypeCashAdvance:
			return "Cash advance reversed"
		case external.CardTransactionTypeTransferToAccount,
			external.CardTransactionTypeTransferFromAccount:
			return "Card transfer reversed"
		case external.CardTransactionTypePreauthorization:
			return "Authorization hold reversed"
		case external.CardTransactionTypePreauthorizationIncremental:
			return "Authorization hold update reversed"
		case external.CardTransactionTypePreauthorizationCompletion:
			return "Payment completion reversed"
		default:
			return ""
		}
	case external.CardTransactionClassificationAuthorization:
		switch txType {
		case external.CardTransactionTypeATMWithdrawal:
			return "ATM cash withdrawal"
		case external.CardTransactionTypeCashAdvance:
			return "Cash advance"
		case external.CardTransactionTypeTransferToAccount,
			external.CardTransactionTypeTransferFromAccount:
			return "Card transfer"
		case external.CardTransactionTypePreauthorization:
			return "Authorization hold — funds reserved until payment is completed"
		case external.CardTransactionTypePreauthorizationIncremental:
			return "Authorization hold updated"
		case external.CardTransactionTypePreauthorizationCompletion:
			return "Payment completed"
		case external.CardTransactionTypeRefundCreditPayment:
			return "Refund"
		case external.CardTransactionTypeCardVerificationInquiry:
			return "Card verification check"
		case external.CardTransactionTypePINUnblock:
			return "PIN unblock"
		case external.CardTransactionTypePINChange:
			return "PIN change"
		default:
			return ""
		}
	default:
		return ""
	}
}

func getMerchantName(tx external.CardTransaction) string {
	if tx.MerchantName != nil && *tx.MerchantName != "" {
		return *tx.MerchantName
	}

	switch tx.Type {
	case external.CardTransactionTypePurchase,
		external.CardTransactionTypePreauthorization,
		external.CardTransactionTypePreauthorizationIncremental,
		external.CardTransactionTypePreauthorizationCompletion:
		return "Purchase"
	case external.CardTransactionTypeATMWithdrawal:
		return "ATM Withdrawal"
	case external.CardTransactionTypeBalanceInquiryOnATM:
		return "Balance Inquiry"
	case external.CardTransactionTypeCashAdvance:
		return "Cash Advance"
	case external.CardTransactionTypeRefundCreditPayment:
		return "Refund"
	case external.CardTransactionTypeTransferToAccount,
		external.CardTransactionTypeTransferFromAccount:
		return "Transfer"
	case external.CardTransactionTypeCardVerificationInquiry:
		return "Card Verification"
	case external.CardTransactionTypePINUnblock:
		return "PIN Unblock"
	case external.CardTransactionTypePINChange:
		return "PIN Change"
	default:
		return "Card Transaction"
	}
}
