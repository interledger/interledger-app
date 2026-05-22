package ops

import (
	"github.com/google/uuid"
	"go.temporal.io/sdk/workflow"
)

// sideEffectUUID generates a new UUID as a Temporal side effect, ensuring deterministic replay.
func sideEffectUUID(ctx workflow.Context) (string, error) {
	var id string
	err := workflow.SideEffect(ctx, func(ctx workflow.Context) any { return uuid.NewString() }).Get(&id)
	return id, err
}
