package grpc

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/identity"
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
	if err := s.b.Validator().Struct(&validateGetOnboarding{
		Id: req.GetId(),
	}); err != nil {
		return nil, ValidationError(err, validateGetOnboardingDescription)
	}

	onboard, err := s.b.Onboarding().GetOnboarding(ctx, &onboarding.GetOnboardingArgs{
		Id: req.Id,
	})
	if err != nil {
		if errors.Is(err, onboarding.ErrNotFound) {
			return nil, NotFoundError("Failed to find onboarding.")
		}
		return nil, InternalError("Get onboarding.")
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
	if err := s.b.Validator().Struct(&validateUpdateOnboarding{
		CountryOfResidence: req.GetCountryOfResidence(),
		Email:              req.GetEmail(),
		Phone:              req.GetPhone(),
	}); err != nil {
		return nil, ValidationError(err, validateUpdateOnboardingDescription)
	}

	onboard, err := s.b.Onboarding().UpdateOnboarding(ctx, &onboarding.UpdateOnboardingArgs{
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
			return nil, NotFoundError("Failed to find onboarding.")
		}
		return nil, InternalError("Update onboarding.")
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

type validateCreateIdentity struct {
	OnboardingId string `validate:"required,uuid"`
}

func validateCreateIdentityDescription(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "Onboarding id is required."
	case "uuid":
		return "Invalid onboarding id."
	}
	return ""
}

func (s *rpcService) CreateIdentity(
	ctx context.Context,
	req *backendv1.CreateIdentityRequest,
) (*backendv1.CreateIdentityResponse, error) {
	if err := s.b.Validator().Struct(&validateCreateIdentity{
		OnboardingId: req.GetOnboardingId(),
	}); err != nil {
		return nil, ValidationError(err, validateCreateIdentityDescription)
	}

	user, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	ob, err := s.b.Onboarding().GetOnboarding(ctx, &onboarding.GetOnboardingArgs{
		Id: req.GetOnboardingId(),
	})
	if err != nil {
		return nil, NotFoundError("failed to find onboarding")
	}

	if !ob.PhoneVerified {
		return nil, ForbiddenError("phone not verified.")
	}

	id, _ := s.b.Identity().Get(ctx, user.ID)
	if id == nil {
		id, err = s.b.Identity().Create(ctx, &identity.CreateArgs{
			ID:           user.ID,
			FirstName:    ob.FirstName,
			LastName:     ob.LastName,
			MobileNumber: ob.Phone,
			Email:        ob.Email,
			Country:      ob.Country,
		})
		if err != nil {
			return nil, InternalError("Failed to create identity.")
		}
	}

	return &backendv1.CreateIdentityResponse{
		IdentityId: id.ID,
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
	if err := s.b.Validator().Struct(&validateSendPhoneVerification{
		To:           req.GetTo(),
		OnboardingId: req.GetOnboardingId(),
	}); err != nil {
		return nil, ValidationError(err, validateSendPhoneVerificationDescription)
	}

	_, err := s.b.Twilio().SendVerificationCode(ctx, req.GetTo())
	if err != nil {
		return nil, InternalError(err.Error())
	}

	_, err = s.b.Onboarding().UpdateOnboarding(ctx, &onboarding.UpdateOnboardingArgs{
		Id:    req.GetOnboardingId(),
		Phone: req.GetTo(),
	})
	if err != nil {
		if errors.Is(err, onboarding.ErrNotFound) {
			return nil, NotFoundError("Failed to find onboarding.")
		}
		return nil, InternalError("Update onboarding.")
	}

	return &backendv1.PhoneVerificationResponse{
		Status: "pending",
	}, nil
}

type validateCheckPhoneVerificationCode struct {
	To           string `validate:"required,e164"`
	Code         string `validate:"required,numeric,len=6"`
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
	if err := s.b.Validator().Struct(&validateCheckPhoneVerificationCode{
		To:           req.GetTo(),
		Code:         req.GetCode(),
		OnboardingId: req.GetOnboardingId(),
	}); err != nil {
		return nil, ValidationError(err, validateCheckPhoneVerificationCodeDescription)
	}

	verification, err := s.b.Twilio().CheckVerificationCode(ctx, &twilio.CheckVerificationCodeArgs{
		PhoneNumber: req.GetTo(),
		Code:        req.GetCode(),
	})
	if err != nil {
		return nil, InternalError(err.Error())
	}
	if verification.Status != "approved" {
		return nil, NewValidationError("Code", "The verification code did not match.")
	}

	// If successful set phoneVerified in onboarding table
	_, err = s.b.Onboarding().UpdateOnboarding(ctx, &onboarding.UpdateOnboardingArgs{
		Id:            req.GetOnboardingId(),
		PhoneVerified: true,
	})
	if err != nil {
		if errors.Is(err, onboarding.ErrNotFound) {
			return nil, NotFoundError("Failed to find onboarding.")
		}
		return nil, InternalError("Update onboarding.")
	}

	return &backendv1.PhoneVerificationResponse{
		Status: "approved",
	}, nil
}
