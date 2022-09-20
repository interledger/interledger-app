package temporal

import (
	mx_activities "gitlab.com/fynbos/backend/providers/mx/activities"
	mx_workflow "gitlab.com/fynbos/backend/providers/mx/workflows"
	"go.temporal.io/sdk/worker"
)

func NewTemporalWorker(b Backends) (worker.Worker, error) {
	w := worker.New(b.Temporal(), "backend", worker.Options{})

	w.RegisterWorkflow(mx_workflow.CreateMxAccountWorkflow)
	createMxBankAccountActivity := mx_activities.NewActivity(b)
	w.RegisterActivity(createMxBankAccountActivity)

	w.RegisterWorkflow(mx_workflow.MxCreateFundingsourceWorkflow)

	return w, nil
}
