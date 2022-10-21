package temporal

import (
	machnet_workflows "gitlab.com/fynbos/backend/providers/machnet/workflows"
	"go.temporal.io/sdk/worker"
)

func NewTemporalWorker(b Backends) (worker.Worker, error) {
	w := worker.New(b.Temporal(), "backend", worker.Options{})

	w.RegisterActivity(machnet_workflows.NewActivity(b))
	w.RegisterWorkflow(machnet_workflows.CreateSendUserWorkflow)
	w.RegisterWorkflow(machnet_workflows.CreateTransactionWorkflow)

	return w, nil
}
