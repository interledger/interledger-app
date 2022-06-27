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
		validator *validator.Validate
		up        Service
		as        accounts.Service
		is        identity.Service
	}

	ActivityArgs struct {
		Up Service          `validate:"required"`
		As accounts.Service `validate:"required"`
		Is identity.Service `validate:"required"`
	}
)

func NewActivity(args *ActivityArgs) (*Activity, error) {
	v := validator.New()
	if err := v.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	return &Activity{
		validator: v,
		up:        args.Up,
		as:        args.As,
		is:        args.Is,
	}, nil
}

func (a *Activity) CreateAccount(ctx context.Context, identityID string, country string) (string, error) {
	// make sure identity exists
	_, err := a.is.Get(ctx, identityID)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	acc, err := a.as.Create(ctx, &accounts.CreateAccountArgs{
		IdentityID: identityID,
		Country:    country,
	})
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	return acc.ID, nil
}

type MapCustomerToAccountArgs struct {
	CustomerID string `validate:"required"`
	Type       string `validate:"required"`
	AccountID  string `validate:"required"`
}

func (a *Activity) MapCustomerToAccount(ctx context.Context, args *MapCustomerToAccountArgs) error {
	if err := a.validator.Struct(args); err != nil {
		return fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	// make sure account exists
	_, err := a.as.Get(ctx, args.AccountID)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	_, err = a.up.CreateCustomer(ctx, &CreateCustomerArgs{
		ID:        args.CustomerID,
		AccountID: args.AccountID,
		Type:      args.Type,
	})
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	return nil
}
