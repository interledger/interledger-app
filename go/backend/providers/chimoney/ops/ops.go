package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/chimoney"
	"gitlab.com/fynbos/backend/providers/chimoney/external"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
)

func CreateWallet(ctx context.Context, b Backends, walletID string) (chimoney.Await, error) {
	// Check that user has an interac email before beginning workflow
	_, err := GetInteracEmail(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	wo := client.StartWorkflowOptions{
		ID:                    "chimoney_create_wallet_" + walletID,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	var workflowStatus enums.WorkflowExecutionStatus
	wflow, err := b.Temporal().DescribeWorkflowExecution(ctx, wo.ID, "")
	switch err.(type) {
	case *serviceerror.Internal,
		*serviceerror.Unavailable,
		*serviceerror.InvalidArgument:
		return nil, fmt.Errorf("%w %s", chimoney.ErrInternal, err)
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
		await, executeErr = b.Temporal().ExecuteWorkflow(ctx, wo, CreateChimoneyUserWorkflow, walletID)
	}
	if executeErr != nil {
		return nil, fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	return await.Get, nil
}

func UpsertInteracEmail(ctx context.Context, b Backends, walletID, email string) (string, error) {
	var mail string
	err := b.DB().GetContext(ctx, &mail, `INSERT INTO chi_money_interac_emails (wallet_id, email)
		VALUES ($1, $2)
		ON CONFLICT (wallet_id) 
		DO UPDATE SET 
			email = EXCLUDED.email,
			updated_at = NOW() RETURNING email;`, walletID, email)
	if err != nil {
		return "", fmt.Errorf("%w failed to insert interloc email: %s", chimoney.ErrInternal, err)
	}

	return mail, nil
}

func GetInteracEmail(ctx context.Context, b Backends, walletID string) (string, error) {
	var email string
	err := b.DB().GetContext(ctx, &email, "SELECT email FROM chi_money_interac_emails WHERE wallet_id=$1", walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w user interac email not found", chimoney.ErrNotFound)
	}
	if err != nil {
		return "", fmt.Errorf("%w %s", chimoney.ErrNotFound, err)
	}

	return email, nil
}

func CreateDepositLink(ctx context.Context, b Backends, ex external.Client, walletID string, amt currency.Amount) (string, error) {
	email, err := GetInteracEmail(ctx, b, walletID)
	if err != nil {
		return "", err
	}

	var chiWallet string
	err = b.DB().GetContext(ctx, &chiWallet, "SELECT external_id FROM chi_money_wallets WHERE wallet_id=$1", walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w no chimoney wallet found for user", chimoney.ErrNotFound)
	}
	if err != nil {
		return "", chimoney.ErrInternal
	}

	resp, err := ex.Deposit(ctx, external.DepositReq{
		Amount:               amt.FormatAmount(),
		Currency:             amt.Currency.String(),
		ChimoneyWallet:       chiWallet,
		Email:                email,
		TurnOffNotifications: true,
	})
	if err != nil {
		return "", fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	return resp.PaymentLink, nil
}

func Withdraw(ctx context.Context, b Backends, ex external.Client, walletID string, amt currency.Amount) error {
	var chiWallet string
	err := b.DB().GetContext(ctx, &chiWallet, "SELECT external_id FROM chi_money_wallets WHERE wallet_id=$1", walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w no chimoney wallet found for user", chimoney.ErrNotFound)
	}
	if err != nil {
		return chimoney.ErrInternal
	}

	email, err := GetInteracEmail(ctx, b, walletID)
	if err != nil {
		return err
	}

	userInfo, err := b.KYC().GetIndividualDetails(ctx, walletID)
	if err != nil {
		return fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	err = ex.Withdraw(ctx, external.WithdrawalReq{
		DebitCurrency:       amt.Currency.String(),
		SubAccount:          chiWallet,
		TurnOffNotification: true,
		Interacs: []external.Interacs{{
			Name:      userInfo.FirstName + " " + userInfo.LastName,
			Email:     email,
			Amount:    amt.Float64(),
			Narration: "Fynbos wallet withdrawal",
		}},
	})
	if err != nil {
		return fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	return nil
}
