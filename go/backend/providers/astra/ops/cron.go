package ops

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	httplogger "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/env"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"gitlab.com/fynbos/backend/providers/astra/external"
	mock_client "gitlab.com/fynbos/backend/providers/astra/external/mock"
	"gitlab.com/fynbos/log"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type opsBackends struct {
	ActivityBackends
	astrExt external.Client
}

func (o opsBackends) External() external.Client {
	return o.astrExt
}

type Activity struct {
	b Backends
}

func NewActivity(ab ActivityBackends) *Activity {
	ex := external.New(&http.Client{
		Transport: otelhttp.NewTransport(
			httplogger.NewTransport(http.DefaultTransport, ab, external.Redact),
		),
	})

	if env.IsTest() {
		ex = mock_client.SetupDevMock(nil)
	}

	return &Activity{b: &opsBackends{
		ActivityBackends: ab,
		astrExt:          ex,
	}}
}

func StartTokenRefreshing(b ActivityBackends) {
	// This workflow ID can be user business logic identifier as well.
	workflowID := "cron_astra_token_refresh"
	workflowOptions := client.StartWorkflowOptions{
		ID:                    workflowID,
		TaskQueue:             "backend",
		CronSchedule:          "0 */12 * * *",                                      // Every 12 hours
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING, // There can be only one
	}

	we, err := b.Temporal().ExecuteWorkflow(context.Background(), workflowOptions, AstraRenewTokensWorkflow)
	if err != nil {
		log.Fatal("Unable to execute workflow", zap.Error(err))
	}
	log.Info("Started workflow", zap.String("WorkflowID", we.GetID()), zap.String("RunID", we.GetRunID()))
}

func AstraRenewTokensWorkflow(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)

	logger.Info("starting refresh astra user tokens")

	var walletIDs []string
	err := workflow.ExecuteActivity(ctx, a.ListExpiringTokens).Get(ctx, &walletIDs)
	if err != nil {
		return err
	}

	for _, walletID := range walletIDs {
		err = workflow.ExecuteActivity(ctx, a.RefreshToken, walletID).Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Activity) ListExpiringTokens(ctx context.Context) ([]string, error) {
	// Select all tokens that expire in the next 12 hours, (we a day grace period built in)
	var walletIDs []string
	err := a.b.DB().SelectContext(ctx, &walletIDs, "SELECT wallet_id FROM astra_access_tokens WHERE refresh_expire_at<$1", time.Now().Add(time.Hour*12))
	if err != nil {
		return nil, err
	}

	return walletIDs, nil
}

func (a *Activity) RefreshToken(ctx context.Context, walletID string) error {
	var token dbToken
	err := a.b.DB().GetContext(ctx, &token, "SELECT * FROM astra_access_tokens WHERE wallet_id=$1", walletID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if token.RefreshExpiresAt.After(time.Now().Add(time.Hour * 13)) {
		// Token has already been refreshed, do nothing
		return nil
	}

	_, err = CreateOrRefreshToken(ctx, a.b, walletID)
	return err
}
