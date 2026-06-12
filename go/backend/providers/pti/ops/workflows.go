package ops

import (
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/providers/pti/external"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/wallets"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func CreateUserWorkflow(ctx workflow.Context, walletID string) (*pti.User, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating pti user.")

	var externalUser pti.User
	err := workflow.ExecuteActivity(ctx, a.GetPtiUser, walletID).Get(ctx, &externalUser)
	if err != nil {
		return nil, err
	}

	if externalUser.ID == "" {
		var externalUserID string
		err = workflow.ExecuteActivity(ctx, a.CreatePtiUser, walletID).Get(ctx, &externalUserID)
		if err != nil {
			return nil, err
		}

		err = workflow.ExecuteActivity(ctx, a.SavePtiUser, externalUserID, walletID).Get(ctx, &externalUser)
		if err != nil {
			return nil, err
		}
	}

	return &externalUser, nil
}

func CreateWalletWorkflow(ctx workflow.Context, args pti.CreateWalletArgs) (*linkedaccounts.LinkedAccount, error) {
	if args.Currency != currency.USD {
		return nil, fmt.Errorf("%w Only supports USD wallets", pti.ErrInternal)
	}

	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating pti wallet.")

	var externalUser pti.User
	err := workflow.ExecuteActivity(ctx, a.GetPtiUser, args.WalletID).Get(ctx, &externalUser)
	if err != nil {
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.CheckUserAssessmentAccepted, args.WalletID).Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	externalWalletID := fmt.Sprintf("%s_%s", args.Currency.String(), args.WalletID)
	err = workflow.ExecuteActivity(ctx, a.CreatePtiWallet, pti.CreateExternalWalletArgs{
		ID:       externalWalletID,
		UserID:   externalUser.ExternalID,
		Currency: args.Currency,
	}).Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	var la linkedaccounts.LinkedAccount
	err = workflow.ExecuteActivity(ctx, a.CreatePtiWalletLinkedAccount, linkedaccounts.CreateArgs{
		WalletID:        args.WalletID,
		Nickname:        args.Nickname,
		Name:            args.Title,
		Provider:        pti.ProviderName,
		ProviderID:      externalWalletID,
		Type:            pti.AccTypeBalance,
		CanReceive:      true,
		ReceiveCountry:  country.US,
		ReceiveCurrency: currency.USD,
		SendCountry:     country.US,
		SendCurrency:    currency.USD,
		CanSend:         true,
		State:           linkedaccounts.Verified,
	}).Get(ctx, &la)
	if err != nil {
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.CreatePTIBalanceAccount, la.ID).Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &la, nil
}

func CreateCardWorkflow(ctx workflow.Context, walletID, tokenID string) (*linkedaccounts.LinkedAccount, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating pti card.")

	var linkedAccount linkedaccounts.LinkedAccount
	err := workflow.ExecuteActivity(ctx, a.CreatePTICard, walletID, tokenID).Get(ctx, &linkedAccount)
	if err != nil {
		return nil, err
	}

	return &linkedAccount, nil
}

func CreatePtiBankAccountWorkflow(ctx workflow.Context, args pti.CreateBankAccountArgs) (*linkedaccounts.LinkedAccount, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating pti bank account.")

	var externalBankDetails external.BankAccountPaymentInformation
	err := workflow.ExecuteActivity(ctx, a.CreatePtiBankAccount, args).Get(ctx, &externalBankDetails)
	if err != nil {
		return nil, err
	}

	var linkedAccount linkedaccounts.LinkedAccount
	err = workflow.ExecuteActivity(ctx, a.SavePtiBankAccount, args.WalletID, externalBankDetails.ID).Get(ctx, &linkedAccount)
	if err != nil {
		return nil, err
	}

	return &linkedAccount, nil
}

func SettleDepositWorkflow(ctx workflow.Context, wh pti.TransactionStatusPayload) (string, error) {
	var a *Activity

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating pti deposit.")

	cc := currency.ParseCurrency(wh.Currency)
	if cc != currency.USD {
		return "", temporal.NewNonRetryableApplicationError("Invalid currency", "ErrInternal", fmt.Errorf("%w invalid currency", pti.ErrInternal))
	}

	amt := currency.FromFloat64(wh.Amount, currency.USD)

	var walletID string
	err := workflow.ExecuteActivity(ctx, a.GetWalletFromPTIUser, wh.UserID).Get(ctx, &walletID)
	if err != nil {
		return "", err
	}

	var txID string
	err = workflow.ExecuteActivity(ctx, a.SettleTransaction, wh.RequestID, walletID, amt).Get(ctx, &txID)
	if err != nil {
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.FinalizePTIDeposit, txID, walletID, amt).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.UpdatePaymentState, wh.RequestID, payments.StateCompleted).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	return txID, nil
}

func ReturnedWorkflow(ctx workflow.Context, wh pti.TransactionStatusPayload) (string, error) {
	var a *Activity

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	webhookCurrency := currency.ParseCurrency(wh.Currency)
	transactionTotalCurrency := currency.ParseCurrency(wh.TransactionTotal.Total.Currency)

	// if currency is not provided in webhook, check transaction total currency. If both are not provided or not USD, return error
	// this is to handle a potential issue (communicated to PTI) where the currency is empty string
	if !(webhookCurrency == currency.USD || transactionTotalCurrency == currency.USD) { // if currency is not provided in webhook, check transaction total currency. If both are not provided or not USD, return error
		return "", temporal.NewNonRetryableApplicationError("Invalid currency", "ErrInternal", fmt.Errorf("%w invalid currency", pti.ErrInternal))
	}

	amt := currency.FromFloat64(wh.Amount, currency.USD)
	originalTransactionID := wh.RequestID
	ptiUserID := wh.UserID

	var walletID string
	err := workflow.ExecuteActivity(ctx, a.GetWalletFromPTIUser, ptiUserID).Get(ctx, &walletID)
	if err != nil {
		return "", err
	}

	var returnTransactionID string
	err = workflow.ExecuteActivity(ctx, a.ReturnTransaction, originalTransactionID, walletID, amt).Get(ctx, &returnTransactionID)
	if err != nil {
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.PostTransfer, returnTransactionID, walletID).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	return returnTransactionID, nil
}

