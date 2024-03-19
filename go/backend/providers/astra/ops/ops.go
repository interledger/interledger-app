package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/currency"

	"gitlab.com/fynbos/backend/providers/astra"
	"gitlab.com/fynbos/backend/providers/astra/external"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
)

func DebitCard(ctx context.Context, b Backends, args astra.CardToAccountArgs) (string, error) {
	token, err := GetToken(ctx, b, args.WalletID)
	if err != nil {
		return "", err
	}

	var accID string
	err = b.DB().GetContext(ctx, &accID, "SELECT account_id FROM astra_accounts WHERE wallet_id=$1", args.WalletID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", astra.ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%w %s", astra.ErrInternal, err)
	}

	userID, err := getUserID(ctx, b, args.WalletID)
	if err != nil {
		return "", err
	}

	tx, err := b.External().CardToAccount(ctx, token, external.CardToAccountArgs{
		IdempotencyKey:      args.IdempotencyKey,
		Name:                args.Name,
		Amount:              args.Amount.Float64(),
		ClientCorrelationID: args.ClientCorrelationID,
		DebitFeePercent:     args.DebitFeePercent,
		Card: external.Source{
			ID: args.CardID,
		},
		Account: external.Destination{
			ID:     accID,
			UserID: userID,
		},
	})
	if err != nil {
		return "", err
	}

	return tx.ID, nil
}

func CreditCard(ctx context.Context, b Backends, args astra.AccountToCardsArgs) (string, error) {
	token, err := GetToken(ctx, b, args.WalletID)
	if err != nil {
		return "", err
	}

	var accID string
	err = b.DB().GetContext(ctx, &accID, "SELECT account_id FROM astra_accounts WHERE wallet_id=$1", args.WalletID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", astra.ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%w %s", astra.ErrInternal, err)
	}

	userID, err := getUserID(ctx, b, args.WalletID)
	if err != nil {
		return "", err
	}

	tx, err := b.External().AccountToCard(ctx, token, external.AccountToCardArgs{
		IdempotencyKey: args.IdempotencyKey,
		Name:           args.Name,
		Amount:         args.Amount.Float64(),
		Card: external.Destination{
			ID:     args.CardID,
			UserID: userID,
		},
		Account: external.Source{
			ID: accID, // TODO: this should be astra account id of the Fynbos op account
		},
		SettlementMode: "net_debit",
	})
	if err != nil {
		return "", err
	}

	return tx.ID, nil
}

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

func LookupTransfer(ctx context.Context, b Backends, walletID, txID string) (*astra.Transfer, error) {
	token, err := GetToken(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	tx, err := b.External().GetTransfer(ctx, token, txID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", astra.ErrInternal, err)
	}

	return &astra.Transfer{
		ID:                    tx.ID,
		RoutineType:           tx.RoutineType,
		RoutineName:           tx.RoutineName,
		RoutineID:             tx.RoutineID,
		ClientCorrelationID:   tx.ClientCorrelationID,
		Amount:                currency.FromFloat64(tx.Amount, currency.USD),
		PaymentType:           tx.PaymentType,
		AstraSettlementReason: tx.AstraSettlementReason,
		FailureReason:         tx.FailureReason,
		Status:                tx.Status,
	}, nil
}

func LookupRoutine(ctx context.Context, b Backends, walletID, routineID string) (*astra.Routine, error) {
	token, err := GetToken(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	routine, err := b.External().GetRoutine(ctx, token, routineID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", astra.ErrInternal, err)
	}

	return &astra.Routine{
		ID:        routine.ID,
		Type:      routine.Type,
		Name:      routine.Name,
		Status:    routine.Status,
		Active:    routine.Active,
		Blocked:   routine.Blocked,
		StartDate: routine.StartDate,
	}, nil
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

func getUserID(ctx context.Context, b Backends, walletID string) (string, error) {
	type intentDB struct {
		UserID   string `db:"user_id"`
		IntentID string `db:"intent_id"`
	}
	var intent intentDB
	err := b.DB().GetContext(ctx, &intent, "SELECT user_id, intent_id FROM astra_user_intents WHERE wallet_id=$1", walletID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w %s", astra.ErrInternal, err)
	}
	if intent.UserID != "" {
		return intent.UserID, nil
	}

	// Lookup the intent and update the userID and status
	exIntent, err := b.External().GetIntent(ctx, intent.IntentID)
	if err != nil {
		return "", fmt.Errorf("%w %s", astra.ErrInternal, err)
	}

	_, err = b.DB().ExecContext(ctx, "UPDATE astra_user_intents SET  user_id=$1, status=$2, updated_at=now() WHERE intent_id=$3", exIntent.UserID, exIntent.Status, exIntent.ID)
	if err != nil {
		return "", fmt.Errorf("%w %s", astra.ErrInternal, err)
	}

	if exIntent.UserID == "" {
		return "", fmt.Errorf("%w astra user not converted for wallet ID", astra.ErrUserNotReady)
	}

	return exIntent.UserID, nil
}
