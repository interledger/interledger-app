package workflows

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/basistheory"
	httplog "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
	"gitlab.com/fynbos/env"
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
		AVS:            args.AVS,
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
	if args.AVS && cardInfo.AVS.CodeAVS != external.AVSResponseCodeY && cardInfo.AVS.CodeAVS != external.AVSResponseCodeA {
		logger.Warn("AVS failed.", "AVSCode", cardInfo.AVS.CodeAVS)
		linkedAccountState = linkedaccounts.OwnershipReviewRequired
	}

	pullNetwork := cardInfo.Card.Pull.Network
	if strings.EqualFold(strings.TrimSpace(pullNetwork), "mastercard") {
		pullNetwork = "Mastercard"
	}
	pushNetwork := cardInfo.Card.Push.Network
	if strings.EqualFold(strings.TrimSpace(pushNetwork), "mastercard") {
		pushNetwork = "Mastercard"
	}
	var tokenizedCard basistheory.Card
	err = workflow.ExecuteActivity(ctx, a.CreateBasisTheoryCard, basistheory.CreateCardArgs{
		WalletID:         args.WalletID,
		TokenID:          args.BasisTheoryTokenID,
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
	}).Get(ctx, &tokenizedCard)
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

	var tabapayReferenceID string
	err = workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		if args.TabapayReferenceID != "" {
			return args.TabapayReferenceID
		}

		return tabapay.NewReferenceID()
	}).Get(&tabapayReferenceID)
	if err != nil {
		logger.Error("Failed to generate tabapay referenceID.")
		return nil, err
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
		RejectDuplicateCard: !env.IsDev(),
		CardNumber:          fmt.Sprintf("{{ %s | json: '$.number' }}", tokenizedCard.TokenID),
		ExpirationDate:      fmt.Sprintf("{{ %s | json: '$.expiration_year' | to_string }}{{ %s | json: '$.expiration_month' | pad_left: 2,'0' }}", tokenizedCard.TokenID, tokenizedCard.TokenID),
		ReferenceID:         tabapayReferenceID,
	}).Get(ctx, &externalAccount)
	if err != nil {
		logger.Error("Failed to create card on tabapay.")
		return nil, err
	}

	// check for duplicate linked cards
	var externalAccountID string
	for _, providerID := range externalAccount.DuplicateAccountIDs {
		var las []linkedaccounts.LinkedAccount
		externalAccountID = providerID
		// list linked accounts whether they are soft deleted
		err = workflow.ExecuteActivity(ctx, a.ListLinkedAccountsByProviderID, tabapay.ProviderName, providerID).Get(ctx, &las)
		if err != nil {
			logger.Error("Failed to check for duplicate linked cards.")
			return nil, err
		}

		for _, la := range las {
			if la.WalletID != args.WalletID {
				logger.Error("Duplicate card found.")
				return nil, temporal.NewNonRetryableApplicationError("tabapay: Duplicate card.", "ErrDuplicateCard", nil)
			} else {
				return &la, nil
			}
		}
	}
	// if no duplicate linked cards found, create a new linked account
	if externalAccountID == "" {
		externalAccountID = externalAccount.AccountID
	}

	mask := cardInfo.Card.Last4
	network := pullNetwork
	if network == "" {
		network = pushNetwork
	}
	err = workflow.ExecuteActivity(ctx, a.CreateLinkedCard, linkedaccounts.CreateArgs{
		ID:                  tokenizedCard.ID,
		WalletID:            args.WalletID,
		ProviderID:          externalAccountID,
		Mask:                mask,
		Name:                fmt.Sprintf("%s %s", network, mask),
		Nickname:            fmt.Sprintf("%s %s", network, mask),
		CanSend:             cardInfo.Card.Pull.Enabled,
		CanReceive:          cardInfo.Card.Push.Enabled,
		State:               linkedAccountState,
		SendCountry:         country.ParseCountry(cardInfo.Card.Pull.Country),
		SendNetwork:         pullNetwork,
		SendCurrency:        currency.ParseCurrency(cardInfo.Card.Pull.Currency),
		SendAvailability:    "Immediate",
		ReceiveCountry:      country.ParseCountry(cardInfo.Card.Push.Country),
		ReceiveCurrency:     currency.ParseCurrency(cardInfo.Card.Push.Currency),
		ReceiveNetwork:      pushNetwork,
		ReceiveAvailability: linkedaccounts.FundsAvailability(cardInfo.Card.Push.Availability),
		Provider:            tabapay.ProviderName,
		Type:                tabapay.TypeCard,
	}).Get(ctx, &la)
	if err != nil {
		logger.Error("Failed to create linked account.")
		return nil, err
	}

	logger.Info("Created tabapay card.")

	return &la, nil
}

func ImportFXRates(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	var filename string
	err := workflow.ExecuteActivity(ctx, a.GetLatestRatesFile).Get(ctx, &filename)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.LoadFXRatesFromS3, filename).Get(ctx, nil)
	return err
}
