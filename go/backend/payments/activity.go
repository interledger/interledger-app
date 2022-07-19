package payments

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/accounts/ops"

	"github.com/go-playground/validator/v10"
	transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/providers/noop"
	"go.temporal.io/sdk/activity"
)

type Activity struct {
	validator *validator.Validate
	ps        Service
	ts        transactions.Service
	as        ops.Service
	noop      noop.Service
}

type ActivityArgs struct {
	Ps Service              `validate:"required"`
	As ops.Service          `validate:"required"`
	Np noop.Service         `validate:"required"`
	Ts transactions.Service `validate:"required"`
}

func NewActivity(args ActivityArgs) (*Activity, error) {
	v := validator.New()
	if err := v.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err.Error())
	}

	return &Activity{
		validator: v,
		as:        args.As,
		ps:        args.Ps,
		ts:        args.Ts,
		noop:      args.Np,
	}, nil
}

func (s *Activity) CreatePendingOutgoingPayment(ctx context.Context, outgoingPaymentId string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating pending transaction")

	outgoingPayment, err := s.ps.Get(ctx, outgoingPaymentId)
	if err != nil {
		return "", err
	}

	acc, err := s.as.Get(ctx, outgoingPayment.AccountID)
	if err != nil {
		return "", err
	}

	// TODO this needs to be better way to handle getting the data.
	trx, err := s.ts.CreatePending(ctx, &transactions.CreatePendingTransactionArgs{
		AccountID:   outgoingPayment.AccountID,
		Type:        "outgoingPayment",
		NetAmount:   outgoingPayment.Amount,
		Description: "Sent to " + outgoingPayment.Destination,
		LedgerTransfers: []transactions.CreateLedgerTransferArgs{
			{
				LedgerID:        s.noop.GetLedgerID(),
				DebitAccountID:  s.noop.GetEquityAccountID(),
				CreditAccountID: acc.LedgerAccountID,
				Amount:          outgoingPayment.Amount,
				// Code: uint16
				Flags: transactions.LedgerTransferFlags{
					TwoPhaseCommit: true,
				},
			},
		},
	})

	// TODO need to work out retryable vs non retryable errors
	if err != nil {
		return "", err
	}

	if err != nil {
		switch {
		case errors.Is(err, transactions.ErrExceedsDebits):
			return "", ErrInsufficientBalance
		default:
			return "", fmt.Errorf("%s %w", err.Error(), ErrInternal)
		}
	}

	return trx.ID, nil
}

func (s *Activity) ProcessNoopOutgoingPayment(ctx context.Context, outgoingPaymentId string) error {

	outgoingPayment, err := s.ps.Get(ctx, outgoingPaymentId)
	if err != nil {
		return err
	}

	_, err = s.as.Get(ctx, outgoingPayment.AccountID)
	if err != nil {
		return err
	}

	err = s.noop.InitiateOutgoingPayment(ctx, &noop.OutgoingPaymentArgs{
		Amount: outgoingPayment.Amount,
		To:     outgoingPayment.Destination,
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *Activity) VoidPendingOutgoingPayment(ctx context.Context, trxId string) error {

	_, err := s.ts.VoidPending(ctx, trxId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Activity) PostPendingOutgoingPayment(ctx context.Context, trxId string) error {

	_, err := s.ts.PostPending(ctx, trxId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Activity) SetOutgoingPaymentState(ctx context.Context, outgoingPaymentId string, state State) error {
	err := s.ps.SetState(ctx, outgoingPaymentId, state)
	if err != nil {
		return err
	}
	return nil
}
