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
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating tabapay card.")

	var cardInfo external.QueryCardResponse
	err := workflow.ExecuteActivity(ctx, a.QueryCard, QueryCard{
		WalletID:       args.WalletID,
		CardNumber:     fmt.Sprintf("{{ %s | json: '$.number' }}", args.BasisTheoryTokenID),
		ExpirationDate: fmt.Sprintf("{{ %s | json: '$.expiration_year' | to_string }}{{ %s | json: '$.expiration_month' | pad_left: 2,'0' }}", args.BasisTheoryTokenID, args.BasisTheoryTokenID),
		AVS:            true,
	}).Get(ctx, &cardInfo)
	if err != nil {
		logger.Error("Failed to query card.")
		return nil, err
	}

	if !cardInfo.Card.Push.Enabled && !cardInfo.Card.Pull.Enabled {
		logger.Info("Unsupported card. Push and pull not enabled.")
		return nil, temporal.NewNonRetryableApplicationError("Unsupported card. Push and pull not enabled.", "ErrUnsupportedCard", fmt.Errorf("%w Unsupported card.", tabapay.ErrInternal))
	}

	// https://developers.tabapay.com/reference/avs-response-codes
	linkedAccountState := linkedaccounts.Verified
	if cardInfo.AVS.CodeAVS != external.AVSResponseCodeY && cardInfo.AVS.CodeAVS != external.AVSResponseCodeA {
		logger.Warn("AVS failed.", "AVSCode", cardInfo.AVS.CodeAVS)
		linkedAccountState = linkedaccounts.OwnershipReviewRequired
	}

	var tokenizedCard basistheory.Card
	err = workflow.ExecuteActivity(ctx, a.CreateBasisTheoryCard, args.WalletID, args.BasisTheoryTokenID).Get(ctx, &tokenizedCard)
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
	if la.ID == tokenizedCard.ID {
		return &la, nil
	}

	newCtx := workflow.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
		Context: fmt.Sprintf("walletID=%s", args.WalletID),
	})
	newCtx = workflow.WithActivityOptions(newCtx, workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 2, // so we don't get blocked by Tabapay
		},
	})
	var externalAccount external.CreateAccountResponse
	err = workflow.ExecuteActivity(newCtx, a.CreateExternalCard, CreateExternalCardArgs{
		WalletID:            args.WalletID,
		RejectDuplicateCard: true,
		CardNumber:          fmt.Sprintf("{{ %s | json: '$.number' }}", tokenizedCard.TokenID),
		ExpirationDate:      fmt.Sprintf("{{ %s | json: '$.expiration_year' | to_string }}{{ %s | json: '$.expiration_month' | pad_left: 2,'0' }}", tokenizedCard.TokenID, tokenizedCard.TokenID),
	}).Get(ctx, &externalAccount)
	if err != nil {
		logger.Error("Failed to create card on tabapay.")
		return nil, err
	}

	mask := cardInfo.Card.Last4
	var network string
	if cardInfo.Card.Push.Network != "" {
		network = cardInfo.Card.Push.Network
	}
	if cardInfo.Card.Pull.Network != "" {
		network = cardInfo.Card.Pull.Network
	}
	err = workflow.ExecuteActivity(ctx, a.CreateLinkedCard, CreateLinkedCardArgs{
		ID:         tokenizedCard.ID,
		WalletID:   args.WalletID,
		ProviderID: externalAccount.AccountID,
		Mask:       mask,
		Name:       fmt.Sprintf("%s %s", network, mask),
		Nickname:   fmt.Sprintf("%s %s", network, mask),
		CanSend:    cardInfo.Card.Pull.Enabled,
		CanReceive: cardInfo.Card.Push.Enabled,
		State:      linkedAccountState,
	}).Get(ctx, &la)
	if err != nil {
		logger.Error("Failed to create linked account.")
		return nil, err
	}

	logger.Info("Created tabapay card.")

	return &la, nil
}
