package unit

import (
	context "context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/identity"
)

type (
	Activity struct {
		validator       *validator.Validate
		unitService     Service
		accountsService accounts.Service
		identityService identity.Service
	}

	ActivityArgs struct {
		UnitService     Service          `validate:"required"`
		AccountsService accounts.Service `validate:"required"`
		IdentityService identity.Service `validate:"required"`
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

	// make sure identity exists
	_, err := a.identityService.Get(ctx, args.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	application, err := a.unitService.CreateApplication(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return application, nil
}

func (a *Activity) UnitCreateAccount(ctx context.Context, identityID string) (string, error) {
	identity, err := a.identityService.Get(ctx, identityID)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	acc, err := a.accountsService.Create(ctx, &accounts.CreateAccountArgs{
		IdentityID: identity.ID,
		Country:    identity.Country,
	})
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	return acc.ID, nil
}

type UnitMapCustomerToAccountArgs struct {
	CustomerID string `validate:"required"`
	Type       string `validate:"required"`
	AccountID  string `validate:"required"`
}

func (a *Activity) UnitMapCustomerToAccount(ctx context.Context, args *UnitMapCustomerToAccountArgs) error {
	if err := a.validator.Struct(args); err != nil {
		return fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	// make sure account exists
	_, err := a.accountsService.Get(ctx, args.AccountID)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	_, err = a.unitService.CreateCustomer(ctx, &CreateCustomerArgs{
		ID:        args.CustomerID,
		AccountID: args.AccountID,
		Type:      args.Type,
	})
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	return nil
}
