package grpc

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/healthcheck"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/providers/unit"
	"gitlab.com/fynbos/backend/user"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func TestGetOnboarding(s *testing.T) {
	ctrl := gomock.NewController(s)
	health, err := healthcheck.NewService()
	if err != nil {
		s.Fatal(err)
	}

	mockOnboardingService := onboarding.NewMockService(ctrl)

	_, _, client := startTestServer(s, &ServerArgs{
		HealthCheckService: health,
		IdentityService:    identity.NewMockService(ctrl),
		AccountsService:    accounts.NewMockService(ctrl),
		AdminAuthService:   auth.NewMockService(),
		UserService:        user.NewMockService(),
		UnitProvider:       unit.NewMockService(ctrl),
		OnboardingService:  mockOnboardingService,
	})

	s.Run("Successfully calls GetOnboarding", func(t *testing.T) {
		ID, name, surname, country, email := uuid.NewString(), faker.FirstName(), faker.LastName(), "US", faker.Email()
		mockOnboardingService.EXPECT().GetOnboarding(gomock.Any(), &onboarding.GetOnboardingArgs{
			Id: ID,
		}).Return(&onboarding.Onboarding{
			ID:               ID,
			FirstName:        name,
			LastName:         surname,
			Country:          country,
			Email:            email,
			Phone:            "",
			PhoneVerified:    false,
			ServiceAgreement: false,
		}, nil).Times(1)

		resp, err := client.GetOnboarding(
			context.Background(),
			&backendv1.GetOnboardingRequest{
				Id: ID,
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, ID, resp.GetId())
		assert.Equal(t, name, resp.GetFirstName())
		assert.False(t, resp.GetPhoneVerified())
	})

	s.Run("Successfully handles errors from GetOnboarding", func(t *testing.T) {
		ID := uuid.NewString()
		mockOnboardingService.EXPECT().GetOnboarding(gomock.Any(), &onboarding.GetOnboardingArgs{
			Id: ID,
		}).Return(nil, onboarding.ErrNotFound).Times(1)

		resp, err := client.GetOnboarding(
			context.Background(),
			&backendv1.GetOnboardingRequest{
				Id: ID,
			},
		)
		if err == nil {
			t.Fatal(err)
		}

		assert.Error(t, err)
		assert.EqualError(t, err, "rpc error: code = NotFound desc = Not found.onboarding: not found.")
		assert.Nil(t, resp)
	})
}

func TestUpdateOnboarding(s *testing.T) {
	ctrl := gomock.NewController(s)
	health, err := healthcheck.NewService()
	if err != nil {
		s.Fatal(err)
	}

	mockOnboardingService := onboarding.NewMockService(ctrl)

	_, _, client := startTestServer(s, &ServerArgs{
		HealthCheckService: health,
		IdentityService:    identity.NewMockService(ctrl),
		AccountsService:    accounts.NewMockService(ctrl),
		AdminAuthService:   auth.NewMockService(),
		UserService:        user.NewMockService(),
		UnitProvider:       unit.NewMockService(ctrl),
		OnboardingService:  mockOnboardingService,
	})

	s.Run("Successfully calls UpdateOnboarding", func(t *testing.T) {
		ID, name, surname, country, email := uuid.NewString(), faker.FirstName(), faker.LastName(), "US", faker.Email()
		mockOnboardingService.EXPECT().UpdateOnboarding(gomock.Any(), &onboarding.UpdateOnboardingArgs{}).Return(&onboarding.Onboarding{
			ID:               ID,
			FirstName:        name,
			LastName:         surname,
			Country:          country,
			Email:            email,
			Phone:            "",
			PhoneVerified:    false,
			ServiceAgreement: false,
		}, nil).Times(1)

		resp, err := client.UpdateOnboarding(
			context.Background(),
			&backendv1.Onboarding{},
		)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, ID, resp.GetId())
		assert.Equal(t, resp.GetFirstName(), name)
		assert.False(t, resp.GetPhoneVerified())
	})

	s.Run("Successfully handles errors from UpdateOnboarding", func(t *testing.T) {
		ID := uuid.NewString()
		mockOnboardingService.EXPECT().UpdateOnboarding(gomock.Any(), &onboarding.UpdateOnboardingArgs{
			Id: ID,
		}).Return(nil, onboarding.ErrInternal).Times(1)

		resp, err := client.UpdateOnboarding(
			context.Background(),
			&backendv1.Onboarding{
				Id: &ID,
			},
		)
		if err == nil {
			t.Fatal(err)
		}

		assert.Error(t, err)
		assert.EqualError(t, err, "rpc error: code = Internal desc = Internal server error.onboarding: internal error.")
		assert.Nil(t, resp)
	})
}

