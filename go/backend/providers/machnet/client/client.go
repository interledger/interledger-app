package client

import (
	"context"

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
	temporal "go.temporal.io/sdk/client"
)

type Backends interface {
	DB() *sqlx.DB
	Users() user.Client
	KYC() kyc.Client
	Temporal() temporal.Client
	LinkedAccounts() linkedaccounts.Client
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
	if env.IsProd() || env.IsSandbox() {
		opsBackends.external = external_client.New(clientID, clientSecret)
	}

	return &client{b: opsBackends, t: b.Temporal()}
}

type client struct {
	b ops.Backends
	t temporal.Client
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

func (c client) HandleEvent(ctx context.Context, event external.Event) error {
	return ops.HandleEvent(ctx, c.b, event)
}

func (c client) CreateSendUser(ctx context.Context, walletID string) (machnet.Await, error) {
	workflowOptions := temporal.StartWorkflowOptions{
		ID:        "machnet_create_send_user_" + walletID,
		TaskQueue: "backend",
	}

	wf, err := c.t.ExecuteWorkflow(ctx, workflowOptions, workflows.CreateSendUserWorkflow, walletID)
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context) error {
		// Wait for the Workflow to complete.
		var externalUserID string
		return wf.Get(ctx, &externalUserID)
	}, nil
}

func (c client) CreateTransaction(ctx context.Context, args machnet.CreateTransactionArgs) (machnet.Await, error) {
	workflowOptions := temporal.StartWorkflowOptions{
		ID:        "machnet_create_transaction_" + args.FromWalletID,
		TaskQueue: "backend",
	}

	wf, err := c.t.ExecuteWorkflow(ctx, workflowOptions, workflows.CreateTransactionWorkflow, args)
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context) error {
		// Wait for the Workflow to complete.
		var trxID string
		return wf.Get(ctx, &trxID)
	}, nil
}

func (c client) CreateReceiveBankAccount(ctx context.Context, args machnet.CreateReceiveBankAccountArgs) (*machnet.ReceiveBankAccount, error) {
	return ops.CreateReceiveBankAccount(ctx, c.b, args)
}

func (c client) GetReceiveBankAccount(ctx context.Context, id string) (*machnet.ReceiveBankAccount, error) {
	return ops.GetReceiveBankAccount(ctx, c.b, id)
}

func (c client) CreateReceiveUser(ctx context.Context, args machnet.CreateReceiveUserArgs) (*machnet.ReceiveUser, error) {
	return ops.CreateReceiveUser(ctx, c.b, args)
}

func (c client) GetReceiveUser(ctx context.Context, args machnet.GetReceiveUserArgs) (*machnet.ReceiveUser, error) {
	return ops.GetReceiveUser(ctx, c.b, args)
}

func (c client) CreateReceiveUserBankAccount(ctx context.Context, args machnet.CreateReceiveUserBankAccountArgs) (*machnet.ReceiveUserBankAccount, error) {
	return ops.CreateReceiveUserBankAccount(ctx, c.b, args)
}

func (c client) GetReceiveUserBankAccount(ctx context.Context, args machnet.GetReceiveUserBankAccountArgs) (*machnet.ReceiveUserBankAccount, error) {
	return ops.GetReceiveUserBankAccount(ctx, c.b, args)
}

func (c client) GetBanks(ctx context.Context, countryCode string) ([]machnet.Bank, error) {
	return ops.GetBanks(ctx, c.b, countryCode)
}
