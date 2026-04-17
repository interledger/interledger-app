package ops

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	httplogger "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/backend/providers/xago"
	"gitlab.com/fynbos/backend/providers/xago/external"
	"gitlab.com/fynbos/pacioli"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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

func NewActivity(ab ActivityBackends, cfg external.Config) *Activity {
	ex := external.New(&http.Client{
		Transport: otelhttp.NewTransport(
			httplogger.NewTransport(http.DefaultTransport, ab, nil),
		),
	}, ab.DB(), cfg)

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
	var depositRef string
	for _, b := range sa.Beneficiaries {
		if strings.EqualFold(b.BeneficiaryType, "rollup") {
			depositRef = b.DepositReference
			break
		}
	}
	if depositRef == "" {
		return fmt.Errorf("%w no deposit reference provided for xago sub account", xago.ErrInternal)
	}
	_, err := a.b.DB().ExecContext(ctx, "INSERT INTO xago_sub_accounts (id, wallet_id, account_id, deposit_address, deposit_tag, deposit_reference) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING ",
		accountID, walletID, sa.AccountID, sa.DepositAddress, strconv.Itoa(sa.DepositTag), depositRef)
	if err != nil && !db.IsErrorCode(err, db.UniqueViolationError) {
		return err
	}
	return nil
}

func UpdateInquiryLinkWorkflow(ctx workflow.Context, args xago.InquiryLinkUpdate) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Update xago inquiry link.")

	err := workflow.ExecuteActivity(ctx, a.UpdateSubAccountActivity, args).Get(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}

func (a *Activity) UpdateSubAccountActivity(ctx context.Context, args xago.InquiryLinkUpdate) error {
	personaInquiry, err := a.b.KYC().GetApprovedPersonaInquiryURL(ctx, args.WalletID)
	if err != nil {
		return fmt.Errorf("xago internal error: wallet does not have an approved persona inquiry")
	}

	kycDetails, err := a.b.KYC().GetIndividualDetails(ctx, args.WalletID)
	if err != nil {
		return err
	}

	idNumber, err := a.b.KYC().GetPersonaZAIDNumber(ctx, args.WalletID)
	if err != nil {
		return err
	}

	return a.b.External().UpdateSubAccount(ctx, args.AccountID, external.UpdateSubAccountRequest{
		ThirdPartyVerificationURL: personaInquiry,
		IDNumber:                  idNumber,
		PhysicalAddress:           kycDetails.Address.String(),
	})
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

	personaInquiry, err := a.b.KYC().GetApprovedPersonaInquiryURL(ctx, walletID)
	if errors.Is(err, kyc.ErrNoKYCInfo) {
		return nil, temporal.NewNonRetryableApplicationError("xago internal error: wallet does not have an approved persona inquiry.", "ErrInternal", err)
	}
	if err != nil {
		return nil, err
	}

	idNumber, err := a.b.KYC().GetPersonaZAIDNumber(ctx, walletID)
	if err != nil {
		return nil, err
	}

	return a.b.External().CreateSubAccount(ctx, ul[0], *id, idNumber, personaInquiry)
}

