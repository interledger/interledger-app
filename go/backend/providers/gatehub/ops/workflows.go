package ops

import (
	"errors"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/gatehub"
	external "gitlab.com/fynbos/backend/providers/gatehub/external"
	"gitlab.com/fynbos/backend/transactions"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func CreateGatehubUserWorkflow(ctx workflow.Context, walletID string) (string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating gatehub wallet.")

	var externalUserID string
	err := workflow.ExecuteActivity(ctx, a.GetGatehubUser, walletID).Get(ctx, &externalUserID)
	var applicationError *temporal.ApplicationError
	if errors.As(err, &applicationError) && applicationError.Type() == "ErrNotFound" {
		innerErr := workflow.ExecuteActivity(ctx, a.CreateGatehubUser, walletID).Get(ctx, &externalUserID)
		if innerErr != nil {
			return "", innerErr
		}
	} else if err != nil {
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.SaveGatehubUser, walletID, externalUserID).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	var la linkedaccounts.LinkedAccount
	err = workflow.ExecuteActivity(ctx, a.CreateGatehubWalletLinkedAccount, walletID).Get(ctx, &la)
	if err != nil {
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.CreateGatehubBalanceAccount, la.ID).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	return externalUserID, nil
}

func CreateGatehubDeposit(ctx workflow.Context, wh DepositWebhook) (string, error) {
	var a *Activity

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating gatehub deposit.")

	cc := currency.ParseCurrency(wh.Data.Currency)
	if cc != currency.EUR {
		return "", temporal.NewNonRetryableApplicationError("Invalid currency", "ErrInternal", fmt.Errorf("%w invalid currency", gatehub.ErrInternal))
	}

	amountValue, err := StringToScaledUInt(wh.Data.Amount)
	if err != nil {
		return "", temporal.NewNonRetryableApplicationError("Invalid amount", "ErrInternal", fmt.Errorf("%w %s", gatehub.ErrInternal, err))
	}

	//get transaction and fee from it
	var providerFeeValue uint64
	args := &FeeFromGhArgs{UserID: wh.UserID, TrxID: wh.Data.TrxID}
	err = workflow.ExecuteActivity(ctx, a.GetFeeFromGatehubTrasaction, args).Get(ctx, &providerFeeValue)
	if err != nil {
		return "", err
	}

	remainder := amountValue - providerFeeValue

	// calculate fee and remaining amount
	providerFee := currency.Amount{
		Value:    providerFeeValue,
		Currency: cc,
	}
	amt := currency.Amount{
		Value:    remainder,
		Currency: cc,
	}
	fullAMount := currency.Amount{
		Value:    amountValue,
		Currency: cc,
	}

	var walletID string
	err = workflow.ExecuteActivity(ctx, a.GetWalletFromGatehubUser, wh.UserID).Get(ctx, &walletID)
	if err != nil {
		return "", err
	}

	var txID string
	err = workflow.ExecuteActivity(ctx, a.CreateGatehubDepositTransaction, wh.Data.TrxID, walletID, fullAMount, providerFee).Get(ctx, &txID)
	if err != nil {
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.FinalizeGatehubDeposit, txID, walletID, amt).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	return txID, nil
}

func CreateGatehubCoverFee(ctx workflow.Context, wh DepositWebhook, tr external.Transaction) (string, error) {
	var a *Activity

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating gatehub cover fee.")

	cc := currency.ParseCurrency(wh.Data.Currency)
	if cc != currency.EUR {
		return "", temporal.NewNonRetryableApplicationError("Invalid currency", "ErrInternal", fmt.Errorf("%w invalid currency", gatehub.ErrInternal))
	}

	amountValue, err := StringToScaledUInt(wh.Data.Amount)
	if err != nil {
		return "", temporal.NewNonRetryableApplicationError("Invalid amount", "ErrInternal", fmt.Errorf("%w %s", gatehub.ErrInternal, err))
	}

	feeValue, err := StringToScaledUInt(tr.Fee)
	if err != nil {
		return "", temporal.NewNonRetryableApplicationError("Invalid amount", "ErrInternal", fmt.Errorf("%w %s", gatehub.ErrInternal, err))
	}

	amt := currency.Amount{
		Value:    amountValue,
		Currency: cc,
	}
	fee := currency.Amount{
		Value:    feeValue,
		Currency: cc,
	}

	var walletID string
	err = workflow.ExecuteActivity(ctx, a.GetWalletFromGatehubUser, wh.UserID).Get(ctx, &walletID)
	if err != nil {
		return "", err
	}

	var txID string
	err = workflow.ExecuteActivity(ctx, a.CreateGatehubDepositTransaction, wh.Data.TrxID, walletID, amt, fee).Get(ctx, &txID)
	if err != nil {
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.FinalizeGatehubDeposit, txID, walletID, amt).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	// process transaction, set corect title and reference id
	err = workflow.ExecuteActivity(ctx, a.ProcessCoverFeeForDeposit, txID, tr).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	return txID, nil
}

func CheckOpsBalance(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Checking available OPS balance.")

	err := workflow.ExecuteActivity(ctx, a.CheckGatehubOpsBalance).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func ProcessGatehubWithdrawal(ctx workflow.Context, walletID, transactionID string) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating gatehub withdrawal.")

	err := workflow.ExecuteActivity(ctx, a.ReserveGatehubBalance, transactionID, walletID).Get(ctx, nil)
	if err != nil {
		handleFailedWithdrawal(ctx, a, walletID, transactionID)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.CheckGatehubWithdrawalComplete, walletID, transactionID).Get(ctx, nil)
	if err != nil {
		handleFailedWithdrawal(ctx, a, walletID, transactionID)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.FinalizeGatehubBalance, transactionID, walletID).Get(ctx, nil)
	if err != nil {
		handleFailedWithdrawal(ctx, a, walletID, transactionID)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.UpdateGatehubWithdrawalState, walletID, transactionID, transactions.StateCompleted).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func handleFailedWithdrawal(ctx workflow.Context, a *Activity, walletID, transactionID string) {
	logger := workflow.GetLogger(ctx)

	err := workflow.ExecuteActivity(ctx, a.UpdateGatehubWithdrawalState, walletID, transactionID, transactions.StateFailed).Get(ctx, nil)
	if err != nil {
		logger.Error("Unable to update transaction state to failed", transactionID)
	}
}
