package ops

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/currency"

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

func (a *Activity) WalletHasSubAccount(ctx context.Context, walletID string) (bool, error) {
	acc, err := LookupSubAccount(ctx, a.b, walletID)
	if err != nil && !errors.Is(err, xago.ErrNotFound) {
		return false, err
	}

	return acc != nil, nil
}

func (a *Activity) SaveSubAccount(ctx context.Context, walletID, accountID string, sa external.SubAccount) error {
	_, err := a.b.DB().ExecContext(ctx, "INSERT INTO xago_sub_accounts (id, wallet_id, account_id, deposit_address, deposit_tag) VALUES ($1, $2, $3, $4, $5)",
		accountID, walletID, sa.AccountID, sa.DepositAddress, strconv.Itoa(sa.DepositTag))
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

func CreateBeneficiaryWorkflow(ctx workflow.Context, bankAcc xago.CreateBankAccountArgs) (*linkedaccounts.LinkedAccount, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating xago beneficiary.")

	var hasSubAcc bool
	err := workflow.ExecuteActivity(ctx, a.WalletHasSubAccount, bankAcc.WalletID).Get(ctx, &hasSubAcc)
	if err != nil {
		return nil, err
	}

	if !hasSubAcc {
		var subAcc external.SubAccount
		err = workflow.ExecuteActivity(ctx, a.CreateSubAccount, bankAcc.WalletID).Get(ctx, &subAcc)
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

		err = workflow.ExecuteActivity(ctx, a.SaveSubAccount, bankAcc.WalletID, accID, subAcc).Get(ctx, nil)
		if err != nil {
			return nil, err
		}
	}

	var ben external.CreateBeneficiaryResp
	err = workflow.ExecuteActivity(ctx, a.CreateExternalBeneficiaries, bankAcc).Get(ctx, &ben)
	if err != nil {
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.SaveBeneficiary, bankAcc.WalletID, ben).Get(ctx, nil)
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
	err = workflow.ExecuteActivity(ctx, a.AddBeneficiaryLinkedAccount, bankAcc.WalletID, ben).Get(ctx, &la)
	if err != nil {
		return nil, err
	}

	return &la, nil
}

func (a *Activity) CreateExternalBeneficiaries(ctx context.Context, bankAcc xago.CreateBankAccountArgs) (*external.CreateBeneficiaryResp, error) {
	details, err := a.b.KYC().GetIndividualDetails(ctx, bankAcc.WalletID)
	if err != nil {
		return nil, err
	}

	reqStruct := external.CreateBeneficiaryReq{
		Name:                details.LastName + " " + details.LastName,
		Scope:               "bank",
		CurrencyCode:        "ZAR",
		AccountNumber:       bankAcc.AccountNumber,
		BranchCode:          bankAcc.BranchCode,
		BankName:            bankAcc.BankName,
		BankCountry:         "ZA",
		AccountName:         details.FirstName,
		BankBeneficiaryType: "IBAN",
		Reference:           details.FirstName + " " + details.LastName[:1],
		Iban:                bankAcc.IBAN,
		Bic:                 bankAcc.BIC,
		AccountType:         "typeAccountNumber",
	}
	if details.Address != nil {
		reqStruct.BeneficiaryPhysicalAddress = details.Address.Line1
		reqStruct.BeneficiaryCity = details.Address.City
		reqStruct.BeneficiaryCountry = details.Address.CountryCode
		reqStruct.BeneficiaryPostalCode = details.Address.ZipCode
		reqStruct.BeneficiaryAddress = details.Address.Line1
	}

	return a.b.External().AddBeneficiary(ctx, reqStruct)
}

func (a *Activity) SaveBeneficiary(ctx context.Context, walletID string, sa external.CreateBeneficiaryResp) error {
	if len(sa.Beneficiaries) != 1 {
		return temporal.NewNonRetryableApplicationError(fmt.Sprintf("incorrect number of beneficiaries, expected 1 got %d", len(sa.Beneficiaries)), "external", nil)
	}
	b := sa.Beneficiaries[0]
	_, err := a.b.DB().ExecContext(ctx, `INSERT INTO xago_beneficiaries (id, wallet_id, address, reference, status, currency, scope, name) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		b.ID, walletID, b.BeneficiaryAddress, b.Reference, b.Status, b.CurrencyCode, b.Scope, b.Name)
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

	cc := currency.ParseCurrency(b.CurrencyCode)
	nation := country.ZA
	if cc == currency.USD {
		nation = country.US
	}

	var mask string
	idx := len(b.AccountNumber) - 3
	for i := range b.AccountNumber {
		if i == 0 || i >= idx {
			mask += string(b.AccountNumber[i])
		} else {
			mask += "*"
		}
	}

	return a.b.LinkedAccounts().Create(ctx, &linkedaccounts.CreateArgs{
		ID:              id,
		WalletID:        walletID,
		Name:            fmt.Sprintf("Xago %s Beneficiary", cc),
		Nickname:        fmt.Sprintf("Xago %s Beneficiary", cc),
		Mask:            mask,
		Provider:        xago.ProviderName,
		ProviderID:      b.ID,
		Type:            xago.AccTypeBank,
		CanSend:         false,
		CanReceive:      false,
		State:           linkedaccounts.Verified,
		SendCountry:     nation,
		SendCurrency:    cc,
		ReceiveCountry:  nation,
		ReceiveCurrency: cc,
	})
}
