package workflows

import (
	"go.temporal.io/sdk/workflow"
	"time"
)

func PublishTwitterProofWorkflow(ctx workflow.Context, identityID string) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("PublishTwitterProofWorkflow workflow started", "identityID", identityID)

	err := workflow.ExecuteActivity(ctx, a.UpdateIdentityState, ctx, UpdateStateArgs{
		IdentityID: identityID,
		Proof:      "",
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to update identity state", "error", err)
		return err
	}

	var tweetUrl string
	err = workflow.ExecuteActivity(ctx, a.PostProofTweet, ctx, identityID).Get(ctx, &tweetUrl)
	if err != nil {
		logger.Error("PostProofTweet activity failed", "error", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.UpdateIdentityState, ctx, UpdateStateArgs{
		IdentityID: identityID,
		Proof:      tweetUrl,
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to update identity state", "error", err)
		return err
	}

	// Kickoff verification workflow
	err = workflow.ExecuteActivity(ctx, a.StartVerification, ctx, identityID, tweetUrl).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to start verification workflow", "error", err)
		return err
	}

	return nil
}