func TestSendPhoneVerification(s *testing.T) {
	ctrl := gomock.NewController(s)
	health, err := healthcheck.NewService()
	if err != nil {
		s.Fatal(err)
	}

	mockOnboardingService := onboarding.NewMockService(ctrl)

	_, _, client := startTestServer(s, &ServerArgs{
		HealthCheckService: health,
		IdentityService:    identity.NewMockService(ctrl),
		AccountsService:    accounts.NewMockService(ctrl),
		AdminAuthService:   auth.NewMockService(),
		UserService:        user.NewMockService(),
		UnitProvider:       unit.NewMockService(ctrl),
		OnboardingService:  mockOnboardingService,
	})

	s.Run("Can send a phone verification token", func(t *testing.T) {
		ID, phone := uuid.NewString(), faker.E164PhoneNumber()
		mockOnboardingService.EXPECT().UpdateOnboarding(gomock.Any(), &onboarding.UpdateOnboardingArgs{
			Id:    ID,
			Phone: phone,
		}).Return(&onboarding.Onboarding{
			ID:               ID,
			FirstName:        "",
			LastName:         "",
			Country:          "",
			Email:            "",
			Phone:            phone,
			PhoneVerified:    false,
			ServiceAgreement: false,
		}, nil).Times(1)
		resp, err := client.SendPhoneVerification(
			context.Background(),
			&backendv1.SendPhoneVerificationRequest{
				To:           phone,
				OnboardingId: ID,
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "pending", resp.GetStatus())
	})

	s.Run("Successfully validates input", func(t *testing.T) {
		ID, phone := uuid.NewString(), faker.Phonenumber()
		mockOnboardingService.EXPECT().UpdateOnboarding(gomock.Any(), &onboarding.UpdateOnboardingArgs{
			Id:            ID,
			PhoneVerified: true,
		}).Return(&onboarding.Onboarding{
			ID:               ID,
			FirstName:        "",
			LastName:         "",
			Country:          "",
			Email:            "",
			Phone:            phone,
			PhoneVerified:    false,
			ServiceAgreement: false,
		}, nil).Times(0)
		resp, err := client.SendPhoneVerification(
			context.Background(),
			&backendv1.SendPhoneVerificationRequest{
				To:           phone,
				OnboardingId: ID,
			},
		)
		if err == nil {
			t.Fatal(err)
		}
		assert.Error(t, err)
		// TODO: Figure out if there's a way to see the entire error validation code from the client.
		assert.EqualError(t, err, "rpc error: code = InvalidArgument desc = Some fields are incorrect.")
		assert.Nil(t, resp)
	})

}

func TestCheckPhoneVerificationCode(s *testing.T) {
	ctrl := gomock.NewController(s)
	health, err := healthcheck.NewService()
	if err != nil {
		s.Fatal(err)
	}

	mockOnboardingService := onboarding.NewMockService(ctrl)

	_, _, client := startTestServer(s, &ServerArgs{
		HealthCheckService: health,
		IdentityService:    identity.NewMockService(ctrl),
		AccountsService:    accounts.NewMockService(ctrl),
		AdminAuthService:   auth.NewMockService(),
		UserService:        user.NewMockService(),
		UnitProvider:       unit.NewMockService(ctrl),
		OnboardingService:  mockOnboardingService,
	})

	s.Run("Can check the verification token", func(t *testing.T) {
		ID, phone, code := uuid.NewString(), faker.E164PhoneNumber(), "948372"
		mockOnboardingService.EXPECT().UpdateOnboarding(gomock.Any(), &onboarding.UpdateOnboardingArgs{
			Id:            ID,
			PhoneVerified: true,
		}).Return(&onboarding.Onboarding{
			ID:               ID,
			FirstName:        "",
			LastName:         "",
			Country:          "",
			Email:            "",
			Phone:            phone,
			PhoneVerified:    false,
			ServiceAgreement: false,
		}, nil).Times(1)
		resp, err := client.CheckPhoneVerificationCode(
			context.Background(),
			&backendv1.CheckPhoneVerificationCodeRequest{
				To:           phone,
				Code:         code,
				OnboardingId: ID,
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "approved", resp.GetStatus())
	})

	s.Run("Successfully validates input", func(t *testing.T) {
		ID, phone := uuid.NewString(), faker.Phonenumber()
		mockOnboardingService.EXPECT().UpdateOnboarding(gomock.Any(), &onboarding.UpdateOnboardingArgs{
			Id:            ID,
			PhoneVerified: true,
		}).Return(&onboarding.Onboarding{
			ID:               ID,
			FirstName:        "",
			LastName:         "",
			Country:          "",
			Email:            "",
			Phone:            phone,
			PhoneVerified:    false,
			ServiceAgreement: false,
		}, nil).Times(0)
		resp, err := client.CheckPhoneVerificationCode(
			context.Background(),
			&backendv1.CheckPhoneVerificationCodeRequest{
				To:           phone,
				OnboardingId: ID,
			},
		)
		if err == nil {
			t.Fatal(err)
		}
		assert.Error(t, err)
		// TODO: Figure out if there's a way to see the entire error validation code from the client.
		assert.EqualError(t, err, "rpc error: code = InvalidArgument desc = Some fields are incorrect.")
		assert.Nil(t, resp)
	})

}
