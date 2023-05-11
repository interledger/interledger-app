package temporal

import (
	"gitlab.com/fynbos/backend/identities/platforms"
	"gitlab.com/fynbos/backend/jobs"
	kyc_workflows "gitlab.com/fynbos/backend/kyc/workflows"
	openpayments_workflows "gitlab.com/fynbos/backend/openpayments/workflows"
	gmt_workflows "gitlab.com/fynbos/backend/providers/gmt/ops"
	tabapay_workflows "gitlab.com/fynbos/backend/providers/tabapay/workflows"
	"go.temporal.io/sdk/worker"
)

func NewTemporalWorker(b Backends) (worker.Worker, error) {
	w := worker.New(b.Temporal(), "backend", worker.Options{})

	w.RegisterActivity(openpayments_workflows.NewActivity(b))
	w.RegisterWorkflow(openpayments_workflows.OutgoingTransactionWorkflow)

	w.RegisterActivity(kyc_workflows.NewActivity(b))
	w.RegisterWorkflow(kyc_workflows.StartKYC)

	w.RegisterActivity(platforms.NewDevActivity(b))
	w.RegisterWorkflow(platforms.DevVerifyWorkflow)
	w.RegisterWorkflow(platforms.TwitterVerifyWorkflow)

	w.RegisterActivity(gmt_workflows.NewActivity(b))
	w.RegisterWorkflow(gmt_workflows.PollNotificationsWorkflow)
	w.RegisterWorkflow(gmt_workflows.OnboardUserWorkflow)
	w.RegisterWorkflow(gmt_workflows.ACH2ACHTransferWorkflow)
	w.RegisterWorkflow(gmt_workflows.Card2ACHTransferWorkflow)
	w.RegisterWorkflow(gmt_workflows.ACH2CardTransferWorkflow)

	w.RegisterActivity(tabapay_workflows.NewActivity(b))
	w.RegisterWorkflow(tabapay_workflows.CreateTabapayCardWorkflow)

	gmt_workflows.StartNotificationsPolling(b)

	// Jobs
	w.RegisterActivity(jobs.NewActivity(b))
	w.RegisterWorkflow(jobs.AddWalletPrivateKeysWorkflow)
	w.RegisterWorkflow(jobs.FixWalletPublicKeysWorkflow)

	return w, nil
}
