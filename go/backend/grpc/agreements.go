package grpc

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/agreements"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

type validateGetAgreement struct {
	ID string `validate:"required"`
}

func validateGetAgreementDescription(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "Agreement id is required."
	}
	return ""
}

func (s *rpcService) GetAgreement(
	ctx context.Context,
	req *backendv1.GetAgreementRequest,
) (*backendv1.Agreement, error) {
	if err := s.validator.Struct(&validateGetAgreement{
		ID: req.GetId(),
	}); err != nil {
		return nil, ValidationError(err, validateGetAgreementDescription)
	}

	agreement, err := s.agreementsService.Get(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, agreements.ErrNotFound) {
			return nil, NotFoundError("Failed to find agreement.")
		}
		return nil, InternalError("Get agreement.")
	}

	return &backendv1.Agreement{
		Content: agreement.Content,
	}, nil
}
