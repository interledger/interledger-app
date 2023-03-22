package ops

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/providers/gmt/external"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

func StartNotificationsPolling(b Backends) {
	// This workflow ID can be user business logic identifier as well.
	workflowID := "cron_" + uuid.NewString()
	workflowOptions := client.StartWorkflowOptions{
		ID:           workflowID,
		TaskQueue:    "cron",
		CronSchedule: "*/5 * * * *", // Every 5 minutes
	}

	we, err := b.Temporal().ExecuteWorkflow(context.Background(), workflowOptions, PollNotificationsWorkflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}

func PollNotificationsWorkflow(ctx workflow.Context) error {
	var a *Activity

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	err := workflow.ExecuteActivity(ctx, a.PollNotifications).Get(ctx, nil)

	return err
}

type WorkflowRef struct {
	ExternalID    string `db:"external_id"`
	WorkflowID    string `db:"workflow_id"`
	WorkflowRunID string `db:"workflow_run_id"`
}

func (a *Activity) PollNotifications(ctx context.Context) error {
	logger := activity.GetLogger(ctx)

	// Count open Transactions awaiting notifications
	var refs []WorkflowRef
	err := a.b.DB().SelectContext(ctx, &refs, "SELECT external_id, workflow_id, workflow_run_id  FROM  gmt_workflow_refs WHERE completed=false")
	if errors.Is(err, sql.ErrNoRows) {
		// No transactions waiting for status updates
		return nil
	}
	if err != nil {
		return err
	}

	if len(refs) == 0 {
		// No transactions waiting for status updates
		return nil
	}

	refMap := make(map[string]WorkflowRef)
	for _, ref := range refs {
		refMap[ref.ExternalID] = ref
	}

	nr, err := a.ext.GetNotificationsContext(ctx, &external.GetNotifications{
		Alias: a.creds.Alias,
		User:  a.creds.User,
		Pass:  a.creds.Password,
	})
	if err != nil {
		return err
	}

	for _, n := range nr.GetNotificationsResult.WsNotifications {
		ref, ok := refMap[n.Password]
		if !ok {
			logger.Error("notification received with no waiting workflow", "id", n.Password, "status", n.Status)
			continue
		}

		err = a.b.Temporal().SignalWorkflow(ctx, ref.WorkflowID, ref.WorkflowRunID, gmtEventsChannel, *n)
		if err != nil {
			logger.Error("failed to signal workflow", "err", err)
		}
	}

	return nil
}
