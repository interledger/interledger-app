package temporal

import (
	"gitlab.com/fynbos/backend/deposits"
	payments_workflow "gitlab.com/fynbos/backend/payments/workflows"
	mx_activities "gitlab.com/fynbos/backend/providers/mx/activities"
	mx_workflow "gitlab.com/fynbos/backend/providers/mx/workflows"
	unit_activities "gitlab.com/fynbos/backend/providers/unit/activities"
	unit_workflows "gitlab.com/fynbos/backend/providers/unit/workflows"
	"gitlab.com/fynbos/backend/withdrawals"
	"go.temporal.io/sdk/worker"
)

func NewTemporalWorker(b Backends) (worker.Worker, error) {
	w := worker.New(b.Temporal(), "backend", worker.Options{})

	// Register Deposits Workflow
	w.RegisterWorkflow(deposits.DepositWorkflow)
	depositActivities, err := deposits.NewActivity(deposits.ActivityArgs{
		Ds: b.Deposits(),
		As: b.Accounts(),
		Np: b.Noop(),
		Ts: b.Transactions(),
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
		As: b.Accounts(),
		Np: b.Noop(),
		Ts: b.Transactions(),
		Ws: b.Withdrawals(),
		Fs: b.FundingSources(),
	})
	if err != nil {
		return nil, err
	}
	w.RegisterActivity(withdrawalActivities)

	// Register onboard unit customer workflow
	w.RegisterWorkflow(unit_workflows.UnitOnboardCustomerWorkflow)
	unitOnboardingActivity := unit_activities.NewActivity(b)
	w.RegisterActivity(unitOnboardingActivity)
	w.RegisterWorkflow(unit_workflows.UnitHandleEventsWorkflow)

	w.RegisterWorkflow(mx_workflow.CreateMxAccountWorkflow)
	createMxBankAccountActivity := mx_activities.NewActivity(b)
	if err != nil {
		return nil, err
	}
	w.RegisterActivity(createMxBankAccountActivity)

	w.RegisterWorkflow(mx_workflow.MxCreateFundingsourceWorkflow)

	return w, nil
}
