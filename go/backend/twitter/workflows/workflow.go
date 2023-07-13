package workflows

import (
	"go.temporal.io/sdk/workflow"
	"time"
)

func PublishTwitterProofWorkflow(ctx workflow.Context, identityID string) (string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("PublishTwitterProofWorkflow workflow started", "identityID", identityID)

	var tweetUrl string
	err := workflow.ExecuteActivity(ctx, a.PostProofTweet, identityID).Get(ctx, &tweetUrl)
	if err != nil {
		logger.Error("PostProofTweet activity failed", "error", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.SetTwitterProof, identityID, tweetUrl).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to update identity state", "error", err)
		return "", err
	}

	return tweetUrl, nil
}
