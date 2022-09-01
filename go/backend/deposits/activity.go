package deposits

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/accounts"
	transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/providers/noop"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
)

type Activity struct {
	validator *validator.Validate
	ds        Service
	ts        transactions.Client
	as        accounts.Client
	noop      noop.Service
}

type ActivityArgs struct {
	Ds Service             `validate:"required"`
	As accounts.Client     `validate:"required"`
	Np noop.Service        `validate:"required"`
	Ts transactions.Client `validate:"required"`
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
		noop:      args.Np,
	}, nil
}

func (s *Activity) CreatePendingDeposit(ctx context.Context, depositId string) (string, error) {
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

	// TODO this needs to be better way to handle getting the data.
	trx, err := s.ts.CreatePending(ctx, &transactions.CreatePendingTransactionArgs{
		AccountID:   deposit.AccountID,
		Type:        "deposit",
		NetAmount:   deposit.Amount,
		Description: fmt.Sprintf("from %s bank account", "test"), // TODO Format to come from FS
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

func (s *Activity) CreateAchDepositTransactions(ctx context.Context, depositID, depositTransferID string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating deposit transactions")

	deposit, err := s.ds.Get(ctx, depositID)
	if err != nil {
		return temporal.NewNonRetryableApplicationError(err.Error(), "ErrInternal", err)
	}

	acc, err := s.as.Get(ctx, deposit.AccountID)
	if err != nil {
		return temporal.NewNonRetryableApplicationError(err.Error(), "ErrInternal", err)
	}

	// TODO: ledger transfers and account transactions should be separated.
	_, err = s.ts.Create(ctx, &transactions.CreateTransactionArgs{
		AccountID:   deposit.AccountID,
		Type:        "deposit",
		NetAmount:   deposit.Amount,
		Description: fmt.Sprintf("from %s bank account", "test"),
		LedgerTransfers: []transactions.CreateLedgerTransferArgs{
			{
				ID:              depositTransferID,
				LedgerID:        s.noop.GetLedgerID(),
				DebitAccountID:  acc.LedgerAccountID,
				CreditAccountID: s.noop.GetEquityAccountID(),
				Amount:          deposit.Amount,
				Code:            1, // TODO: define ledger transfer codes.
			},
		},
	})
	// TODO: need to separate account transaction and ledger transfers.
	// Can't determine if it's safe to retry
	if err != nil {
		return temporal.NewNonRetryableApplicationError(err.Error(), "ErrInternal", err)
	}

	return nil
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

func (s *Activity) VoidPendingDeposit(ctx context.Context, trxId string) error {

	_, err := s.ts.VoidPending(ctx, trxId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Activity) PostPendingDeposit(ctx context.Context, trxId string) error {

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

func (a *Activity) GetDeposit(ctx context.Context, id string) (*Deposit, error) {
	deposit, err := a.ds.Get(ctx, id)
	if errors.Is(err, ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "ErrNotFound", err)
	}
	if err != nil {
		return nil, err
	}

	return deposit, err
}
