package ops

import (
	"context"
	"fmt"
	"time"

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
