package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/payments/workflows"
	"gitlab.com/fynbos/backend/twilio"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

// requiresOTP is currently a stub for when OTP may not be required for flows initiated from Rafiki or for low value payments
func requiresOTP(_ context.Context, _ payments.InitiateOutgoingPaymentArgs) bool {
	return true
}

func InitiateOutgoingPayment(
	ctx context.Context,
	b Backends,
	args payments.InitiateOutgoingPaymentArgs,
) (*payments.OutgoingPayment, error) {
	err := b.Validator().Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err, payments.ErrInvalidArgument)
	}

	id, err := b.Identity().Get(ctx, args.UserID)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err, payments.ErrInternal)
	}

	acc, err := b.Accounts().GetByIdentityID(ctx, id.ID)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err, payments.ErrInternal)
	}

	if !b.Accounts().CanMakeOutgoingPayment(acc, id.ID) {
		return nil, fmt.Errorf("%w", payments.ErrUnauthorized)
	}

	if acc.AvailableBalance < int64(args.Amount) {
		return nil, fmt.Errorf("%w", payments.ErrInsufficientBalance)
	}

	// Check the OTP
	if requiresOTP(ctx, args) {
		verify, err := b.Twilio().CheckVerificationCode(ctx, &twilio.CheckVerificationCodeArgs{
			PhoneNumber: id.MobileNumber,
			Code:        args.OTP,
		})
		if err != nil {
			return nil, err
		}

		if !verify.IsValid() {
			return nil, fmt.Errorf("invalid OTP provided %w", twilio.ErrInvalidOTP)
		}
	}

	var outgoingPayment payments.OutgoingPayment
	err = b.DB().GetContext(ctx, &outgoingPayment, `INSERT INTO outgoing_payments
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

func Get(ctx context.Context, b Backends, id, userID string) (*payments.OutgoingPayment, error) {
	var outgoingPayment payments.OutgoingPayment

	acc, err := b.Accounts().GetByIdentityID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err, payments.ErrInternal)
	}

	err = b.DB().GetContext(ctx, &outgoingPayment, `select * from outgoing_payments where id = $1 and account_id = $2 LIMIT 1`, id, acc.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, payments.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s %w", err, payments.ErrInternal)
	}

	return &outgoingPayment, nil
}

func GetUnauthenticated(ctx context.Context, b Backends, id string) (*payments.OutgoingPayment, error) {
	var outgoingPayment payments.OutgoingPayment
	err := b.DB().GetContext(ctx, &outgoingPayment, `select * from outgoing_payments where id = $1 LIMIT 1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, payments.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s %w", err, payments.ErrInternal)
	}

	return &outgoingPayment, nil
}

func SetState(ctx context.Context, b Backends, id string, state payments.State) error {
	outgoingPayment, err := GetUnauthenticated(ctx, b, id)
	if err != nil {
		return err
	}

	_, err = b.DB().ExecContext(ctx, "update outgoing_payments set state = $1 where id = $2", state.String(), outgoingPayment.ID)
	if err != nil {
		return err
	}

	return nil
}
