package workflows

import (
	"fmt"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/twitter"
	"go.temporal.io/sdk/workflow"
	"time"
)

func PublishTwitterProofWorkflow(ctx workflow.Context, identity *identities.Identity, connection twitter.Connection) (string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("PublishTwitterProofWorkflow workflow started", "identity", identity, "connection", connection)

	var paymentPointerURL string
	err := workflow.ExecuteActivity(ctx, a.GetWalletPaymentPointerURL, ctx, identity.WalletID).Get(ctx, &paymentPointerURL)
	if err != nil {
		logger.Error("GetWalletPaymentPointerURL activity failed", "error", err)
		return "", err
	}

	var tweetID string
	err = workflow.ExecuteActivity(ctx, a.PostProofTweet, ctx, identity.SignatureHash, connection.ID, paymentPointerURL).Get(ctx, &tweetID)
	if err != nil {
		logger.Error("PostProofTweet activity failed", "error", err)
		return "", err
	}

	proofUrl := fmt.Sprintf("https://twitter.com/%s/status/%s", connection.Username, tweetID)

	return proofUrl, nil
}
