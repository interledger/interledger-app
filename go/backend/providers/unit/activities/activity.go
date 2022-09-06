package activities

import (
	context "context"
	"errors"
	"fmt"
	"reflect"

	"go.temporal.io/sdk/temporal"

	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/providers/unit"
	"gitlab.com/fynbos/backend/providers/unit/ops"
)

type Activity struct {
	b Backends
}

func NewActivity(b Backends) *Activity {
	return &Activity{b: b}
}

func (a *Activity) UnitCreateApplication(ctx context.Context, args *unit.CreateApplicationArgs) (*unit.Application, error) {
	if err := a.b.Val().Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", unit.ErrInvalidArgument, err)
	}

	// TODO: retryable/non-retryable errors
	// make sure identity exists
	_, err := a.b.Identity().Get(ctx, args.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", unit.ErrInternal, err)
	}

	application, err := a.b.Unit().CreateApplication(ctx, args)
	if err != nil {
		if ops.IsRetryableError(err) {
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
	identity, err := a.b.Identity().Get(ctx, args.IdentityID)
	if err != nil {
		return "", fmt.Errorf("%w %s", unit.ErrInternal, err)
	}

	acc, err := a.b.Accounts().Create(ctx, &accounts.CreateAccountArgs{
		IdentityID: identity.ID,
		Provider:   "unit",
		ProviderID: args.DepositAccountID,
	})
	if err != nil {
		return "", fmt.Errorf("%w %s", unit.ErrInternal, err)
	}

	return acc.ID, nil
}

type UnitCreateCustomerArgs struct {
	CustomerID string `validate:"required"`
	Type       string `validate:"required"`
	IdentityID string `validate:"required"`
}

func (a *Activity) UnitCreateCustomer(ctx context.Context, args *UnitCreateCustomerArgs) error {
	if err := a.b.Val().Struct(args); err != nil {
		return fmt.Errorf("%w %s", unit.ErrInvalidArgument, err)
	}

	_, err := a.b.Unit().CreateCustomer(ctx, &unit.CreateCustomerArgs{
		ID:         args.CustomerID,
		IdentityID: args.IdentityID,
		Type:       args.Type,
	})
	if err != nil {
		return fmt.Errorf("%w %s", unit.ErrInternal, err)
	}

	return nil
}

func (a *Activity) UnitCreateDepositAccount(
	ctx context.Context,
	customerID string,
) (*unit.DepositAccount, error) {
	ret, err := a.b.Unit().CreateDepositAccount(ctx, customerID)
	if err != nil {
		if errors.Is(err, accounts.ErrNotFound) || errors.Is(err, unit.ErrNotFound) {
			return nil, temporal.NewNonRetryableApplicationError(err.Error(), "ErrInternal", err)
		}
		if errors.Is(err, unit.ErrInvalidArgument) {
			return nil, temporal.NewNonRetryableApplicationError(err.Error(), "ErrInvalidArgument", err)
		}
		return nil, err
	}

	return ret, nil
}

func (a *Activity) UnitInitiateUserDeposit(
	ctx context.Context,
	args *unit.InitiateUserDepositArgs,
) (*unit.UserAchDeposit, error) {
	achDeposit, err := a.b.Unit().InitiateUserDeposit(ctx, &unit.InitiateUserDepositArgs{
		DepositID:       args.DepositID,
		AccountID:       args.AccountID,
		FundingsourceID: args.FundingsourceID,
		Amount:          args.Amount,
		Description:     args.Description,
	})
	if err != nil {
		if !ops.IsRetryableError(err) {
			return nil, temporal.NewNonRetryableApplicationError(err.Error(), reflect.TypeOf(err).String(), err)
		}
		return nil, err
	}

	return achDeposit, nil
}
