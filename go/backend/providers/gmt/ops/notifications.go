package ops

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

func StartNotificationsPolling(b Backends) {
	scheduleID := "schedule_gmt_notifications"
	workflowID := "cron_gmt_notifications"
	schedule, err := b.Temporal().ScheduleClient().Create(context.Background(), client.ScheduleOptions{
		ID: scheduleID,
		Spec: client.ScheduleSpec{
			Intervals: []client.ScheduleIntervalSpec{
				{
					Every: 2 * time.Minute,
				},
			},
		},
		Action: &client.ScheduleWorkflowAction{
			ID:        workflowID,
			TaskQueue: "backend",
			Workflow:  PollNotificationsWorkflow,
		},
		Overlap: enums.SCHEDULE_OVERLAP_POLICY_CANCEL_OTHER,
	})
	if err != nil {
		log.Fatalln("Unable to start gmt notifications schedule", err)
	}

	log.Println("Started schedule gmt notifications schedule", "ScheduleID", schedule.GetID())
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

	nr, err := a.ext.GetNotifications(ctx)
	if err != nil {
		return err
	}

	for _, n := range nr {
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
