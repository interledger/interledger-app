package ops

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
	"gitlab.com/fynbos/backend/providers/tabapay/workflows"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

func CreateCard(ctx context.Context, b Backends, args tabapay.CreateCardArgs) (tabapay.Await, error) {
	cardQuery, err := b.External().QueryCard(ctx, external.QueryCardArgs{
		Card: &external.Card{
			AccountNumber:  args.CardNumber,
			ExpirationDate: args.ExpirationDate,
			SecurityCode:   args.CVV,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}
	if cardQuery.Card.Pull.Type != external.CardTypeDebit || cardQuery.Card.Push.Type != external.CardTypeDebit {
		return nil, fmt.Errorf("%w Not a debit card.", tabapay.ErrInvalidCard)
	}

	wo := client.StartWorkflowOptions{
		ID:                       "tabapay_create_card" + args.IdempotencyKey,
		TaskQueue:                "backend",
		WorkflowExecutionTimeout: 2 * time.Minute,
		WorkflowIDReusePolicy:    enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}
	await, err := b.Temporal().ExecuteWorkflow(ctx, wo, workflows.CreateTabapayCardWorkflow, args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	return await.Get, nil
}

func PullFromCard(ctx context.Context, b Backends, args PullFromCardArgs) (string, error) {
	transactionResponse, err := b.External().CreateTransaction(ctx, external.CreateTransactionArgs{
		ReferenceID: args.ReferenceID,
		Type:        external.TransactionTypePull,
		Currency:    args.Amount.Currency.String(),
		Amount:      fmt.Sprintf("%f", args.Amount.Float64()),
		Accounts: external.CreateTransactionAccounts{
			SourceAccountID:      args.ProviderID,
			DestinationAccountID: args.SettlementAccountID,
		},
	})
	if err != nil {
		return "", fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	return transactionResponse.TransactionID, nil
}

func PushToCard(ctx context.Context, b Backends, args PullFromCardArgs) (string, error) {
	transactionResponse, err := b.External().CreateTransaction(ctx, external.CreateTransactionArgs{
		ReferenceID: args.ReferenceID,
		Type:        external.TransactionTypePush,
		Currency:    args.Amount.Currency.String(),
		Amount:      fmt.Sprintf("%f", args.Amount.Float64()),
		Accounts: external.CreateTransactionAccounts{
			SourceAccountID:      args.SettlementAccountID,
			DestinationAccountID: args.ProviderID,
		},
	})
	if err != nil {
		return "", fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	return transactionResponse.TransactionID, nil
}

func GetTransaction(ctx context.Context, b Backends, id string) (*tabapay.Transaction, error) {
	trxResp, err := b.External().RetrieveTransaction(ctx, id)
	if errors.Is(err, external.ErrNotFound) {
		return nil, tabapay.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	floatAmount, err := strconv.ParseFloat(trxResp.AmountUSD, 64)
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	return &tabapay.Transaction{
		ID:             id,
		ReferenceID:    trxResp.ReferenceID,
		Status:         trxResp.Status,
		OriginalStatus: trxResp.Originally,
		Amount:         currency.FromFloat64(floatAmount, "USD"),
		ReversalStatus: trxResp.ReversalStatus,
	}, nil
}
