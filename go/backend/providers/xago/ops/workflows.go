package ops

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/xago"
	"gitlab.com/fynbos/backend/providers/xago/external"
	"go.temporal.io/sdk/temporal"
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

func CreateSubAccountWorkflow(ctx workflow.Context, walletID string) (*xago.SubAccount, error) {
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

	var accID string
	err = workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		return uuid.NewString()
	}).Get(&accID)
	if err != nil {
		logger.Error("error generating sub account ID as side effect", "Error", err)
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.SaveSubAccount, walletID, accID, subAcc).Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &xago.SubAccount{
		ID:             accID,
		AccountID:      subAcc.AccountID,
		DepositAddress: subAcc.DepositAddress,
		DepositTag:     subAcc.DepositTag,
		WalletID:       walletID,
	}, nil
}

func (a *Activity) SaveSubAccount(ctx context.Context, walletID, accountID string, sa external.SubAccount) error {
	_, err := a.b.DB().ExecContext(ctx, "INSERT INTO xago_sub_accounts (id, wallet_id, account_id, deposit_address, deposit_tag) VALUES ($1, $2, $3, $4, $5)",
		accountID, walletID, sa.AccountID, sa.DepositAddress, sa.DepositTag)
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

func CreateBeneficiaryWorkflow(ctx workflow.Context, walletID string) (*linkedaccounts.LinkedAccount, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating xago beneficiary.")

	var ben external.CreateBeneficiaryResp
	err := workflow.ExecuteActivity(ctx, a.CreateExternalBeneficiaries, walletID).Get(ctx, &ben)
	if err != nil {
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.SaveBeneficiary, walletID, ben).Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	var laID string
	err = workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		return uuid.NewString()
	}).Get(&laID)
	if err != nil {
		logger.Error("error generating linked account ID as side effect for xago beneficiary", "Error", err)
		return nil, err
	}

	var la linkedaccounts.LinkedAccount
	err = workflow.ExecuteActivity(ctx, a.AddBeneficiaryLinkedAccount, walletID, ben).Get(ctx, &la)
	if err != nil {
		return nil, err
	}

	return &la, nil
}

func (a *Activity) CreateExternalBeneficiaries(ctx context.Context, walletID string) (*external.CreateBeneficiaryResp, error) {
	id, err := a.b.KYC().GetIndividualDetails(ctx, walletID)
	if err != nil {
		return nil, err
	}

	return a.b.External().AddBeneficiary(ctx, *id)
}

func (a *Activity) SaveBeneficiary(ctx context.Context, walletID string, sa external.CreateBeneficiaryResp) error {
	if len(sa.Beneficiaries) != 1 {
		return temporal.NewNonRetryableApplicationError(fmt.Sprintf("incorrect number of beneficiaries, expected 1 got %d", len(sa.Beneficiaries)), "external", nil)
	}
	b := sa.Beneficiaries[0]
	_, err := a.b.DB().ExecContext(ctx, `INSERT INTO xago_beneficiaries (id, wallet_id, address, reference, bank_name, branch_code, account_number, status, currency, scope, name) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		b.ID, walletID, b.BeneficiaryAddress, b.Reference, b.BankName, b.BranchCode, b.AccountNumber, b.Status, b.CurrencyCode, b.Scope, b.Name)
	if db.IsErrorCode(err, db.UniqueViolationError) {
		return nil
	}
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) AddBeneficiaryLinkedAccount(ctx context.Context, walletID, id string, sa external.CreateBeneficiaryResp) (*linkedaccounts.LinkedAccount, error) {
	if len(sa.Beneficiaries) != 1 {
		return nil, temporal.NewNonRetryableApplicationError(fmt.Sprintf("incorrect number of beneficiaries, expected 1 got %d", len(sa.Beneficiaries)), "external", nil)
	}
	b := sa.Beneficiaries[0]

	la, _ := a.b.LinkedAccounts().Get(ctx, id)
	if la != nil {
		return la, nil
	}

	return a.b.LinkedAccounts().Create(ctx, &linkedaccounts.CreateArgs{
		ID:              id,
		WalletID:        walletID,
		Name:            "Xago Beneficiary",
		Nickname:        "Xago Beneficiary",
		Mask:            "",
		Provider:        "xago",
		ProviderID:      b.ID,
		Type:            "bank_account",
		CanSend:         false,
		CanReceive:      true,
		State:           linkedaccounts.Verified,
		SendCountry:     "ZA",
		SendCurrency:    "ZAR",
		ReceiveCountry:  "ZA",
		ReceiveCurrency: "ZAR",
	})
}
