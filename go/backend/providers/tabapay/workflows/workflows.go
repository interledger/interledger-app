package workflows

import (
	"errors"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/basistheory"
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
	err := workflow.ExecuteActivity(ctx, a.MarkCardNotDeleted, args.BasisTheoryCardID).Get(ctx, &la)
	var applicationError *temporal.ApplicationError
	if errors.As(err, &applicationError) && applicationError.Type() != "NotFound" {
		return nil, err
	}
	if la.ID == args.BasisTheoryCardID {
		return &la, nil
	}

	var tokenizedCard basistheory.Card
	err = workflow.ExecuteActivity(ctx, a.GetBasisTheoryCard, args.BasisTheoryCardID).Get(ctx, &tokenizedCard)
	if err != nil {
		logger.Error("Failed to get basis theory card.")
		return nil, err
	}
	if tokenizedCard.WalletID != args.WalletID {
		logger.Error("Card does not belong to wallet.")
		return nil, err
	}

	var externalAccount external.CreateAccountResponse
	err = workflow.ExecuteActivity(ctx, a.CreateExternalCard, CreateExternalCardArgs{
		BasisTheoryCardID: args.BasisTheoryCardID,
		WalletID:          args.WalletID,
		CardNumber:        fmt.Sprintf("{{ %s | json: '$.number' }}", tokenizedCard.TokenID),
		ExpirationDate:    fmt.Sprintf("{{ %s | json: '$.expiration_year''$.expiration_month' }}", tokenizedCard.TokenID),
	}).Get(ctx, &externalAccount)
	if err != nil {
		logger.Error("Failed to create card on tabapay.")
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.CreateLinkedCard, CreateLinkedCardArgs{
		ID:         args.BasisTheoryCardID,
		WalletID:   args.WalletID,
		ProviderID: externalAccount.AccountID,
		Mask:       tokenizedCard.TokenizedNumber,
		Name:       tokenizedCard.TokenizedNumber,
		Nickname:   tokenizedCard.TokenizedNumber,
	}).Get(ctx, &la)
	if err != nil {
		logger.Error("Failed to create linked account.")
		return nil, err
	}

	logger.Info("Created tabapay card.")

	return &la, nil
}
