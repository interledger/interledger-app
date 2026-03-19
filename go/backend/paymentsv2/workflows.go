package paymentsv2

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/providers/xago"
	"gitlab.com/fynbos/pacioli"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type Activity struct {
	service *Service
}

func NewActivity(service *Service) *Activity {
	return &Activity{
		service: service,
	}
}

func (a *Activity) CreateTransfer(ctx context.Context, transferID string, payment *Payment) (err error) {
	_, err = a.service.LedgerClient.CreateTransfers(context.Background(), []pacioli.CreateTransferArgs{
		{
			ID:              transferID,
			DebitAccountID:  payment.SenderAccountID,
			CreditAccountID: payment.ReceiverAccountID,
			Amount:          payment.SenderCurrency.RawAmount().Uint64(),
			Pending:         false,
			Code:            1,
			Ledger:          xago.LedgerIDZAR,
		},
	})

	return
}

func (a *Activity) Update(ctx context.Context, transferID string, payment *Payment) (err error) {
	payment.SetState("COMPLETED")
	payment.AddTrasnsfer(transferID)

	return a.service.Repo.Update(context.Background(), payment)
}

func PaymentWorkflowV2(ctx workflow.Context, payment *Payment) (err error) {
	if payment == nil {
		panic("payment is nil")
	}

	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			NonRetryableErrorTypes: []string{"ErrInvalidStateTransition", "ErrNotFound"},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Executing payment", "id", payment.ID)

	transferID := uuid.NewString()

	err = workflow.ExecuteActivity(ctx, a.CreateTransfer, transferID, payment).Get(ctx, nil)
	if err != nil {
		return
	}
	err = workflow.ExecuteActivity(ctx, a.Update, transferID, payment).Get(ctx, nil)
	if err != nil {
		return
	}

	return
}