func CreateBalanceAccountWorkflow(ctx workflow.Context, args xago.CreateBalanceAccArgs) (*linkedaccounts.LinkedAccount, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating xago beneficiary.")

	var hasSubAcc bool
	err := workflow.ExecuteActivity(ctx, a.WalletHasSubAccount, args.WalletID).Get(ctx, &hasSubAcc)
	if err != nil {
		return nil, err
	}

	if !hasSubAcc {
		var subAcc external.SubAccount
		err = workflow.ExecuteActivity(ctx, a.CreateSubAccount, args.WalletID).Get(ctx, &subAcc)
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

		err = workflow.ExecuteActivity(ctx, a.SaveSubAccount, args.WalletID, accID, subAcc).Get(ctx, nil)
		if err != nil {
			return nil, err
		}
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
	err = workflow.ExecuteActivity(ctx, a.AddBalanceLinkedAccount, laID, args).Get(ctx, &la)
	if err != nil {
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.AddBalanceAccount, laID, args).Get(ctx, nil)

	return &la, err
}

func (a *Activity) AddBalanceAccount(ctx context.Context, id string, args xago.CreateBalanceAccArgs) error {

	ledgerID := xago.LedgerIDUSD
	if args.Currency == currency.ZAR {
		ledgerID = xago.LedgerIDZAR
	}

	accs, err := a.b.Pacioli().ConfigureAccounts(ctx, []pacioli.ConfigureAccountArgs{
		{
			ID:                         id,
			LedgerID:                   ledgerID,
			Code:                       1,
			DebitsMustNotExceedCredits: true,
			CreditsMustNotExceedDebits: false,
		},
	})
	if err != nil {
		return err
	}

	if len(accs) == 0 {
		// No error codes to speak of
		return nil
	}

	if accs[0].Code != pacioli.AccountOK && accs[0].Code != pacioli.AccountExists {
		return fmt.Errorf("%w failed to setup account status(%s)", xago.ErrInternal, accs[0].Code)
	}

	return nil
}

func (a *Activity) AddBalanceLinkedAccount(ctx context.Context, id string, args xago.CreateBalanceAccArgs) (*linkedaccounts.LinkedAccount, error) {
	la, _ := a.b.LinkedAccounts().Get(ctx, id)
	if la != nil {
		return la, nil
	}

	if args.Currency != currency.ZAR && args.Currency != currency.USD {
		return nil, temporal.NewNonRetryableApplicationError("invalid currency", "invalid_argument", xago.ErrInternal, "currency", args.Currency)
	}

	cc := args.Currency
	nation := country.ZA
	if cc == currency.USD {
		nation = country.US
	}

	return a.b.LinkedAccounts().Create(ctx, &linkedaccounts.CreateArgs{
		ID:              id,
		WalletID:        args.WalletID,
		Name:            args.Title,
		Nickname:        args.Nickname,
		Provider:        xago.ProviderName,
		ProviderID:      fmt.Sprintf("xago_%s_%s", cc.String(), args.WalletID), // Deterministic providerID stops duplicate accounts from being created.
		Type:            xago.AccTypeBalance,
		CanSend:         true,
		CanReceive:      true,
		State:           linkedaccounts.Verified,
		SendCountry:     nation,
		SendCurrency:    cc,
		ReceiveCountry:  nation,
		ReceiveCurrency: cc,
	})
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

	var beneficiary external.AccountBeneficiaries
	err = workflow.ExecuteActivity(ctx, a.CreateExternalBeneficiaries, bankAcc).Get(ctx, &beneficiary)
	if err != nil {
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.SaveBeneficiary, bankAcc.WalletID, beneficiary).Get(ctx, nil)
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
	err = workflow.ExecuteActivity(ctx, a.AddBeneficiaryLinkedAccount, bankAcc.WalletID, laID, beneficiary).Get(ctx, &la)
	if err != nil {
		return nil, err
	}

	return &la, nil
}

func (a *Activity) CreateExternalBeneficiaries(ctx context.Context, bankAcc xago.CreateBankAccountArgs) (*external.AccountBeneficiaries, error) {
	subAccount, err := LookupSubAccount(ctx, a.b, bankAcc.WalletID)
	if err != nil {
		return nil, err
	}

	details, err := a.b.KYC().GetIndividualDetails(ctx, bankAcc.WalletID)
	if err != nil {
		return nil, err
	}

	userList, err := a.b.Users().ListUsers(ctx, bankAcc.WalletID)
	if err != nil {
		return nil, err
	}
	if len(userList) < 1 {
		err = fmt.Errorf("%w wallet has (%d) users associated", xago.ErrInternal, len(userList))
		return nil, err
	}

	reqStruct := external.CreateBeneficiaryReq{
		Name:          details.FirstName + " " + details.LastName,
		Scope:         "bank",
		CurrencyCode:  "ZAR",
		AccountNumber: bankAcc.AccountNumber,
		BranchCode:    bankAcc.BranchCode,
		BankName:      bankAcc.BankName,
		BankCountry:   "ZA",
		AccountName:   details.FirstName,
		Reference:     details.FirstName + " " + details.LastName[:1],
		Iban:          bankAcc.IBAN,
		Bic:           bankAcc.BIC,
		AccountType:   "typeAccountNumber",
		KYCRequest: external.CreateBeneficiaryKYCRequest{
			SubAccountID: subAccount.AccountID,
		},
		MobileNumber: userList[0].PhoneNumber,
	}
	if details.Address != nil {
		reqStruct.BeneficiaryPhysicalAddress = details.Address.Line1
		reqStruct.BeneficiaryCity = details.Address.City
		reqStruct.BeneficiaryCountry = details.Address.CountryCode
		reqStruct.BeneficiaryPostalCode = details.Address.ZipCode
		reqStruct.BeneficiaryAddress = details.Address.Line1
		reqStruct.BeneficiaryDistrict = details.Address.State
	}

	beneficiary, err := a.b.External().AddBeneficiary(ctx, reqStruct)
	if err != nil {
		return nil, err
	}

	return beneficiary, nil
}

// Not used at the moment; can be removed once the Xago integration is updated
func (a *Activity) GetExternalBeneficiary(ctx context.Context, externalID string) (*external.AccountBeneficiaries, error) {
	var err error
	var response *external.ListBeneficiariesResponse
	limit, page := uint(20), uint(1)
	hasNextPage := true
	for {
		if !hasNextPage {
			break
		}

		response, err = a.b.External().ListBeneficiaries(ctx, limit, page)
		if err != nil {
			return nil, err
		}

		for _, acc := range response.Beneficiaries {
			if acc.ID == externalID {
				return &acc, nil
			}
		}

		hasNextPage = response.Pagination.NumberOfPages >= response.Pagination.PageNumber
		page++
	}

	return nil, temporal.NewNonRetryableApplicationError("xago beneficiary not found in list", "external", nil)
}

func (a *Activity) SaveBeneficiary(ctx context.Context, walletID string, sa external.AccountBeneficiaries) error {
	_, err := a.b.DB().ExecContext(ctx, `INSERT INTO xago_beneficiaries (id, wallet_id, address, reference, status, currency, scope, name) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		sa.ID, walletID, sa.BeneficiaryAddress, sa.Reference, sa.Status, sa.CurrencyCode, sa.Scope, sa.Name)
	if db.IsErrorCode(err, db.UniqueViolationError) {
		return nil
	}

	return err
}

func (a *Activity) AddBeneficiaryLinkedAccount(ctx context.Context, walletID, id string, b external.AccountBeneficiaries) (*linkedaccounts.LinkedAccount, error) {

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
		Mask:            mask,
		Provider:        xago.ProviderName,
		ProviderID:      b.ID,
		Type:            xago.AccTypeBank,
		CanSend:         false,
		CanReceive:      true,
		State:           linkedaccounts.Verified,
		SendCountry:     nation,
		SendCurrency:    cc,
		ReceiveCountry:  nation,
		ReceiveCurrency: cc,
		ReceiveNetwork:  b.BankName,
	})
}

func FundEUROLiquidityAccountWorkflow(ctx workflow.Context, amount string) error {
	currencyAmount, err := currency.FromString(amount, currency.EUR)
	if err != nil {
		return err
	}

	var a *Activity
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	if err := workflow.ExecuteActivity(ctx, a.FundEUROLiquidityAccountW, currencyAmount).Get(ctx, nil); err != nil {
		return err
	}

	return nil
}

func (a *Activity) FundEUROLiquidityAccountW(ctx context.Context, amount currency.Amount) error {
	_, err := a.b.Pacioli().CreateTransfers(ctx, []pacioli.CreateTransferArgs{
		{
			ID:              uuid.NewString(),
			Amount:          amount.Value,
			CreditAccountID: xago.EURLiquidityAccount,
			DebitAccountID:  xago.EUROpsAccount,
			Pending:         false,
			Code:            1, // TODO magic number, but what does it mean, double rainbow?
			Ledger:          xago.LedgerIDEUR,
		},
	})
	if err != nil {
		return fmt.Errorf("%w %s", xago.ErrInternal, err)
	}

	// TODO check if this is needed
	// if len(tx) > 0 {

	// 	if tx[0].Code == pacioli.TransferExists {
	// 		return nil
	// 	}
	// 	if tx[0].Code == pacioli.TransferExceedsCredits || tx[0].Code == pacioli.TransferExceedsDebits || tx[0].Code == pacioli.TransferExceedsPendingTransferAmount {
	// 		return fmt.Errorf("%w insufficient balance code (%s)", xago.ErrInsufficientBalance, tx[0].Code.String())
	// 	}
	// 	if tx[0].Code != 0 {
	// 		return fmt.Errorf("%w non success code (%s)", xago.ErrInternal, tx[0].Code.String())
	// 	}
	// }

	return nil
}
