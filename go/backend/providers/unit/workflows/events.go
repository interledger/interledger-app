package workflows

import (
	"encoding/json"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/providers/unit/activities"
	"gitlab.com/fynbos/backend/providers/unit/external"
	"go.temporal.io/sdk/workflow"
)

// Handles unit webhook events in one at a time. Workflow will fail if at least 1 event
// was not parsed/handled successfully.
func UnitHandleEventsWorkflow(ctx workflow.Context, rawEvents []json.RawMessage) error {
	var a *activities.Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout:    20 * time.Second,
		ScheduleToCloseTimeout: 40 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)

	logger.Info("Handling unit events")

	err := workflow.ExecuteActivity(ctx, a.UnitStoreEvents, rawEvents).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to store events.")
		return err
	}

	var didFail bool
	count := 0
	for i, rawEvent := range rawEvents {
		var event external.Event
		if err := json.Unmarshal(rawEvent, &event); err != nil {
			logger.Error("Failed to parse event number=", i, "in event array.", err)
			didFail = true
			continue
		}

		switch event.Type {
		case external.CUSTOMER_CREATED:
			var customerCreatedEvent external.CustomerCreatedEvent
			if err = json.Unmarshal(rawEvent, &customerCreatedEvent); err != nil {
				logger.Error("Failed to parse customerCreatedEvent id=", event.ID)
				didFail = true
				continue
			}

			if err = workflow.ExecuteActivity(ctx, a.UnitNotifyCustomerCreated, customerCreatedEvent).Get(ctx, nil); err != nil {
				logger.Error("Failed to notify unit customer created eventID=", event.ID)
				didFail = true
				continue
			}
			count++
		case external.APPLICATION_DENIED:
			var applicationDeniedEvent external.ApplicationDeniedEvent
			if err = json.Unmarshal(rawEvent, &applicationDeniedEvent); err != nil {
				logger.Error("Failed to parse applicationDeniedEvent id=", event.ID)
				didFail = true
				continue
			}

			if err = workflow.ExecuteActivity(ctx, a.UnitNotifyApplicationDenied, applicationDeniedEvent).Get(ctx, nil); err != nil {
				logger.Error("Failed to notify unit application denied eventID=", event.ID)
				didFail = true
				continue
			}
			count++
		case external.PAYMENT_CREATED, external.PAYMENT_CLEARING, external.PAYMENT_SENT,
			external.PAYMENT_REJECTED, external.PAYMENT_RETURNED, external.PAYMENT_CANCELED, external.PAYMENT_PENDING_REVIEW:
			var achPaymentEvent external.AchPayment
			if err = json.Unmarshal(rawEvent, &achPaymentEvent); err != nil {
				logger.Error("Failed to parse achPaymentEvent id=", event.ID)
				didFail = true
				continue
			}

			if err = workflow.ExecuteActivity(ctx, a.UnitNotifyAchPaymentEvent, &achPaymentEvent).Get(ctx, nil); err != nil {
				logger.Error("Failed to notify unit achPayment event eventID=", event.ID)
				didFail = true
				continue
			}
			count++
		default:
			logger.Warn(fmt.Sprintf("Unknown unit event. eventID=%s eventType=%s", event.ID, event.Type))
		}
	}

	if didFail {
		return fmt.Errorf("Failed to handle all unit events. (%d/%d handled)", count, len(rawEvents))
	}

	return nil
}
