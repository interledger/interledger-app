package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/gatehub/external"
	httplogger "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/pacioli"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.temporal.io/sdk/temporal"
)

type Activity struct {
	b        Backends
	external external.Client
}

func NewActivity(b Backends) *Activity {
	ec := external.NewClient(
		os.Getenv("GATEHUB_APP_ID"),
		os.Getenv("GATEHUB_SECRET"),
		&http.Client{
			Transport: otelhttp.NewTransport(
				httplogger.NewTransport(http.DefaultTransport, b, nil),
			),
		},
	)

	return &Activity{
		b:        b,
		external: ec,
	}
}

func (a *Activity) CreateGatehubUser(ctx context.Context, walletID string) (string, error) {
	ul, err := a.b.Users().ListUsers(ctx, walletID)
	if err != nil {
		return "", fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	if len(ul) < 1 {
		return "", fmt.Errorf("%w No Fynbos user found for walletID", gatehub.ErrInternal)
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
		ProviderID:      primaryWallet.UUID,
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
		return fmt.Errorf("%w failed to setup account status(%s)", pti.ErrInternal, accs[0].Code)
	}

	return nil
}

func (a *Activity) CreateGatehubDepositTransaction(ctx context.Context, webhookID, walletID string, amount currency.Amount) (string, error) {
	existingDeposit, err := a.b.Transactions().GetTransactionByForeignID(ctx, walletID, webhookID)
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
		return "", temporal.NewNonRetryableApplicationError("Invalid currency", "ErrNotFound", fmt.Errorf("%w No wallet found for gatehub user", gatehub.ErrNotFound))
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
		ID:                      webhookID,
		WalletID:                walletID,
		ForeignID:               webhookID,
		ForeignType:             transactions.TransactionTypeDeposit,
		Provider:                gatehub.ProviderName,
		State:                   transactions.StateCompleted,
		Source:                  wallet.AddressString(),
		Destination:             wallet.AddressString(),
		Title:                   "Deposit",
		DestinationIdentity:     walletID,
		DestinationIdentityType: payments.IdentityTypeWalletID.String(),
		Amount:                  amount,
		LinkedAccountTitle:      "EUR Balance",
		Transfers: []transactions.TransferArgs{
			{
				LinkedAccountID: eurBalance.ID,
				ForeignID:       webhookID,
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

func (a *Activity) CreateGatehubWithdrawalTransaction(ctx context.Context, walletID, externalTransactionID string) (string, error) {
	existingWithdrawal, err := a.b.Transactions().GetTransactionByForeignID(ctx, walletID, externalTransactionID)
	if err != nil && !errors.Is(err, transactions.ErrNotFound) {
		return "", err
	}
	if existingWithdrawal != nil {
		return existingWithdrawal.ID, nil
	}

	externalUserID, err := getExternalUserID(ctx, a.b, walletID)
	if errors.Is(err, gatehub.ErrNotFound) {
		return "", temporal.NewNonRetryableApplicationError("Gatehub user not found", "ErrNotFound", err)
	}
	if err != nil {
		return "", err
	}

	trx, err := a.external.GetTransaction(ctx, externalUserID, externalTransactionID)
	if errors.Is(err, external.ErrNotFound) {
		return "", temporal.NewNonRetryableApplicationError("External transaction not found", "ErrNotFound", err)
	}
	if err != nil {
		return "", err
	}

	cc := currency.ParseCurrency(trx.Vault.AssetCode)
	if cc != currency.EUR {
		return "", temporal.NewNonRetryableApplicationError("Invalid currency", "ErrInternal", fmt.Errorf("%w invalid currency", gatehub.ErrInternal))
	}

	parts := strings.Split(trx.Amount, ".")
	if len(parts) < 1 {
		return "", temporal.NewNonRetryableApplicationError("Invalid amount", "ErrInternal", fmt.Errorf("%w invalid amount", gatehub.ErrInternal))
	}

	value, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return "", temporal.NewNonRetryableApplicationError("Invalid amount", "ErrInternal", fmt.Errorf("%w %s", gatehub.ErrInternal, err))
	}
	amount := currency.Amount{
		Value:    value * 100, // EUR scale = 2
		Currency: cc,
	}

	wallet, err := a.b.Wallets().Get(ctx, walletID)
	if errors.Is(err, gatehub.ErrNotFound) {
		return "", temporal.NewNonRetryableApplicationError("Invalid currency", "ErrNotFound", fmt.Errorf("%w No wallet found for gatehub user", gatehub.ErrNotFound))
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
		WalletID:                walletID,
		ForeignID:               externalTransactionID,
		ForeignType:             transactions.TransactionTypeDeposit,
		Provider:                gatehub.ProviderName,
		State:                   transactions.StateCompleted,
		Source:                  wallet.AddressString(),
		Destination:             wallet.AddressString(),
		Title:                   "Withdrawal",
		DestinationIdentity:     walletID,
		DestinationIdentityType: payments.IdentityTypeWalletID.String(),
		Amount:                  amount,
		LinkedAccountTitle:      "EUR Balance",
		Transfers: []transactions.TransferArgs{
			{
				LinkedAccountID: eurBalance.ID,
				ForeignID:       externalTransactionID,
				Amount:          amount,
				Type:            transactions.TransferTypeDebitBalance,
				State:           transactions.StateCompleted,
			},
		},
	})
	if err != nil {
		return "", err
	}

	return tx, nil
}

func (a *Activity) AssignGatehubWithdrawal(ctx context.Context, id, walletID string) error {
	tx, err := a.b.Transactions().GetTransaction(ctx, walletID, id)
	if errors.Is(err, transactions.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError("Transaction not found", "ErrInternal", fmt.Errorf("%w transaction not found", gatehub.ErrInternal))
	}
	if tx.Amount.Currency != currency.EUR {
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
	ledgerTx, err := a.b.Pacioli().CreateTransfers(ctx, []pacioli.CreateTransferArgs{
		{
			ID:              id,
			Amount:          tx.Amount.Value,
			CreditAccountID: opsAcc,
			DebitAccountID:  eurBalance.ID,
			Pending:         false,
			Code:            1,
			Ledger:          ledger,
		},
	})
	if err != nil {
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	if len(ledgerTx) > 0 {
		if ledgerTx[0].Code == pacioli.TransferExists {
			return nil
		}
		if ledgerTx[0].Code == pacioli.TransferExceedsCredits || ledgerTx[0].Code == pacioli.TransferExceedsDebits || ledgerTx[0].Code == pacioli.TransferExceedsPendingTransferAmount {
			return fmt.Errorf("%w insufficient balance code (%s)", pti.ErrInsufficientBalance, ledgerTx[0].Code.String())
		}
		if ledgerTx[0].Code != 0 {
			return fmt.Errorf("%w non success code (%s)", pti.ErrInternal, ledgerTx[0].Code.String())
		}
	}

	return nil
}

func (a *Activity) FinalizeGatehubDeposit(ctx context.Context, id, walletID string, amount currency.Amount) error {
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
			return fmt.Errorf("%w insufficient balance code (%s)", pti.ErrInsufficientBalance, tx[0].Code.String())
		}
		if tx[0].Code != 0 {
			return fmt.Errorf("%w non success code (%s)", pti.ErrInternal, tx[0].Code.String())
		}
	}

	return nil
}

func (a *Activity) GetWalletFromGatehubUser(ctx context.Context, externalUserID string) (string, error) {
	walletID, err := getWalletID(ctx, a.b, externalUserID)
	if errors.Is(err, gatehub.ErrNotFound) {
		return "", temporal.NewNonRetryableApplicationError("Invalid currency", "ErrNotFound", fmt.Errorf("%w No wallet found for gatehub user", gatehub.ErrNotFound))
	}
	if err != nil {
		return "", err
	}

	return walletID, nil
}

func (a *Activity) ValidateGatehubWithdrawal(ctx context.Context, walletID string, externalTransactionID string) error {
	externalUserID, err := getExternalUserID(ctx, a.b, walletID)
	if errors.Is(err, gatehub.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError("Gatehub user not found", "ErrNotFound", err)
	}
	if err != nil {
		return err
	}

	trx, err := a.external.GetTransaction(ctx, externalUserID, externalTransactionID)
	if errors.Is(err, external.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError("External transaction not found", "ErrNotFound", err)
	}
	if err != nil {
		return err
	}
	if trx.Type != external.TransactionTypeWithdrawal {
		return temporal.NewNonRetryableApplicationError("External transaction is not a withdrawal", "ErrInternal", fmt.Errorf("%w External transaction is not a withdrawal", gatehub.ErrInternal))
	}

	cc := currency.ParseCurrency(trx.Vault.AssetCode)
	if cc != currency.EUR {
		return temporal.NewNonRetryableApplicationError("Invalid currency", "ErrInternal", fmt.Errorf("%w invalid currency", gatehub.ErrInternal))
	}

	balances, err := a.b.LinkedAccounts().ListBalances(ctx, walletID)
	if err != nil {
		return err
	}

	var balance *linkedaccounts.LinkedAccount
	for _, bal := range balances {
		if bal.Provider == gatehub.ProviderName && bal.Type == gatehub.AccTypeBalance {
			balance = &bal
			break
		}
	}
	if balance == nil {
		return temporal.NewNonRetryableApplicationError("Gatehub balance linked account not found", "ErrNotFound", fmt.Errorf("%w Gatehub balance linked account not found", gatehub.ErrNotFound))
	}
	if balance.ProviderID != trx.SendingWallet.UUID {
		return temporal.NewNonRetryableApplicationError("Gatehub withdrawal not for this Fynbos wallet", "ErrInternal", fmt.Errorf("%w Gatehub withdrawal not for this Fynbos wallet", gatehub.ErrInternal))
	}

	return nil
}

func (a *Activity) CheckGatehubTransactionComplete(ctx context.Context, walletID, externalTransactionID string) error {
	externalUserID, err := getExternalUserID(ctx, a.b, walletID)
	if errors.Is(err, gatehub.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError("Gatehub user not found", "ErrNotFound", err)
	}
	if err != nil {
		return err
	}

	trx, err := a.external.GetTransaction(ctx, externalUserID, externalTransactionID)
	if errors.Is(err, external.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError("External transaction not found", "ErrNotFound", err)
	}
	if err != nil {
		return err
	}

	if trx.Status != external.TransactionStatusCompleted {
		return fmt.Errorf("%w Gatehub transaction not completed", gatehub.ErrInternal)
	}

	return nil
}