func MarkTransactionStateWorkflow(ctx workflow.Context, wh pti.TransactionStatusPayload, state transactions.State) error {
	var a *Activity

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	var walletID string
	if err := workflow.ExecuteActivity(ctx, a.GetWalletFromPTIUser, wh.UserID).Get(ctx, &walletID); err != nil {
		return err
	}

	var tx *transactions.Transaction
	err := workflow.ExecuteActivity(ctx, a.GetTransactionByForeignID, walletID, wh.RequestID).Get(ctx, &tx)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.UpdateTransactionState, walletID, tx.ID, transactions.StateFailed).Get(ctx, nil)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.UpdatePaymentState, wh.RequestID, payments.StateFailed).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func DepositWorkflow(ctx workflow.Context, payment *payments.Payment, wallet *wallets.Wallet) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	err := workflow.ExecuteActivity(ctx, a.ValidateLinkedAccounts, payment, wallet.ID).Get(ctx, nil)
	if err != nil {
		return temporal.NewNonRetryableApplicationError("Invalid linked account", "ErrInternal", err)		
	}

	pitArgs := pti.TransactionArgs{
		PaymentID:       payment.ID,
		WalletID:        wallet.ID,	
		Amount:          payment.ReceiverAmount,
		LinkedAccountID: payment.SenderAccount,
	}
	var externalTxID string
	err = workflow.ExecuteActivity(ctx, a.PTIDeposit, pitArgs).Get(ctx, &externalTxID)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.CreateTransaction, wallet.ID, payment, externalTxID).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func ReservePtiBalance(ctx workflow.Context, walletID, transactionID string) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating gatehub withdrawal.")
	err := workflow.ExecuteActivity(ctx, a.ReservePTIBalance, walletID, transactionID).Get(ctx, nil)
	if err != nil {
		handlePtiFailedWithdrawal(ctx, a, walletID, transactionID)
		return err
	}

	return nil
}

func RevertWithdrawWorkflow(ctx workflow.Context, wh pti.TransactionStatusPayload) error {
	var a *Activity

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating pti deposit.")

	cc := currency.ParseCurrency(wh.Currency)
	if cc != currency.USD {
		return temporal.NewNonRetryableApplicationError("Invalid currency", "ErrInternal", fmt.Errorf("%w invalid currency", pti.ErrInternal))
	}

	amt := currency.FromFloat64(wh.Amount, currency.USD)

	var walletID string
	err := workflow.ExecuteActivity(ctx, a.GetWalletFromPTIUser, wh.UserID).Get(ctx, &walletID)
	if err != nil {
		return err
	}

	var tx *transactions.Transaction
	err = workflow.ExecuteActivity(ctx, a.GetTransactionByForeignID, walletID, wh.RequestID).Get(ctx, &tx)
	if err != nil {
		return err
	}

	var txID string
	err = workflow.ExecuteActivity(ctx, a.RollbackPTIBalance, tx.ID, walletID, amt).Get(ctx, &txID)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.UpdateTransactionState, walletID, tx.ID, transactions.StateFailed).Get(ctx, nil)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.UpdatePaymentState, wh.RequestID, payments.StateFailed).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func SettleWithdrawWorkflow(ctx workflow.Context, wh pti.TransactionStatusPayload) (string, error) {
	var a *Activity

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating pti withdraw.")
	curr := wh.Currency
	if curr == "" {
		curr = wh.TransactionTotal.Total.Currency
	}
	if curr == "" {
		curr = wh.Total.Total.Currency
	}

	cc := currency.ParseCurrency(curr)
	if cc != currency.USD {
		return "", temporal.NewNonRetryableApplicationError("Invalid currency", "ErrInternal", fmt.Errorf("%w invalid currency", pti.ErrInternal))
	}

	amt := currency.FromFloat64(wh.Amount, currency.USD)

	var walletID string
	err := workflow.ExecuteActivity(ctx, a.GetWalletFromPTIUser, wh.UserID).Get(ctx, &walletID)
	if err != nil {
		return "", err
	}
	// if payment is complied returns error
	err = workflow.ExecuteActivity(ctx, a.CheckPaymentState, wh.RequestID).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	var tx *transactions.Transaction
	err = workflow.ExecuteActivity(ctx, a.GetTransactionByForeignID, walletID, wh.RequestID).Get(ctx, &tx)
	if err != nil {
		return "", err
	}

	var txID string
	err = workflow.ExecuteActivity(ctx, a.FinalizePTIBalance, tx.ID, walletID, amt).Get(ctx, &txID)
	if err != nil {
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.UpdateTransactionState, walletID, tx.ID, transactions.StateCompleted).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.UpdatePaymentState, wh.RequestID, payments.StateCompleted).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	return txID, nil
}

func handlePtiFailedWithdrawal(ctx workflow.Context, a *Activity, walletID, transactionID string) {
	logger := workflow.GetLogger(ctx)

	err := workflow.ExecuteActivity(ctx, a.UpdateTransactionState, walletID, transactionID, transactions.StateFailed).Get(ctx, nil)
	if err != nil {
		logger.Error("Unable to update transaction state to failed", transactionID)
	}
}
