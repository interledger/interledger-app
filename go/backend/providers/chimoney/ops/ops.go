package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/pacioli"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/chimoney"
	"gitlab.com/fynbos/backend/providers/chimoney/external"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
)

func CreateWallet(ctx context.Context, b Backends, walletID string) (chimoney.Await, error) {
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

func WatchForSuccessfulKYC(ctx context.Context, b Backends, walletID string) error {
	wo := client.StartWorkflowOptions{
		ID:                    "chimoney_poll_wallet_kyc" + walletID,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	var workflowStatus enums.WorkflowExecutionStatus
	wflow, err := b.Temporal().DescribeWorkflowExecution(ctx, wo.ID, "")
	switch err.(type) {
	case *serviceerror.Internal,
		*serviceerror.Unavailable,
		*serviceerror.InvalidArgument:
		return fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	case *serviceerror.NotFound:
		// do nothing
	default:
		if wflow != nil {
			workflowStatus = wflow.GetWorkflowExecutionInfo().Status
		}
	}

	// return workflow if it's running
	var executeErr error
	if workflowStatus == enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
		// Do nothing
	} else {
		_, executeErr = b.Temporal().ExecuteWorkflow(ctx, wo, ChimomeyWatchForSuccessfulKYC, walletID)
	}

	return executeErr
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
	userList, err := b.Users().ListUsers(ctx, walletID)
	if err != nil {
		return "", err
	}
	if len(userList) < 1 {
		return "", fmt.Errorf("%w No user found for wallet", chimoney.ErrInternal)
	}

	chiWallet, err := GetChiWallet(ctx, b, walletID)
	if err != nil {
		return "", err
	}

	resp, err := ex.Deposit(ctx, external.DepositReq{
		Amount:               amt.FormatAmount(),
		Currency:             amt.Currency.String(),
		ChimoneyWallet:       chiWallet,
		Email:                userList[0].Email,
		TurnOffNotifications: true,
	})
	if err != nil {
		return "", fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	return resp.PaymentLink, nil
}

func Withdraw(ctx context.Context, b Backends, ex external.Client, walletID string, amt currency.Amount) error {
	chiWallet, err := GetChiWallet(ctx, b, walletID)
	if err != nil {
		return err
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

func CreateWithdrawal(ctx context.Context, b Backends, walletID string, amount currency.Amount) (chimoney.Await, error) {
	// Check that user has a chimoney wallet before beginning workflow
	_, err := GetChiWallet(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	// Check that user has an interac email before beginning workflow
	_, err = GetInteracEmail(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	wo := client.StartWorkflowOptions{
		ID:                    "chimoney_create_withdrawal_" + walletID + "_" + amount.FormatAmount(),
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
		await, executeErr = b.Temporal().ExecuteWorkflow(ctx, wo, CreateChimoneyWithdrawalWorkflow, walletID, amount)
	}
	if executeErr != nil {
		return nil, fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	return await.Get, nil
}

func GetChiWallet(ctx context.Context, b Backends, walletID string) (string, error) {
	var chiWallet string
	err := b.DB().GetContext(ctx, &chiWallet, "SELECT external_id FROM chi_money_wallets WHERE wallet_id=$1", walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w no chimoney wallet found for user", chimoney.ErrNotFound)
	}
	if err != nil {
		return "", chimoney.ErrInternal
	}

	return chiWallet, err
}

func Transfer(ctx context.Context, b Backends, ex external.Client, args chimoney.TransferArgs) error {
	sender, err := GetChiWallet(ctx, b, args.SendingWalletID)
	if err != nil {
		return err
	}

	receiver, err := GetChiWallet(ctx, b, args.SendingWalletID)
	if err != nil {
		return err
	}

	err = ex.Transfer(ctx, external.TransferReq{
		SenderSubAccount:   sender,
		ReceiverSubAccount: receiver,
		Amount:             args.Amount,
	})
	if err != nil {
		return fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	return err
}

func ReserveBalance(ctx context.Context, b Backends, linkedAccountID, txID string, amt currency.Amount, timeout time.Duration) (*chimoney.Balance, error) {
	if amt.Currency != currency.CAD {
		return nil, fmt.Errorf("%w %s not supported", chimoney.ErrInternal, amt.Currency)
	}

	la, err := b.LinkedAccounts().Get(ctx, linkedAccountID)
	if err != nil {
		return nil, err
	}
	if la.Provider != chimoney.ProviderName || la.Type != chimoney.AccTypeBalance {
		return nil, fmt.Errorf("%w Not a Chimoney linked account", chimoney.ErrInternal)
	}

	opsAcc := chimoney.CADOpsAccount
	ledger := chimoney.LedgerIDCAD
	tx, err := b.Pacioli().CreateTransfers(ctx, []pacioli.CreateTransferArgs{
		{
			ID:              txID,
			Amount:          amt.Value,
			DebitAccountID:  la.ID,
			CreditAccountID: opsAcc,
			Pending:         true,
			Code:            1,
			Timeout:         uint64(timeout),
			Ledger:          ledger,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}
	if len(tx) > 0 {
		if tx[0].Code == pacioli.TransferExceedsCredits || tx[0].Code == pacioli.TransferExceedsDebits || tx[0].Code == pacioli.TransferExceedsPendingTransferAmount {
			return nil, fmt.Errorf("%w insufficiens balance cod (%s)", chimoney.ErrInsufficientBalance, tx[0].Code.String())
		}
		if tx[0].Code != 0 {
			return nil, fmt.Errorf("%w non success code (%s)", chimoney.ErrInternal, tx[0].Code.String())
		}
	}

	accs, err := b.Pacioli().GetAccounts(ctx, []string{la.ID})
	if err != nil {
		return nil, fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	if len(accs) != 1 {
		return nil, fmt.Errorf("%w account not found", chimoney.ErrNotFound)
	}

	return &chimoney.Balance{
		Total:     currency.FromUInt64(accs[0].CreditsPosted-accs[0].DebitsPosted, la.SendCurrency),
		Available: currency.FromUInt64(accs[0].CreditsPosted-accs[0].DebitsPosted-accs[0].DebitsPending, la.SendCurrency),
	}, nil
}

func FinaliseReserve(ctx context.Context, b Backends, trxID string) error {
	tx, err := b.Pacioli().PostTransfers(ctx, []string{trxID})
	if err != nil {
		return fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}
	if len(tx) == 0 {
		return nil
	}
	if tx[0].Code == pacioli.TransferPendingTransferAlreadyPosted {
		return nil
	}
	if tx[0].Code == pacioli.TransferExceedsCredits || tx[0].Code == pacioli.TransferExceedsDebits || tx[0].Code == pacioli.TransferExceedsPendingTransferAmount {
		return fmt.Errorf("%w insufficiens balance cod (%s)", chimoney.ErrInsufficientBalance, tx[0].Code.String())
	}
	if tx[0].Code != 0 {
		return fmt.Errorf("%w non success code (%s)", chimoney.ErrInternal, tx[0].Code.String())
	}

	return nil
}

func AssignBalance(ctx context.Context, b Backends, linkedAccountID, txID string, amt currency.Amount) (*chimoney.Balance, error) {
	la, err := b.LinkedAccounts().Get(ctx, linkedAccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	if la.Provider != chimoney.ProviderName || la.Type != chimoney.AccTypeBalance {
		return nil, fmt.Errorf("%w chimoney account not correct type", chimoney.ErrNotFound)
	}

	opsAcc := chimoney.CADOpsAccount
	ledger := chimoney.LedgerIDCAD
	tx, err := b.Pacioli().CreateTransfers(ctx, []pacioli.CreateTransferArgs{
		{
			ID:              txID,
			Amount:          amt.Value,
			CreditAccountID: la.ID,
			DebitAccountID:  opsAcc,
			Pending:         false,
			Code:            1,
			Ledger:          ledger,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}
	if len(tx) != 0 {
		if tx[0].Code == pacioli.TransferExceedsCredits || tx[0].Code == pacioli.TransferExceedsDebits || tx[0].Code == pacioli.TransferExceedsPendingTransferAmount {
			return nil, fmt.Errorf("%w insufficiens balance cod (%s)", chimoney.ErrInsufficientBalance, tx[0].Code.String())
		}
		if tx[0].Code != 0 {
			return nil, fmt.Errorf("%w non success code (%s)", chimoney.ErrInternal, tx[0].Code.String())
		}
	}

	accs, err := b.Pacioli().GetAccounts(ctx, []string{la.ID})
	if err != nil {
		return nil, fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	if len(accs) != 1 {
		return nil, fmt.Errorf("%w account not found", chimoney.ErrNotFound)
	}

	return &chimoney.Balance{
		Total:     currency.FromUInt64(accs[0].CreditsPosted-accs[0].DebitsPosted, la.SendCurrency),
		Available: currency.FromUInt64(accs[0].CreditsPosted-accs[0].DebitsPosted-accs[0].DebitsPending, la.SendCurrency),
	}, nil
}

func RollbackReserve(ctx context.Context, b Backends, txID string) error {
	tx, err := b.Pacioli().VoidTransfers(ctx, []string{txID})
	if err != nil {
		return fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}
	if len(tx) == 0 {
		return nil
	}
	if tx[0].Code != 0 {
		return fmt.Errorf("%w non success code (%s)", chimoney.ErrInternal, tx[0].Code.String())
	}

	return nil
}

func GetKYCWidget(ctx context.Context, b Backends, walletID string) (string, error) {
	externalID, err := GetChiWallet(ctx, b, walletID)
	if errors.Is(err, chimoney.ErrNotFound) {
		await, innerErr := CreateWallet(ctx, b, walletID)
		if innerErr != nil {
			return "", innerErr
		}

		innerErr = await(ctx, &externalID)
		if innerErr != nil {
			return "", innerErr
		}
	} else if err != nil {
		return "", err
	}

	baseURL := "https://dash.chimoney.io"
	if !env.IsProd() {
		baseURL = "https://sandbox.chimoney.io"
	}

	widgetURL := fmt.Sprintf("%s/verify/kyc/%s", baseURL, externalID)

	return widgetURL, nil
}

func CreateDeposit(ctx context.Context, b Backends, ex external.Client, walletID, issueID string) (chimoney.Await, error) {
	externalID, err := GetChiWallet(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	depositDetails, err := ex.VerifyPayment(ctx, external.VerifyPaymentReq{
		ChiWallet: externalID,
		IssueID:   issueID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	amount, err := currency.FromString(depositDetails.Amount, currency.ParseCurrency(depositDetails.Currency))
	if err != nil {
		return nil, fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	wo := client.StartWorkflowOptions{
		ID:                    "chimoney_deposit_" + issueID,
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
		await, executeErr = b.Temporal().ExecuteWorkflow(ctx, wo, CreateChimoneyDepositWorkflow, walletID, issueID, amount)
	}
	if executeErr != nil {
		return nil, fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	return await.Get, nil
}
