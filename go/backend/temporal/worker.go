package temporal

import (
	"gitlab.com/fynbos/backend/identities/platforms"
	"gitlab.com/fynbos/backend/jobs"
	kyc_workflows "gitlab.com/fynbos/backend/kyc/workflows"
	payments_workflows "gitlab.com/fynbos/backend/payments/ops"
	gmt_workflows "gitlab.com/fynbos/backend/providers/gmt/ops"
	tabapay_workflows "gitlab.com/fynbos/backend/providers/tabapay/workflows"
	rafiki_workflows "gitlab.com/fynbos/backend/rafiki/ops"
	twitter_workflows "gitlab.com/fynbos/backend/twitter/workflows"
	"go.temporal.io/sdk/worker"
)

func NewTemporalWorker(b Backends) (worker.Worker, error) {
	w := worker.New(b.Temporal(), "backend", worker.Options{})

	w.RegisterActivity(kyc_workflows.NewActivity(b))
	w.RegisterWorkflow(kyc_workflows.StartKYC)

	//Identities
	w.RegisterActivity(platforms.NewTwitterActivity(b))
	w.RegisterWorkflow(platforms.TwitterVerifyWorkflow)
	w.RegisterActivity(platforms.NewDomainActivity(b))
	w.RegisterWorkflow(platforms.DomainVerifyWorkflow)

	//Twitter
	w.RegisterActivity(twitter_workflows.NewActivity(b))
	w.RegisterWorkflow(twitter_workflows.PublishTwitterProofWorkflow)

	w.RegisterActivity(gmt_workflows.NewActivity(b))
	// w.RegisterWorkflow(gmt_workflows.PollNotificationsWorkflow) // disabled until we support ACH
	w.RegisterWorkflow(gmt_workflows.OnboardUserWorkflow)
	w.RegisterWorkflow(gmt_workflows.ACH2ACHTransferWorkflow)
	w.RegisterWorkflow(gmt_workflows.Card2ACHTransferWorkflow)
	w.RegisterWorkflow(gmt_workflows.ACH2CardTransferWorkflow)
	w.RegisterWorkflow(gmt_workflows.Card2CardTransferWorkflow)
	w.RegisterWorkflow(gmt_workflows.NotifyGMTCard2CardWorkflow)
	w.RegisterWorkflow(gmt_workflows.RollbackTabapayPullWorkflow)
	w.RegisterWorkflow(gmt_workflows.GMTNotifyCompleted)
	w.RegisterWorkflow(gmt_workflows.GMTComplianceChecksWorkflow)

	w.RegisterActivity(tabapay_workflows.NewActivity(b))
	w.RegisterWorkflow(tabapay_workflows.CreateTabapayCardWorkflow)

	gmt_workflows.StartNotificationsPolling(b)

	// Jobs
	w.RegisterActivity(jobs.NewActivity(b))
	w.RegisterWorkflow(jobs.AddWalletPrivateKeysWorkflow)
	w.RegisterWorkflow(jobs.FixWalletPublicKeysWorkflow)
	w.RegisterWorkflow(jobs.RunGMTCertification)
	w.RegisterWorkflow(jobs.TabapayCertificationWorkflow)
	w.RegisterWorkflow(jobs.UpdateBasisTheoryCardDetailsWorkflow)
	w.RegisterWorkflow(jobs.MigratePaymentPointers)
	w.RegisterWorkflow(jobs.MigrateOpenPaymentsObjects)
	w.RegisterWorkflow(jobs.ClearOrphanedGMTTransactions)
	w.RegisterWorkflow(jobs.BackfillLinkedCardCurrencyInfo)
	w.RegisterWorkflow(jobs.BackfillTransactionsRefundState)

	// Payment Engine
	w.RegisterActivity(payments_workflows.NewActivity(b))
	w.RegisterWorkflow(payments_workflows.RollbackPayInWorkflow)
	w.RegisterWorkflow(payments_workflows.PayinWorkflow)
	w.RegisterWorkflow(payments_workflows.PayoutWorkflow)
	w.RegisterWorkflow(payments_workflows.PaymentWorkflow)
	w.RegisterWorkflow(payments_workflows.AwaitReceiverWorkflow)
	w.RegisterWorkflow(payments_workflows.CreateReferralsWorkflow)

	// Rafiki
	w.RegisterActivity(rafiki_workflows.NewActivity(b))
	w.RegisterWorkflow(rafiki_workflows.PayoutIncomingPaymentsWorkflow)

	rafiki_workflows.StartRafikiIncomingPaymentsPolling(b)

	return w, nil
}
