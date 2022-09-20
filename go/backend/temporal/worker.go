package temporal

import (
	mx_activities "gitlab.com/fynbos/backend/providers/mx/activities"
	mx_workflow "gitlab.com/fynbos/backend/providers/mx/workflows"
	unit_activities "gitlab.com/fynbos/backend/providers/unit/activities"
	unit_workflows "gitlab.com/fynbos/backend/providers/unit/workflows"
	"go.temporal.io/sdk/worker"
)

func NewTemporalWorker(b Backends) (worker.Worker, error) {
	w := worker.New(b.Temporal(), "backend", worker.Options{})

	// Register onboard unit customer workflow
	w.RegisterWorkflow(unit_workflows.UnitOnboardCustomerWorkflow)
	unitOnboardingActivity := unit_activities.NewActivity(b)
	w.RegisterActivity(unitOnboardingActivity)
	w.RegisterWorkflow(unit_workflows.UnitHandleEventsWorkflow)

	w.RegisterWorkflow(mx_workflow.CreateMxAccountWorkflow)
	createMxBankAccountActivity := mx_activities.NewActivity(b)
	w.RegisterActivity(createMxBankAccountActivity)

	w.RegisterWorkflow(mx_workflow.MxCreateFundingsourceWorkflow)

	return w, nil
}
