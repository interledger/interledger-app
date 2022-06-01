package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/onboarding"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *rpcService) GetOnboarding(
	ctx context.Context,
	req *backendv1.GetOnboardingRequest,
) (*backendv1.Onboarding, error) {

	onboarding, err := s.onboardingService.GetOnboarding(ctx, &onboarding.GetOnboardingArgs{
		Id: req.Id,
	})
	if err != nil {
		return nil, status.Error(codes.NotFound, "Not found.")
	}

	return &backendv1.Onboarding{
		Id:                 onboarding.ID,
		FirstName:          onboarding.FirstName,
		LastName:           onboarding.LastName,
		CountryOfResidence: onboarding.Country,
		Email:              onboarding.Email,
		Phone:              onboarding.Phone,
		PhoneVerified:      onboarding.PhoneVerified,
		ServiceAgreement:   onboarding.ServiceAgreement,
	}, nil
}

// rpc UpdateOnboarding (Onboarding) returns (Onboarding);
func (s *rpcService) UpdateOnboarding(
	ctx context.Context,
	req *backendv1.Onboarding,
) (*backendv1.Onboarding, error) {

	onboard, err := s.onboardingService.UpdateOnboarding(ctx, &onboarding.UpdateOnboardingArgs{
		Id:               req.Id,
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		Country:          req.CountryOfResidence,
		Email:            req.Email,
		Phone:            req.Phone,
		PhoneVerified:    req.PhoneVerified,
		ServiceAgreement: req.ServiceAgreement,
	})
	if err != nil {
		if err == onboarding.ErrNotFound {
			return nil, status.Error(codes.NotFound, "Not found.")
		}
		return nil, status.Error(codes.Internal, "Internal server error.")
	}

	return &backendv1.Onboarding{
		Id:                 onboard.ID,
		FirstName:          onboard.FirstName,
		LastName:           onboard.LastName,
		CountryOfResidence: onboard.Country,
		Email:              onboard.Email,
		Phone:              onboard.Phone,
		PhoneVerified:      onboard.PhoneVerified,
		ServiceAgreement:   onboard.ServiceAgreement,
	}, nil
}
