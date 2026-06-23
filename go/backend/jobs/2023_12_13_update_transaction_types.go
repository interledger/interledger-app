package jobs

import (
	"context"
	"github.com/interledger/interledger-app/go/backend/transactions"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func UpdateTransactionTypes(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	err := workflow.ExecuteActivity(ctx, a.UpdateTransactionsIncomingToReceived).Get(ctx, nil)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.UpdateTransactionsOutgoingToSent).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) UpdateTransactionsIncomingToReceived(ctx context.Context) error {
	_, err := a.b.DB().ExecContext(
		ctx,
		"UPDATE backend.public.transactions SET type = $1 where backend.public.transactions.type = 'incoming'",
		transactions.TransactionTypeReceived,
	)
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) UpdateTransactionsOutgoingToSent(ctx context.Context) error {
	_, err := a.b.DB().ExecContext(
		ctx,
		"UPDATE backend.public.transactions SET type = $1 where backend.public.transactions.type = 'outgoing'",
		transactions.TransactionTypeSent,
	)
	if err != nil {
		return err
	}

	return nil
}
