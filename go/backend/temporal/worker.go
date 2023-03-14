package temporal

import (
	"gitlab.com/fynbos/backend/identities/platforms"
	openpayments_workflows "gitlab.com/fynbos/backend/openpayments/workflows"
	"go.temporal.io/sdk/worker"
)

func NewTemporalWorker(b Backends) (worker.Worker, error) {
	w := worker.New(b.Temporal(), "backend", worker.Options{})

	w.RegisterActivity(openpayments_workflows.NewActivity(b))
	w.RegisterWorkflow(openpayments_workflows.OutgoingTransactionWorkflow)

	w.RegisterActivity(platforms.NewDevActivity(b))
	w.RegisterWorkflow(platforms.DevVerifyWorkflow)

	return w, nil
}
