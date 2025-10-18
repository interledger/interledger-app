package ops

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/providers/pti/external"
	"gitlab.com/fynbos/backend/transactions"

	"gitlab.com/fynbos/env"

	"gitlab.com/fynbos/backend/slack"

	"gitlab.com/fynbos/pacioli"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
)

var (
	userFields       = "id, external_id, wallet_id, status, assessment_status, created_at, updated_at"
	userInsertFields = "external_id, wallet_id"
)

func CreateUser(ctx context.Context, b Backends, walletID string) (pti.Await, error) {
	wo := client.StartWorkflowOptions{
		ID:                    "pti_create_user_" + walletID,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	var workflowStatus enums.WorkflowExecutionStatus
	wflow, err := b.Temporal().DescribeWorkflowExecution(ctx, wo.ID, "")
	switch err.(type) {
	case *serviceerror.Internal,
		*serviceerror.Unavailable,
		*serviceerror.InvalidArgument:
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
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
		await, executeErr = b.Temporal().ExecuteWorkflow(ctx, wo, CreateUserWorkflow, walletID)
	}
	if executeErr != nil {
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	return await.Get, nil
}

func CreateWallet(ctx context.Context, b Backends, args pti.CreateWalletArgs) (pti.Await, error) {
	wo := client.StartWorkflowOptions{
		ID:                    "pti_create_wallet_" + args.WalletID + "_" + args.Currency.String(),
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	var workflowStatus enums.WorkflowExecutionStatus
	wflow, err := b.Temporal().DescribeWorkflowExecution(ctx, wo.ID, "")
	switch err.(type) {
	case *serviceerror.Internal,
		*serviceerror.Unavailable,
		*serviceerror.InvalidArgument:
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
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
		await, executeErr = b.Temporal().ExecuteWorkflow(ctx, wo, CreateWalletWorkflow, args)
	}
	if executeErr != nil {
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	return await.Get, nil
}

func GetUser(ctx context.Context, b Backends, walletID string) (*pti.User, error) {
	var user pti.User
	err := b.DB().GetContext(ctx, &user, fmt.Sprintf("SELECT %s from pti_users where wallet_id=$1;", userFields), walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, pti.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func GetUserFromExternalID(ctx context.Context, b Backends, externalID string) (*pti.User, error) {
	var user pti.User
	err := b.DB().GetContext(ctx, &user, fmt.Sprintf("SELECT %s from pti_users where external_id=$1;", userFields), externalID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, pti.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func GetWallet(ctx context.Context, b Backends, external external.Client, linkedAccountID string) (*pti.Wallet, error) {
	la, err := b.LinkedAccounts().Get(ctx, linkedAccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
	}
	if la.Provider != pti.ProviderName {
		return nil, pti.ErrNotFound
	}

	externalUser, err := GetUser(ctx, b, la.WalletID)
	if err != nil {
		return nil, err
	}

	w, err := external.GetWallet(ctx, externalUser.ExternalID, la.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	return &pti.Wallet{
		ID:        w.WalletID,
		UserID:    externalUser.ExternalID,
		Reference: w.Reference,
	}, nil
}

func GetBalance(ctx context.Context, b Backends, linkedAccountID string) (*pti.Balance, error) {
	la, err := b.LinkedAccounts().Get(ctx, linkedAccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	if la.Provider != pti.ProviderName || la.Type != pti.AccTypeBalance {
		return nil, fmt.Errorf("%w linked account not correct type", pti.ErrNotFound)
	}

	accs, err := b.Pacioli().GetAccounts(ctx, []string{la.ID})
	if err != nil {
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	if len(accs) != 1 {
		return nil, fmt.Errorf("%w account not found", pti.ErrNotFound)
	}

	return &pti.Balance{
		Total:     currency.FromUInt64(accs[0].CreditsPosted-accs[0].DebitsPosted, la.SendCurrency),
		Available: currency.FromUInt64(accs[0].CreditsPosted-accs[0].DebitsPosted-accs[0].DebitsPending, la.SendCurrency),
	}, nil
}

func DepositToWallet(ctx context.Context, b Backends, ec external.Client, args pti.TransactionArgs) (string, error) {
	externalUser, err := GetUser(ctx, b, args.WalletID)
	if err != nil {
		return "", err
	}

	las, err := b.LinkedAccounts().ListByWalletId(ctx, args.WalletID)
	if err != nil {
		return "", fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	var balance *linkedaccounts.LinkedAccount
	for _, la := range las {
		if la.Provider == pti.ProviderName && la.Type == pti.AccTypeBalance {
			balance = &la
			break
		}
	}

	// only allow depositing from card or bank
	var source *linkedaccounts.LinkedAccount
	var sourceType string
	for _, la := range las {
		if la.Provider == pti.ProviderName && la.Type == pti.TypeCard && la.ID == args.LinkedAccountID {
			source = &la
			sourceType = "card"
			break
		}

		if la.Provider == pti.ProviderName && la.Type == pti.TypeBank && la.ID == args.LinkedAccountID {
			source = &la
			sourceType = "bank"
			break
		}
	}
	if balance == nil {
		return "", fmt.Errorf("%w balance account not found", pti.ErrNotFound)
	}
	if source == nil {
		return "", fmt.Errorf("%w source account not found", pti.ErrNotFound)
	}

	kycDetails, err := b.KYC().GetIndividualDetails(ctx, args.WalletID)
	if err != nil {
		return "", err
	}

	txID, err := ec.WalletDeposit(ctx, external.DepositArgs{
		RequestID:                 args.PaymentID,
		ScenarioID:                pti.ScenarioDeposit,
		SessionID:                 args.PaymentID,
		UserID:                    externalUser.ExternalID,
		ExternalWalletID:          balance.ProviderID,
		ExternalPaymentMethodID:   source.ProviderID,
		ExternalPaymentMethodType: sourceType,
		Amount:                    args.Amount,
		AccountHolderName:         kycDetails.FirstName + " " + kycDetails.LastName,
	})
	if errors.Is(err, external.ErrUnprocessableEntity) {
		return "", err
	}
	if err != nil {
		return "", fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	return txID, nil
}

func WithdrawFromWallet(ctx context.Context, b Backends, ec external.Client, args pti.TransactionArgs) (string, error) {
	externalUser, err := GetUser(ctx, b, args.WalletID)
	if err != nil {
		return "", err
	}

	las, err := b.LinkedAccounts().ListByWalletId(ctx, args.WalletID)
	if err != nil {
		return "", fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	var balance *linkedaccounts.LinkedAccount
	for _, la := range las {
		if la.Provider == pti.ProviderName && la.Type == pti.AccTypeBalance {
			balance = &la
			break
		}
	}

	// only allow withdrawing from bank
	var bank *linkedaccounts.LinkedAccount
	for _, la := range las {
		if la.Provider == pti.ProviderName && la.Type == pti.TypeBank && la.ID == args.LinkedAccountID {
			bank = &la
			break
		}
	}
	if balance == nil {
		return "", fmt.Errorf("%w balance account not found", pti.ErrNotFound)
	}
	if bank == nil {
		return "", fmt.Errorf("%w source account not found or is not a bank account", pti.ErrNotFound)
	}

	kycDetails, err := b.KYC().GetIndividualDetails(ctx, args.WalletID)
	if err != nil {
		return "", err
	}

	txID, err := ec.WalletWithdrawal(ctx, external.WithdrawalArgs{
		RequestID:             args.PaymentID,
		SessionID:             args.PaymentID,
		ScenarioID:            pti.ScenarioWithdrawal,
		UserID:                externalUser.ExternalID,
		ExternalWalletID:      balance.ProviderID,
		ExternalBankAccountID: bank.ProviderID,
		Amount:                args.Amount,
		AccountHolderName:     kycDetails.FirstName + " " + kycDetails.LastName,
	})
	if errors.Is(err, external.ErrUnprocessableEntity) {
		return "", err
	}
	if err != nil {
		return "", fmt.Errorf("%w %s", pti.ErrInternal, err)
	}
	return txID.ID, nil
}

func UpdateTransactionStatus(ctx context.Context, b Backends, ex external.Client, args pti.TransactionStatusArgs) error {
	payload, err := json.Marshal(external.StatusPayload{
		ProviderName: "UNKNOWN",
		Status:       string(args.Status),
		PaymentTotal: external.PaymentTotal{
			Subtotal: external.Subtotal{
				Amount: args.Amount.Float64(),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	_, err = ex.UpdateTransactionStatus(ctx, external.UpdateTxStatusArgs{
		RequestID:     args.PaymentID,
		TransactionID: args.TransactionID,
		Feedback:      string(args.Status),
		Date:          time.Now(),
		ProviderName:  "UNKNOWN",
		Payload:       string(payload),
	})
	if err != nil {
		return fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	return nil
}

func ReserveBalance(ctx context.Context, b Backends, linkedAccountID, txID string, amt currency.Amount, timeout time.Duration) (*pti.Balance, error) {
	if amt.Currency != currency.USD {
		return nil, fmt.Errorf("%w %s not supported.", pti.ErrInternal, amt.Currency)
	}

	la, err := b.LinkedAccounts().Get(ctx, linkedAccountID)
	if err != nil {
		return nil, err
	}
	if la.Provider != pti.ProviderName || la.Type != pti.AccTypeBalance {
		return nil, fmt.Errorf("%w Not a PTI linked account", pti.ErrInternal)
	}

	opsAcc := pti.USDOpsAccount
	ledger := pti.LedgerIDUSD
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
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
	}
	if len(tx) > 0 {
		if tx[0].Code == pacioli.TransferExceedsCredits || tx[0].Code == pacioli.TransferExceedsDebits || tx[0].Code == pacioli.TransferExceedsPendingTransferAmount {
			return nil, fmt.Errorf("%w insufficiens balance cod (%s)", pti.ErrInsufficientBalance, tx[0].Code.String())
		}
		if tx[0].Code != 0 {
			return nil, fmt.Errorf("%w non success code (%s)", pti.ErrInternal, tx[0].Code.String())
		}
	}

	accs, err := b.Pacioli().GetAccounts(ctx, []string{la.ID})
	if err != nil {
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	if len(accs) != 1 {
		return nil, fmt.Errorf("%w account not found", pti.ErrNotFound)
	}

	return &pti.Balance{
		Total:     currency.FromUInt64(accs[0].CreditsPosted-accs[0].DebitsPosted, la.SendCurrency),
		Available: currency.FromUInt64(accs[0].CreditsPosted-accs[0].DebitsPosted-accs[0].DebitsPending, la.SendCurrency),
	}, nil
}

func FinaliseReserve(ctx context.Context, b Backends, trxID string) error {
	tx, err := b.Pacioli().PostTransfers(ctx, []string{trxID})
	if err != nil {
		return fmt.Errorf("%w %s", pti.ErrInternal, err)
	}
	if len(tx) == 0 {
		return nil
	}
	if tx[0].Code == pacioli.TransferPendingTransferAlreadyPosted {
		return nil
	}
	if tx[0].Code == pacioli.TransferExceedsCredits || tx[0].Code == pacioli.TransferExceedsDebits || tx[0].Code == pacioli.TransferExceedsPendingTransferAmount {
		return fmt.Errorf("%w insufficient balance code (%s)", pti.ErrInsufficientBalance, tx[0].Code.String())
	}
	if tx[0].Code == pacioli.TransferPendingTransferExpired {
		return fmt.Errorf("%w transfer timed out code (%s)", pti.ErrTimedOut, tx[0].Code.String())
	}
	if tx[0].Code == pacioli.TransferPendingTransferNotPending {
		return fmt.Errorf("%w pending transfer not found code (%s)", pti.ErrNotFound, tx[0].Code.String())
	}
	if tx[0].Code != 0 {
		return fmt.Errorf("%w non success code (%s)", pti.ErrInternal, tx[0].Code.String())
	}

	return nil
}

func RollbackReserve(ctx context.Context, b Backends, txID string) error {
	tx, err := b.Pacioli().VoidTransfers(ctx, []string{txID})
	if err != nil {
		slack.SendToChannel(ctx, slack.ChannelNotifyErrors, "wallet-info-bot", fmt.Sprintf("*:::[Fiant ERROR]:::* \n *RollbackReserve txID:* %s,\n *error:* %s", txID, err))
		return fmt.Errorf("%w %s", pti.ErrInternal, err)
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
		return fmt.Errorf("%w non success code (%s)", pti.ErrInternal, tx[0].Code.String())
	}

	return nil
}

func AssignBalance(ctx context.Context, b Backends, linkedAccountID, txID string, amt currency.Amount) (*pti.Balance, error) {
	la, err := b.LinkedAccounts().Get(ctx, linkedAccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	if la.Provider != pti.ProviderName || la.Type != pti.AccTypeBalance {
		return nil, fmt.Errorf("%w linked account not correct type", pti.ErrNotFound)
	}

	opsAcc := pti.USDOpsAccount
	ledger := pti.LedgerIDUSD
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
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
	}
	if len(tx) != 0 {
		if tx[0].Code == pacioli.TransferExceedsCredits || tx[0].Code == pacioli.TransferExceedsDebits || tx[0].Code == pacioli.TransferExceedsPendingTransferAmount {
			return nil, fmt.Errorf("%w insufficiens balance cod (%s)", pti.ErrInsufficientBalance, tx[0].Code.String())
		}
		if tx[0].Code != 0 {
			return nil, fmt.Errorf("%w non success code (%s)", pti.ErrInternal, tx[0].Code.String())
		}
	}

	accs, err := b.Pacioli().GetAccounts(ctx, []string{la.ID})
	if err != nil {
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	if len(accs) != 1 {
		return nil, fmt.Errorf("%w account not found", pti.ErrNotFound)
	}

	return &pti.Balance{
		Total:     currency.FromUInt64(accs[0].CreditsPosted-accs[0].DebitsPosted, la.SendCurrency),
		Available: currency.FromUInt64(accs[0].CreditsPosted-accs[0].DebitsPosted-accs[0].DebitsPending, la.SendCurrency),
	}, nil
}

func ReserveTransfer(ctx context.Context, b Backends, fromAccount, toAccount, txID string, amt currency.Amount, timeout time.Duration) error {
	if amt.Currency != currency.USD {
		return fmt.Errorf("%w %s not supported.", pti.ErrInternal, amt.Currency)
	}

	fromAcc, err := b.LinkedAccounts().Get(ctx, fromAccount)
	if err != nil {
		return err
	}
	if fromAcc.Provider != pti.ProviderName || fromAcc.Type != pti.AccTypeBalance {
		return fmt.Errorf("%w from account not a PTI linked account", pti.ErrInternal)
	}

	toAcc, err := b.LinkedAccounts().Get(ctx, toAccount)
	if err != nil {
		return err
	}
	if toAcc.Provider != pti.ProviderName || toAcc.Type != pti.AccTypeBalance {
		return fmt.Errorf("%w to account not a PTI linked account", pti.ErrInternal)
	}

	tx, err := b.Pacioli().CreateTransfers(ctx, []pacioli.CreateTransferArgs{
		{
			ID:              txID,
			Amount:          amt.Value,
			DebitAccountID:  fromAccount,
			CreditAccountID: toAccount,
			Pending:         true,
			Code:            1,
			Timeout:         uint64(timeout),
			Ledger:          pti.LedgerIDUSD,
		},
	})
	if err != nil {
		return fmt.Errorf("%w %s", pti.ErrInternal, err)
	}
	if len(tx) > 0 {
		if tx[0].Code == pacioli.TransferExceedsCredits || tx[0].Code == pacioli.TransferExceedsDebits || tx[0].Code == pacioli.TransferExceedsPendingTransferAmount {
			return fmt.Errorf("%w insufficiens balance cod (%s)", pti.ErrInsufficientBalance, tx[0].Code.String())
		}
		if tx[0].Code != 0 {
			return fmt.Errorf("%w non success code (%s)", pti.ErrInternal, tx[0].Code.String())
		}
	}

	return nil
}

func GetKYCWidget(ctx context.Context, b Backends, walletID string) (*pti.WidgetDetails, error) {
	externalUser, err := GetUser(ctx, b, walletID)
	if errors.Is(err, pti.ErrNotFound) {
		await, innerErr := CreateUser(ctx, b, walletID)
		if innerErr != nil {
			return nil, err
		}

		innerErr = await(ctx, &externalUser)

		if innerErr != nil {
			return nil, innerErr
		}
	} else if err != nil {
		return nil, err
	}

	sdkUrl := "https://sdk.staging.fiant.io/latest/index.js"
	formsUrl := "https://forms.staging.fiant.io"
	if env.IsProd() {
		sdkUrl = "https://sdk.platform.fiant.io/0.0.23/index.js"
		formsUrl = "https://forms.platform.fiant.io"
	}

	return &pti.WidgetDetails{
		// ScenarioID:        pti.ScenarioDeposit,
		ScenarioID:        pti.ScenarioWithdrawal,
		RequestID:         uuid.NewString(),
		UserID:            externalUser.ExternalID,
		ClientID:          os.Getenv("PTI_CLIENT_ID"),
		GenerateTokenPath: fmt.Sprintf("%s/api/pti/token", env.GetUrl()),
		SdkUrl:            sdkUrl,
		FormsUrl:          formsUrl,
		SessionID:         walletID,
	}, nil
}

func CreateCard(ctx context.Context, b Backends, walletID, tokenID string) (pti.Await, error) {
	wo := client.StartWorkflowOptions{
		ID:                    "pti_create_card_" + walletID + "_" + tokenID,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	var workflowStatus enums.WorkflowExecutionStatus
	wflow, err := b.Temporal().DescribeWorkflowExecution(ctx, wo.ID, "")
	switch err.(type) {
	case *serviceerror.Internal,
		*serviceerror.Unavailable,
		*serviceerror.InvalidArgument:
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
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
		await, executeErr = b.Temporal().ExecuteWorkflow(ctx, wo, CreateCardWorkflow, walletID, tokenID)
	}
	if executeErr != nil {
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	return await.Get, nil
}

func GetLinkedAccountCardDetails(ctx context.Context, b Backends, ex external.Client, id string) (*pti.EncryptedCreditCardPaymentInformation, error) {
	la, err := b.LinkedAccounts().Get(ctx, id)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return nil, fmt.Errorf("%w %s", pti.ErrNotFound, err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	ptiUser, err := GetUser(ctx, b, la.WalletID)
	if err != nil {
		return nil, err
	}

	details, err := ex.GetUsersPaymentInformation(ctx, ptiUser.ExternalID, la.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	var card external.EncryptedCreditCardPaymentInformation
	err = json.Unmarshal(details, &card)
	if err != nil {
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	network := card.CreditCardType
	if len(network) > 0 {
		network = strings.ToUpper(network[0:1]) + network[1:]
	}
	card.CreditCardType = network

	return &card, nil
}

func CreateBankAccount(ctx context.Context, b Backends, args pti.CreateBankAccountArgs) (pti.Await, error) {
	wo := client.StartWorkflowOptions{
		ID:                    "pti_create_bank_" + args.WalletID + "_" + args.AccountNumber,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	var workflowStatus enums.WorkflowExecutionStatus
	wflow, err := b.Temporal().DescribeWorkflowExecution(ctx, wo.ID, "")
	switch err.(type) {
	case *serviceerror.Internal,
		*serviceerror.Unavailable,
		*serviceerror.InvalidArgument:
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
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
		await, executeErr = b.Temporal().ExecuteWorkflow(ctx, wo, CreatePtiBankAccountWorkflow, args)
	}
	if executeErr != nil {
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	return await.Get, nil
}

func CreateDeposit(ctx context.Context, b Backends, la *linkedaccounts.LinkedAccount, amount currency.Amount, note string) error {
	workflowOptions := client.StartWorkflowOptions{
		ID:                       "pti_create_deposit_" + la.WalletID,
		TaskQueue:                "backend",
		WorkflowExecutionTimeout: time.Hour,
	}

	_, err := b.Temporal().ExecuteWorkflow(ctx, workflowOptions, DepositWorkflow, la, amount, note)
	if err != nil {
		return err
	}

	return nil
}

func ConfirmWithdrawal(ctx context.Context, b Backends, ec external.Client, walletID string, payment *payments.Payment) (string, error) {
	externalUser, err := GetUser(ctx, b, walletID)
	if err != nil {
		return "", err
	}
	las, err := b.LinkedAccounts().ListByWalletId(ctx, walletID)
	if err != nil {
		return "", fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	var balance *linkedaccounts.LinkedAccount
	for _, la := range las {
		if la.Provider == pti.ProviderName && la.Type == pti.AccTypeBalance {
			balance = &la
			break
		}
	}
	// only allow withdrawing from bank
	var bank *linkedaccounts.LinkedAccount
	for _, la := range las {
		if la.Provider == pti.ProviderName && la.Type == pti.TypeBank {
			bank = &la
			break
		}
	}
	if balance == nil {
		return "", fmt.Errorf("%w balance account not found", pti.ErrNotFound)
	}
	if bank == nil {
		return "", fmt.Errorf("%w source account not found or is not a bank account", pti.ErrNotFound)
	}

	kycDetails, err := b.KYC().GetIndividualDetails(ctx, walletID)
	if err != nil {
		return "", err
	}

	withdrawTx, err := ec.WalletWithdrawal(ctx, external.WithdrawalArgs{
		RequestID:             payment.ID,
		ScenarioID:            pti.ScenarioWithdrawal,
		UserID:                externalUser.ExternalID,
		ExternalWalletID:      balance.ProviderID,
		ExternalBankAccountID: bank.ProviderID,
		Amount:                payment.ReceiverAmount,
		AccountHolderName:     kycDetails.FirstName + " " + kycDetails.LastName,
	})
	if errors.Is(err, external.ErrUnprocessableEntity) {
		return "", err
	}
	if err != nil {
		return "", fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	_, err = RegisterWithdrawal(ctx, b, walletID, withdrawTx, balance.ID, payment.Note)
	if err != nil {
		return "", fmt.Errorf("%w %s", pti.ErrInternal, err)
	}
	return withdrawTx.ID, nil
}

func RegisterWithdrawal(ctx context.Context, b Backends, walletID string, withdrawalDetails *external.WithdrawDetails, linkedAccountID, note string) (string, error) {

	existingWithdrawal, err := b.Transactions().GetTransactionByForeignID(ctx, walletID, withdrawalDetails.ID)
	if err != nil && !errors.Is(err, transactions.ErrNotFound) {
		return "", fmt.Errorf("%w %s", pti.ErrInternal, err)
	}
	if existingWithdrawal != nil {
		_, err = processWithdrawal(ctx, b, walletID, existingWithdrawal.ID)
		if err != nil {
			return "", fmt.Errorf("%w %s", pti.ErrInternal, err)
		}
		return existingWithdrawal.ID, nil
	}

	wallet, err := b.Wallets().Get(ctx, walletID)
	if err != nil {
		return "", fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	trxID, err := b.Transactions().CreateTransaction(ctx, transactions.CreateTransactionArgs{
		WalletID:                walletID,
		ForeignID:               withdrawalDetails.ID,
		Provider:                pti.ProviderName,
		State:                   transactions.StatePending,
		ForeignType:             transactions.TransactionTypeWithdrawal,
		Source:                  wallet.AddressString(),
		Destination:             wallet.AddressString(),
		Title:                   "Withdrawal",
		DestinationIdentity:     walletID,
		ProviderFee:             nil,
		DestinationIdentityType: payments.IdentityTypeWalletID.String(),
		Amount:                  withdrawalDetails.Amount,
		Note:                    note,
		LinkedAccountTitle:      "USD Balance",
		Transfers: []transactions.TransferArgs{
			{
				LinkedAccountID: linkedAccountID,
				ForeignID:       withdrawalDetails.ID,
				Amount:          withdrawalDetails.Amount,
				Type:            transactions.TransferTypeDebitBalance,
				State:           transactions.StatePending,
			},
		},
	})
	if err != nil {

		return "", fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	_, err = processWithdrawal(ctx, b, walletID, trxID)
	if err != nil {

		return "", fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	return trxID, err
}

func processWithdrawal(ctx context.Context, b Backends, walletID, transactionID string) (pti.Await, error) {

	wo := client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("fiant_reserve balance_withdrawal_%s", transactionID),
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	var workflowStatus enums.WorkflowExecutionStatus
	wflow, err := b.Temporal().DescribeWorkflowExecution(ctx, wo.ID, "")
	switch err.(type) {
	case *serviceerror.Internal,
		*serviceerror.Unavailable,
		*serviceerror.InvalidArgument:
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
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
		await, executeErr = b.Temporal().ExecuteWorkflow(ctx, wo, ProcessPTIWithdrawal, walletID, transactionID)
	}
	if executeErr != nil {
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	return await.Get, nil
}
