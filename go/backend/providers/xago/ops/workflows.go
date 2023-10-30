package ops

import (
	"context"
	"time"

	"gitlab.com/fynbos/backend/db"

	"gitlab.com/fynbos/backend/providers/xago/external"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"go.temporal.io/sdk/workflow"
)

type opsBackends struct {
	ActivityBackends
	xagoExt external.Client
}

func (o opsBackends) External() external.Client {
	return o.xagoExt
}

type Activity struct {
	b Backends
}

func NewActivity(ab ActivityBackends) *Activity {
	ex := external.New()

	return &Activity{b: &opsBackends{
		ActivityBackends: ab,
		xagoExt:          ex,
	}}
}

func CreateSubAccountWorkflow(ctx workflow.Context, walletID string) (*linkedaccounts.LinkedAccount, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating xago sub account.")

	var subAcc external.SubAccount
	err := workflow.ExecuteActivity(ctx, a.CreateSubAccount, walletID).Get(ctx, &subAcc)
	if err != nil {
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.SaveSubAccount, walletID, subAcc).Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	var la linkedaccounts.LinkedAccount
	err = workflow.ExecuteActivity(ctx, a.AddLinkedAccount, walletID, subAcc).Get(ctx, &la)
	if err != nil {
		return nil, err
	}

	return &la, nil
}

func (a *Activity) AddLinkedAccount(ctx context.Context, walletID string, sa external.SubAccount) (*linkedaccounts.LinkedAccount, error) {
	return a.b.LinkedAccounts().Create(ctx, &linkedaccounts.CreateArgs{
		WalletID:        walletID,
		Name:            "Xago Account",
		Nickname:        "Xago Account",
		Mask:            "",
		Provider:        "xago",
		ProviderID:      sa.DepositAddress,
		Type:            "bank_account",
		CanSend:         true,
		CanReceive:      true,
		State:           linkedaccounts.Verified,
		SendCountry:     "ZA",
		SendCurrency:    "ZAR",
		ReceiveCountry:  "ZA",
		ReceiveCurrency: "ZAR",
	})
}

func (a *Activity) SaveSubAccount(ctx context.Context, walletID string, sa external.SubAccount) error {
	_, err := a.b.DB().ExecContext(ctx, "INSERT INTO xago_sub_accounts (wallet_id, account_id, deposit_address, deposit_tag) VALUES ($1, $2, $3, $4)",
		walletID, sa.AccountID, sa.DepositAddress, sa.DepositTag)
	if db.IsErrorCode(err, db.UniqueViolationError) {
		return nil
	}
	return err
}

func (a *Activity) CreateSubAccount(ctx context.Context, walletID string) (*external.SubAccount, error) {
	ul, err := a.b.Users().ListUsers(ctx, walletID)
	if err != nil {
		return nil, err
	}

	id, err := a.b.KYC().GetIndividualDetails(ctx, walletID)
	if err != nil {
		return nil, err
	}

	idNum, err := a.b.KYC().GetPersonaIDNumbers(ctx, walletID)
	if err != nil {
		return nil, err
	}

	return a.b.External().CreateSubAccount(ctx, ul[0], *id, *idNum)
}
