package ops

import (
	"errors"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/gatehub/external"
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

func LinkGatehubUserToGatewayWorkflow(ctx workflow.Context, externalUserID string) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Link gatehub user to gateway.")
	err := workflow.ExecuteActivity(ctx, a.LinkGatehubUserToGateway, externalUserID).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
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
	var providerFeeValue int64
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

	err = workflow.ExecuteActivity(ctx, a.FinalizeGatehubDeposit, txID, walletID, amt, providerFee).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	return txID, nil
}

// TODO: remove once old withdrawal workflows are drained from the Temporal queue
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

func BackfillAccountWorkflow(ctx workflow.Context, walletID string) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Backfill gatehub account.")

	// The activity will check if the config has SendingUserID set
	var externalID string
	err := workflow.ExecuteActivity(ctx, a.CheckIfBackfillWasDone, walletID).Get(ctx, &externalID)
	if err != nil {
		return err
	}

	if externalID != "" {
		var balance gatehub.Balance
		err = workflow.ExecuteActivity(ctx, a.BackfillPaywiserBalanceAfterKYC, walletID).Get(ctx, &balance)
		if err != nil {
			return err
		}

		err = workflow.ExecuteActivity(ctx, a.MarkBackfillUser, walletID, externalID, balance).Get(ctx, nil)
		if err != nil {
			// if this errors we need to go in manually and update the table will
			return err
		}
	}

	err = workflow.ExecuteActivity(ctx, a.SetKYCApprovedForGatehub, walletID).Get(ctx, nil)
	if err != nil {
		// if this  we need to go in manually and update the table will
		return err
	}
	return nil
}

func NotifyWithdrawalSCTITimeoutWorkflow(ctx workflow.Context, wh MoreBridgeWithdrawalSCTITimeoutWebhook) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Notify withdrawal SCTI timeout.")

	var walletID string
	err := workflow.ExecuteActivity(ctx, a.GetWalletFromGatehubUser, wh.UserID).Get(ctx, &walletID)
	if err != nil {
		return err
	}

	var tx transactions.Transaction
	err = workflow.ExecuteActivity(ctx, a.GetGateHubTransactionByForeignID, walletID, wh.Data.TransactionID).Get(ctx, &tx)
	if err != nil {
		return err
	}

	amount := fmt.Sprintf("%s %s", wh.Data.Amount, wh.Data.Currency)
	err = workflow.ExecuteActivity(ctx, a.SendWithdrawalSCTITimeoutEmail, tx.ID, walletID, amount, wh.Data.CounterpartyName, wh.Data.CounterpartyIBAN, wh.Data.Timestamp).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}


func CompleteGatehubWithdrawalWorkflow(ctx workflow.Context, userID, externalTxID string) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Completing gatehub withdrawal.", "externalTxID", externalTxID)

	var walletID string
	if err := workflow.ExecuteActivity(ctx, a.GetWalletFromGatehubUser, userID).Get(ctx, &walletID); err != nil {
		return err
	}

	var internalTx *transactions.Transaction
	err := workflow.ExecuteActivity(ctx, a.GetGateHubTransactionByForeignID, walletID, externalTxID).Get(ctx, &internalTx)
	if err != nil {
		return err
	}

	if err = workflow.ExecuteActivity(ctx, a.FinalizeGatehubWithdrawal, internalTx.ID).Get(ctx, nil); err != nil {
		return err
	}

	return nil
}

func RejectGatehubWithdrawalWorkflow(ctx workflow.Context, wh MoreBridgeWithdrawalRejectedWebhookData) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Rejecting gatehub withdrawal.", "externalTxID", wh.TxID)

	var walletID string
	if err := workflow.ExecuteActivity(ctx, a.GetWalletFromGatehubUser, wh.UserID).Get(ctx, &walletID); err != nil {
		return err
	}

	var internalTx *transactions.Transaction
	err := workflow.ExecuteActivity(ctx, a.GetGateHubTransactionByForeignID, walletID, wh.TxID).Get(ctx, &internalTx)
	if err != nil {
		return err
	}

	if err = workflow.ExecuteActivity(ctx, a.RollbackGatehubWithdrawal, internalTx.ID).Get(ctx, nil); err != nil {
		return err
	}

	var externalTx *external.Transaction
	if err = workflow.ExecuteActivity(ctx, a.FetchGatehubTransaction, wh.UserID, wh.TxID).Get(ctx, &externalTx); err != nil {
		return err
	}

	return workflow.ExecuteActivity(ctx, a.SendWithdrawalRejectedEmail, internalTx.ID, walletID, wh.Amount, wh.Currency, wh.IBAN, externalTx.Account.LegalName).Get(ctx, nil)
}

func NotifyWithdrawalReroutedWorkflow(ctx workflow.Context, wh MoreBridgeWithdrawalReroutedWebhook) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Notify withdrawal SCT reroute.")

	var walletID string
	err := workflow.ExecuteActivity(ctx, a.GetWalletFromGatehubUser, wh.UserID).Get(ctx, &walletID)
	if err != nil {
		return err
	}

	var tx transactions.Transaction
	err = workflow.ExecuteActivity(ctx, a.GetGateHubTransactionByForeignID, walletID, wh.Data.ID).Get(ctx, &tx)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.SendWithdrawalSCTITimeoutEmail, tx.ID, walletID).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}