package unit

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// The state passed into the workflow,
// which can be mutated by the workflow,
// and will be persisted for the duration of the workflow.
type UnitOnboardCustomerState struct {
	CustomerID      string
	Type            string
	IdentityID      string
	AccountID       string
	ApplicationArgs CreateApplicationArgs // TODO Change to key from vault
}

func UnitOnboardCustomerWorkflow(ctx workflow.Context, state UnitOnboardCustomerState) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout:    10 * time.Second,
		ScheduleToCloseTimeout: 15 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)

	logger.Info("Onboarding unit customer")

	// Channels that webhooks will use to communicate with the workflow through.
	customerCreatedChannel := workflow.GetSignalChannel(ctx, "onboard-unit-customer-created")
	applicationDeniedChannel := workflow.GetSignalChannel(ctx, "onboard-unit-application-denied")

	approved, denied := false, false

	var application Application
	err := workflow.ExecuteActivity(
		ctx,
		a.UnitCreateApplication,
		state.ApplicationArgs,
	).Get(ctx, &application)
	if err != nil {
		logger.Error("Failed to create unit application.", err)
		return err
	}

	// Catch the case where an application is immediately approved/denied - skip waiting for webhooks
	if application.Status == "Approved" {
		approved = true
		state.CustomerID = application.CustomerID
		state.Type = application.Type
	} else if application.Status == "Denied" {
		denied = true
	}

	for {
		if approved || denied {
			// We don't need to wait for webhooks if the application is approved or denied.
			break
		}

		// Create a new Selector on each iteration of the loop means Temporal will pick the first
		// event that occurs each time: receiving whichever signal comes first.
		selector := workflow.NewSelector(ctx)

		selector.AddReceive(customerCreatedChannel, func(c workflow.ReceiveChannel, _ bool) {
			var signal CustomerCreatedEvent
			c.Receive(ctx, &signal)

			// Handle signals for customer created (application success)
			approved = true
			state.CustomerID = signal.Relationships.Customer.Data.ID
			state.Type = signal.Relationships.Customer.Data.Type
		})

		selector.AddReceive(applicationDeniedChannel, func(c workflow.ReceiveChannel, _ bool) {
			var signal interface{}
			c.Receive(ctx, &signal)

			// Handle signals for application denied
			denied = true
		})

		selector.Select(ctx)
	}

	if denied {
		// Don't create an account.
		// TODO: Execute new denied activity that provides information to the customer that their application was denied.
		return nil
	}

	err = workflow.ExecuteActivity(
		ctx,
		a.UnitCreateAccount,
		state.IdentityID,
	).Get(ctx, &state.AccountID)
	if err != nil {
		logger.Error("Failed to create Fynbos account.", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.UnitCreateCustomer, &UnitCreateCustomerArgs{
		CustomerID: state.CustomerID,
		IdentityID: state.IdentityID,
		Type:       state.Type,
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to create unit customer.", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.UnitCreateDepositAccount, state.CustomerID).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to create unit deposit account.", err)
		return err
	}

	return nil
}
