package ops

import (
	"context"
	"fmt"

	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/payments/workflows"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

func InitiateOutgoingPayment(
	ctx context.Context,
	b Backends,
	args payments.InitiateOutgoingPaymentArgs,
) (*payments.OutgoingPayment, error) {
	err := b.Validator().Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err, payments.ErrInvalidArgument)
	}

	id, err := b.Identity().Get(ctx, args.IdentityID)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err, payments.ErrInternal)
	}
	acc, err := b.Accounts().Get(ctx, args.AccountID)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err, payments.ErrInternal)
	}
	if !b.Accounts().CanMakeOutgoingPayment(acc, id.ID) {
		return nil, fmt.Errorf("%w", payments.ErrUnauthorized)
	}
	if acc.AvailableBalance < int64(args.Amount) {
		return nil, fmt.Errorf("%w", payments.ErrInsufficientBalance)
	}

	var outgoingPayment payments.OutgoingPayment
	err = b.DB().Get(&outgoingPayment, `INSERT INTO outgoing_payments
	(account_id, amount, destination, state) VALUES ($1, $2, $3, $4) RETURNING *`, acc.ID, args.Amount, args.To, payments.Created)

	if err != nil {
		return nil, fmt.Errorf("failed to insert into db %s %w", err, payments.ErrInternal)
	}

	// TODO add mechanism to handle if deposit is created but workflow is not
	workflowOptions := client.StartWorkflowOptions{
		ID:                    "outgoingPayment_" + outgoingPayment.ID,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}
	_, err = b.Temporal().ExecuteWorkflow(context.Background(), workflowOptions, workflows.OutgoingPaymentWorkflow, outgoingPayment.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to insert execute workflow %s %w", err.Error(), payments.ErrInternal)
	}

	return &outgoingPayment, nil
}

func Get(ctx context.Context, b Backends, id string) (*payments.OutgoingPayment, error) {
	var outgoingPayment payments.OutgoingPayment
	err := b.DB().GetContext(ctx, &outgoingPayment, `select * from outgoing_payments where id = $1 LIMIT 1`, id)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err, payments.ErrInternal)
	}

	return &outgoingPayment, nil
}

func SetState(ctx context.Context, b Backends, id string, state payments.State) error {
	outgoingPayment, err := Get(ctx, b, id)

	if err != nil {
		return err
	}

	_, err = b.DB().ExecContext(ctx, "update outgoing_payments set state = $1 where id = $2", state.String(), outgoingPayment.ID)
	if err != nil {
		return err
	}

	return nil
}
