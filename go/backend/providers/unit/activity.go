package unit

import (
	context "context"
	"errors"
	"fmt"
	"reflect"

	"go.temporal.io/sdk/temporal"

	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/identity"
)

type (
	Activity struct {
		validator       *validator.Validate
		unitService     Service
		accountsService accounts.Client
		identityService identity.Client
	}

	ActivityArgs struct {
		UnitService     Service         `validate:"required"`
		AccountsService accounts.Client `validate:"required"`
		IdentityService identity.Client `validate:"required"`
	}
)

func NewActivity(args *ActivityArgs) (*Activity, error) {
	v := validator.New()
	if err := v.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	return &Activity{
		validator:       v,
		unitService:     args.UnitService,
		accountsService: args.AccountsService,
		identityService: args.IdentityService,
	}, nil
}

func (a *Activity) UnitCreateApplication(ctx context.Context, args *CreateApplicationArgs) (*Application, error) {
	if err := a.validator.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	// TODO: retryable/non-retryable errors
	// make sure identity exists
	_, err := a.identityService.Get(ctx, args.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	application, err := a.unitService.CreateApplication(ctx, args)
	if err != nil {
		if IsRetryableError(err) {
			return nil, err
		} else {
			return nil, temporal.NewNonRetryableApplicationError("failed to create application", "unit", err)
		}
	}

	return application, nil
}

type UnitCreateAccountArgs struct {
	IdentityID       string
	DepositAccountID string
}

func (a *Activity) UnitCreateAccount(ctx context.Context, args *UnitCreateAccountArgs) (string, error) {
	identity, err := a.identityService.Get(ctx, args.IdentityID)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	acc, err := a.accountsService.Create(ctx, &accounts.CreateAccountArgs{
		IdentityID: identity.ID,
		Provider:   "unit",
		ProviderID: args.DepositAccountID,
	})
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	return acc.ID, nil
}

type UnitCreateCustomerArgs struct {
	CustomerID string `validate:"required"`
	Type       string `validate:"required"`
	IdentityID string `validate:"required"`
}

func (a *Activity) UnitCreateCustomer(ctx context.Context, args *UnitCreateCustomerArgs) error {
	if err := a.validator.Struct(args); err != nil {
		return fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	_, err := a.unitService.CreateCustomer(ctx, &CreateCustomerArgs{
		ID:         args.CustomerID,
		IdentityID: args.IdentityID,
		Type:       args.Type,
	})
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	return nil
}

func (a *Activity) UnitCreateDepositAccount(
	ctx context.Context,
	customerID string,
) (*DepositAccount, error) {
	ret, err := a.unitService.CreateDepositAccount(ctx, customerID)
	if err != nil {
		if errors.Is(err, accounts.ErrNotFound) || errors.Is(err, ErrNotFound) {
			return nil, temporal.NewNonRetryableApplicationError(err.Error(), "ErrInternal", err)
		}
		if errors.Is(err, ErrInvalidArgument) {
			return nil, temporal.NewNonRetryableApplicationError(err.Error(), "ErrInvalidArgument", err)
		}
		return nil, err
	}

	return ret, nil
}

func (a *Activity) UnitInitiateUserDeposit(
	ctx context.Context,
	args *InitiateUserDepositArgs,
) (*UserAchDeposit, error) {
	achDeposit, err := a.unitService.InitiateUserDeposit(ctx, &InitiateUserDepositArgs{
		DepositID:       args.DepositID,
		AccountID:       args.AccountID,
		FundingsourceID: args.FundingsourceID,
		Amount:          args.Amount,
		Description:     args.Description,
	})
	if err != nil {
		if !IsRetryableError(err) {
			return nil, temporal.NewNonRetryableApplicationError(err.Error(), reflect.TypeOf(err).String(), err)
		}
		return nil, err
	}

	return achDeposit, nil
}
