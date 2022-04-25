package deposits

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/accounts"
	transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/providers/noop"
	"go.temporal.io/sdk/activity"
)

type Activity struct {
	validator *validator.Validate
	ds        Service
	ts        transactions.Service
	as        accounts.Service
	fs        fundingsources.Service
	noop      noop.Service
}

type ActivityArgs struct {
	Ds Service                `validate:"required"`
	As accounts.Service       `validate:"required"`
	Np noop.Service           `validate:"required"`
	Ts transactions.Service   `validate:"required"`
	Fs fundingsources.Service `validate:"required"`
}

func NewActivity(args ActivityArgs) (*Activity, error) {
	v := validator.New()
	if err := v.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err.Error())
	}

	return &Activity{
		validator: v,
		as:        args.As,
		ds:        args.Ds,
		ts:        args.Ts,
		fs:        args.Fs,
		noop:      args.Np,
	}, nil
}

func (s *Activity) CreatePendingTransaction(ctx context.Context, depositId string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating pending transaction")

	deposit, err := s.ds.Get(ctx, depositId)
	if err != nil {
		return "", err
	}

	acc, err := s.as.Get(ctx, deposit.AccountID)
	if err != nil {
		return "", err
	}

	// get funding source
	fs, err := s.fs.Get(ctx, deposit.FundingSourceId)
	if err != nil {
		return "", fmt.Errorf("%w: %s", errors.New("deposit service: internal error"), err.Error())
	}

	// TODO this needs to be better way to handle getting the data.
	trx, err := s.ts.CreatePending(ctx, &transactions.CreatePendingTransactionArgs{
		AccountID:   deposit.AccountID,
		Type:        "deposit",
		NetAmount:   deposit.Amount,
		Description: fmt.Sprintf("from %s bank account", fs.Mask), // TODO Format to come from FS
		LedgerTransfers: []transactions.CreateLedgerTransferArgs{
			{
				LedgerID:        s.noop.GetLedgerID(),
				DebitAccountID:  acc.LedgerAccountID,
				CreditAccountID: s.noop.GetEquityAccountID(),
				Amount:          deposit.Amount,
				// Code: "1", // TODO: define ledger transfer codes.
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

	return trx.ID, nil
}

func (s *Activity) ProcessNoopDeposit(ctx context.Context, depositId string) error {

	deposit, err := s.ds.Get(ctx, depositId)
	if err != nil {
		return err
	}

	acc, err := s.as.Get(ctx, deposit.AccountID)
	if err != nil {
		return err
	}

	err = s.noop.InitiateBankDeposit(ctx, &noop.BankDepositArgs{
		IdentityID:      acc.IdentityID,
		FundingSourceID: deposit.FundingSourceId,
		Amount:          deposit.Amount,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Activity) VoidPendingTransaction(ctx context.Context, trxId string) error {

	_, err := s.ts.VoidPending(ctx, trxId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Activity) PostPendingTransaction(ctx context.Context, trxId string) error {

	_, err := s.ts.PostPending(ctx, trxId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Activity) SetDepositState(ctx context.Context, depositId string, state State) error {
	err := s.ds.SetState(ctx, depositId, state)
	if err != nil {
		return err
	}
	return nil
}
