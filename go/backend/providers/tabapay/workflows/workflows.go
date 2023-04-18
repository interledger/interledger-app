package workflows

import (
	"errors"
	"time"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func CreateTabapayCardWorkflow(ctx workflow.Context, args tabapay.CreateCardArgs) (*linkedaccounts.LinkedAccount, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating tabapay card.")

	var la linkedaccounts.LinkedAccount
	err := workflow.ExecuteActivity(ctx, a.MarkCardNotDeleted, args.IdempotencyKey).Get(ctx, &la)
	var applicationError *temporal.ApplicationError
	if errors.As(err, &applicationError) && applicationError.Type() != "NotFound" {
		return nil, err
	}
	if la.ID == args.IdempotencyKey {
		return &la, nil
	}

	var externalAccount external.CreateAccountResponse
	err = workflow.ExecuteActivity(ctx, a.CreateExternalCard, CreateExternalCardArgs{
		LinkedAccountID: args.IdempotencyKey,
		WalletID:        args.WalletID,
		Name:            args.Name,
		CardNumber:      args.CardNumber,
		CVV:             args.CVV,
		ExpirationDate:  args.ExpirationDate,
	}).Get(ctx, &externalAccount)
	if err != nil {
		logger.Error("Failed to create card on tabapay.")
		return nil, err
	}

	last4 := args.CardNumber
	if len(args.CardNumber) > 4 {
		last4 = last4[len(args.CardNumber)-4:]
	}
	err = workflow.ExecuteActivity(ctx, a.CreateLinkedCard, CreateLinkedCardArgs{
		ID:         args.IdempotencyKey,
		WalletID:   args.WalletID,
		ProviderID: externalAccount.AccountID,
		Mask:       last4,
		Name:       args.Name,
		Nickname:   args.Name,
	}).Get(ctx, &la)
	if err != nil {
		logger.Error("Failed to create linked account.")
		return nil, err
	}

	logger.Info("Created tabapay card.")

	return &la, nil
}
