package grpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/twilio"
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

type validateSendPhoneVerification struct {
	To           string `validate:"required,e164"`
	OnboardingId string `validate:"required"`
}

func validateSendPhoneVerificationDescription(err validator.FieldError) string {
	switch err.Tag() {
	case "e164":
		return "Phone number is invalid."
	case "required":
		return "Required."
	}
	return ""
}

func (s *rpcService) SendPhoneVerification(
	ctx context.Context,
	req *backendv1.SendPhoneVerificationRequest,
) (*backendv1.PhoneVerificationResponse, error) {
	if err := s.validator.Struct(&validateSendPhoneVerification{
		To:           req.GetTo(),
		OnboardingId: req.GetOnboardingId(),
	}); err != nil {
		return nil, ValidationError(err, validateSendPhoneVerificationDescription)
	}

	_, err := s.twilioService.SendVerificationCode(ctx, req.GetTo())
	if err != nil {
		return nil, InternalError(err)
	}

	_, err = s.onboardingService.UpdateOnboarding(ctx, &onboarding.UpdateOnboardingArgs{
		Id:    req.GetOnboardingId(),
		Phone: req.GetTo(),
	})
	if err != nil {
		if errors.Is(err, onboarding.ErrNotFound) {
			return nil, NotFoundError(err)
		}
		return nil, InternalError(err)
	}

	return &backendv1.PhoneVerificationResponse{
		Status: "pending",
	}, nil
}

type validateCheckPhoneVerificationCode struct {
	To           string `validate:"required,e164"`
	Code         string `validate:"required"`
	OnboardingId string `validate:"required"`
}

func validateCheckPhoneVerificationCodeDescription(err validator.FieldError) string {
	switch err.Tag() {
	case "e164":
		return "Phone number is invalid."
	case "required":
		return "Required."
	}
	return ""
}

func (s *rpcService) CheckPhoneVerificationCode(
	ctx context.Context,
	req *backendv1.CheckPhoneVerificationCodeRequest,
) (*backendv1.PhoneVerificationResponse, error) {
	if err := s.validator.Struct(&validateCheckPhoneVerificationCode{
		To:           req.GetTo(),
		Code:         req.GetCode(),
		OnboardingId: req.GetOnboardingId(),
	}); err != nil {
		return nil, ValidationError(err, validateCheckPhoneVerificationCodeDescription)
	}

	verification, err := s.twilioService.CheckVerificationCode(ctx, &twilio.CheckVerificationCodeArgs{
		PhoneNumber: req.GetTo(),
		Code:        req.GetCode(),
	})
	if err != nil {
		return nil, InternalError(err)
	}
	if verification.Status != "approved" {
		return nil, InternalError(fmt.Errorf(" verification code status is not approved: %s", verification.Status))
	}

	// If successful set phoneVerified in onboarding table
	_, err = s.onboardingService.UpdateOnboarding(ctx, &onboarding.UpdateOnboardingArgs{
		Id:            req.GetOnboardingId(),
		PhoneVerified: true,
	})
	if err != nil {
		if errors.Is(err, onboarding.ErrNotFound) {
			return nil, NotFoundError(err)
		}
		return nil, InternalError(err)
	}

	return &backendv1.PhoneVerificationResponse{
		Status: "approved",
	}, nil
}
