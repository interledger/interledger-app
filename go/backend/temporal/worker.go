package temporal

import (
	openpayments_workflows "gitlab.com/fynbos/backend/openpayments/workflows"
	machnet_workflows "gitlab.com/fynbos/backend/providers/machnet/workflows"
	"go.temporal.io/sdk/worker"
)

func NewTemporalWorker(b Backends) (worker.Worker, error) {
	w := worker.New(b.Temporal(), "backend", worker.Options{})

	w.RegisterActivity(machnet_workflows.NewActivity(b))
	w.RegisterWorkflow(machnet_workflows.CreateSendUserWorkflow)
	w.RegisterWorkflow(machnet_workflows.CreateTransactionWorkflow)
	w.RegisterWorkflow(machnet_workflows.CreateWalletTopupWorkflow)
	w.RegisterWorkflow(machnet_workflows.CreateWalletWithdrawalWorkflow)
	w.RegisterWorkflow(machnet_workflows.ExecuteWalletWithdrawalWorkflow)
	w.RegisterWorkflow(machnet_workflows.ExecuteWalletTopupWorkflow)
	w.RegisterWorkflow(machnet_workflows.DeleteAccountWorkflow)
	w.RegisterWorkflow(machnet_workflows.EmailStatements)

	w.RegisterActivity(openpayments_workflows.NewActivity(b))
	w.RegisterWorkflow(openpayments_workflows.OutgoingTransactionWorkflow)

	return w, nil
}
