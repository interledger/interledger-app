package ops

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/providers/tabapay/workflows"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

func CreateCard(ctx context.Context, b Backends, args tabapay.CreateCardArgs) (tabapay.Await, error) {
	// TODO: check that this is a debit card

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
