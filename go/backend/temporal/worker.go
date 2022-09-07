package temporal

import (
	"gitlab.com/fynbos/backend/accounts"
	transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/deposits"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/payments"
	payments_workflow "gitlab.com/fynbos/backend/payments/workflows"
	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/noop"
	"gitlab.com/fynbos/backend/providers/unit"
	unit_activities "gitlab.com/fynbos/backend/providers/unit/activities"
	unit_workflows "gitlab.com/fynbos/backend/providers/unit/workflows"
	"gitlab.com/fynbos/backend/withdrawals"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// TODO replace worker args with Backends.
type WorkerArgs struct {
	Client client.Client
	Ps     payments.Client
	Ds     deposits.Service
	As     accounts.Client
	Np     noop.Service
	Ts     transactions.Client
	Ws     withdrawals.Service
	Fs     fundingsources.Client
	Os     onboarding.Client
	Up     unit.Client
	Mx     mx.Service
	Is     identity.Client
}

func NewTemporalWorker(args WorkerArgs, b Backends) (worker.Worker, error) {
	w := worker.New(args.Client, "backend", worker.Options{})

	// Register Deposits Workflow
	w.RegisterWorkflow(deposits.DepositWorkflow)
	depositActivities, err := deposits.NewActivity(deposits.ActivityArgs{
		Ds: args.Ds,
		As: args.As,
		Np: args.Np,
		Ts: args.Ts,
	})
	if err != nil {
		return nil, err
	}
	w.RegisterActivity(depositActivities)

	// Register Outgoing Payments Workflow
	w.RegisterWorkflow(payments_workflow.OutgoingPaymentWorkflow)
	paymentsActivities := payments_workflow.NewActivity(b)
	w.RegisterActivity(paymentsActivities)

	// Register Withdrawals Workflow
	w.RegisterWorkflow(withdrawals.WithdrawalWorkflow)
	withdrawalActivities, err := withdrawals.NewActivity(withdrawals.ActivityArgs{
		As: args.As,
		Np: args.Np,
		Ts: args.Ts,
		Ws: args.Ws,
		Fs: args.Fs,
	})
	if err != nil {
		return nil, err
	}
	w.RegisterActivity(withdrawalActivities)

	// Register onboard unit customer workflow
	w.RegisterWorkflow(unit_workflows.UnitOnboardCustomerWorkflow)
	unitOnboardingActivity := unit_activities.NewActivity(b)
	w.RegisterActivity(unitOnboardingActivity)

	w.RegisterWorkflow(mx.CreateMxAccountWorkflow)
	createMxBankAccountActivity, err := mx.NewActivity(&mx.ActivityArgs{
		Mx:                   args.Mx,
		Unit:                 args.Up,
		AccountService:       args.As,
		IdentityService:      args.Is,
		FundingSourceService: args.Fs,
	})
	if err != nil {
		return nil, err
	}
	w.RegisterActivity(createMxBankAccountActivity)

	return w, nil
}
