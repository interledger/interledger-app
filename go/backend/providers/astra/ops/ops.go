package ops

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/providers/astra"
	"gitlab.com/fynbos/backend/providers/astra/external"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
)

func CreateCard(ctx context.Context, b Backends, args astra.CreateCardArgs) (astra.Await, error) {
	wo := client.StartWorkflowOptions{
		ID:                       "astra_create_card_" + args.BasisTheoryTokenID,
		TaskQueue:                "backend",
		WorkflowExecutionTimeout: 2 * time.Minute,
		WorkflowIDReusePolicy:    enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	var workflowStatus enums.WorkflowExecutionStatus
	wflow, err := b.Temporal().DescribeWorkflowExecution(ctx, wo.ID, "")
	switch err.(type) {
	case *serviceerror.Internal,
		*serviceerror.Unavailable,
		*serviceerror.InvalidArgument:
		return nil, fmt.Errorf("%w %s", astra.ErrInternal, err)
	case *serviceerror.NotFound:
		// do nothing
	default:
		if wflow != nil {
			workflowStatus = wflow.GetWorkflowExecutionInfo().Status
		}
	}

	// return workflow if it's running
	var await client.WorkflowRun
	var executeErr error
	if workflowStatus == enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
		await = b.Temporal().GetWorkflow(ctx, wo.ID, "")
	} else {
		await, executeErr = b.Temporal().ExecuteWorkflow(ctx, wo, CreateAstraCardWorkflow, args)
	}
	if executeErr != nil {
		return nil, fmt.Errorf("%w %s", astra.ErrInternal, err)
	}

	return await.Get, nil
}

func CreateIntent(ctx context.Context, b Backends, walletID string) error {
	id, err := b.KYC().GetIndividualDetails(ctx, walletID)
	if err != nil {
		return fmt.Errorf("%w %s", astra.ErrInternal, err)
	}
	if id.Address == nil {
		return fmt.Errorf("%w incomplete KYC for astra, missing address", astra.ErrNotFound)
	}

	idNums, err := b.KYC().GetPersonaIDNumbers(ctx, walletID)
	if err != nil {
		return fmt.Errorf("%w %s", astra.ErrInternal, err)
	}

	ul, err := b.Users().ListUsers(ctx, walletID)
	if err != nil {
		return fmt.Errorf("%w %s", astra.ErrInternal, err)
	}

	if len(ul) <= 0 {
		return fmt.Errorf("%w no users found for wallet", astra.ErrNotFound)
	}

	u := ul[0]

	state := id.Address.State
	if len(state) > 2 {
		state = id.Address.State[len(state)-2:]
	}

	phone := u.PhoneNumber
	if !strings.HasPrefix(phone, "+") {
		phone = "+" + phone
	}

	args := external.CreateIntentReq{
		Email:          u.Email,
		Phone:          phone,
		FirstName:      id.FirstName,
		LastName:       id.LastName,
		Address1:       id.Address.Line1,
		Address2:       id.Address.Line2,
		City:           id.Address.City,
		State:          state,
		PostalCode:     id.Address.ZipCode,
		DateOfBirth:    id.DateOfBirth.Format(time.DateOnly),
		SocialSecurity: strings.ReplaceAll(idNums.SocialSecurity, "-", ""),
		IPAddress:      id.IPAddress,
	}

	intentID, err := b.External().CreateIntent(ctx, args)
	if err != nil {
		return err
	}

	_, err = b.DB().ExecContext(ctx, "INSERT INTO astra_user_intents (wallet_id, intent_id, status, user_id) VALUES ($1, $2, 'unknown', '')", walletID, intentID)
	if err != nil {
		return fmt.Errorf("%w %s", astra.ErrInternal, err)
	}

	return nil
}
