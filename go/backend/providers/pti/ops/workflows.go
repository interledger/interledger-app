package ops

import (
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/pti"

	"go.temporal.io/sdk/workflow"
)

func CreateUserWorkflow(ctx workflow.Context, walletID string) (*pti.User, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating pti user.")

	var externalUser pti.User
	err := workflow.ExecuteActivity(ctx, a.GetPtiUser, walletID).Get(ctx, &externalUser)
	if err != nil {
		return nil, err
	}

	if externalUser.ID == "" {
		var externalUserID string
		err = workflow.ExecuteActivity(ctx, a.CreatePtiUser, walletID).Get(ctx, &externalUserID)
		if err != nil {
			return nil, err
		}

		err = workflow.ExecuteActivity(ctx, a.SavePtiUser, externalUserID, walletID).Get(ctx, &externalUser)
		if err != nil {
			return nil, err
		}
	}

	return &externalUser, nil
}

func CreateWalletWorkflow(ctx workflow.Context, args pti.CreateWalletArgs) (*linkedaccounts.LinkedAccount, error) {
	if args.Currency != currency.USD {
		return nil, fmt.Errorf("%w Only supports USD wallets", pti.ErrInternal)
	}

	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating pti wallet.")

	var externalUser pti.User
	err := workflow.ExecuteActivity(ctx, a.GetPtiUser, args.WalletID).Get(ctx, &externalUser)
	if err != nil {
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.CheckUserAssessmentAccepted, args.WalletID).Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	externalWalletID := fmt.Sprintf("%s_%s", args.Currency.String(), args.WalletID)
	err = workflow.ExecuteActivity(ctx, a.CreatePtiWallet, pti.CreateExternalWalletArgs{
		ID:       externalWalletID,
		UserID:   externalUser.ExternalID,
		Currency: args.Currency,
	}).Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	var la linkedaccounts.LinkedAccount
	err = workflow.ExecuteActivity(ctx, a.CreatePtiWalletLinkedAccount, linkedaccounts.CreateArgs{
		WalletID:        args.WalletID,
		Nickname:        args.Nickname,
		Name:            args.Title,
		Provider:        pti.ProviderName,
		ProviderID:      externalWalletID,
		Type:            pti.AccTypeBalance,
		CanReceive:      true,
		ReceiveCountry:  country.US,
		ReceiveCurrency: currency.USD,
		SendCountry:     country.US,
		SendCurrency:    currency.USD,
		CanSend:         true,
		State:           linkedaccounts.Verified,
	}).Get(ctx, &la)
	if err != nil {
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.CreatePTIBalanceAccount, la.ID).Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &la, nil
}

func CreateCardWorkflow(ctx workflow.Context, walletID, tokenID string) (*linkedaccounts.LinkedAccount, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating pti card.")

	var linkedAccount linkedaccounts.LinkedAccount
	err := workflow.ExecuteActivity(ctx, a.CreatePTICard, walletID, tokenID).Get(ctx, &linkedAccount)
	if err != nil {
		return nil, err
	}

	return &linkedAccount, nil
}
