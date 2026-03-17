package jobs

import (
	"context"
	"time"

	"gitlab.com/fynbos/backend/providers/gatehub/external"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type updateOrganizationArgs struct {
	APIBaseURL string `json:"apiBaseUrl"`
	TwoFAType  string `json:"twoFAType"`
}

func UpdateGateHubOrganizationConfig(ctx workflow.Context, args updateOrganizationArgs) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)

	var res external.UpdateOrganizationConfigurationResponse
	err := workflow.ExecuteActivity(ctx, a.UpdateOrganizationConfig, args).Get(ctx, &res)
	if err != nil {
		return err
	}

	logger.Info("updated GateHub organization configuration", "api-base-url", res.APIBaseURL, "2fa-type", res.TwoFAType)

	return nil
}

func (a *Activity) UpdateOrganizationConfig(ctx context.Context, args updateOrganizationArgs) (*external.UpdateOrganizationConfigurationResponse, error) {
	res, err := a.b.Gatehub().UpdateOrganizationConfiguration(ctx, args.APIBaseURL, args.TwoFAType)
	if err != nil {
		return nil, err
	}

	return res, nil
}
