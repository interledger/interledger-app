package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"gitlab.com/fynbos/backend/db"

	"gitlab.com/fynbos/backend/providers/mx/workflows"

	"gitlab.com/fynbos/backend/providers/mx"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/env"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

func CreateAccount(
	ctx context.Context,
	b Backends,
	args mx.CreateAccountArgs,
) (*mx.Account, error) {
	if err := b.Validator().Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", mx.ErrInvalidArgument, err)
	}

	ret := &mx.Account{}
	err := b.DB().GetContext(
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
		if db.IsErrorCode(err, db.UniqueViolationError) {
			return nil, mx.ErrDuplicate
		}
		return nil, fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	return ret, nil
}

func GetAccount(ctx context.Context, b Backends, mxAccountGuid string) (*mx.Account, error) {
	ret := &mx.Account{}
	err := b.DB().GetContext(ctx, ret, "SELECT * FROM mx_accounts WHERE guid=$1", mxAccountGuid)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%w %s", mx.ErrNotFound, fmt.Sprintf("mxAccountGuid=%s", mxAccountGuid))
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	return ret, nil
}

func GetAccountByFundingsource(ctx context.Context, b Backends, fundingsourceID string) (*mx.Account, error) {
	ret := &mx.Account{}
	err := b.DB().GetContext(ctx, ret, "SELECT * FROM mx_accounts WHERE fundingsource_id=$1", fundingsourceID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%w %s", mx.ErrNotFound, fmt.Sprintf("fundingsourceID=%s", fundingsourceID))
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	return ret, nil
}

func StartIdentityAggregation(ctx context.Context, b Backends, mxUserGuid, mxMemberGuid string) (*mx.Member, error) {
	member, err := b.MXExternal().AggregateIdentity(ctx, mxUserGuid, mxMemberGuid)
	if err != nil {
		return nil, fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	return &mx.Member{
		Guid:                     member.Guid,
		UserGuid:                 member.UserGuid,
		AggregatedAt:             member.AggregatedAt,
		IsBeingAggregated:        member.IsBeingAggregated,
		SuccessfullyAggregatedAt: member.SuccessfullyAggregatedAt,
		ConnectionStatus:         member.ConnectionStatus,
	}, nil
}

func GetMemberStatus(ctx context.Context, b Backends, mxUserGuid, mxMemberGuid string) (*mx.Member, error) {
	member, err := b.MXExternal().GetMemberStatus(ctx, mxUserGuid, mxMemberGuid)
	if err != nil {
		return nil, fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	return &mx.Member{
		Guid:                     member.Guid,
		UserGuid:                 member.UserGuid,
		AggregatedAt:             member.AggregatedAt,
		IsBeingAggregated:        member.IsBeingAggregated,
		SuccessfullyAggregatedAt: member.SuccessfullyAggregatedAt,
		ConnectionStatus:         member.ConnectionStatus,
	}, nil
}

func GetAccountOwner(
	ctx context.Context,
	b Backends,
	args mx.GetAccountOwnerArgs,
) (*mx.AccountOwner, error) {
	owners, err := b.MXExternal().GetAccountOwners(ctx, args.MxUserGuid, args.MxMemberGuid)
	if err != nil {
		return nil, fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	var ret *mx.AccountOwner = nil
	for _, owner := range owners {
		if owner.AccountGuid == args.MxAccountGuid {
			ret = &mx.AccountOwner{
				OwnerName:   owner.OwnerName,
				AccountGuid: owner.AccountGuid,
			}
			break
		}
	}
	if ret == nil {
		return nil, fmt.Errorf(
			"%w No account owner details found for mx account guid=%s",
			mx.ErrNotFound,
			args.MxAccountGuid,
		)
	}

	return ret, nil
}

func ReadAccount(ctx context.Context, b Backends, mxAccountGuid string) (*mx.AccountDetails, error) {
	mxAccount, err := GetAccount(ctx, b, mxAccountGuid)
	if err != nil {
		return nil, err
	}

	acc, err := b.MXExternal().ReadAccount(ctx, mxAccount.UserGuid, mxAccount.Guid)
	if err != nil {
		return nil, fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	return &mx.AccountDetails{
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

func GetSelectedAccountGuid(ctx context.Context, b Backends, mxUserGuid string, mxMemberGuid string) (string, error) {
	accountNumbers, err := b.MXExternal().GetAccountNumbers(ctx, mxUserGuid, mxMemberGuid)
	if err != nil {
		return "", fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	// there should be only 1 entry.
	if len(accountNumbers) != 1 {
		return "", fmt.Errorf(
			"%w Unable to find account user selected. %d accounts were returned.",
			mx.ErrInternal,
			len(accountNumbers),
		)
	}

	return accountNumbers[0].AccountGuid, nil
}

func GetMxUserByAccountID(ctx context.Context, b Backends, accountID string) (string, error) {
	mxUserGuids := []string{}
	err := b.DB().SelectContext(
		ctx,
		&mxUserGuids,
		"SELECT DISTINCT user_guid FROM mx_accounts WHERE account_id=$1;",
		accountID,
	)
	if err != nil {
		return "", fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	if len(mxUserGuids) == 0 {
		return "", fmt.Errorf("%w", mx.ErrNotFound)
	}

	if len(mxUserGuids) != 1 {
		return "", fmt.Errorf("%w There are %d mx users linked to accountID=%s", mx.ErrInternal, len(mxUserGuids), accountID)
	}

	return mxUserGuids[0], nil
}

func VerifyOwnership(ctx context.Context, b Backends, args mx.VerifyOwnershipArgs) error {
	acc, err := b.Accounts().Get(ctx, args.AccountID)
	if err != nil {
		return fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	user, err := b.Identity().Get(ctx, acc.IdentityID)
	if err != nil {
		return fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	ownerDetails, err := GetAccountOwner(ctx, b, mx.GetAccountOwnerArgs{
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
		return mx.ErrOwnershipCheckFailed
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

func GetConnectWidget(
	ctx context.Context,
	b Backends,
	accountID string,
	identityID string,
) (string, error) {
	acc, err := b.Accounts().Get(ctx, accountID)
	if err != nil {
		return "", fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	mxUserGuid := ""
	mxUserGuid, err = GetMxUserByAccountID(ctx, b, acc.ID)
	if errors.Is(err, mx.ErrNotFound) {
		mxUserGuid, err = b.MXExternal().CreateUser(ctx)
		if err != nil {
			return "", fmt.Errorf("%w %s", mx.ErrInternal, err)
		}
	} else if err != nil {
		return "", fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	url, err := b.MXExternal().GetWidgetUrl(ctx, mxUserGuid)
	if err != nil {
		return "", fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	return url, nil
}

func InitiateCreateAccount(
	ctx context.Context,
	b Backends,
	args mx.InitiateCreateAccountArgs,
) (string, error) {
	if err := b.Validator().Struct(args); err != nil {
		return "", fmt.Errorf("%w %s", mx.ErrInvalidArgument, err)
	}

	acc, err := b.Accounts().Get(ctx, args.AccountID)
	if err != nil {
		return "", fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	if acc.IdentityID != args.IdentityID {
		return "", mx.ErrUnauthorized
	}

	workflowUuid := uuid.NewString()
	_, err = b.Temporal().ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:                    "create_mx_bank_account_" + workflowUuid,
			TaskQueue:             "backend",
			WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
		},
		workflows.CreateMxAccountWorkflow,
		&workflows.CreateMxAccountWorkflowArgs{
			ID:         workflowUuid,
			IdentityID: args.IdentityID,
			AccountID:  args.AccountID,
			UserGuid:   args.UserGuid,
			MemberGuid: args.MemberGuid,
		},
	)
	if err != nil {
		return "", fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	return workflowUuid, nil
}

func WaitForCreateAccount(ctx context.Context, b Backends, fundingsourceID string) error {
	err := b.Temporal().GetWorkflow(ctx, "create_mx_bank_account_"+fundingsourceID, "").Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	return nil
}

func InitiateCreateFundingsource(ctx context.Context, b Backends, args mx.InitiateCreateFundingsourceArgs) error {
	if err := b.Validator().Struct(args); err != nil {
		return fmt.Errorf("%w %s", mx.ErrInvalidArgument, err)
	}

	mxAccount, err := GetAccount(ctx, b, args.MxAccountGuid)
	if err != nil {
		return err
	}

	acc, err := b.Accounts().Get(ctx, args.AccountID)
	if err != nil {
		return fmt.Errorf("%w %s", mx.ErrInternal, err)
	}
	if mxAccount.AccountID != acc.ID {
		return mx.ErrUnauthorized
	}

	user, err := b.Identity().Get(ctx, acc.IdentityID)
	if err != nil {
		return fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	v, err := b.Twilio().CheckVerificationCode(ctx, &twilio.CheckVerificationCodeArgs{
		Code:        args.Otp,
		PhoneNumber: user.MobileNumber,
	})
	if err != nil {
		return fmt.Errorf("%w %s", mx.ErrInternal, err)
	}
	if !v.IsValid() {
		return fmt.Errorf("%w %s", mx.ErrUnauthorized, "2FA failed.")
	}

	_, err = b.Temporal().ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:                    "mx_create_fundingsource_" + uuid.NewString(),
			TaskQueue:             "backend",
			WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
		},
		workflows.MxCreateFundingsourceWorkflow,
		&workflows.MxCreateFundingsourceWorkflowArgs{
			MxAccountGuid: mxAccount.Guid,
			AccountID:     acc.ID,
			Name:          args.Name,
		},
	)
	if err != nil {
		return fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	return nil
}

func StartBalanceAggregation(ctx context.Context, b Backends, mxAccountGuid string) (*mx.Member, error) {
	mxAcc, err := GetAccount(ctx, b, mxAccountGuid)
	if err != nil {
		return nil, err
	}

	member, err := b.MXExternal().AggregateBalance(ctx, mxAcc.UserGuid, mxAcc.MemberGuid)
	if err != nil {
		return nil, fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	return &mx.Member{
		Guid:                     member.Guid,
		UserGuid:                 member.UserGuid,
		AggregatedAt:             member.AggregatedAt,
		IsBeingAggregated:        member.IsBeingAggregated,
		SuccessfullyAggregatedAt: member.SuccessfullyAggregatedAt,
		ConnectionStatus:         member.ConnectionStatus,
	}, nil
}

func GetAccountBalance(ctx context.Context, b Backends, mxAccountGuid string) (*mx.AccountBalance, error) {
	mxAcc, err := GetAccount(ctx, b, mxAccountGuid)
	if err != nil {
		return nil, err
	}

	accountDetails, err := b.MXExternal().ReadAccount(ctx, mxAcc.UserGuid, mxAcc.Guid)
	if err != nil {
		return nil, fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	balanceInCents := accountDetails.AvailableBalance * 100 // assume assetScale=2
	balance := int64(balanceInCents)                        // truncate after cents

	return &mx.AccountBalance{
		AssetCode:  accountDetails.CurrencyCode,
		AssetScale: 2,
		Value:      balance,
	}, nil
}
