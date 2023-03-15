package temporal

import (
	kyc_workflows "gitlab.com/fynbos/backend/kyc/workflows"
	openpayments_workflows "gitlab.com/fynbos/backend/openpayments/workflows"
	"go.temporal.io/sdk/worker"
)

func NewTemporalWorker(b Backends) (worker.Worker, error) {
	w := worker.New(b.Temporal(), "backend", worker.Options{})

	w.RegisterActivity(openpayments_workflows.NewActivity(b))
	w.RegisterWorkflow(openpayments_workflows.OutgoingTransactionWorkflow)

	w.RegisterActivity(kyc_workflows.NewActivity(b))
	w.RegisterWorkflow(kyc_workflows.StartKYC)

	return w, nil
}
