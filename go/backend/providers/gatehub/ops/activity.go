package ops

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/notify"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/gatehub/external"
	httplogger "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/backend/slack"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/log"
	"gitlab.com/fynbos/pacioli"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.uber.org/zap"
)

type Activity struct {
	b        Backends
	external external.Client
	config   gatehub.Config
}

func NewActivity(b Backends, cfg gatehub.Config) *Activity {
	// Validate vault ID is available for activities
	if cfg.PaywiserEuroVaultID == "" {
		log.Warn("PaywiserEuroVaultID is not set in Gatehub configuration")
	}

	ec := external.NewClient(
		cfg.AppID,
		cfg.Secret,
		cfg.CardAppID,
		cfg.GatewayID,
		cfg.CardAccountProductCode,
		cfg.PaywiserEuroVaultID,
		cfg.OnOffRampClientID,
		cfg.OnboardingClientID,
		cfg.ExchangeClientID,
		cfg.APIBaseURL,
		cfg.OnboardingBaseURL,
		cfg.OnOffRampBaseURL,
		cfg.OrganizationID,
		&http.Client{
			Transport: otelhttp.NewTransport(
				httplogger.NewTransport(http.DefaultTransport, b, nil),
			),
		},
	)

	return &Activity{
		b:        b,
		external: ec,
		config:   cfg,
	}
}

