package mx

//go:generate mockgen -destination=./mock.go -package=mx -source=./service.go

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/mx/external"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/env"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

var (
	ErrInvalidArgument      = errors.New("mx provider: invalid argument.")
	ErrInternal             = errors.New("mx provider: internal error.")
	ErrNotFound             = errors.New("mx provider: not found.")
	ErrDuplicate            = errors.New("mx provider: duplicate.")
	ErrOwnershipCheckFailed = errors.New("mx provider: ownership check failed.")
	ErrUnauthorized         = errors.New("mx provider: unauthorized.")
)

type (
	Service interface {
		CreateAccount(ctx context.Context, args *CreateAccountArgs) (*Account, error)
		GetAccount(ctx context.Context, mxAccountGuid string) (*Account, error)
		GetAccountByFundingsource(ctx context.Context, fundingsourceID string) (*Account, error)
		StartIdentityAggregation(ctx context.Context, mxUserGuid, mxMemberGuid string) (*Member, error)
		GetMemberStatus(ctx context.Context, mxUserGuid, mxMemberGuid string) (*Member, error)
		// This will fetch the account owner information for the specified mx account. The identity
		// aggregation has to have been completed first.
		GetAccountOwner(ctx context.Context, args *GetAccountOwnerArgs) (*AccountOwner, error)
		ReadAccount(ctx context.Context, mxAccountGuid string) (*AccountDetails, error)
		// The mx connect widget will allow the user to log into their bank and select an account.
		// They do not pass this to us on the front end and so we need to call out to find out the
		// mx account guid of the account that was selected.
		// Calling the users/:users/members/:members/account_numbers should only have the account selected
		// by the user.
		GetSelectedAccountGuid(ctx context.Context, mxUserGuid string, mxMemberGuid string) (string, error)
		GetMxUserByAccountID(ctx context.Context, accountID string) (string, error)
		VerifyOwnership(ctx context.Context, args *VerifyOwnershipArgs) error
		GetConnectWidget(ctx context.Context, accountID string, identityID string) (string, error)
		InitiateCreateAccount(ctx context.Context, args *InitiateCreateAccountArgs) (string, error)
		// Blocks till the workflow is complete and returns the result.
		WaitForCreateAccount(ctx context.Context, fundingsourceID string) error
		InitiateCreateFundingsource(ctx context.Context, args *InitiateCreateFundingsourceArgs) error
		StartBalanceAggregation(ctx context.Context, mxAccountGuid string) (*Member, error)
		GetAccountBalance(ctx context.Context, mxAccountGuid string) (*AccountBalance, error)
	}

	Account struct {
		Guid            string //from mx
		UserGuid        string `db:"user_guid"`   // from mx
		MemberGuid      string `db:"member_guid"` // from mx
		AccountID       string `db:"account_id"`  // Fynbos account id
		FundingsourceID string `db:"fundingsource_id"`
		CreatedAt       string `db:"created_at"`
		UpdatedAt       string `db:"updated_at"`
	}

	Member struct {
		Guid                     string
		UserGuid                 string
		AggregatedAt             string `json:"aggregated_at"`
		IsBeingAggregated        bool   `json:"is_being_aggregated"`
		SuccessfullyAggregatedAt string `json:"successfully_aggregated_at"`
		ConnectionStatus         string `json:"connection_status"`
	}

	AccountOwner struct {
		AccountGuid string
		OwnerName   string
	}

	AccountDetails struct {
		Guid              string
		UserGuid          string `json:"user_guid"`
		MemberGuid        string `json:"member_guid"`
		AccountNumber     string `json:"account_number"`
		InstitutionNumber string `json:"institution_number"`
		RoutingNumber     string `json:"routing_number"`
		TransitNumber     string `json:"transit_number"`
		CurrencyCode      string `json:"currency_code"`
		Type              string
	}

	AccountBalance struct {
		AssetCode  string
		AssetScale uint8
		Value      int64
	}

	ServiceArgs struct {
		ExternalClient  external.Mx     `validate:"required"`
		Db              *sqlx.DB        `validate:"required"`
		AccountsService accounts.Client `validate:"required"`
		IdentityService identity.Client `validate:"required"`
		Temporal        client.Client   `validate:"required"`
		Twilio          twilio.Service  `validate:"required"`
	}

	service struct {
		v               *validator.Validate
		externalClient  external.Mx
		db              *sqlx.DB
		accountsService accounts.Client
		identityService identity.Client
		temporal        client.Client
		twilio          twilio.Service
	}
)

