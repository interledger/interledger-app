package ops

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/gatehub"
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

	parts := strings.Split(wh.Data.Amount, ".")
	if len(parts) < 1 {
		return "", temporal.NewNonRetryableApplicationError("Invalid amount", "ErrInternal", fmt.Errorf("%w invalid amount", gatehub.ErrInternal))
	}

	value, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return "", temporal.NewNonRetryableApplicationError("Invalid amount", "ErrInternal", fmt.Errorf("%w %s", gatehub.ErrInternal, err))
	}
	amt := currency.Amount{
		Value:    value * 100, // EUR scale = 2
		Currency: cc,
	}

	var walletID string
	err = workflow.ExecuteActivity(ctx, a.GetWalletFromGatehubUser, wh.UserID).Get(ctx, &walletID)
	if err != nil {
		return "", err
	}

	var txID string
	err = workflow.ExecuteActivity(ctx, a.CreateGatehubDepositTransaction, wh.ID, walletID, amt).Get(ctx, &txID)
	if err != nil {
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.FinalizeGatehubDeposit, txID, walletID, amt).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	return txID, nil
}

func CreateGatehubWithdrawal(ctx workflow.Context, walletID, externalTransactionID string) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating gatehub withdrawal.")

	err := workflow.ExecuteActivity(ctx, a.ValidateGatehubWithdrawal, walletID, externalTransactionID).Get(ctx, nil)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.CheckGatehubTransactionComplete, walletID, externalTransactionID).Get(ctx, nil)
	if err != nil {
		return err
	}

	var trxID string
	err = workflow.ExecuteActivity(ctx, a.CreateGatehubWithdrawalTransaction, walletID, externalTransactionID).Get(ctx, &trxID)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.AssignGatehubWithdrawal, trxID, walletID).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}