func (a *Activity) CreateGatehubUser(ctx context.Context, walletID string) (string, error) {
	ul, err := a.b.Users().ListUsers(ctx, walletID)
	if err != nil {
		return "", fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	if len(ul) < 1 {
		return "", fmt.Errorf("%w No Interledger user found for walletID", gatehub.ErrInternal)
	}

	resp, err := a.external.CreateUser(ctx, ul[0].Email)
	if err != nil {
		return "", fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return resp.ID, nil
}

func (a *Activity) SaveGatehubUser(ctx context.Context, walletID, externalID string) error {
	_, err := a.b.DB().ExecContext(ctx, "INSERT INTO gatehub_users (external_id, wallet_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;", externalID, walletID)
	if err != nil {
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return nil
}

func (a *Activity) GetGatehubUser(ctx context.Context, walletID string) (string, error) {
	var externalID string
	err := a.b.DB().GetContext(ctx, &externalID, "SELECT external_id FROM gatehub_users WHERE wallet_id=$1;", walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", temporal.NewNonRetryableApplicationError("Gatehub user not found for wallet", "ErrNotFound", err)
	} else if err != nil {
		return "", err
	}

	return externalID, nil
}

func (a *Activity) CreateGatehubWalletLinkedAccount(ctx context.Context, walletID string) (*linkedaccounts.LinkedAccount, error) {
	w, err := a.b.Wallets().Get(ctx, walletID)
	if errors.Is(err, wallets.ErrNoWalletFound) {
		return nil, temporal.NewNonRetryableApplicationError("Wallet not found", "ErrNotFound", err)
	} else if err != nil {
		return nil, err
	}

	externalID, err := getExternalUserID(ctx, a.b, w.ID)
	if errors.Is(err, gatehub.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError("Gatehub user not found for wallet", "ErrNotFound", err)
	} else if err != nil {
		return nil, err
	}

	userWallets, err := a.external.GetUserWallets(ctx, externalID)
	if err != nil {
		return nil, err
	}

	var primaryWallet *external.Wallet
	for _, w := range userWallets.Wallets {
		if w.Primary {
			primaryWallet = &w
			break
		}
	}
	if primaryWallet == nil {
		return nil, fmt.Errorf("%w Could not find a primary wallet for gatehub user", gatehub.ErrInternal)
	}

	las, err := a.b.LinkedAccounts().ListBalances(ctx, w.ID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	for _, la := range las {
		if la.Provider == gatehub.ProviderName && la.Type == gatehub.AccTypeBalance {
			return &la, nil
		}
	}

	la, err := a.b.LinkedAccounts().Create(ctx, &linkedaccounts.CreateArgs{
		WalletID:        w.ID,
		Type:            gatehub.AccTypeBalance,
		Provider:        gatehub.ProviderName,
		ProviderID:      primaryWallet.Address,
		Name:            "EUR Balance",
		Nickname:        "EUR Balance",
		CanReceive:      true,
		ReceiveCountry:  w.Country,
		ReceiveCurrency: currency.EUR,
		SendCountry:     w.Country,
		SendCurrency:    currency.EUR,
		CanSend:         true,
		State:           linkedaccounts.Verified,
	})
	if err != nil {
		return nil, err
	}

	return la, nil
}

func (a *Activity) CreateGatehubBalanceAccount(ctx context.Context, id string) error {
	accs, err := a.b.Pacioli().ConfigureAccounts(ctx, []pacioli.ConfigureAccountArgs{
		{
			ID:                         id,
			LedgerID:                   gatehub.LedgerIDEUR,
			Code:                       1,
			DebitsMustNotExceedCredits: true,
			CreditsMustNotExceedDebits: false,
		},
	})
	if err != nil {
		return err
	}

	if len(accs) == 0 {
		// No error codes to speak of
		return nil
	}

	if accs[0].Code != pacioli.AccountOK && accs[0].Code != pacioli.AccountExists {
		return fmt.Errorf("%w failed to setup account status(%s)", gatehub.ErrInternal, accs[0].Code)
	}

	return nil
}

func (a *Activity) CreateGatehubDepositTransaction(ctx context.Context, transactionID, walletID string, amount, providerFee currency.Amount) (string, error) {
	existingDeposit, err := a.b.Transactions().GetTransactionByForeignID(ctx, walletID, transactionID)
	if err != nil && !errors.Is(err, transactions.ErrNotFound) {
		return "", err
	}
	if existingDeposit != nil {
		return existingDeposit.ID, nil
	}

	if amount.Currency != currency.EUR {
		return "", temporal.NewNonRetryableApplicationError("Invalid currency", "ErrInternal", fmt.Errorf("%w invalid currency", gatehub.ErrInternal))
	}

	wallet, err := a.b.Wallets().Get(ctx, walletID)
	if errors.Is(err, gatehub.ErrNotFound) {
		return "", temporal.NewNonRetryableApplicationError("Wallet not found", "ErrNotFound", fmt.Errorf("%w No wallet found for gatehub user", gatehub.ErrNotFound))
	}
	if err != nil {
		return "", err
	}

	las, err := a.b.LinkedAccounts().ListBalances(ctx, walletID)
	if err != nil {
		return "", err
	}

	var eurBalance *linkedaccounts.LinkedAccount
	for _, la := range las {
		if la.Provider == gatehub.ProviderName && la.Type == gatehub.AccTypeBalance {
			eurBalance = &la
			break
		}
	}
	if eurBalance == nil {
		return "", temporal.NewNonRetryableApplicationError("Gatehub EUR balance account not found", "ErrInternal", fmt.Errorf("%w Gatehub EUR balance account not found", gatehub.ErrInternal))
	}

	tx, err := a.b.Transactions().CreateTransaction(ctx, transactions.CreateTransactionArgs{
		ID:                      transactionID,
		WalletID:                walletID,
		ForeignID:               transactionID,
		ForeignType:             transactions.TransactionTypeDeposit,
		Provider:                gatehub.ProviderName,
		State:                   transactions.StateCompleted,
		Source:                  wallet.AddressString(),
		Destination:             wallet.AddressString(),
		Title:                   "Deposit",
		DestinationIdentity:     walletID,
		DestinationIdentityType: payments.IdentityTypeWalletID.String(),
		Amount:                  amount,
		ProviderFee:             &providerFee,
		LinkedAccountTitle:      "EUR Balance",
		Transfers: []transactions.TransferArgs{
			{
				LinkedAccountID: eurBalance.ID,
				ForeignID:       transactionID,
				Amount:          amount,
				Type:            transactions.TransferTypeCreditBalance,
				State:           transactions.StateCompleted,
			},
		},
	})
	if err != nil {
		return "", err
	}

	return tx, nil
}

func (a *Activity) UpdateGatehubWithdrawalState(ctx context.Context, walletID, transactionID string, state transactions.State) error {
	info := activity.GetInfo(ctx)
	if info.Attempt == 1 && state == transactions.StateFailed {
		slack.SendToChannel(ctx, slack.ChannelNotifyEvents, "wallet-info-bot", fmt.Sprintf("Gatehub withdrawal failed. %s/wallet/%s/transactions/%s", env.AdminURL(), walletID, transactionID))
	}

	trx, err := a.b.Transactions().GetTransaction(ctx, walletID, transactionID)
	if errors.Is(err, transactions.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError("Gatehub withdrawal not found", "ErrNotFound", fmt.Errorf("%w %s", gatehub.ErrInternal, err))
	}
	if err != nil {
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	transfers, err := a.b.Transactions().ListTransfers(ctx, trx.ID)
	if err != nil {
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	for _, t := range transfers {
		err = a.b.Transactions().SetTransferState(ctx, t.ID, state)
		if err != nil {
			return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
		}
	}

	return a.b.Transactions().SetTransactionState(ctx, trx.ID, state)
}

func (a *Activity) ReserveGatehubBalance(ctx context.Context, id, walletID string) error {
	tx, err := a.b.Transactions().GetTransaction(ctx, walletID, id)
	if errors.Is(err, transactions.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError("Transaction not found", "ErrInternal", fmt.Errorf("%w transaction not found", gatehub.ErrInternal))
	}
	if tx.Amount.Currency != currency.EUR {
		return temporal.NewNonRetryableApplicationError("Invalid currency", "ErrInternal", fmt.Errorf("%w invalid currency", gatehub.ErrInternal))
	}

	transfers, err := a.b.Transactions().ListTransfers(ctx, tx.ID)
	if err != nil {
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	if len(transfers) < 1 {
		return fmt.Errorf("%w Unable to reserve balance. No linked account specified.", gatehub.ErrInternal)
	}

	timeout := time.Hour * 24 * 365 // Pending transfers must have a timeout.

	amountWithFee := currency.Amount{
		Value:    tx.Amount.Value + tx.ProviderFee.Value,
		Currency: tx.Amount.Currency,
		Scale:    tx.Amount.Scale,
	}

	_, err = ReserveBalance(ctx, a.b, transfers[0].LinkedAccountID, tx.ID, amountWithFee, timeout)
	return err
}

func (a *Activity) FinalizeGatehubBalance(ctx context.Context, id, walletID string) error {
	tx, err := a.b.Transactions().GetTransaction(ctx, walletID, id)
	if errors.Is(err, transactions.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError("Transaction not found", "ErrInternal", fmt.Errorf("%w transaction not found", gatehub.ErrInternal))
	}
	if tx.Amount.Currency != currency.EUR {
		return temporal.NewNonRetryableApplicationError("Invalid currency", "ErrInternal", fmt.Errorf("%w invalid currency", gatehub.ErrInternal))
	}

	return FinaliseReserve(ctx, a.b, tx.ID)
}

func (a *Activity) FinalizeGatehubDeposit(ctx context.Context, id, walletID string, amount, providerFee currency.Amount) error {
	if amount.Currency != currency.EUR {
		return temporal.NewNonRetryableApplicationError("Invalid currency", "ErrInternal", fmt.Errorf("%w invalid currency", gatehub.ErrInternal))
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

	opsAcc := gatehub.EUROpsAccount
	ledger := gatehub.LedgerIDEUR
	tx, err := a.b.Pacioli().CreateTransfers(ctx, []pacioli.CreateTransferArgs{
		{
			ID:              id,
			Amount:          amount.Value,
			CreditAccountID: eurBalance.ID,
			DebitAccountID:  opsAcc,
			Pending:         false,
			Code:            1,
			Ledger:          ledger,
		},
	})
	if err != nil {
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	if len(tx) > 0 {
		if tx[0].Code == pacioli.TransferExists {
			return nil
		}
		if tx[0].Code == pacioli.TransferExceedsCredits || tx[0].Code == pacioli.TransferExceedsDebits || tx[0].Code == pacioli.TransferExceedsPendingTransferAmount {
			return fmt.Errorf("%w insufficient balance code (%s)", gatehub.ErrInsufficientBalance, tx[0].Code.String())
		}
		if tx[0].Code != 0 {
			return fmt.Errorf("%w non success code (%s)", gatehub.ErrInternal, tx[0].Code.String())
		}
	}

	return nil
}

type FeeFromGhArgs struct {
	UserID string
	TrxID  string
}

func (a *Activity) GetFeeFromGatehubTrasaction(ctx context.Context, args FeeFromGhArgs) (uint64, error) {
	if strings.TrimSpace(args.TrxID) == "" || strings.TrimSpace(args.UserID) == "" {
		return 0, fmt.Errorf("%w missing args", gatehub.ErrInternal)
	}

	et, err := a.external.GetTransaction(ctx, args.UserID, args.TrxID)
	if err != nil {
		return 0, err
	}
	// 23,35 -> 2335
	providerFee, err := StringToScaledUInt(et.Fee)
	if err != nil {
		return 0, err
	}

	return providerFee, nil
}

func (a *Activity) GetWalletFromGatehubUser(ctx context.Context, externalUserID string) (string, error) {
	walletID, err := getWalletID(ctx, a.b, externalUserID)
	if errors.Is(err, gatehub.ErrNotFound) {
		return "", temporal.NewNonRetryableApplicationError("Could not find wallet", "ErrNotFound", fmt.Errorf("%w No wallet found for gatehub user", gatehub.ErrNotFound))
	}
	if err != nil {
		return "", err
	}

	return walletID, nil
}

func (a *Activity) CheckGatehubWithdrawalComplete(ctx context.Context, walletID, transactionID string) error {
	trx, err := a.b.Transactions().GetTransaction(ctx, walletID, transactionID)
	if errors.Is(err, transactions.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError("Transaction not found", "ErrNotFound", err)
	}
	if err != nil {
		return err
	}

	externalUserID, err := getExternalUserID(ctx, a.b, walletID)
	if errors.Is(err, gatehub.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError("Gatehub user not found", "ErrNotFound", err)
	}
	if err != nil {
		return err
	}

	externalTrx, err := a.external.GetTransaction(ctx, externalUserID, trx.ForeignID)
	if errors.Is(err, external.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError("External transaction not found", "ErrNotFound", err)
	}
	if err != nil {
		return err
	}

	if externalTrx.Status != external.TransactionStatusCompleted {
		return fmt.Errorf("%w Gatehub transaction not completed", gatehub.ErrInternal)
	}

	return nil
}

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

func (a *Activity) LinkGatehubUserToGateway(ctx context.Context, externalUser string) error {
	return a.external.LinkUserToGateway(ctx, externalUser)
}

func (a *Activity) GetLinkedAccount(ctx context.Context, walletID string) (linkedaccounts.LinkedAccount, error) {
	var linkedAccount linkedaccounts.LinkedAccount
	la, err := a.b.LinkedAccounts().ListByWalletId(ctx, walletID)
	if err != nil {
		return linkedAccount, err
	}

	for _, l := range la {
		if l.Provider == gatehub.ProviderName && l.Type == gatehub.AccTypeBalance {
			linkedAccount = l
			break
		}
	}

	return linkedAccount, nil
}

func (a *Activity) BackfillPaywiserBalanceAfterKYC(ctx context.Context, walletID string) (*gatehub.Balance, error) {
	if a.config.SendingUserID == "" {
		return nil, fmt.Errorf("missing SendingUserID in Gatehub configuration")
	}
	if a.config.SendingUserAddress == "" {
		return nil, fmt.Errorf("missing SendingUserAddress in Gatehub configuration")
	}

	la, err := a.b.LinkedAccounts().ListByWalletId(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("missing linked account")
	}
	var linkedAccount linkedaccounts.LinkedAccount
	for _, l := range la {
		if l.Provider == gatehub.ProviderName && l.Type == gatehub.AccTypeBalance {
			linkedAccount = l
		}
	}
	if linkedAccount.ID == "" {
		return nil, fmt.Errorf("%w missing linked account for %s", gatehub.ErrInternal, walletID)
	}
	balance, err := GetBalance(ctx, a.b, linkedAccount.ID)
	if err != nil {
		return nil, err
	}
	transfer := balance.Total.Float64()
	if transfer <= 0 {
		log.Info("no balance to transfer", zap.String("wallet_id", walletID))
		return balance, nil
	}
	if a.config.SendingUserAddress == linkedAccount.ProviderID {
		return balance, nil
	}
	externalTx, err := a.external.CreateTransaction(ctx, external.CreateTransactionRequest{
		SendingUserID:    a.config.SendingUserID,
		SendingAddress:   a.config.SendingUserAddress,
		ReceivingAddress: linkedAccount.ProviderID,
		Amount:           transfer,
		Message:          "backfill transfer",
		Type:             external.TransactionTypeHosted,
		VaultID:          a.external.GetVaultID(),
	})

	if err != nil {
		return nil, err
	}
	log.Info("created external transaction", zap.Any("externalTx", externalTx), zap.Float64("amount", transfer))

	return balance, nil
}

func (a *Activity) CheckIfBackfillWasDone(ctx context.Context, walletID string) (string, error) {
	var externalID string
	err := a.b.DB().GetContext(ctx, &externalID, "SELECT external_id FROM gatehub_users WHERE wallet_id=$1", walletID)
	if err != nil {
		return "", err
	}

	if externalID == "" {
		return "", nil
	}
	var retrievedWalletID string
	err = a.b.DB().GetContext(ctx, &retrievedWalletID, "SELECT wallet_id FROM gatehub_backfill_users WHERE wallet_id=$1 AND external_id=$2", walletID, externalID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return externalID, nil
		}
		return "", err
	}

	if retrievedWalletID != "" {
		return "", nil
	}

	return externalID, nil

}

func (a *Activity) MarkBackfillUser(ctx context.Context, walletID, externalUserID string, balance *gatehub.Balance) error {
	_, err := a.b.DB().ExecContext(ctx, `INSERT INTO gatehub_backfill_users(wallet_id, external_id, unscaled_value)
													VALUES ( $1, $2, $3);`, walletID, externalUserID, balance.Total.Value)
	if err != nil {
		return err
	}
	return nil

}

func (a *Activity) SetKYCApprovedForGatehub(ctx context.Context, walletID string) error {
	return a.b.KYC().SetKYCStatus(ctx, kyc.StatusUpdateArgs{WalletID: walletID, Status: kyc.StatusLevel1})
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

func getMerchantName(tx external.CardTransaction) string {
	switch tx.Type {
	case external.CardTransactionTypeATMWithdrawal:
		if tx.MerchantName != nil && *tx.MerchantName != "" {
			return fmt.Sprintf("ATM Withdrawal (%s)", *tx.MerchantName)
		}
		return "ATM Withdrawal"
	default:
		if tx.MerchantName != nil && *tx.MerchantName != "" {
			return *tx.MerchantName
		}
		return "N/A"
	}
}