func NewService(args *ServiceArgs) (Service, error) {
	v := validator.New()
	if err := v.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	return &service{
		v:               v,
		externalClient:  args.ExternalClient,
		db:              args.Db,
		accountsService: args.AccountsService,
		identityService: args.IdentityService,
		temporal:        args.Temporal,
		twilio:          args.Twilio,
	}, nil
}

type CreateAccountArgs struct {
	Guid            string `validate:"required"` // from mx
	UserGuid        string `validate:"required"` // from mx
	MemberGuid      string `validate:"required"` // from mx
	AccountID       string `validate:"uuid4"`
	FundingsourceID string `validate:"uuid4"`
}

func (s *service) CreateAccount(
	ctx context.Context,
	args *CreateAccountArgs,
) (*Account, error) {
	if err := s.v.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	ret := &Account{}
	err := s.db.GetContext(
		ctx,
		ret,
		`
		INSERT INTO mx_accounts (
			guid,
			user_guid,
			member_guid,
			account_id,
			fundingsource_id
		)
		VALUES ($1, $2, $3, $4, $5) RETURNING *;
		`,
		args.Guid,
		args.UserGuid,
		args.MemberGuid,
		args.AccountID,
		args.FundingsourceID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint \"primary\"") {
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return ret, nil
}

func (s service) GetAccount(ctx context.Context, mxAccountGuid string) (*Account, error) {
	ret := &Account{}
	err := s.db.GetContext(ctx, ret, "SELECT * FROM mx_accounts WHERE guid=$1", mxAccountGuid)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%w %s", ErrNotFound, fmt.Sprintf("mxAccountGuid=%s", mxAccountGuid))
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return ret, nil
}

func (s service) GetAccountByFundingsource(ctx context.Context, fundingsourceID string) (*Account, error) {
	ret := &Account{}
	err := s.db.GetContext(ctx, ret, "SELECT * FROM mx_accounts WHERE fundingsource_id=$1", fundingsourceID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%w %s", ErrNotFound, fmt.Sprintf("fundingsourceID=%s", fundingsourceID))
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return ret, nil
}

func (s *service) StartIdentityAggregation(ctx context.Context, mxUserGuid, mxMemberGuid string) (*Member, error) {
	member, err := s.externalClient.AggregateIdentity(ctx, mxUserGuid, mxMemberGuid)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &Member{
		Guid:                     member.Guid,
		UserGuid:                 member.UserGuid,
		AggregatedAt:             member.AggregatedAt,
		IsBeingAggregated:        member.IsBeingAggregated,
		SuccessfullyAggregatedAt: member.SuccessfullyAggregatedAt,
		ConnectionStatus:         member.ConnectionStatus,
	}, nil
}

func (s *service) GetMemberStatus(ctx context.Context, mxUserGuid, mxMemberGuid string) (*Member, error) {
	member, err := s.externalClient.GetMemberStatus(ctx, mxUserGuid, mxMemberGuid)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &Member{
		Guid:                     member.Guid,
		UserGuid:                 member.UserGuid,
		AggregatedAt:             member.AggregatedAt,
		IsBeingAggregated:        member.IsBeingAggregated,
		SuccessfullyAggregatedAt: member.SuccessfullyAggregatedAt,
		ConnectionStatus:         member.ConnectionStatus,
	}, nil
}

type GetAccountOwnerArgs struct {
	MxUserGuid    string
	MxMemberGuid  string
	MxAccountGuid string
}

func (s service) GetAccountOwner(
	ctx context.Context,
	args *GetAccountOwnerArgs,
) (*AccountOwner, error) {
	owners, err := s.externalClient.GetAccountOwners(ctx, args.MxUserGuid, args.MxMemberGuid)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	var ret *AccountOwner = nil
	for _, owner := range owners {
		if owner.AccountGuid == args.MxAccountGuid {
			ret = &AccountOwner{
				OwnerName:   owner.OwnerName,
				AccountGuid: owner.AccountGuid,
			}
			break
		}
	}
	if ret == nil {
		return nil, fmt.Errorf(
			"%w No account owner details found for mx account guid=%s",
			ErrNotFound,
			args.MxAccountGuid,
		)
	}

	return ret, nil
}

func (s service) ReadAccount(ctx context.Context, mxAccountGuid string) (*AccountDetails, error) {
	mxAccount, err := s.GetAccount(ctx, mxAccountGuid)
	if err != nil {
		return nil, err
	}

	acc, err := s.externalClient.ReadAccount(ctx, mxAccount.UserGuid, mxAccount.Guid)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &AccountDetails{
		Guid:              acc.Guid,
		UserGuid:          acc.UserGuid,
		MemberGuid:        acc.MemberGuid,
		AccountNumber:     acc.AccountNumber,
		InstitutionNumber: acc.InstitutionNumber,
		RoutingNumber:     acc.RoutingNumber,
		TransitNumber:     acc.TransitNumber,
		CurrencyCode:      acc.CurrencyCode,
		Type:              acc.Type,
	}, nil
}

func (s service) GetSelectedAccountGuid(ctx context.Context, mxUserGuid string, mxMemberGuid string) (string, error) {
	accountNumbers, err := s.externalClient.GetAccountNumbers(ctx, mxUserGuid, mxMemberGuid)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	// there should be only 1 entry.
	if len(accountNumbers) != 1 {
		return "", fmt.Errorf(
			"%w Unable to find account user selected. %d accounts were returned.",
			ErrInternal,
			len(accountNumbers),
		)
	}

	return accountNumbers[0].AccountGuid, nil
}

func (s service) GetMxUserByAccountID(ctx context.Context, accountID string) (string, error) {
	mxUserGuids := []string{}
	err := s.db.SelectContext(
		ctx,
		&mxUserGuids,
		"SELECT DISTINCT user_guid FROM mx_accounts WHERE account_id=$1;",
		accountID,
	)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	if len(mxUserGuids) == 0 {
		return "", fmt.Errorf("%w", ErrNotFound)
	}

	if len(mxUserGuids) != 1 {
		return "", fmt.Errorf("%w There are %d mx users linked to accountID=%s", ErrInternal, len(mxUserGuids), accountID)
	}

	return mxUserGuids[0], nil
}

type VerifyOwnershipArgs struct {
	AccountID     string
	MxUserGuid    string
	MxMemberGuid  string
	MxAccountGuid string
}

func (s *service) VerifyOwnership(ctx context.Context, args *VerifyOwnershipArgs) error {
	acc, err := s.accountsService.Get(ctx, args.AccountID)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	user, err := s.identityService.Get(ctx, acc.IdentityID)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	ownerDetails, err := s.GetAccountOwner(ctx, &GetAccountOwnerArgs{
		MxUserGuid:    args.MxUserGuid,
		MxMemberGuid:  args.MxMemberGuid,
		MxAccountGuid: args.MxAccountGuid,
	})
	if err != nil {
		return err
	}

	userFullName := strings.TrimSpace(fmt.Sprintf("%s %s", user.FirstName, user.LastName))
	userFullName = strings.ToUpper(userFullName)
	ownerName := strings.TrimSpace(ownerDetails.OwnerName)
	ownerName = strings.ToUpper(ownerName)

	// we do the auto verify here so we test getting the account owner details.
	if autoVerifyOwnership(userFullName) {
		return nil
	}
	if userFullName != ownerName {
		return ErrOwnershipCheckFailed
	}

	return nil
}

func autoVerifyOwnership(name string) bool {
	allowedNames := []string{"MX USER"}
	if !env.IsProd() {
		for _, allowedName := range allowedNames {
			if strings.ToUpper(strings.TrimSpace(name)) == allowedName {
				return true
			}
		}
	}
	return false
}

func (s *service) GetConnectWidget(
	ctx context.Context,
	accountID string,
	identityID string,
) (string, error) {
	acc, err := s.accountsService.Get(ctx, accountID)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	mxUserGuid := ""
	mxUserGuid, err = s.GetMxUserByAccountID(ctx, acc.ID)
	if errors.Is(err, ErrNotFound) {
		mxUserGuid, err = s.externalClient.CreateUser(ctx)
		if err != nil {
			return "", fmt.Errorf("%w %s", ErrInternal, err)
		}
	} else if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	url, err := s.externalClient.GetWidgetUrl(ctx, mxUserGuid)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	return url, nil
}

type InitiateCreateAccountArgs struct {
	UserGuid   string `validate:"required"` // from mx
	MemberGuid string `validate:"required"` // from mx
	AccountID  string `validate:"uuid4"`
	IdentityID string `validate:"uuid4"`
}

func (s *service) InitiateCreateAccount(
	ctx context.Context,
	args *InitiateCreateAccountArgs,
) (string, error) {
	if err := s.v.Struct(args); err != nil {
		return "", fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	acc, err := s.accountsService.Get(ctx, args.AccountID)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	if acc.IdentityID != args.IdentityID {
		return "", ErrUnauthorized
	}

	workflowUuid := uuid.NewString()
	_, err = s.temporal.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:                    "create_mx_bank_account_" + workflowUuid,
			TaskQueue:             "backend",
			WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
		},
		CreateMxAccountWorkflow,
		&CreateMxAccountWorkflowArgs{
			ID:         workflowUuid,
			IdentityID: args.IdentityID,
			AccountID:  args.AccountID,
			UserGuid:   args.UserGuid,
			MemberGuid: args.MemberGuid,
		},
	)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	return workflowUuid, nil
}

func (s *service) WaitForCreateAccount(ctx context.Context, fundingsourceID string) error {
	err := s.temporal.GetWorkflow(ctx, "create_mx_bank_account_"+fundingsourceID, "").Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	return nil
}

type InitiateCreateFundingsourceArgs struct {
	AccountID     string `validate:"required"`
	Otp           string `validate:"required"`
	Name          string `validate:"required"`
	MxAccountGuid string `validate:"uuid4"`
}

func (s *service) InitiateCreateFundingsource(ctx context.Context, args *InitiateCreateFundingsourceArgs) error {
	if err := s.v.Struct(args); err != nil {
		return fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	mxAccount, err := s.GetAccount(ctx, args.MxAccountGuid)
	if err != nil {
		return err
	}

	acc, err := s.accountsService.Get(ctx, args.AccountID)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}
	if mxAccount.AccountID != acc.ID {
		return ErrUnauthorized
	}

	user, err := s.identityService.Get(ctx, acc.IdentityID)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	v, err := s.twilio.CheckVerificationCode(ctx, &twilio.CheckVerificationCodeArgs{
		Code:        args.Otp,
		PhoneNumber: user.MobileNumber,
	})
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}
	if !v.IsValid() {
		return fmt.Errorf("%w %s", ErrUnauthorized, "2FA failed.")
	}

	_, err = s.temporal.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:                    "mx_create_fundingsource_" + uuid.NewString(),
			TaskQueue:             "backend",
			WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
		},
		MxCreateFundingsourceWorkflow,
		&MxCreateFundingsourceWorkflowArgs{
			MxAccountGuid: mxAccount.Guid,
			AccountID:     acc.ID,
			Name:          args.Name,
		},
	)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	return nil
}

