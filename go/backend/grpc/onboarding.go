package grpc

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/onboarding"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

type validateGetOnboarding struct {
	Id string `validate:"omitempty,uuid"`
}

func validateGetOnboardingDescription(err validator.FieldError) string {
	switch err.Tag() {
	case "uuid":
		return "uuid is invlaid."
	}
	return ""
}

func (s *rpcService) GetOnboarding(
	ctx context.Context,
	req *backendv1.GetOnboardingRequest,
) (*backendv1.Onboarding, error) {
	if err := s.validator.Struct(&validateGetOnboarding{
		Id: req.GetId(),
	}); err != nil {
		return nil, ValidationError(err, validateGetOnboardingDescription)
	}

	onboard, err := s.onboardingService.GetOnboarding(ctx, &onboarding.GetOnboardingArgs{
		Id: req.Id,
	})
	if err != nil {
		if errors.Is(err, onboarding.ErrNotFound) {
			return nil, NotFoundError(err)
		}
		return nil, InternalError(err)
	}

	return &backendv1.Onboarding{
		Id:                 &onboard.ID,
		FirstName:          &onboard.FirstName,
		LastName:           &onboard.LastName,
		CountryOfResidence: &onboard.Country,
		Email:              &onboard.Email,
		Phone:              &onboard.Phone,
		PhoneVerified:      &onboard.PhoneVerified,
		ServiceAgreement:   &onboard.ServiceAgreement,
	}, nil
}

type validateUpdateOnboarding struct {
	CountryOfResidence string `validate:"omitempty,iso3166_1_alpha2"`
	Email              string `validate:"omitempty,email"`
	Phone              string `validate:"omitempty,e164"`
}

func validateUpdateOnboardingDescription(err validator.FieldError) string {
	switch err.Tag() {
	case "iso3166_1_alpha2":
		return "Country is required."
	case "email":
		return "Please provide a valid email address."
	}
	return ""
}

func (s *rpcService) UpdateOnboarding(
	ctx context.Context,
	req *backendv1.Onboarding,
) (*backendv1.Onboarding, error) {
	if err := s.validator.Struct(&validateUpdateOnboarding{
		CountryOfResidence: req.GetCountryOfResidence(),
		Email:              req.GetEmail(),
		Phone:              req.GetPhone(),
	}); err != nil {
		return nil, ValidationError(err, validateUpdateOnboardingDescription)
	}

	onboard, err := s.onboardingService.UpdateOnboarding(ctx, &onboarding.UpdateOnboardingArgs{
		Id:               req.GetId(),
		FirstName:        req.GetFirstName(),
		LastName:         req.GetLastName(),
		Country:          req.GetCountryOfResidence(),
		Email:            req.GetEmail(),
		Phone:            req.GetPhone(),
		PhoneVerified:    req.GetPhoneVerified(),
		ServiceAgreement: req.GetServiceAgreement(),
	})

	if err != nil {
		if errors.Is(err, onboarding.ErrNotFound) {
			return nil, NotFoundError(err)
		}
		return nil, InternalError(err)
	}

	return &backendv1.Onboarding{
		Id:                 &onboard.ID,
		FirstName:          &onboard.FirstName,
		LastName:           &onboard.LastName,
		CountryOfResidence: &onboard.Country,
		Email:              &onboard.Email,
		Phone:              &onboard.Phone,
		PhoneVerified:      &onboard.PhoneVerified,
		ServiceAgreement:   &onboard.ServiceAgreement,
	}, nil
}
