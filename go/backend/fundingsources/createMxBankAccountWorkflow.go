package fundingsources

import (
	"go.temporal.io/sdk/workflow"
)

type CreateMxBankAccountWorkflowArgs struct {
	FundingSourceID string `validate:"required"`
	MxUserGuid      string `validate:"required"`
	MxMemberGuid    string `validate:"required"`
}

type AccountInfo struct{}

func CreateMxBankAccountWorkflow(ctx workflow.Context, args *CreateMxBankAccountWorkflowArgs) error {

	return nil
}
