package temporal

import (
	"gitlab.com/fynbos/backend/accounts"
	transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/deposits"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/noop"
	"gitlab.com/fynbos/backend/providers/unit"
	"gitlab.com/fynbos/backend/withdrawals"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

type WorkerArgs struct {
	Client client.Client
	Ps     payments.Service
	Ds     deposits.Service
	As     accounts.Client
	Np     noop.Service
	Ts     transactions.Client
	Ws     withdrawals.Service
	Fs     fundingsources.Service
	Os     onboarding.Service
	Up     unit.Service
	Mx     mx.Service
	Is     identity.Service
}

func NewTemporalWorker(args WorkerArgs) (worker.Worker, error) {
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
	w.RegisterWorkflow(payments.OutgoingPaymentWorkflow)
	paymentsActivities, err := payments.NewActivity(payments.ActivityArgs{
		Ps: args.Ps,
		As: args.As,
		Np: args.Np,
		Ts: args.Ts,
	})
	if err != nil {
		return nil, err
	}
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
	w.RegisterWorkflow(unit.UnitOnboardCustomerWorkflow)
	unitOnboardingActivity, err := unit.NewActivity(&unit.ActivityArgs{
		UnitService:     args.Up,
		AccountsService: args.As,
		IdentityService: args.Is,
	})
	if err != nil {
		return nil, err
	}
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
