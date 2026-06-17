package ops

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	"github.com/interledger/interledger-app/go/backend/payments"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub/external"
	"github.com/interledger/interledger-app/go/backend/slack"
	"github.com/interledger/interledger-app/go/backend/transactions"
	"github.com/interledger/interledger-app/go/env"
	"github.com/interledger/interledger-app/go/pacioli"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

func CreateUser(ctx context.Context, b Backends, walletID string) (gatehub.Await, error) {
	wo := client.StartWorkflowOptions{
		ID:                    "gatehub_create_user_" + walletID,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	var workflowStatus enums.WorkflowExecutionStatus
	wflow, err := b.Temporal().DescribeWorkflowExecution(ctx, wo.ID, "")
	switch err.(type) {
	case *serviceerror.Internal,
		*serviceerror.Unavailable,
		*serviceerror.InvalidArgument:
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	case *serviceerror.NotFound:
		// do nothing
	default:
		if wflow != nil {
			workflowStatus = wflow.GetWorkflowExecutionInfo().Status
		}
	}

	// return workflow if it's running
	var await client.WorkflowRun
	var executeErr error
	if workflowStatus == enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
		await = b.Temporal().GetWorkflow(ctx, wo.ID, "")
	} else {
		await, executeErr = b.Temporal().ExecuteWorkflow(ctx, wo, CreateGatehubUserWorkflow, walletID)
	}
	if executeErr != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return await.Get, nil
}

