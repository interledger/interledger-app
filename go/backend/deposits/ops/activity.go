package ops

import (
	"context"
	"fmt"

	"gitlab.com/fynbos/backend/deposits"

	transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/providers/noop"
	"go.temporal.io/sdk/activity"
)

type Activity struct {
	b Backends
	/*validator *validator.Validate
	ds        Service
	ts        transactions.Service
	as        accounts.Service
	noop      noop.Service*/
}

/*
type ActivityArgs struct {
	Ds Service              `validate:"required"`
	As accounts.Service     `validate:"required"`
	Np noop.Service         `validate:"required"`
	Ts transactions.Service `validate:"required"`
}*/

func NewActivity(b Backends) (*Activity, error) {
	return &Activity{
		b: b,
	}, nil
}

func (a *Activity) CreatePendingDeposit(ctx context.Context, depositId string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating pending transaction")

	deposit, err := Get(ctx, a.b, depositId)
	if err != nil {
		return "", err
	}

	acc, err := a.b.Accounts().Get(ctx, deposit.AccountID)
	if err != nil {
		return "", err
	}

	// TODO this needs to be better way to handle getting the data.
	trx, err := a.b.Transactions().CreatePending(ctx, &transactions.CreatePendingTransactionArgs{
		AccountID:   deposit.AccountID,
		Type:        "deposit",
		NetAmount:   deposit.Amount,
		Description: fmt.Sprintf("from %s bank account", "test"), // TODO Format to come from FS
		LedgerTransfers: []transactions.CreateLedgerTransferArgs{
			{
				LedgerID:        a.b.Noop().GetLedgerID(),
				DebitAccountID:  acc.LedgerAccountID,
				CreditAccountID: a.b.Noop().GetEquityAccountID(),
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

func (a *Activity) ProcessNoopDeposit(ctx context.Context, depositId string) error {

	deposit, err := Get(ctx, a.b, depositId)
	if err != nil {
		return err
	}

	acc, err := a.b.Accounts().Get(ctx, deposit.AccountID)
	if err != nil {
		return err
	}

	err = a.b.Noop().InitiateBankDeposit(ctx, &noop.BankDepositArgs{
		IdentityID:      acc.IdentityID,
		FundingSourceID: deposit.FundingSourceId,
		Amount:          deposit.Amount,
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) VoidPendingDeposit(ctx context.Context, trxId string) error {

	_, err := a.b.Transactions().VoidPending(ctx, trxId)
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) PostPendingDeposit(ctx context.Context, trxId string) error {

	_, err := a.b.Transactions().PostPending(ctx, trxId)
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) SetDepositState(ctx context.Context, depositId string, state deposits.State) error {
	err := SetState(ctx, a.b, depositId, state)
	if err != nil {
		return err
	}
	return nil
}
