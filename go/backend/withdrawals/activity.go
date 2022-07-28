package withdrawals

import (
	"context"
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
	ws        Service
	ts        transactions.Service
	as        accounts.Client
	noop      noop.Service
	fs        fundingsources.Service
}

type ActivityArgs struct {
	Ws Service                `validate:"required"`
	As accounts.Client        `validate:"required"`
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
		ws:        args.Ws,
		ts:        args.Ts,
		noop:      args.Np,
		fs:        args.Fs,
	}, nil
}

func (s *Activity) CreatePendingWithdrawalTransaction(ctx context.Context, withdrawalId string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating pending transaction")

	withdrawal, err := s.ws.Get(ctx, withdrawalId)
	if err != nil {
		return "", err
	}

	acc, err := s.as.Get(ctx, withdrawal.AccountID)
	if err != nil {
		return "", err
	}

	fs, err := s.fs.Get(ctx, withdrawal.FundingSourceId)
	if err != nil {
		return "", err
	}

	// TODO this needs to be better way to handle getting the data.
	trx, err := s.ts.CreatePending(ctx, &transactions.CreatePendingTransactionArgs{
		AccountID:   withdrawal.AccountID,
		Type:        "withdrawal",
		NetAmount:   withdrawal.Amount,
		Description: fmt.Sprintf("to %s bank account", fs.Name), // TODO Format to come from FS
		LedgerTransfers: []transactions.CreateLedgerTransferArgs{
			{
				LedgerID:        s.noop.GetLedgerID(),
				CreditAccountID: acc.LedgerAccountID,
				DebitAccountID:  s.noop.GetEquityAccountID(),
				Amount:          withdrawal.Amount,
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

func (s *Activity) ProcessNoopWithdrawal(ctx context.Context, withdrawalId string) error {

	withdrawal, err := s.ws.Get(ctx, withdrawalId)
	if err != nil {
		return err
	}

	err = s.noop.InitiateBankWithdrawal(ctx, &noop.BankWithdrawalArgs{
		Amount: withdrawal.Amount,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Activity) VoidPendingWithdrawalTransaction(ctx context.Context, trxId string) error {

	_, err := s.ts.VoidPending(ctx, trxId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Activity) PostPendingWithdrawalTransaction(ctx context.Context, trxId string) error {

	_, err := s.ts.PostPending(ctx, trxId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Activity) SetWithdrawalState(ctx context.Context, withdrawalId string, state State) error {
	err := s.ws.SetState(ctx, withdrawalId, state)
	if err != nil {
		return err
	}
	return nil
}