func GetUser(ctx context.Context, b Backends, ec external.Client, walletID string) (*gatehub.User, error) {
	externalUserID, err := getExternalUserID(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	usr, err := ec.GetUser(ctx, externalUserID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return usr, nil
}

func GetOnboardingWidget(ctx context.Context, b Backends, ec external.Client, walletID string) (string, error) {
	externalUserID, err := getExternalUserID(ctx, b, walletID)
	if errors.Is(err, gatehub.ErrNotFound) {
		await, innerErr := CreateUser(ctx, b, walletID)
		if innerErr != nil {
			return "", innerErr
		}

		innerErr = await(ctx, &externalUserID)
		if innerErr != nil {
			return "", fmt.Errorf("%w %s", gatehub.ErrInternal, innerErr)
		}
	} else if err != nil {
		return "", err
	}

	widget, err := ec.GetOnboardingWidget(ctx, externalUserID)
	if err != nil {
		return "", fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return widget, nil
}

func GetOnOffRampWidget(ctx context.Context, b Backends, ec external.Client, walletID string, isDeposit bool) (string, error) {
	externalUserID, err := getExternalUserID(ctx, b, walletID)
	if errors.Is(err, gatehub.ErrNotFound) {
		await, innerErr := CreateUser(ctx, b, walletID)
		if innerErr != nil {
			return "", innerErr
		}

		innerErr = await(ctx, &externalUserID)
		if innerErr != nil {
			return "", fmt.Errorf("%w %s", gatehub.ErrInternal, innerErr)
		}
	} else if err != nil {
		return "", err
	}

	widget, err := ec.GetOnOffRampWidget(ctx, externalUserID, isDeposit)
	if err != nil {
		return "", fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return widget, nil
}

func getExternalUserID(ctx context.Context, b Backends, walletID string) (string, error) {
	var externalID string
	err := b.DB().GetContext(ctx, &externalID, "SELECT external_id FROM gatehub_users WHERE wallet_id=$1;", walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w %s", gatehub.ErrNotFound, err)
	} else if err != nil {
		return "", fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return externalID, err
}

func GetExternalIDs(ctx context.Context, b Backends, walletID string) (*gatehub.ExternalIDs, error) {
	var externalIDs gatehub.ExternalIDs
	err := b.DB().GetContext(ctx, &externalIDs, "SELECT external_id, external_customer_id, external_customer_source_id, external_account_id, external_account_source_id FROM gatehub_users WHERE wallet_id = $1;", walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", gatehub.ErrNotFound, err)
	} else if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return &externalIDs, nil
}

func getExternalIDsByUserID(ctx context.Context, b Backends, userID string) (*gatehub.ExternalIDs, error) {
	var externalIDs gatehub.ExternalIDs
	err := b.DB().GetContext(ctx, &externalIDs, "SELECT external_id, external_customer_id, external_customer_source_id, external_account_id, external_account_source_id FROM gatehub_users WHERE external_id = $1;", userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", gatehub.ErrNotFound, err)
	} else if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return &externalIDs, nil
}

func getWalletID(ctx context.Context, b Backends, externalUserID string) (string, error) {
	var walletID string
	err := b.DB().GetContext(ctx, &walletID, "SELECT wallet_id FROM gatehub_users WHERE external_id=$1;", externalUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w %s", gatehub.ErrNotFound, err)
	} else if err != nil {
		return "", fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return walletID, nil
}

func GetBalance(ctx context.Context, b Backends, linkedAccountID string) (*gatehub.Balance, error) {
	la, err := b.LinkedAccounts().Get(ctx, linkedAccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	if la.Provider != gatehub.ProviderName || la.Type != gatehub.AccTypeBalance {
		return nil, fmt.Errorf("%w linked account not correct type", gatehub.ErrNotFound)
	}

	accs, err := b.Pacioli().GetAccounts(ctx, []string{la.ID})
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	if len(accs) != 1 {
		return nil, fmt.Errorf("%w account not found", gatehub.ErrNotFound)
	}

	return &gatehub.Balance{
		Total:     currency.FromUInt64(int64(accs[0].CreditsPosted)-int64(accs[0].DebitsPosted), la.SendCurrency),
		Available: currency.FromUInt64(int64(accs[0].CreditsPosted)-int64(accs[0].DebitsPosted)-int64(accs[0].DebitsPending), la.SendCurrency),
	}, nil
}

func CreateWithdrawal(ctx context.Context, b Backends, ec external.Client, walletID, externalTransactionID string) (string, error) {
	existingWithdrawal, err := b.Transactions().GetTransactionByForeignID(ctx, walletID, externalTransactionID)
	if err != nil && !errors.Is(err, transactions.ErrNotFound) {
		return "", fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	if existingWithdrawal != nil {
		return existingWithdrawal.ID, nil
	}

	amount, balanceAccount, fee, err := validateWithdrawal(ctx, b, ec, walletID, externalTransactionID)
	if err != nil {
		return "", err
	}

	wallet, err := b.Wallets().Get(ctx, walletID)
	if err != nil {
		return "", fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	trxID := uuid.NewString()

	amountWithFee := currency.Amount{
		Value:    amount.Value + fee.Value,
		Currency: amount.Currency,
		Scale:    amount.Scale,
	}
	if _, err = ReserveBalance(ctx, b, balanceAccount.ID, trxID, amountWithFee, 365*24*time.Hour); err != nil {
		slack.SendToChannel(ctx, slack.ChannelError, "wallet-info-bot",
			fmt.Sprintf("*:::[GateHub ERROR]:::* Withdrawal ledger reserve failed\n*Wallet ID:* %s\n*GateHub TX ID:* %s\n*Error:* %s",
				walletID, externalTransactionID, err))
		return "", fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	if _, err = b.Transactions().CreateTransaction(ctx, transactions.CreateTransactionArgs{
		ID:                      trxID,
		WalletID:                walletID,
		ForeignID:               externalTransactionID,
		Provider:                gatehub.ProviderName,
		State:                   transactions.StatePending,
		ForeignType:             transactions.TransactionTypeWithdrawal,
		Source:                  wallet.AddressString(),
		Destination:             wallet.AddressString(),
		Title:                   "Withdrawal",
		DestinationIdentity:     walletID,
		ProviderFee:             &fee,
		DestinationIdentityType: payments.IdentityTypeWalletID.String(),
		Amount:                  amount,
		LinkedAccountTitle:      "EUR Balance",
	}); err != nil {
		slack.SendToChannel(ctx, slack.ChannelError, "wallet-info-bot",
			fmt.Sprintf("*:::[GateHub ERROR]:::* Withdrawal transaction record creation failed\n*Wallet ID:* %s\n*Transaction ID:* %s\n*GateHub TX ID:* %s\n*Error:* %s",
				walletID, trxID, externalTransactionID, err))
		return "", fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return trxID, nil
}

func validateWithdrawal(ctx context.Context, b Backends, ec external.Client, walletID, externalTransactionID string) (currency.Amount, *linkedaccounts.LinkedAccount, currency.Amount, error) {
	externalUserID, err := getExternalUserID(ctx, b, walletID)
	if err != nil {
		return currency.Amount{}, nil, currency.Amount{}, err
	}

	trx, err := ec.GetTransaction(ctx, externalUserID, externalTransactionID)
	if err != nil {
		return currency.Amount{}, nil, currency.Amount{}, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	fee, err := StringToScaledUInt(trx.Fee)
	if err != nil {
		return currency.Amount{}, nil, currency.Amount{}, fmt.Errorf("%w invalid fee", gatehub.ErrInternal)
	}

	if trx.Type != external.TransactionTypeWithdrawal {
		return currency.Amount{}, nil, currency.Amount{}, fmt.Errorf("%w Transaction is not a withdrawal", gatehub.ErrInternal)
	}
	cc := currency.ParseCurrency(trx.Vault.AssetCode)
	if cc != currency.EUR {
		return currency.Amount{}, nil, currency.Amount{}, fmt.Errorf("%w Invalid currency", gatehub.ErrInternal)
	}

	balances, err := b.LinkedAccounts().ListBalances(ctx, walletID)
	if err != nil {
		return currency.Amount{}, nil, currency.Amount{}, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	var balance *linkedaccounts.LinkedAccount
	for _, bal := range balances {
		if bal.Provider == gatehub.ProviderName && bal.Type == gatehub.AccTypeBalance {
			balance = &bal
			break
		}
	}
	if balance == nil {
		return currency.Amount{}, nil, currency.Amount{}, fmt.Errorf("%w Gatehub balance linked account not found", gatehub.ErrNotFound)
	}
	if balance.ProviderID != trx.SendingWallet.Address {
		return currency.Amount{}, nil, currency.Amount{}, fmt.Errorf("%w Gatehub withdrawal is not for this Interledger wallet", gatehub.ErrInternal)
	}

	parts := strings.Split(trx.Amount, ".")
	if len(parts) < 1 {
		return currency.Amount{}, nil, currency.Amount{}, fmt.Errorf("%w invalid amount", gatehub.ErrInternal)
	}

	value, err := StringToScaledUInt(trx.Amount)
	if err != nil {
		return currency.Amount{}, nil, currency.Amount{}, fmt.Errorf("%w invalid amount", gatehub.ErrInternal)
	}

	return currency.Amount{
			Value:    value, // EUR scale = 2
			Currency: cc,
		}, balance, currency.Amount{
			Value:    fee,
			Currency: cc,
			Scale:    2,
		}, nil
}

func ReserveBalance(ctx context.Context, b Backends, linkedAccountID, txID string, amt currency.Amount, timeout time.Duration) (*gatehub.Balance, error) {
	if amt.Currency != currency.EUR {
		return nil, fmt.Errorf("%w %s not supported.", gatehub.ErrInternal, amt.Currency)
	}

	la, err := b.LinkedAccounts().Get(ctx, linkedAccountID)
	if err != nil {
		return nil, err
	}
	if la.Provider != gatehub.ProviderName || la.Type != gatehub.AccTypeBalance {
		return nil, fmt.Errorf("%w Not a Gatehub linked account", gatehub.ErrInternal)
	}

	opsAcc := gatehub.EUROpsAccount
	ledger := gatehub.LedgerIDEUR
	tx, err := b.Pacioli().CreateTransfers(ctx, []pacioli.CreateTransferArgs{
		{
			ID:              txID,
			Amount:          amt.Value,
			DebitAccountID:  la.ID,
			CreditAccountID: opsAcc,
			Pending:         true,
			Code:            1,
			Timeout:         uint64(timeout),
			Ledger:          ledger,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	if len(tx) > 0 {
		if tx[0].Code == pacioli.TransferExceedsCredits || tx[0].Code == pacioli.TransferExceedsDebits || tx[0].Code == pacioli.TransferExceedsPendingTransferAmount {
			return nil, fmt.Errorf("%w insufficiens balance cod (%s)", gatehub.ErrInsufficientBalance, tx[0].Code.String())
		}
		if tx[0].Code != 0 {
			return nil, fmt.Errorf("%w non success code (%s)", gatehub.ErrInternal, tx[0].Code.String())
		}
	}

	accs, err := b.Pacioli().GetAccounts(ctx, []string{la.ID})
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	if len(accs) != 1 {
		return nil, fmt.Errorf("%w account not found", gatehub.ErrNotFound)
	}

	return &gatehub.Balance{
		Total:     currency.FromUInt64(int64(accs[0].CreditsPosted)-int64(accs[0].DebitsPosted), la.SendCurrency),
		Available: currency.FromUInt64(int64(accs[0].CreditsPosted)-int64(accs[0].DebitsPosted)-int64(accs[0].DebitsPending), la.SendCurrency),
	}, nil
}

func FinaliseReserve(ctx context.Context, b Backends, trxID string) error {
	tx, err := b.Pacioli().PostTransfers(ctx, []string{trxID})
	if err != nil {
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	if len(tx) == 0 {
		return nil
	}
	if tx[0].Code == pacioli.TransferPendingTransferAlreadyPosted {
		return nil
	}
	if tx[0].Code == pacioli.TransferExceedsCredits || tx[0].Code == pacioli.TransferExceedsDebits || tx[0].Code == pacioli.TransferExceedsPendingTransferAmount {
		return fmt.Errorf("%w insufficient balance code (%s)", gatehub.ErrInsufficientBalance, tx[0].Code.String())
	}
	if tx[0].Code == pacioli.TransferPendingTransferExpired {
		return fmt.Errorf("%w transfer timed out code (%s)", gatehub.ErrTimedOut, tx[0].Code.String())
	}
	if tx[0].Code == pacioli.TransferPendingTransferNotPending {
		return fmt.Errorf("%w pending transfer not found code (%s)", gatehub.ErrNotFound, tx[0].Code.String())
	}
	if tx[0].Code != 0 {
		return fmt.Errorf("%w non success code (%s)", gatehub.ErrInternal, tx[0].Code.String())
	}

	return nil
}

func RollbackReserve(ctx context.Context, b Backends, txID string) error {
	tx, err := b.Pacioli().VoidTransfers(ctx, []string{txID})
	if err != nil {
		slack.SendToChannel(ctx, slack.ChannelError, "wallet-info-bot", fmt.Sprintf("*:::[GateHub ERROR]:::* \n *RollbackReserve  txID:* %s,\n *error:* %s", txID, err))
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	if len(tx) == 0 {
		return nil
	}
	if tx[0].Code == pacioli.TransferPendingTransferNotFound ||
		tx[0].Code == pacioli.TransferPendingTransferAlreadyVoided ||
		tx[0].Code == pacioli.TransferPendingTransferExpired {
		return nil
	}
	if tx[0].Code != 0 {
		return fmt.Errorf("%w non success code (%s)", gatehub.ErrInternal, tx[0].Code.String())
	}

	return nil
}

func AssignBalance(ctx context.Context, b Backends, linkedAccountID, txID string, amt currency.Amount) (*gatehub.Balance, error) {
	la, err := b.LinkedAccounts().Get(ctx, linkedAccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	if la.Provider != gatehub.ProviderName || la.Type != gatehub.AccTypeBalance {
		return nil, fmt.Errorf("%w linked account not correct type", gatehub.ErrNotFound)
	}

	opsAcc := gatehub.EUROpsAccount
	ledger := gatehub.LedgerIDEUR
	tx, err := b.Pacioli().CreateTransfers(ctx, []pacioli.CreateTransferArgs{
		{
			ID:              txID,
			Amount:          amt.Value,
			CreditAccountID: la.ID,
			DebitAccountID:  opsAcc,
			Pending:         false,
			Code:            1,
			Ledger:          ledger,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	if len(tx) != 0 {
		if tx[0].Code == pacioli.TransferExceedsCredits || tx[0].Code == pacioli.TransferExceedsDebits || tx[0].Code == pacioli.TransferExceedsPendingTransferAmount {
			return nil, fmt.Errorf("%w insufficient balance cod (%s)", gatehub.ErrInsufficientBalance, tx[0].Code.String())
		}
		if tx[0].Code != 0 {
			return nil, fmt.Errorf("%w non success code (%s)", gatehub.ErrInternal, tx[0].Code.String())
		}
	}

	accs, err := b.Pacioli().GetAccounts(ctx, []string{la.ID})
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	if len(accs) != 1 {
		return nil, fmt.Errorf("%w account not found", gatehub.ErrNotFound)
	}

	return &gatehub.Balance{
		Total:     currency.FromUInt64(int64(accs[0].CreditsPosted)-int64(accs[0].DebitsPosted), la.SendCurrency),
		Available: currency.FromUInt64(int64(accs[0].CreditsPosted)-int64(accs[0].DebitsPosted)-int64(accs[0].DebitsPending), la.SendCurrency),
	}, nil
}

func CreateTransfer(ctx context.Context, b Backends, ec external.Client, args gatehub.CreateTransferArgs) (*external.Transaction, error) {
	sendLA, err := b.LinkedAccounts().Get(ctx, args.SendingLinkedAccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	recvLA, err := b.LinkedAccounts().Get(ctx, args.ReceivingLinkedAccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	recvWallet, err := b.Wallets().Get(ctx, recvLA.WalletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	sendingUser, err := getExternalUserID(ctx, b, sendLA.WalletID)
	if err != nil {
		return nil, err
	}

	// Validate vault ID is available before making the external call
	vaultID := ec.GetVaultID()
	if vaultID == "" && !env.IsTestExecution() {
		return nil, temporal.NewNonRetryableApplicationError(
			"Gatehub vault ID not configured",
			"ConfigurationError",
			fmt.Errorf("GATEHUB_PAYWISER_EURO_VAULT_ID environment variable is not set"),
		)
	}

	externalTx, err := ec.CreateTransaction(ctx, external.CreateTransactionRequest{
		SendingUserID:    sendingUser,
		SendingAddress:   sendLA.ProviderID,
		ReceivingAddress: recvLA.ProviderID,
		Amount:           args.Amount.Float64(),
		Message:          fmt.Sprintf("Payment to %s", recvWallet.Name),
		Type:             external.TransactionTypeHosted,
		VaultID:          vaultID,
	})
	if errors.Is(err, external.ErrNotFound) {
		return nil, fmt.Errorf("%w %s", gatehub.ErrNotFound, err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return externalTx, nil
}

func GetTransaction(ctx context.Context, b Backends, ec external.Client, walletID, id string) (*external.Transaction, error) {
	externalUser, err := getExternalUserID(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	return ec.GetTransaction(ctx, externalUser, id)
}

func ListDeliveryAddresses(ctx context.Context, b Backends, ec external.Client, walletID string) ([]external.CustomerDeliveryAddress, error) {
	externalIDs, err := GetExternalIDs(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	if !externalIDs.CustomerID.Valid {
		return nil, fmt.Errorf("%w attempted to list delivery addresses for non-existing customer", gatehub.ErrInternal)
	}

	return ec.GetDeliveryAddresses(ctx, externalIDs.UserID, externalIDs.CustomerID.String)
}

func ListCards(ctx context.Context, b Backends, ec external.Client, externalIDs gatehub.ExternalIDs) ([]external.Card, error) {
	if !externalIDs.CustomerID.Valid {
		return []external.Card{}, nil
	}

	cardsRes, err := ec.ListCards(ctx, externalIDs.UserID, externalIDs.CustomerID.String)
	if err != nil {
		return []external.Card{}, nil
	}

	return cardsRes.Data, nil
}

func GetCardApplicationProducts(ctx context.Context, b Backends, ec external.Client) ([]external.CardApplicationProduct, error) {
	return ec.GetCardApplicationProducts(ctx)
}

func OrderCard(ctx context.Context, b Backends, ec external.Client, args gatehub.OrderCardArgs) error {
	las, err := b.LinkedAccounts().ListByWalletId(ctx, args.Wallet.ID)
	if err != nil {
		return err
	}

	var la linkedaccounts.LinkedAccount
	for _, acc := range las {
		if acc.Provider == gatehub.ProviderName && acc.Type == gatehub.AccTypeBalance {
			la = acc
			break
		}
	}

	nameOnCard, err := getNameOnCard(args.Wallet.AddressShortString())
	if err != nil {
		return err
	}

	if args.ExternalIDs.IsCustomerCreated() {
		wo := client.StartWorkflowOptions{
			ID:                    "gatehub_create_card_" + args.ExternalIDs.UserID + "_" + time.Now().UTC().Format(gatehub.TimeLayout),
			TaskQueue:             "backend",
			WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
		}

		var workflowStatus enums.WorkflowExecutionStatus

		wflow, err := b.Temporal().DescribeWorkflowExecution(ctx, wo.ID, "")
		switch err.(type) {
		case *serviceerror.Internal,
			*serviceerror.Unavailable,
			*serviceerror.InvalidArgument:

			return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
		case *serviceerror.NotFound:
			// do nothing
		default:
			if wflow != nil {
				workflowStatus = wflow.GetWorkflowExecutionInfo().Status
			}
		}

		if workflowStatus != enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
			_, err = b.Temporal().ExecuteWorkflow(ctx, wo, CreateGateHubCardWorkflow, CreateCardWorkflowArgs{
				WalletID:           args.Wallet.ID,
				ExternalIDs:        args.ExternalIDs,
				Currency:           "PW_" + currency.EUR.String(),
				NameOnCard:         nameOnCard,
				WalletAddress:      la.ProviderID,
				CardProductCode:    args.CardProductCode,
				DeliveryAddressID:  args.DeliveryAddressID,
				NewDeliveryAddress: args.NewDeliveryAddress,
				ShouldOrderPlastic: args.ShouldOrderPlastic,
			})

			if err != nil {
				return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
			}
		}

		return nil
	}

	customer, err := ec.CreateCustomerAndCard(ctx, args.ExternalIDs.UserID, external.CreateCustomerAndCardArgs{
		WalletAddress: la.ProviderID,
		NameOnCard:    nameOnCard,
		Delivery:      args.NewDeliveryAddress,
		Account: external.CardAccount{
			Currency: "PW_" + currency.EUR.String(),
			Card: external.NewCardArgs{
				ProductCode: args.CardProductCode,
			},
		},
	})

	if err != nil {
		return err
	}

	err = updateCustomerSourceIDs(ctx, b, args.ExternalIDs.UserID, customer.SourceID, customer.Accounts[0].SourceID, args.ShouldOrderPlastic)
	if err != nil {
		return err
	}

	return nil
}

func ValidateCardProductCode(ctx context.Context, ec external.Client, cardProductCode string) error {
	var err error
	products, err := ec.GetCardApplicationProducts(ctx)
	if err != nil {
		return err
	}

	for _, p := range products {
		if p.Code == cardProductCode {
			return nil
		}
	}

	return fmt.Errorf("%w invalid card product code: %s", gatehub.ErrInternal, cardProductCode)
}

func updateCustomerSourceIDs(ctx context.Context, b Backends, userID, customerSourceID, accountSourceID string, shouldOrderPlastic bool) error {
	_, err := b.DB().ExecContext(ctx, "UPDATE gatehub_users SET external_customer_source_id = $1, external_account_source_id = $2, is_first_card_plastic = $3 WHERE external_id = $4;", customerSourceID, accountSourceID, shouldOrderPlastic, userID)
	if err != nil {
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	return nil
}

func updateCustomerIDsReturningPlasticFlag(ctx context.Context, b Backends, userID, customerID, accountID string) (sql.NullBool, error) {
	var flag sql.NullBool
	err := b.DB().GetContext(ctx, &flag, "UPDATE gatehub_users SET external_customer_id = $1, external_account_id = $2 WHERE external_id = $3 RETURNING is_first_card_plastic;", customerID, accountID, userID)
	if err != nil {
		return sql.NullBool{}, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return flag, nil
}

func updateFirstCardProcessTime(ctx context.Context, b Backends, userID string) error {
	_, err := b.DB().ExecContext(ctx, "UPDATE gatehub_users SET first_card_processed_at = NOW(), updated_at = NOW() WHERE external_id = $1", userID)
	if err != nil {
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return nil
}

func LinkUserToGatewayByWalletID(ctx context.Context, b Backends, ec external.Client, walletID string) error {
	wo := client.StartWorkflowOptions{
		ID:                    "gatehub_link_user_to_gateway_" + walletID,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	var workflowStatus enums.WorkflowExecutionStatus
	wflow, err := b.Temporal().DescribeWorkflowExecution(ctx, wo.ID, "")
	switch err.(type) {
	case *serviceerror.Internal,
		*serviceerror.Unavailable,
		*serviceerror.InvalidArgument:
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	case *serviceerror.NotFound:
		// do nothing
	default:
		if wflow != nil {
			workflowStatus = wflow.GetWorkflowExecutionInfo().Status
		}
	}

	externalUserID, err := getExternalUserID(ctx, b, walletID)
	// return workflow if it's running
	var await client.WorkflowRun
	var executeErr error
	if workflowStatus == enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
		await = b.Temporal().GetWorkflow(ctx, wo.ID, "")
	} else {
		await, executeErr = b.Temporal().ExecuteWorkflow(ctx, wo, LinkGatehubUserToGatewayWorkflow, externalUserID)
	}
	if executeErr != nil {
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return await.Get(ctx, nil)
}

func LinkUserToGatewayByExternalID(ctx context.Context, ec external.Client, externalID string) error {
	return ec.LinkUserToGateway(ctx, externalID)
}

func GetCardToken(ctx context.Context, b Backends, ec external.Client, args gatehub.GetCardTokenArgs) (*external.TokenResponse, error) {
	return ec.GetCardToken(ctx, args.UserID, args.TokenType, external.GetCardTokenArgs{
		CardID:    args.CardID,
		PublicKey: args.PublicKey,
	})
}

func FreezeCard(ctx context.Context, b Backends, ec external.Client, args gatehub.FreezeCardArgs) error {
	return ec.FreezeCard(ctx, args.UserID, args.CardID, external.FreezeCardArgs{
		ReasonCode: args.ReasonCode,
		Note:       args.Note,
	})
}

func UnfreezeCard(ctx context.Context, b Backends, ec external.Client, args gatehub.UnfreezeCardArgs) error {
	return ec.UnfreezeCard(ctx, args.UserID, args.CardID, external.UnfreezeCardArgs{
		Note: args.Note,
	})
}

func BlockCard(ctx context.Context, b Backends, ec external.Client, args gatehub.BlockCardArgs) error {
	return ec.BlockCard(ctx, args.UserID, args.CardID, external.BlockCardArgs{
		ReasonCode: args.ReasonCode,
	})
}

func GetPendingThreeDSConfirmations(ctx context.Context, ec external.Client, userID string) ([]external.PendingThreeDSConfirmation, error) {
	return ec.GetPendingThreeDSConfirmations(ctx, userID)
}

func ThreeDSPaymentConfirmation(ctx context.Context, ec external.Client, userID, txID string, confirmed bool) error {
	return ec.ThreeDSPaymentConfirmation(ctx, userID, external.ThreeDSPaymentConfirmationArgs{
		TransactionID: txID,
		Confirmed:     confirmed,
		AuthMethod:    "08", // 08 - login ; 07 - biometric
	})
}

func getNameOnCard(walletAddress string) (string, error) {
	// For non production environments we generate a random card name.
	// This might be counter intuitive but we have the limitation of 26 chars
	// for the name on the card.
	//
	// For production we already now the exact length of the wallet address:
	//	* $ilp.link/ - 10 charactes
	//  * walletAddressPath - 16 characters
	//
	// For the other environments (sandbox.ilp.link, local.ilp.link, etc...), this
	// is going to be hard to manage. Therefore we generate a unique name
	// that fulfills the requirements.
	if !env.IsProd() {
		url := env.OpenPaymentsURL()
		url = strings.Replace(url, "https://", "", 1)
		url = gatehub.DollarSignPlaceholder + url + "/"

		if gatehub.NameOnCardMaxLength <= len(url) {
			return url[:gatehub.NameOnCardMaxLength], nil
		}

		b := make([]byte, 32)
		_, err := rand.Read(b)
		if err != nil {
			return "", fmt.Errorf("%w could not generate random string for name on card: %s", gatehub.ErrInternal, err)
		}

		randomStr := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
		randomStr = strings.ToLower(randomStr)

		nameOnCard := url + randomStr
		if len(nameOnCard) > gatehub.NameOnCardMaxLength {
			nameOnCard = nameOnCard[:gatehub.NameOnCardMaxLength]
		}

		return nameOnCard, nil
	}

	nameOnCard := gatehub.DollarSignPlaceholder + walletAddress
	if len(nameOnCard) > gatehub.NameOnCardMaxLength {
		return "", fmt.Errorf("%w attempted to order card with a name that is longer than %d", gatehub.ErrInternal, gatehub.NameOnCardMaxLength)
	}

	return nameOnCard, nil
}

func BackfillAccountAndSetKYC(ctx context.Context, b Backends, walletID, webhookID string) error {
	wo := client.StartWorkflowOptions{
		ID:                    "gatehub_backfill_account_" + webhookID,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
	}

	var workflowStatus enums.WorkflowExecutionStatus

	wflow, err := b.Temporal().DescribeWorkflowExecution(ctx, wo.ID, "")
	switch err.(type) {
	case *serviceerror.Internal,
		*serviceerror.Unavailable,
		*serviceerror.InvalidArgument:

		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	case *serviceerror.NotFound:
		// do nothing
	default:
		if wflow != nil {
			workflowStatus = wflow.GetWorkflowExecutionInfo().Status
		}
	}

	if workflowStatus != enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
		_, err := b.Temporal().ExecuteWorkflow(ctx, wo, BackfillAccountWorkflow, walletID)
		if err != nil {
			return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
		}
	}

	return nil
}

type pendingTransaction struct {
	ID     string `db:"id"`
	UserID string `db:"user_id"`
}

func GetPendingClearingCardTransactions(ctx context.Context, b Backends) ([]pendingTransaction, error) {
	sqlStmt := "SELECT id, user_id FROM gatehub_card_transactions WHERE status NOT IN ($1, $2) AND type NOT IN ($3, $4) AND raw_transaction ->> 'transactionClassification' != $5 AND (raw_transaction ->> 'createdAt')::timestamptz::date < CURRENT_DATE ORDER BY created_at ASC"
	sqlArgs := []any{
		external.CardTransactionStatusCompleted,
		external.CardTransactionStatusFailed,
		external.CardTransactionTypeTransferToAccount,
		external.CardTransactionTypeTransferFromAccount,
		external.CardTransactionClassificationReversal,
	}
	return getPendingCardTransactions(ctx, b, sqlStmt, sqlArgs)
}

func GetPendingRealtimeCardTransactions(ctx context.Context, b Backends) ([]pendingTransaction, error) {
	sqlStmt := "SELECT id, user_id FROM gatehub_card_transactions WHERE status NOT IN ($1, $2) AND (type IN ($3, $4) OR raw_transaction ->> 'transactionClassification' = $5) ORDER BY created_at ASC"
	sqlArgs := []any{
		external.CardTransactionStatusCompleted,
		external.CardTransactionStatusFailed,
		external.CardTransactionTypeTransferToAccount,
		external.CardTransactionTypeTransferFromAccount,
		external.CardTransactionClassificationReversal,
	}
	return getPendingCardTransactions(ctx, b, sqlStmt, sqlArgs)
}

func getPendingCardTransactions(ctx context.Context, b Backends, sqlStmt string, sqlArgs []any) ([]pendingTransaction, error) {
	var txs []pendingTransaction
	if err := b.DB().SelectContext(ctx, &txs, sqlStmt, sqlArgs...); err != nil {
		return nil, fmt.Errorf("%w %s", transactions.ErrInternal, err)
	}
	return txs, nil
}

func UpdateOrganizationConfiguration(ctx context.Context, ec external.Client, apiBaseURL, twoFAType string) (*external.UpdateOrganizationConfigurationResponse, error) {
	baseURL, err := url.Parse(apiBaseURL)
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	if !env.IsLocal() && baseURL.Scheme != "https" {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, "api url must be https")
	}

	if env.IsLocal() && baseURL.Scheme != "http" && baseURL.Scheme != "https" {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, "api url must be http or https")
	}

	t, err := external.ParseTwoFA(twoFAType)
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return ec.UpdateOrganizationConfiguration(ctx, external.UpdateOrganizationConfigurationArgs{
		APIBaseURL: baseURL.String(),
		TwoFAType:  t,
	})
}

func updateCardTransactionStatus(ctx context.Context, b Backends, txID, status string) error {
	_, err := b.DB().ExecContext(ctx, "UPDATE gatehub_card_transactions SET status = $1, updated_at = NOW() WHERE id = $2", status, txID)
	if err != nil {
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	return nil
}