func (s *service) StartBalanceAggregation(ctx context.Context, mxAccountGuid string) (*Member, error) {
	mxAcc, err := s.GetAccount(ctx, mxAccountGuid)
	if err != nil {
		return nil, err
	}

	member, err := s.externalClient.AggregateBalance(ctx, mxAcc.UserGuid, mxAcc.MemberGuid)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &Member{
		Guid:                     member.Guid,
		UserGuid:                 member.UserGuid,
		AggregatedAt:             member.AggregatedAt,
		IsBeingAggregated:        member.IsBeingAggregated,
		SuccessfullyAggregatedAt: member.SuccessfullyAggregatedAt,
		ConnectionStatus:         member.ConnectionStatus,
	}, nil
}

func (s *service) GetAccountBalance(ctx context.Context, mxAccountGuid string) (*AccountBalance, error) {
	mxAcc, err := s.GetAccount(ctx, mxAccountGuid)
	if err != nil {
		return nil, err
	}

	accountDetails, err := s.externalClient.ReadAccount(ctx, mxAcc.UserGuid, mxAcc.Guid)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	balanceInCents := accountDetails.AvailableBalance * 100 // assume assetScale=2
	balance := int64(balanceInCents)                        // truncate after cents

	return &AccountBalance{
		AssetCode:  accountDetails.CurrencyCode,
		AssetScale: 2,
		Value:      balance,
	}, nil
}

func CanAggregate(memberConnectionStatus string) bool {
	switch memberConnectionStatus {
	case external.CONNECTION_STATUS_CONNECTED,
		external.CONNECTION_STATUS_CREATED,
		external.CONNECTION_STATUS_DEGRADED,
		external.CONNECTION_STATUS_DISCONNECTED,
		external.CONNECTION_STATUS_EXPIRED,
		external.CONNECTION_STATUS_FAILED,
		external.CONNECTION_STATUS_IMPEDED,
		external.CONNECTION_STATUS_RECONNECTED,
		external.CONNECTION_STATUS_UPDATED,
		external.CONNECTION_STATUS_DELAYED,
		external.CONNECTION_STATUS_REJECTED,
		external.CONNECTION_STATUS_RESUMED:
		return true
	default:
		return false
	}
}

func IsSavings(accountType string) bool {
	return strings.ToLower(strings.TrimSpace(accountType)) == "savings"
}
