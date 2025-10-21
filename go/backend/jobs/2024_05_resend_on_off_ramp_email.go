package jobs

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/providers/xago"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func ResendOnOffRampEmailJob(ctx workflow.Context, paymentID string) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	err := workflow.ExecuteActivity(ctx, a.JobSendDepositEmail, paymentID).Get(ctx, nil)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.JobSendWithdrawalEmail, paymentID).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) JobSendDepositEmail(ctx context.Context, paymentID string) error {
	payment, err := a.b.Payments().Lookup(ctx, paymentID)
	if err != nil {
		if errors.Is(err, payments.ErrNotFound) {
			return temporal.NewNonRetryableApplicationError(err.Error(), "ErrNotFound", err)
		}
		return err
	}

	if payment.Type != payments.TypeDeposit {
		return nil
	}

	la, err := a.b.LinkedAccounts().Get(ctx, payment.SenderAccount)
	if err != nil {
		if errors.Is(err, linkedaccounts.ErrNotFound) {
			return temporal.NewNonRetryableApplicationError(err.Error(), "ErrNotFound", err)
		}
		return err
	}

	sourceAccountName := la.Name
	if la.Provider == pti.ProviderName && la.Type == pti.TypeCard {
		sourceAccountName = "Card ending " + strings.Replace(la.Mask, "*", "", -1)
	}
	if la.Provider == xago.ProviderName && la.Type == xago.AccTypeBank {
		sourceAccountName = fmt.Sprintf("%s %s", la.ReceiveNetwork, la.Mask)
	}

	if payment.State == payments.StateCompleted {
		a.b.Email().SendDepositReceivedEmail(ctx, la.WalletID, payment.SenderAmount, sourceAccountName, payment.UpdatedAt.Format("02 Jan 2006"))
	}

	return nil
}

func (a *Activity) JobSendWithdrawalEmail(ctx context.Context, paymentID string) error {
	payment, err := a.b.Payments().Lookup(ctx, paymentID)
	if err != nil {
		if errors.Is(err, payments.ErrNotFound) {
			return temporal.NewNonRetryableApplicationError(err.Error(), "ErrNotFound", err)
		}
		return err
	}

	if payment.Type != payments.TypeWithdrawal {
		return nil
	}

	la, err := a.b.LinkedAccounts().Get(ctx, payment.ReceiverAccount)
	if err != nil {
		if errors.Is(err, linkedaccounts.ErrNotFound) {
			return temporal.NewNonRetryableApplicationError(err.Error(), "ErrNotFound", err)
		}
		return err
	}

	destinationAccountName := la.Name
	if la.Provider == pti.ProviderName && la.Type == pti.TypeCard {
		destinationAccountName = "Card ending " + strings.Replace(la.Mask, "*", "", -1)
	}
	if la.Provider == xago.ProviderName && la.Type == xago.AccTypeBank {
		destinationAccountName = fmt.Sprintf("%s %s", la.ReceiveNetwork, la.Mask)
	}

	if payment.State == payments.StateCompleted {
		a.b.Email().SendWithdrawalEmail(ctx, la.WalletID, payment.SenderAmount, destinationAccountName, payment.UpdatedAt.Format("02 Jan 2006"))
	}
	return nil
}
