package temporal

import (
	"crypto/rand"
	"crypto/rsa"
	"os"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/log"

	"gitlab.com/fynbos/backend/identities/platforms"
	"gitlab.com/fynbos/backend/jobs"
	kyc_workflows "gitlab.com/fynbos/backend/kyc/ops"
	payments_workflows "gitlab.com/fynbos/backend/payments/ops"
	asta_workflows "gitlab.com/fynbos/backend/providers/astra/ops"
	chimoney_workflows "gitlab.com/fynbos/backend/providers/chimoney/ops"
	gatehub_workflows "gitlab.com/fynbos/backend/providers/gatehub/ops"
	pti_workflows "gitlab.com/fynbos/backend/providers/pti/ops"
	xago_workflows "gitlab.com/fynbos/backend/providers/xago/ops"
	rafiki_workflows "gitlab.com/fynbos/backend/rafiki/ops"
	twitter_workflows "gitlab.com/fynbos/backend/twitter/workflows"
	"go.temporal.io/sdk/worker"
)

func NewTemporalWorker(b Backends) (worker.Worker, error) {
	w := worker.New(b.Temporal(), "backend", worker.Options{})

	w.RegisterActivity(kyc_workflows.NewActivity(b))
	w.RegisterWorkflow(kyc_workflows.SetKYCStatusWorkflow)

	//Identities
	w.RegisterActivity(platforms.NewTwitterActivity(b))
	w.RegisterWorkflow(platforms.TwitterVerifyWorkflow)
	w.RegisterActivity(platforms.NewDomainActivity(b))
	w.RegisterWorkflow(platforms.DomainVerifyWorkflow)

	//Twitter
	w.RegisterActivity(twitter_workflows.NewActivity(b))
	w.RegisterWorkflow(twitter_workflows.PublishTwitterProofWorkflow)

	// Jobs
	w.RegisterActivity(jobs.NewActivity(b))
	w.RegisterWorkflow(jobs.AddWalletPrivateKeysWorkflow)
	w.RegisterWorkflow(jobs.FixWalletPublicKeysWorkflow)
	w.RegisterWorkflow(jobs.MigratePaymentPointers)
	w.RegisterWorkflow(jobs.MigrateOpenPaymentsObjects)
	w.RegisterWorkflow(jobs.BackfillTransactionsRefundState)
	w.RegisterWorkflow(jobs.BackfillTransactionsTitle)
	w.RegisterWorkflow(jobs.UpdateTransactionTypes)
	w.RegisterWorkflow(jobs.GenerateWalletPaymentPointersJob)
	w.RegisterWorkflow(jobs.MigrateUSWalletsToPTIJob)
	w.RegisterWorkflow(jobs.CreateAstraBusinessProfile)
	w.RegisterWorkflow(jobs.ExchangeAstraBusinessProfileCode)
	w.RegisterWorkflow(jobs.ResendOnOffRampEmailJob)
	w.RegisterWorkflow(jobs.CreateRafikiPaymentPointersJob)
	w.RegisterWorkflow(jobs.MigrateWalletAddressesToIlpLinkJob)
	w.RegisterWorkflow(jobs.RemoveCustodialKeysJob)
	w.RegisterWorkflow(jobs.TransformKeysToBase64URLJob)
	w.RegisterWorkflow(jobs.MigrateWalletAddressesToLowercaseJob)
	w.RegisterWorkflow(jobs.UpdateAccountEnabledJob)

	// Payment Engine
	w.RegisterActivity(payments_workflows.NewActivity(b))
	w.RegisterWorkflow(payments_workflows.RollbackPayInWorkflow)
	w.RegisterWorkflow(payments_workflows.PayinWorkflow)
	w.RegisterWorkflow(payments_workflows.PayoutWorkflow)
	w.RegisterWorkflow(payments_workflows.PaymentWorkflow)
	w.RegisterWorkflow(payments_workflows.AwaitReceiverWorkflow)

	// Rafiki
	w.RegisterActivity(rafiki_workflows.NewActivity(b))
	w.RegisterWorkflow(rafiki_workflows.WebMonetizationPaymentsWorkflow)

	rafiki_workflows.StartRafikiIncomingPaymentsPolling(b)

	// Xago
	w.RegisterActivity(xago_workflows.NewActivity(b))
	w.RegisterWorkflow(xago_workflows.CreateBeneficiaryWorkflow)
	w.RegisterWorkflow(xago_workflows.CreateBalanceAccountWorkflow)
	w.RegisterWorkflow(xago_workflows.XagoDepositPollWorkflow)

	xago_workflows.StartDepositsPolling(b)

	var ptiPrivateKey jwk.Key
	if env.IsLocal() {
		privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			log.Fatalln(err)
		}
		ptiPrivateKey, err = jwk.FromRaw(privateKey)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		var err error
		ptiPrivateKey, err = jwk.ParseKey([]byte(os.Getenv("PTI_JWK")))
		if err != nil {
			log.Fatalln(err)
		}
	}
	w.RegisterActivity(pti_workflows.NewActivity(b, ptiPrivateKey))
	w.RegisterWorkflow(pti_workflows.CreateWalletWorkflow)

	// Astra
	w.RegisterActivity(asta_workflows.NewActivity(b))
	w.RegisterWorkflow(asta_workflows.AstraRenewTokensWorkflow)
	w.RegisterWorkflow(asta_workflows.CreateAstraCardWorkflow)

	asta_workflows.StartTokenRefreshing(b)

	// Gatehub
	w.RegisterActivity(gatehub_workflows.NewActivity(b))
	w.RegisterWorkflow(gatehub_workflows.CreateGatehubUserWorkflow)
	w.RegisterWorkflow(gatehub_workflows.CreateGatehubDeposit)
	w.RegisterWorkflow(gatehub_workflows.ProcessGatehubWithdrawal)

	// Chimoney
	w.RegisterActivity(chimoney_workflows.NewActivity(b))
	w.RegisterWorkflow(chimoney_workflows.CreateChimoneyUserWorkflow)
	w.RegisterWorkflow(chimoney_workflows.ChimomeyWatchForSuccessfulKYC)
	w.RegisterWorkflow(chimoney_workflows.CreateChimoneyDepositWorkflow)
	w.RegisterWorkflow(chimoney_workflows.ExecuteChimoneyWithdrawalWorkflow)

	return w, nil
}
