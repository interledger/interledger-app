package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/pacioli"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/chimoney"
	"gitlab.com/fynbos/backend/providers/chimoney/external"
	"gitlab.com/fynbos/backend/slack"
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

func ExecuteCompleteKYCWorkflow(ctx context.Context, b Backends, externalID string, kycStatus kyc.Status) error {

	walletID, err := GetWalletID(ctx, b, externalID)
	if err != nil {
		return fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	wo := client.StartWorkflowOptions{
		ID:                    "chimoney_kyc_" + walletID + "_" + kycStatus.String(),
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
		_, executeErr = b.Temporal().ExecuteWorkflow(ctx, wo, ChimomeyCompleteKYC, walletID, kycStatus)
	}

	if executeErr != nil {
		return fmt.Errorf("%w %s", chimoney.ErrInternal, executeErr)
	}

	return nil
}

func SetInteracEmail(ctx context.Context, b Backends, walletID, email string) (*linkedaccounts.LinkedAccount, error) {
	las, err := b.LinkedAccounts().ListByWalletId(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	var interacAcc *linkedaccounts.LinkedAccount
	for _, la := range las {
		if la.Provider == chimoney.ProviderName && la.Type == chimoney.AccTypeInterac && la.DeletedAt.Time.IsZero() {
			interacAcc = &la
			break
		}
	}
	if interacAcc != nil && interacAcc.ProviderID != email {
		return nil, chimoney.ErrInteracAlreadyLinked
	} else if interacAcc != nil {
		return interacAcc, nil
	}

	return b.LinkedAccounts().Create(ctx, &linkedaccounts.CreateArgs{
		WalletID:            walletID,
		Name:                "Interac",
		Nickname:            "Interac",
		Mask:                email,
		Type:                chimoney.AccTypeInterac,
		Provider:            chimoney.ProviderName,
		ProviderID:          email,
		CanSend:             false,
		State:               linkedaccounts.Verified,
		SendCountry:         country.CA,
		SendCurrency:        currency.CAD,
		SendAvailability:    linkedaccounts.Few,
		SendNetwork:         "interac",
		CanReceive:          true,
		ReceiveCountry:      country.CA,
		ReceiveCurrency:     currency.CAD,
		ReceiveAvailability: linkedaccounts.Few,
		ReceiveNetwork:      "interac",
	})
}

func GetInteracEmail(ctx context.Context, b Backends, walletID string) (string, error) {
	las, err := b.LinkedAccounts().ListByWalletId(ctx, walletID)
	if err != nil {
		return "", fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	var interacAcc *linkedaccounts.LinkedAccount
	for _, la := range las {
		if la.Provider == chimoney.ProviderName && la.Type == chimoney.AccTypeInterac && la.DeletedAt.Time.IsZero() {
			interacAcc = &la
			break
		}
	}
	if interacAcc == nil {
		return "", fmt.Errorf("%w interac email not found", chimoney.ErrNotFound)
	}

	return interacAcc.ProviderID, nil
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
		RedirectURL:          fmt.Sprintf("%s/callbacks/chimoney", env.GetUrl()),
	})
	if err != nil {
		return "", fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	return resp.PaymentLink, nil
}

func ExecuteFinishWithdraw(ctx context.Context, b Backends, ec external.Client, IssueID string, status string, chiWalletID string) error {
	wo := client.StartWorkflowOptions{
		ID:                    "finish_chimoney_withdrawal_" + status + "_" + IssueID,
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
		_ = b.Temporal().GetWorkflow(ctx, wo.ID, "")
	} else {
		_, executeErr = b.Temporal().ExecuteWorkflow(ctx, wo, ExecuteChimoneyFinishWithdrawalWorkflow, IssueID, chiWalletID, status)
	}
	if executeErr != nil {
		return fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	return nil
}

func ExecuteWithdraw(ctx context.Context, b Backends, walletID, transactionID string) error {
	wo := client.StartWorkflowOptions{
		ID:                    "initiate_chimoney_withdrawal_" + walletID + "_" + transactionID,
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
		_ = b.Temporal().GetWorkflow(ctx, wo.ID, "")
	} else {
		_, executeErr = b.Temporal().ExecuteWorkflow(ctx, wo, ExecuteChimoneyWithdrawalWorkflow, walletID, transactionID)
	}
	if executeErr != nil {
		return fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	return nil
}

func Withdraw(ctx context.Context, b Backends, ex external.Client, walletID string, amt currency.Amount) (*external.WithdrawResponse, error) {
	chiWallet, err := GetChiWallet(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	email, err := GetInteracEmail(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	ul, err := b.Users().ListUsers(ctx, walletID)
	if err != nil {
		return nil, err
	}
	if len(ul) < 1 {
		return nil, fmt.Errorf("%w No user infomration found", chimoney.ErrInternal)
	}

	resp, err := ex.Withdraw(ctx, external.WithdrawalReq{
		DebitCurrency:       amt.Currency.String(),
		SubAccount:          chiWallet,
		TurnOffNotification: true,
		Interacs: []external.Interacs{{
			Name:      fmt.Sprintf("%s %s", ul[0].FirstName, ul[0].LastName),
			Email:     email,
			Amount:    amt.Float64(),
			Narration: "Interledger wallet withdrawal",
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	return resp, nil
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

func GetWalletID(ctx context.Context, b Backends, chiWalletID string) (string, error) {
	var walletID string
	err := b.DB().GetContext(ctx, &walletID, "SELECT wallet_ID FROM chi_money_wallets WHERE external_id=$1;", chiWalletID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w no wallet found for chiWalletID", chimoney.ErrNotFound)
	}
	if err != nil {
		return "", chimoney.ErrInternal
	}

	return walletID, nil
}

func Transfer(ctx context.Context, b Backends, ex external.Client, args chimoney.TransferArgs) error {
	sender, err := GetChiWallet(ctx, b, args.SendingWalletID)
	if err != nil {
		return err
	}

	receiver, err := GetChiWallet(ctx, b, args.ReceivingWalletID)
	if err != nil {
		return err
	}

	err = ex.Transfer(ctx, external.TransferReq{
		SenderSubAccount:    sender,
		ReceiverSubAccount:  receiver,
		Amount:              args.Amount,
		TurnOffNotification: true,
	})
	if err != nil {
		return fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	return err
}

func GetBalance(ctx context.Context, b Backends, linkedAccountID string) (*chimoney.Balance, error) {
	la, err := b.LinkedAccounts().Get(ctx, linkedAccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	if la.Provider != chimoney.ProviderName || la.Type != chimoney.AccTypeBalance {
		return nil, fmt.Errorf("%w linked account not correct type", chimoney.ErrNotFound)
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
	if tx[0].Code == pacioli.TransferPendingTransferExpired {
		return fmt.Errorf("%w transfer timed out code (%s)", chimoney.ErrTimedOut, tx[0].Code.String())
	}
	if tx[0].Code == pacioli.TransferPendingTransferNotPending {
		return fmt.Errorf("%w pending transfer not found code (%s)", chimoney.ErrNotFound, tx[0].Code.String())
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
		slack.SendToChannel(ctx, slack.ChannelNotifyErrors, "wallet-info-bot", fmt.Sprintf("*:::[Chimoney ERROR]:::* \n *RollbackReserve txID:* %s,\n *error:* %s", txID, err))
		return fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}
	if len(tx) == 0 {
		return nil
	}
	if tx[0].Code == pacioli.TransferPendingTransferNotFound ||
		tx[0].Code == pacioli.TransferPendingTransferAlreadyVoided ||
		tx[0].Code == pacioli.TransferPendingTransferExpired {
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
	redirectURL := fmt.Sprintf("%s/callbacks/chimoney?kyc", env.GetUrl())
	if !env.IsProd() {
		baseURL = "https://sandbox.chimoney.io"
	}

	widgetURL := fmt.Sprintf("%s/verify/kyc/%s?redirect=%s", baseURL, externalID, redirectURL)

	return widgetURL, nil
}

func CreateDeposit(ctx context.Context, b Backends, ex external.Client, issueID string) (chimoney.Await, error) {
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
		await, executeErr = b.Temporal().ExecuteWorkflow(ctx, wo, CreateChimoneyDepositWorkflow, issueID)
	}
	if executeErr != nil {
		return nil, fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	return await.Get, nil
}
