package temporal

import (
	"fmt"

	"github.com/interledger/interledger-app/go/log"
	"github.com/lestrrat-go/jwx/v3/jwk"

	"github.com/interledger/interledger-app/go/backend/identities/platforms"
	"github.com/interledger/interledger-app/go/backend/jobs"
	kyc_workflows "github.com/interledger/interledger-app/go/backend/kyc/ops"
	payments_workflows "github.com/interledger/interledger-app/go/backend/payments/ops"
	chimoney_workflows "github.com/interledger/interledger-app/go/backend/providers/chimoney/ops"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub"
	gatehub_workflows "github.com/interledger/interledger-app/go/backend/providers/gatehub/ops"
	pti_workflows "github.com/interledger/interledger-app/go/backend/providers/pti/ops"
	xago_external "github.com/interledger/interledger-app/go/backend/providers/xago/external"
	xago_workflows "github.com/interledger/interledger-app/go/backend/providers/xago/ops"
	rafiki_workflows "github.com/interledger/interledger-app/go/backend/rafiki/ops"
	"github.com/interledger/interledger-app/go/backend/slack"
	twitter_workflows "github.com/interledger/interledger-app/go/backend/twitter/workflows"
	"go.temporal.io/sdk/worker"
)

func NewTemporalWorker(b Backends, gatehubConfig gatehub.Config, xagoConfig xago_external.Config, ptiJWK, ptiBaseURL, ptiClientID, chimoneyToken string, jobsCfg jobs.Config) (worker.Worker, error) {
	w := worker.New(b.Temporal(), "backend", worker.Options{})

	w.RegisterActivity(slack.SendToChannelActivity)

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
	w.RegisterActivity(jobs.NewActivity(b, gatehubConfig, jobsCfg))
	w.RegisterWorkflow(jobs.AddWalletPrivateKeysWorkflow)
	w.RegisterWorkflow(jobs.FixWalletPublicKeysWorkflow)
	w.RegisterWorkflow(jobs.MigratePaymentPointers)
	w.RegisterWorkflow(jobs.MigrateOpenPaymentsObjects)
	w.RegisterWorkflow(jobs.BackfillTransactionsRefundState)
	w.RegisterWorkflow(jobs.BackfillTransactionsTitle)
	w.RegisterWorkflow(jobs.UpdateTransactionTypes)
	w.RegisterWorkflow(jobs.GenerateWalletPaymentPointersJob)
	w.RegisterWorkflow(jobs.MigrateUSWalletsToPTIJob)
	w.RegisterWorkflow(jobs.ResendOnOffRampEmailJob)
	w.RegisterWorkflow(jobs.CreateRafikiPaymentPointersJob)
	w.RegisterWorkflow(jobs.MigrateWalletAddressesToIlpLinkJob)
	w.RegisterWorkflow(jobs.RemoveCustodialKeysJob)
	w.RegisterWorkflow(jobs.TransformKeysToBase64URLJob)
	w.RegisterWorkflow(jobs.MigrateWalletAddressesToLowercaseJob)
	w.RegisterWorkflow(jobs.BalanceDiscrepanciesJob)
	w.RegisterWorkflow(jobs.UpdateRafikiWalletEnabledJob)
	w.RegisterWorkflow(jobs.SetGatehubGatewayToPaywiserJob)
	w.RegisterWorkflow(jobs.RestartKYCstatusForXagoJob)
	w.RegisterWorkflow(jobs.BackfillPaywiserAccountsJob)
	w.RegisterWorkflow(jobs.PtiSettleDepositAndWithdrawsForWallet)
	w.RegisterWorkflow(jobs.PtiDropOldDataWorkflow)
	w.RegisterWorkflow(jobs.EnableSendVerificationEmailToUnverifiedUserJob)
	w.RegisterWorkflow(jobs.UpdateGateHubOrganizationConfig)
	w.RegisterWorkflow(jobs.NotifyAgreementChangedWorkflow)
	w.RegisterWorkflow(jobs.DisabledAccountsTabWorkflow)

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
	w.RegisterActivity(xago_workflows.NewActivity(b, xagoConfig))
	w.RegisterWorkflow(xago_workflows.CreateBeneficiaryWorkflow)
	w.RegisterWorkflow(xago_workflows.CreateBalanceAccountWorkflow)
	w.RegisterWorkflow(xago_workflows.XagoDepositPollWorkflow)
	w.RegisterWorkflow(xago_workflows.UpdateInquiryLinkWorkflow)

	xago_workflows.StartDepositsPolling(b)

	// PTI
	if ptiJWK == "" {
		if ptiBaseURL != "" || ptiClientID != "" {
			return nil, fmt.Errorf("PTI_JWK is required when PTI is configured")
		}
		log.Warn("PTI not configured: skipping PTI workflow/activity registration")
	} else {
		w.RegisterWorkflow(pti_workflows.DepositWorkflow)
		w.RegisterWorkflow(pti_workflows.SettleDepositWorkflow)
		w.RegisterWorkflow(pti_workflows.MarkTransactionStateWorkflow)
		w.RegisterWorkflow(pti_workflows.ReservePtiBalance)
		w.RegisterWorkflow(pti_workflows.SettleWithdrawWorkflow)
		w.RegisterWorkflow(pti_workflows.RevertWithdrawWorkflow)
		w.RegisterWorkflow(pti_workflows.ReturnedWorkflow)

		ptiPrivateKey, err := jwk.ParseKey([]byte(ptiJWK))
		if err != nil {
			return nil, fmt.Errorf("invalid PTI_JWK: %w", err)
		}

		w.RegisterActivity(pti_workflows.NewActivity(b, ptiPrivateKey, ptiBaseURL, ptiClientID))
		w.RegisterWorkflow(pti_workflows.CreateWalletWorkflow)
		w.RegisterWorkflow(pti_workflows.CreateUserWorkflow)
		w.RegisterWorkflow(pti_workflows.CreateCardWorkflow)
		w.RegisterWorkflow(pti_workflows.CreatePtiBankAccountWorkflow)
	}

	// Gatehub
	w.RegisterActivity(gatehub_workflows.NewActivity(b, gatehubConfig))
	w.RegisterWorkflow(gatehub_workflows.CreateGatehubUserWorkflow)
	w.RegisterWorkflow(gatehub_workflows.CreateGatehubDeposit)
	w.RegisterWorkflow(gatehub_workflows.ProcessGatehubWithdrawal) // TODO: remove once old withdrawal workflows are drained from the Temporal queue
	w.RegisterWorkflow(gatehub_workflows.CompleteGatehubWithdrawalWorkflow)
	w.RegisterWorkflow(gatehub_workflows.RejectGatehubWithdrawalWorkflow)
	w.RegisterWorkflow(gatehub_workflows.LinkGatehubUserToGatewayWorkflow)
	w.RegisterWorkflow(gatehub_workflows.ProcessCardCreationWorkflow)
	w.RegisterWorkflow(gatehub_workflows.CreateGateHubCardWorkflow)
	w.RegisterWorkflow(gatehub_workflows.BackfillAccountWorkflow)
	w.RegisterWorkflow(gatehub_workflows.CreateCardTransaction)
	w.RegisterWorkflow(gatehub_workflows.GatehubClearingCardTransactionsPollWorkflow)
	w.RegisterWorkflow(gatehub_workflows.GatehubRealtimeCardTransactionsPollWorkflow)
	w.RegisterWorkflow(gatehub_workflows.NotifyWithdrawalSCTITimeoutWorkflow)
	w.RegisterWorkflow(gatehub_workflows.NotifyWithdrawalReroutedWorkflow)

	gatehub_workflows.StartClearingCardTransactionsPolling(b)
	gatehub_workflows.StartRealtimeCardTransactionsPolling(b)

	// Chimoney
	w.RegisterActivity(chimoney_workflows.NewActivity(b, chimoneyToken))
	w.RegisterWorkflow(chimoney_workflows.CreateChimoneyUserWorkflow)
	w.RegisterWorkflow(chimoney_workflows.ChimomeyCompleteKYC)
	w.RegisterWorkflow(chimoney_workflows.CreateChimoneyDepositWorkflow)
	w.RegisterWorkflow(chimoney_workflows.FinishChimoneyDepositWorkflow)
	w.RegisterWorkflow(chimoney_workflows.ExecuteChimoneyWithdrawalWorkflow)
	w.RegisterWorkflow(chimoney_workflows.ExecuteChimoneyFinishWithdrawalWorkflow)

	return w, nil
}
