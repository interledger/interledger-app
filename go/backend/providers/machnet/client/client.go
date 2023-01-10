package client

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"gitlab.com/fynbos/backend/statements"
	"gitlab.com/fynbos/backend/transactions"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	external_client "gitlab.com/fynbos/backend/providers/machnet/external/client"
	inmemory_external_client "gitlab.com/fynbos/backend/providers/machnet/external/client/inmemory"
	"gitlab.com/fynbos/backend/providers/machnet/ops"
	"gitlab.com/fynbos/backend/providers/machnet/workflows"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/log"
	"go.temporal.io/api/enums/v1"
	temporal "go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

var _ machnet.Client = client{}

type Backends interface {
	DB() *sqlx.DB
	Users() user.Client
	KYC() kyc.Client
	Temporal() temporal.Client
	LinkedAccounts() linkedaccounts.Client
	Transactions() transactions.Client
	Statements() statements.Client
}

type opsBackends struct {
	Backends
	external external.Client
}

func (b opsBackends) External() external.Client {
	return b.external
}

func New(b Backends, clientID, clientSecret string) machnet.Client {
	opsBackends := opsBackends{
		Backends: b,
		external: inmemory_external_client.New(),
	}
	if !env.IsLocal() || (clientID != "" && clientSecret != "") {
		opsBackends.external = external_client.New(clientID, clientSecret)
	}

	return &client{b: opsBackends, t: b.Temporal(), externalApi: opsBackends.external}
}

type client struct {
	b           ops.Backends
	t           temporal.Client
	externalApi external.Client
}

func (c client) External() external.Client {
	return c.externalApi
}

func (c client) GetKYCStatus(ctx context.Context, walletID string) (*machnet.UserKYC, error) {
	return ops.GetKYCStatus(ctx, c.b, walletID)
}

func (c client) GetUserByWalletID(ctx context.Context, walletID string) (*machnet.User, error) {
	return ops.GetUserByWalletID(ctx, c.b, walletID)
}

func (c client) GetUserByID(ctx context.Context, id string) (*machnet.User, error) {
	return ops.GetUserByID(ctx, c.b, id)
}

func (c client) CreateUser(ctx context.Context, args machnet.CreateArgs) (*machnet.User, error) {
	return ops.CreateUser(ctx, c.b, args)
}

func (c client) GetWidgetToken(ctx context.Context, walletID string) (*machnet.WidgetToken, error) {
	return ops.GetWidgetToken(ctx, c.b, walletID)
}

func (c client) StartSendUserKYC(ctx context.Context, walletID string) (machnet.Await, error) {
	workflowOptions := temporal.StartWorkflowOptions{
		ID:                    "machnet_create_send_user_" + walletID,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	u, err := ops.GetUserByWalletID(ctx, c.b, walletID)
	if err != nil && !errors.Is(err, machnet.ErrNotFound) {
		return nil, err
	}

	// Return immediately if the user already exists and the user's state is anything but Retry
	if u != nil && u.KYCStatus != machnet.KYCStatusRetry {
		return func(ctx context.Context, out interface{}) error {
			rf := reflect.ValueOf(out)
			if rf.Type().Kind() != reflect.Ptr {
				return errors.New("value parameter is not a pointer")
			}
			rf.SetString(u.ID)
			return nil
		}, nil
	}

	wf, err := c.t.ExecuteWorkflow(ctx, workflowOptions, workflows.CreateSendUserWorkflow, walletID)
	if err != nil {
		return nil, err
	}

	// Set the user KYC status to in progress as soon as possible for retries and in case the consumer doesn't wait.
	if u != nil {
		err = ops.SetKYCInProgress(ctx, c.b, u.ID)
		if err != nil {
			log.Warn("failed to update in machnet user KYC status to InProgress for retry", zap.Error(err))
		}
	}

	return func(ctx context.Context, out interface{}) error {
		// Wait for the Workflow to complete.
		return wf.Get(ctx, &out)
	}, nil
}

func (c client) CreateTransaction(ctx context.Context, args machnet.CreateTransactionArgs) (machnet.Await, error) {
	workflowOptions := temporal.StartWorkflowOptions{
		ID:        "machnet_create_transaction_" + args.FromLinkedAccountID + "_" + args.ToLinkedAccountID,
		TaskQueue: "backend",
	}

	wf, err := c.t.ExecuteWorkflow(ctx, workflowOptions, workflows.CreateTransactionWorkflow, args)
	if err != nil {
		return nil, err
	}

	return wf.Get, nil
}

func (c client) CreateReceiveBankAccount(ctx context.Context, args machnet.CreateReceiveBankAccountArgs) (*machnet.ReceiveBankAccount, error) {
	return ops.CreateReceiveBankAccount(ctx, c.b, args)
}

func (c client) GetBanks(ctx context.Context, countryCode string) ([]machnet.Bank, error) {
	return ops.GetBanks(ctx, c.b, countryCode)
}

func (c client) CreateWallet(ctx context.Context, args machnet.CreateWalletArgs) (*linkedaccounts.LinkedAccount, error) {
	return ops.CreateWallet(ctx, c.b, args)
}

func (c client) GetWallet(ctx context.Context, id string) (*machnet.Wallet, error) {
	return ops.GetWallet(ctx, c.b, id)
}

func (c client) WithdrawFromWallet(ctx context.Context, args machnet.WithdrawFromWalletArgs) (machnet.Await, error) {
	idempotencyKey := args.IdempotencyKey
	if idempotencyKey == "" {
		idempotencyKey = uuid.NewString()
	}
	workflowOptions := temporal.StartWorkflowOptions{
		ID:                    "machnet_withdraw_from_wallet_" + idempotencyKey,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}

	wf, err := c.t.ExecuteWorkflow(ctx, workflowOptions, workflows.CreateWalletWithdrawalWorkflow, args)
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context, _ interface{}) error {
		// Wait for the Workflow to complete.
		return wf.Get(ctx, nil)
	}, nil
}

func (c client) DeleteFundSource(ctx context.Context, linkedAccID string) (machnet.Await, error) {
	workflowOptions := temporal.StartWorkflowOptions{
		ID:        "machnet_delete_func_source" + linkedAccID,
		TaskQueue: "backend",
	}

	wf, err := c.t.ExecuteWorkflow(ctx, workflowOptions, workflows.DeleteAccountWorkflow, linkedAccID)
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context, _ interface{}) error {
		// Wait for the Workflow to complete.
		return wf.Get(ctx, nil)
	}, nil
}

func (c client) StartWalletTopup(ctx context.Context, args machnet.StartWalletTopupArgs) (machnet.Await, error) {
	idempotencyKey := args.IdempotencyKey
	if idempotencyKey == "" {
		idempotencyKey = uuid.NewString()
	}
	opts := temporal.StartWorkflowOptions{
		ID:                       "machnet_create_wallet_topup_" + idempotencyKey,
		TaskQueue:                "backend",
		WorkflowExecutionTimeout: time.Hour * 24 * 7, // Workflow has 7 days to complete
		WorkflowIDReusePolicy:    enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}
	topupWorkflow, err := c.b.Temporal().ExecuteWorkflow(ctx, opts, workflows.CreateWalletTopupWorkflow, args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return topupWorkflow.Get, nil
}

func (c client) GetCurrentStatement(ctx context.Context, walletID string) ([]byte, error) {
	return ops.GetCurrentStatement(ctx, c.b, walletID)
}
