package workflows

import (
	"time"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
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

	var linkedAccountID string
	err := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		return uuid.NewString()
	}).Get(&linkedAccountID)
	if err != nil {
		logger.Error("Failed to generate linkedAccountID.")
		return nil, err
	}

	var externalAccount external.CreateAccountResponse
	err = workflow.ExecuteActivity(ctx, a.CreateExternalCard, CreateExternalCardArgs{
		LinkedAccountID: linkedAccountID,
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
		ID:         linkedAccountID,
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
