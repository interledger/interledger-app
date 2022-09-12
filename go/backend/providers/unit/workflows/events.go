package workflows

import (
	"fmt"

	"gitlab.com/fynbos/backend/providers/unit/external"
	"go.temporal.io/sdk/workflow"
)

func UnitHandleEventsWorkflow(ctx workflow.Context, events []external.Event) error {
	fmt.Println(events)
	return nil
}
