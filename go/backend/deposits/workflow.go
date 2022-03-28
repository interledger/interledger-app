package deposits

import (
	"go.temporal.io/sdk/workflow"
)

func Deposit(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting deposit")

	var a *Activity

	err := workflow.ExecuteActivity(ctx, a.InitiateProviderTransfer).Get(ctx, nil)
	if err != nil {
		return err
	}

	//err = workflow.ExecuteActivity(ctx, a.CreateAccount).Get(ctx, nil)
	if err != nil {
		return err
	}
	logger.Info("Workflow complete")

	return nil
}
