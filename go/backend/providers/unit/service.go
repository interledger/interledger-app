package unit

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

var (
	ErrInternal = errors.New("unit: internal error")
)

type Service interface {
}

type service struct {
	validator *validator.Validate
	token     string
}

type ServiceArgs struct {
	Token string `validate:"required"`
}

func NewService(args ServiceArgs) (Service, error) {
	validator := validator.New()
	err := validator.Struct(args)
	if err != nil {
		return nil, err
	}

	return &service{
		validator: validator,
		token:     args.Token,
	}, nil
}
