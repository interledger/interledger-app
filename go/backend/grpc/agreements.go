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
	if err := s.b.Validator().Struct(&validateGetAgreement{
		ID: req.GetId(),
	}); err != nil {
		return nil, ValidationError(err, validateGetAgreementDescription)
	}

	agreement, err := s.b.Agreements().Get(ctx, req.GetId())
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

type validateSignAgreements struct {
	AgreementIDs []string `validate:"required,min=1"`
	IdentityID   string   `validate:"required,uuid"`
	IPAddress    string   `validate:"required,ip_addr"`
}

func validateSignAgreementsDescription(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "Missing required fields."
	case "min":
		return "You have to provide at least one agreement."
	case "uuid":
		return "Identity id must be a valid uuid."
	case "ip_addr":
		return "IP address must be a valid ip address."
	}
	return ""
}

func (s *rpcService) SignAgreements(
	ctx context.Context,
	req *backendv1.SignAgreementsRequest,
) (*backendv1.SignAgreementsResponse, error) {
	if err := s.b.Validator().Struct(&validateSignAgreements{
		AgreementIDs: req.GetAgreementIds(),
		IdentityID:   req.GetIdentityId(),
		IPAddress:    req.GetIpAddress(),
	}); err != nil {
		return nil, ValidationError(err, validateSignAgreementsDescription)
	}

	err := s.b.Agreements().Sign(ctx, &agreements.SignArgs{
		AgreementIDs: req.GetAgreementIds(),
		IdentityID:   req.GetIdentityId(),
		IPAddress:    req.GetIpAddress(),
	})
	if err != nil {
		if errors.Is(err, agreements.ErrNotFound) {
			return nil, NotFoundError("Agreement not found.")
		}
		return nil, InternalError("Sign agreements.")
	}

	return &backendv1.SignAgreementsResponse{
		Signed: true,
	}, nil
}
