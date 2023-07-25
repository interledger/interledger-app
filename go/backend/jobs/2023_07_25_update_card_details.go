package jobs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/providers/basistheory"
	tabapay_external "gitlab.com/fynbos/backend/providers/tabapay/external"
	tabapay_workflow "gitlab.com/fynbos/backend/providers/tabapay/workflows"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func (a *Activity) UpdateCardDetails(ctx context.Context, new basistheory.UpdateCardArgs) (*basistheory.Card, error) {
	return a.b.BasisTheory().UpdateCard(ctx, new)
}

func (a *Activity) ListBasisTheoryCards(ctx context.Context) ([]basistheory.Card, error) {
	return a.b.BasisTheory().ListCards(ctx)
}

func UpdateBasisTheoryCardDetailsWorkflow(ctx workflow.Context) error {
	var a *Activity
	var tabapayActivity *tabapay_workflow.Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)

	var cards []basistheory.Card
	err := workflow.ExecuteActivity(ctx, a.ListBasisTheoryCards).Get(ctx, &cards)
	if err != nil {
		logger.Error("Failed to list basistheory cards.")
		return err
	}

	for _, c := range cards {
		var cardInfo tabapay_external.QueryCardResponse
		err := workflow.ExecuteActivity(ctx, tabapayActivity.QueryCard, tabapay_workflow.QueryCard{
			WalletID:       c.WalletID,
			CardNumber:     fmt.Sprintf("{{ %s | json: '$.number' }}", c.TokenID),
			ExpirationDate: fmt.Sprintf("{{ %s | json: '$.expiration_year' | to_string }}{{ %s | json: '$.expiration_month' | pad_left: 2,'0' }}", c.TokenID, c.TokenID),
			AVS:            false,
		}).Get(ctx, &cardInfo)
		if err != nil {
			logger.Error("Failed to query card on tabapay.")
			return err
		}

		pullNetwork := cardInfo.Card.Pull.Network
		if strings.EqualFold(strings.TrimSpace(pullNetwork), "mastercard") {
			pullNetwork = "Mastercard"
		}
		pushNetwork := cardInfo.Card.Push.Network
		if strings.EqualFold(strings.TrimSpace(pushNetwork), "mastercard") {
			pushNetwork = "Mastercard"
		}
		err = workflow.ExecuteActivity(ctx, a.UpdateCardDetails, basistheory.UpdateCardArgs{
			ID:               c.ID,
			Bin:              cardInfo.Card.Bin,
			PullNetwork:      pullNetwork,
			PullEnabled:      cardInfo.Card.Pull.Enabled,
			PullType:         string(cardInfo.Card.Pull.Type),
			PullCountry:      cardInfo.Card.Pull.Country,
			PushNetwork:      pushNetwork,
			PushEnabled:      cardInfo.Card.Push.Enabled,
			PushType:         string(cardInfo.Card.Push.Type),
			PushAvailability: cardInfo.Card.Push.Availability,
			PushCountry:      cardInfo.Card.Push.Country,
		}).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to update basistheory card.")
			return err
		}
	}

	return nil
}
