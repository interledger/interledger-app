package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/astra"
	"gitlab.com/fynbos/backend/providers/astra/external"
	"gitlab.com/fynbos/backend/providers/basistheory"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func CreateAstraCardWorkflow(ctx workflow.Context, args astra.CreateCardArgs) (*linkedaccounts.LinkedAccount, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating astra card.")

	var cardInfo external.UserCard
	err := workflow.ExecuteActivity(ctx, a.AddCardToAstra, args).Get(ctx, &cardInfo)
	if err != nil {
		logger.Error("Failed to add card to astra.", zap.Error(err))
		return nil, err
	}

	var tokenizedCard basistheory.Card
	err = workflow.ExecuteActivity(ctx, a.CreateBasisTheoryCard, args, cardInfo).Get(ctx, &tokenizedCard)
	if err != nil {
		logger.Error("Failed to save card info to basis theory", zap.Error(err))
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

	err = workflow.ExecuteActivity(ctx, a.CreateCardLinkedAccount, args, cardInfo, tokenizedCard).Get(ctx, &la)
	if err != nil {
		logger.Error("Failed to create linked account for card", zap.Error(err))
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.CreateAccount, args.WalletID).Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &la, nil
}

func (a *Activity) CreateAccount(ctx context.Context, walletID string) error {
	var accID string
	err := a.b.DB().GetContext(ctx, &accID, "SELECT account_id FROM astra_accounts WHERE wallet_id=$1", walletID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if accID != "" {
		// We already have an account for this user, do nothing.
		return nil
	}

	token, err := GetToken(ctx, a.b, walletID)
	if err != nil {
		return err
	}

	acc, err := a.b.External().AddAccount(ctx, token, external.CreateAccountArgs{
		BankAccountType: external.AccountTypeChecking,
		Name:            "Universal",
		AccountNumber:   "TODO",
		RoutingNumber:   "TODO",
	})
	if err != nil {
		return err
	}

	_, err = a.b.DB().ExecContext(ctx, "INSERT INTO astra_accounts(wallet_id, account_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", walletID, acc.ID)
	return err
}

func (a *Activity) CreateCardLinkedAccount(ctx context.Context, args astra.CreateCardArgs, card external.UserCard, tokenCard basistheory.Card) (*linkedaccounts.LinkedAccount, error) {
	mask := card.LastFourDigits
	network := tokenCard.PullNetwork
	if network == "" {
		network = tokenCard.PushNetwork
	}

	la, err := a.b.LinkedAccounts().Create(ctx, &linkedaccounts.CreateArgs{
		WalletID:            args.WalletID,
		Name:                fmt.Sprintf("%s %s", network, mask),
		Nickname:            fmt.Sprintf("%s %s", network, mask),
		Mask:                mask,
		Provider:            astra.ProviderName,
		ProviderID:          card.ID,
		Type:                astra.TypeCard,
		CanSend:             true,
		CanReceive:          true,
		State:               linkedaccounts.Verified,
		SendCountry:         country.US,
		SendCurrency:        currency.USD,
		SendAvailability:    linkedaccounts.Immediate,
		SendNetwork:         tokenCard.PullNetwork,
		ReceiveCountry:      country.US,
		ReceiveCurrency:     currency.USD,
		ReceiveAvailability: linkedaccounts.Immediate,
		ReceiveNetwork:      tokenCard.PushNetwork,
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", astra.ErrInternal, err)
	}

	return la, nil
}

func (a *Activity) CreateBasisTheoryCard(ctx context.Context, args astra.CreateCardArgs, card external.UserCard) (*basistheory.Card, error) {
	bin, err := a.b.External().GetCardBin(ctx, fmt.Sprintf("{{ %s | json: '$.bin' }}", args.BasisTheoryTokenID))
	if err != nil {
		return nil, err
	}

	basisCard, err := a.b.BasisTheory().CreateCard(ctx, basistheory.CreateCardArgs{
		WalletID:         args.WalletID,
		TokenID:          args.BasisTheoryTokenID,
		Bin:              bin.Bin,
		PullNetwork:      card.CardCompany,
		PullEnabled:      card.PullEnabled,
		PullType:         bin.CardType,
		PullCountry:      "US", // Astra is US Based
		PushNetwork:      card.CardCompany,
		PushEnabled:      card.PushEnabled,
		PushType:         bin.CardType,
		PushAvailability: "Immediate",
		PushCountry:      "US",
	})
	if err != nil {
		return nil, err
	}

	return basisCard, nil
}

func (a *Activity) AddCardToAstra(ctx context.Context, args astra.CreateCardArgs) (*external.UserCard, error) {
	owner, err := a.b.KYC().GetIndividualDetails(ctx, args.WalletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", astra.ErrInternal, err)
	}
	if owner.Address == nil {
		return nil, fmt.Errorf("%w require address for wallet", astra.ErrInternal)
	}

	token, err := GetToken(ctx, a.b, args.WalletID)
	if err != nil {
		return nil, err
	}

	state := owner.Address.State
	if len(state) > 2 {
		state = owner.Address.State[len(state)-2:]
	}

	card, err := a.b.External().AddCard(ctx, token, external.CreateCardArgs{
		CardNumber:       fmt.Sprintf("{{ %s | json: '$.number' }}", args.BasisTheoryTokenID),
		CardSecurityCode: fmt.Sprintf("{{ %s | json: '$.cvc' }}", args.BasisTheoryTokenID),
		ExpirationDate:   fmt.Sprintf("{{ %s | json: '$.expiration_year' | to_string }}/{{ %s | json: '$.expiration_month' | pad_left: 2,'0' }}", args.BasisTheoryTokenID, args.BasisTheoryTokenID),
		FirstName:        owner.FirstName,
		LastName:         owner.LastName,
		StreetLine1:      owner.Address.Line1,
		StreetLine2:      owner.Address.Line2,
		City:             owner.Address.City,
		State:            state,
		ZipCode:          owner.Address.ZipCode,
		AddedByUser:      true,
	})
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (a *Activity) MarkCardNotDeleted(ctx context.Context, id string) (*linkedaccounts.LinkedAccount, error) {
	la, err := a.b.LinkedAccounts().MarkNotDeleted(ctx, id)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	return la, nil
}
