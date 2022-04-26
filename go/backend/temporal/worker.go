package temporal

import (
	"gitlab.com/fynbos/backend/accounts"
	transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/deposits"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/noop"
	"gitlab.com/fynbos/backend/withdrawals"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

type WorkerArgs struct {
	Client client.Client
	Ps     payments.Service
	Ds     deposits.Service
	As     accounts.Service
	Np     noop.Service
	Ts     transactions.Service
	Ws     withdrawals.Service
	Fs     fundingsources.Service
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

	// Register Deposits Workflow
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

	return w, nil
}
