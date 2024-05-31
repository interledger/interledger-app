package ops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gitlab.com/fynbos/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type Activity struct {
	b Backends
}

func NewActivity(b Backends) *Activity {
	return &Activity{b}
}

func StartTransactionsPolling(b Backends) {
	// This workflow ID can be user business logic identifier as well.
	workflowID := "cron_wealth_tfsa_poll"
	workflowOptions := client.StartWorkflowOptions{
		ID:                    workflowID,
		TaskQueue:             "backend",                                           // TODO maybe a new queue
		CronSchedule:          "0 2 * * *",                                         // Every day at 2
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING, // There can be only one
	}

	we, err := b.Temporal().ExecuteWorkflow(context.Background(), workflowOptions, GetEasyTFSATransactionsWorkflow)
	if err != nil {
		log.Fatal("Unable to execute workflow", zap.Error(err))
	}
	log.Info("Started workflow", zap.String("WorkflowID", we.GetID()), zap.String("RunID", we.GetRunID()))
}

func (a *Activity) ListWealthUsers(ctx context.Context) ([]int64, error) {
	var ids []int64
	err := a.b.DB().SelectContext(ctx, &ids, "SELECT external_id FROM wealth_users")

	return ids, err
}

func GetEasyTFSATransactionsWorkflow(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var wealthUserIDs []int64
	err := workflow.ExecuteActivity(ctx, a.ListWealthUsers).Get(ctx, &wealthUserIDs)
	if err != nil {
		return err
	}

	for _, wealthUser := range wealthUserIDs {
		var filename string
		err = workflow.ExecuteActivity(ctx, a.DownloadTransactions, wealthUser).Get(ctx, &filename)
		if err != nil {
			return err
		}

		err = workflow.ExecuteActivity(ctx, a.PostDepositsToWealth, wealthUser, filename).Get(ctx, nil)
		if err != nil {
			return err
		}

		err = workflow.ExecuteActivity(ctx, a.DeleteTransactionsFile, filename).Get(ctx, nil)
		return err
	}

	return nil
}

func (a *Activity) DownloadTransactions(_ context.Context, wealthUser int64) (string, error) {
	// TODO: get these values from Vault (possibly)
	var username, password string

	session, err := Login(username, password)
	if err != nil {
		return "", err
	}

	if session.credentialsValid == false || session.hasMFA {
		return "", temporal.NewNonRetryableApplicationError("can't use credentials to login", "authentication", fmt.Errorf("invalid login"), "hasMFA", session.hasMFA, "valid credentials", session.credentialsValid)
	}

	return DownloadTFSATransactions(wealthUser, session)
}

func (a *Activity) PostDepositsToWealth(_ context.Context, wealthUser int64, filename string) error {
	deposits, err := ParseTXHistory(filename)
	if err != nil {
		return err
	}

	if len(deposits) == 0 {
		return nil
	}

	body, err := json.Marshal(deposits)
	if err != nil {
		return temporal.NewNonRetryableApplicationError("cannot marshall wealth json", "internal", err)
	}

	// TODO: change address on env
	_, err = otelhttp.DefaultClient.Post("https://wealh.fynbos.app/api/easy/tfsa/deposits", "application/json", bytes.NewReader(body))
	return err
}

func (a *Activity) DeleteTransactionsFile(_ context.Context, filename string) error {
	// Best effort here, if the file is already gone then no harm done
	_ = os.Remove(filename)
	return nil
}
