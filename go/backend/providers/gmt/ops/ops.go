package ops

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/providers/gmt"
)

func Authenticate3DS(ctx context.Context, b Backends, args gmt.Authenticate3DSArgs) error {
	var ref WorkflowRef
	err := b.DB().GetContext(
		ctx,
		&ref,
		"SELECT external_id, workflow_id, workflow_run_id  FROM  gmt_workflow_refs WHERE completed=false AND external_id=$1 AND activity_name=$2;",
		args.OutgoingPaymentID,
		card2AchAuthenticate3ds,
	)
	if errors.Is(err, gmt.ErrNotFound) {
		return gmt.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("%w %s", gmt.ErrInternal, err)
	}

	err = b.Temporal().SignalWorkflow(ctx, ref.WorkflowID, ref.WorkflowRunID, gmt3DSChannel, args)
	if err != nil {
		return fmt.Errorf("%w %s", gmt.ErrInternal, err)
	}

	return nil
}
