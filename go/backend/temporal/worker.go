package temporal

import (
	"gitlab.com/fynbos/backend/accounts"
	transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/deposits"
	"gitlab.com/fynbos/backend/providers/noop"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

type WorkerArgs struct {
	Client client.Client
	Ds     deposits.Service
	As     accounts.Service
	Np     noop.Service
	Ts     transactions.Service
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

	return w, nil
}
