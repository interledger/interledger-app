package client

import (
	"context"

	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/mx/external"
	"gitlab.com/fynbos/backend/providers/mx/ops"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

var _ mx.Client = client{}

type client struct {
	b ops.Backends
}

func New(b Backends, mxClientID, mxApiKey string) mx.Client {

	ob := &opsBackends{
		Backends:   b,
		mxExternal: external.NewClient(mxClientID, mxApiKey),
	}

	return &client{
		b: ob,
	}
}

func (c client) CreateAccount(ctx context.Context, args mx.CreateAccountArgs) (*mx.Account, error) {
	return ops.CreateAccount(ctx, c.b, args)
}

func (c client) GetAccount(ctx context.Context, mxAccountGuid string) (*mx.Account, error) {
	return ops.GetAccount(ctx, c.b, mxAccountGuid)
}

func (c client) GetAccountByFundingsource(ctx context.Context, fundingsourceID string) (*mx.Account, error) {
	return ops.GetAccountByFundingsource(ctx, c.b, fundingsourceID)
}

func (c client) StartIdentityAggregation(ctx context.Context, mxUserGuid, mxMemberGuid string) (*mx.Member, error) {
	return ops.StartIdentityAggregation(ctx, c.b, mxUserGuid, mxMemberGuid)
}

func (c client) GetMemberStatus(ctx context.Context, mxUserGuid, mxMemberGuid string) (*mx.Member, error) {
	return ops.GetMemberStatus(ctx, c.b, mxUserGuid, mxMemberGuid)
}

func (c client) GetAccountOwner(ctx context.Context, args mx.GetAccountOwnerArgs) (*mx.AccountOwner, error) {
	return ops.GetAccountOwner(ctx, c.b, args)
}

func (c client) ReadAccount(ctx context.Context, mxAccountGuid string) (*mx.AccountDetails, error) {
	return ops.ReadAccount(ctx, c.b, mxAccountGuid)
}

func (c client) GetSelectedAccountGuid(ctx context.Context, mxUserGuid string, mxMemberGuid string) (string, error) {
	return ops.GetSelectedAccountGuid(ctx, c.b, mxUserGuid, mxMemberGuid)
}

func (c client) GetMxUserByAccountID(ctx context.Context, accountID string) (string, error) {
	return ops.GetMxUserByAccountID(ctx, c.b, accountID)
}

func (c client) VerifyOwnership(ctx context.Context, args mx.VerifyOwnershipArgs) error {
	return ops.VerifyOwnership(ctx, c.b, args)
}

func (c client) GetConnectWidget(ctx context.Context, accountID string, identityID string) (string, error) {
	return ops.GetConnectWidget(ctx, c.b, accountID, identityID)
}

func (c client) InitiateCreateAccount(ctx context.Context, args mx.InitiateCreateAccountArgs) (string, error) {
	return ops.InitiateCreateAccount(ctx, c.b, args)
}

func (c client) WaitForCreateAccount(ctx context.Context, fundingsourceID string) error {
	return ops.WaitForCreateAccount(ctx, c.b, fundingsourceID)
}

func (c client) InitiateCreateFundingsource(ctx context.Context, args mx.InitiateCreateFundingsourceArgs) (err error) {
	defer func() {
		if err != nil {
			log.Error("Failed to initiate creating mx funding source", zap.String("err", err.Error()))
			return
		}

		log.Debug(
			"Initiating creating mx funding source",
			zap.String("accountID", args.AccountID),
			zap.String("mxAccountID", args.MxAccountGuid),
			zap.String("name", args.Name),
		)
	}()
	return ops.InitiateCreateFundingsource(ctx, c.b, args)
}

func (c client) StartBalanceAggregation(ctx context.Context, mxAccountGuid string) (*mx.Member, error) {
	return ops.StartBalanceAggregation(ctx, c.b, mxAccountGuid)
}

func (c client) GetAccountBalance(ctx context.Context, mxAccountGuid string) (*mx.AccountBalance, error) {
	return ops.GetAccountBalance(ctx, c.b, mxAccountGuid)
}
