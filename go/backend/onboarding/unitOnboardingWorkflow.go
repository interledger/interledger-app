package onboarding

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func OnboardUnitCustomerWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout:    10 * time.Second,
		ScheduleToCloseTimeout: 15 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)

	logger.Info("Onboarding unit customer")

	var a *Activity

	err := workflow.ExecuteActivity(ctx, a.CreateAccount).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to create Fynbos account.", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.MapCustomerToAccount).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to map Unit customer to Fynbos account.", err)
		return err
	}

	return nil
}
