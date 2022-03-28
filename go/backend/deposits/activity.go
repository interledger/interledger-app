package deposits

import (
	"context"
	transactions "gitlab.com/fynbos/backend/accounttransactions"
	"go.temporal.io/sdk/activity"
)

type Activity struct {
	ts transactions.Service
}

func (s *Activity) PrepareWithdrawUserFunds(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Running create Identity")
	//trx, err := s.ts.Create(ctx, &transactions.CreateTransactionArgs{
	//	AccountID:   acc.ID,
	//	Type:        "deposit",
	//	NetAmount:   args.Amount,
	//	Description: fmt.Sprintf("from %s bank account", fundingSource.Mask), // TODO Format to come from FS
	//	LedgerTransfers: []transactions.CreateLedgerTransferArgs{
	//		{
	//			LedgerID:        s.noop.GetLedgerID(),
	//			DebitAccountID:  acc.LedgerAccountID,
	//			CreditAccountID: s.noop.GetEquityAccountID(),
	//			Amount:          args.Amount,
	//			// Code: "1", // TODO: define ledger transfer codes.
	//			Flags: transactions.LedgerTransferFlags{
	//				TwoPhaseCommit: true,
	//			},
	//		},
	//	},
	//})
	return nil
}

func (s *Activity) InitiateProviderTransfer(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Running create Identity")
	return nil
}

func (s *Activity) CommitWithdrawUserFunds(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Running create Identity")
	return nil
}
