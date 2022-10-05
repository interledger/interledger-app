package client

import (
	"context"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	inmemory_external_client "gitlab.com/fynbos/backend/providers/machnet/external/client/inmemory"
	"gitlab.com/fynbos/backend/providers/machnet/ops"
	"gitlab.com/fynbos/backend/providers/machnet/workflows"
	"gitlab.com/fynbos/backend/user"
	"go.temporal.io/api/enums/v1"
	temporal "go.temporal.io/sdk/client"
)

// const (
// 	sandboxUrl = "https://v4sandbox.machpay.com/v4"
// 	prodUrl    = "https://machpay.com"
// )

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

func New(b Backends) machnet.Client {
	// TODO: http client

	opsBackends := opsBackends{
		Backends: b,
		external: inmemory_external_client.New(),
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

func (c client) CreateSendUser(ctx context.Context, walletID string) error {
	workflowOptions := temporal.StartWorkflowOptions{
		ID:                    "machnet_create_send_user_" + walletID,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}

	_, err := c.t.ExecuteWorkflow(ctx, workflowOptions, workflows.CreateSendUserWorkflow, c.b, walletID)
	return err
}
