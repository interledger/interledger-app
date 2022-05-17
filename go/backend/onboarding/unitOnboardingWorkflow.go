package onboarding

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

type OnboardUnitCustomerArgs struct {
	CustomerID string
	Type       string
	IdentityID string
	Country    string
}

func OnboardUnitCustomerWorkflow(ctx workflow.Context, args *OnboardUnitCustomerArgs) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout:    10 * time.Second,
		ScheduleToCloseTimeout: 15 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)

	logger.Info("Onboarding unit customer")

	var accountID string
	err := workflow.ExecuteActivity(
		ctx,
		a.CreateAccount,
		args.IdentityID,
		args.Country,
	).Get(ctx, &accountID)
	if err != nil {
		logger.Error("Failed to create Fynbos account.", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.MapCustomerToAccount, &MapCustomerToAccountArgs{
		CustomerID: args.CustomerID,
		AccountID:  accountID,
		Type:       args.Type,
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to map Unit customer to Fynbos account.", err)
		return err
	}

	return nil
}
