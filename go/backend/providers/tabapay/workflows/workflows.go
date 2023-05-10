package workflows

import (
	"errors"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/basistheory"
	httplog "gitlab.com/fynbos/backend/providers/http"
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

	var tokenizedCard basistheory.Card
	err := workflow.ExecuteActivity(ctx, a.CreateBasisTheoryCard, args.WalletID, args.BasisTheoryTokenID).Get(ctx, &tokenizedCard)
	if err != nil {
		logger.Error("Failed to create basis theory card.")
		return nil, err
	}

	var la linkedaccounts.LinkedAccount
	err = workflow.ExecuteActivity(ctx, a.MarkCardNotDeleted, tokenizedCard.ID).Get(ctx, &la)
	var applicationError *temporal.ApplicationError
	if errors.As(err, &applicationError) && applicationError.Type() != "NotFound" {
		return nil, err
	}
	if la.ID == args.BasisTheoryTokenID {
		return &la, nil
	}

	newCtx := workflow.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
		Context: fmt.Sprintf("walletID=%s", args.WalletID),
	})
	var externalAccount external.CreateAccountResponse
	err = workflow.ExecuteActivity(newCtx, a.CreateExternalCard, CreateExternalCardArgs{
		WalletID:       args.WalletID,
		CardNumber:     fmt.Sprintf("{{ %s | json: '$.number' }}", tokenizedCard.TokenID),
		ExpirationDate: fmt.Sprintf("{{ %s | json: '$.expiration_year' | to_string }}{{ %s | json: '$.expiration_month' | pad_left: 2,'0' }}", tokenizedCard.TokenID, tokenizedCard.TokenID),
	}).Get(ctx, &externalAccount)
	if err != nil {
		logger.Error("Failed to create card on tabapay.")
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.CreateLinkedCard, CreateLinkedCardArgs{
		ID:         tokenizedCard.ID,
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
