package onboarding

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/providers/unit"
)

type (
	Activity struct {
		validator *validator.Validate
		up        unit.Service
	}

	ActivityArgs struct {
		Up unit.Service `validate:"required"`
	}
)

func NewActivity(args *ActivityArgs) (*Activity, error) {
	v := validator.New()
	if err := v.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	return &Activity{v, args.Up}, nil
}

func (a *Activity) CreateAccount() error {
	return nil
}

func (a *Activity) MapCustomerToAccount() error {
	return nil
}
