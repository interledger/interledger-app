package grpc

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/twilio"
	_user "gitlab.com/fynbos/backend/user"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func TestGetOnboarding(s *testing.T) {
	ctrl := gomock.NewController(s)
	c := NewTestContainer(s, ctrl)
	_, _, client := startTestServer(s, c)

	s.Run("Successfully calls GetOnboarding", func(t *testing.T) {
		ID, name, surname, country, email := uuid.NewString(), faker.FirstName(), faker.LastName(), "US", faker.Email()
		c.OnboardingService.EXPECT().GetOnboarding(gomock.Any(), &onboarding.GetOnboardingArgs{
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
		c.OnboardingService.EXPECT().GetOnboarding(gomock.Any(), &onboarding.GetOnboardingArgs{
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
		assert.EqualError(t, err, "rpc error: code = NotFound desc = Not found: Failed to find onboarding.")
		assert.Nil(t, resp)
	})
}

func TestUpdateOnboarding(s *testing.T) {
	ctrl := gomock.NewController(s)
	c := NewTestContainer(s, ctrl)
	_, _, client := startTestServer(s, c)

	s.Run("Successfully calls UpdateOnboarding", func(t *testing.T) {
		ID, name, surname, country, email := uuid.NewString(), faker.FirstName(), faker.LastName(), "US", faker.Email()
		c.OnboardingService.EXPECT().UpdateOnboarding(gomock.Any(), &onboarding.UpdateOnboardingArgs{}).Return(&onboarding.Onboarding{
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
		c.OnboardingService.EXPECT().UpdateOnboarding(gomock.Any(), &onboarding.UpdateOnboardingArgs{
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
		assert.EqualError(t, err, "rpc error: code = Internal desc = Internal server error: Update onboarding.")
		assert.Nil(t, resp)
	})
}

func TestCreateIdentity(s *testing.T) {
	ctrl := gomock.NewController(s)
	c := NewTestContainer(s, ctrl)
	_, _, client := startTestServer(s, c)
	userId := uuid.NewString()

	s.Run("Creates identity from onboarding", func(t *testing.T) {
		ID, name, surname, country, email, phone := uuid.NewString(), faker.FirstName(), faker.LastName(), "US", faker.Email(), faker.E164PhoneNumber()
		c.OnboardingService.EXPECT().GetOnboarding(gomock.Any(), &onboarding.GetOnboardingArgs{
			Id: ID,
		}).Return(&onboarding.Onboarding{
			ID:               ID,
			FirstName:        name,
			LastName:         surname,
			Country:          country,
			Email:            email,
			Phone:            phone,
			PhoneVerified:    true,
			ServiceAgreement: true,
		}, nil).Times(1)

		// This error is swallowed if there is no identity, so we don't need to worry about it here.
		c.IdentityService.EXPECT().Get(gomock.Any(), userId).Return(nil, nil).Times(1)

		c.IdentityService.EXPECT().Create(gomock.Any(), &identity.CreateArgs{
			ID:           userId,
			FirstName:    name,
			LastName:     surname,
			Country:      country,
			Email:        email,
			MobileNumber: phone,
		}).Return(&identity.Identity{
			ID:           userId,
			FirstName:    name,
			LastName:     surname,
			Country:      country,
			Email:        email,
			MobileNumber: phone,
		}, nil).Times(1)

		resp, err := client.CreateIdentity(
			_user.ActingAsContext(t, context.Background(), &_user.User{
				ID:    userId,
				Email: email,
			}),
			&backendv1.CreateIdentityRequest{
				OnboardingId: ID,
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, userId, resp.GetIdentityId())
	})

	s.Run("Already has identity, so just returns it", func(t *testing.T) {
		ID, name, surname, country, email, phone := uuid.NewString(), faker.FirstName(), faker.LastName(), "US", faker.Email(), faker.E164PhoneNumber()
		c.OnboardingService.EXPECT().GetOnboarding(gomock.Any(), &onboarding.GetOnboardingArgs{
			Id: ID,
		}).Return(&onboarding.Onboarding{
			ID:               ID,
			FirstName:        name,
			LastName:         surname,
			Country:          country,
			Email:            email,
			Phone:            phone,
			PhoneVerified:    true,
			ServiceAgreement: true,
		}, nil).Times(1)

		c.IdentityService.EXPECT().Get(gomock.Any(), userId).Return(&identity.Identity{
			ID: userId,
		}, nil).Times(1)

		// Shouldn't need to call this if an identity exists for the user.
		c.IdentityService.EXPECT().Create(gomock.Any(), nil).Return(nil, nil).Times(0)

		resp, err := client.CreateIdentity(
			_user.ActingAsContext(t, context.Background(), &_user.User{
				ID:    userId,
				Email: email,
			}),
			&backendv1.CreateIdentityRequest{
				OnboardingId: ID,
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, userId, resp.GetIdentityId())
	})
}

func TestSendPhoneVerification(s *testing.T) {
	ctrl := gomock.NewController(s)
	c := NewTestContainer(s, ctrl)
	_, _, client := startTestServer(s, c)

	s.Run("Can send a phone verification token", func(t *testing.T) {
		ID, phone := uuid.NewString(), faker.E164PhoneNumber()
		c.OnboardingService.EXPECT().UpdateOnboarding(gomock.Any(), &onboarding.UpdateOnboardingArgs{
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
		c.TwilioService.EXPECT().SendVerificationCode(gomock.Any(), phone).Return(&twilio.Verification{
			Status:      "pending",
			PhoneNumber: phone,
			Sid:         "",
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
		c.OnboardingService.EXPECT().UpdateOnboarding(gomock.Any(), &onboarding.UpdateOnboardingArgs{
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
	c := NewTestContainer(s, ctrl)
	_, _, client := startTestServer(s, c)

	s.Run("Can check the verification token", func(t *testing.T) {
		ID, phone, code := uuid.NewString(), faker.E164PhoneNumber(), "948372"
		c.OnboardingService.EXPECT().UpdateOnboarding(gomock.Any(), &onboarding.UpdateOnboardingArgs{
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
		c.TwilioService.EXPECT().CheckVerificationCode(gomock.Any(), &twilio.CheckVerificationCodeArgs{
			PhoneNumber: phone,
			Code:        code,
		}).Return(&twilio.Verification{
			Status:      "approved",
			PhoneNumber: phone,
			Sid:         "",
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

	s.Run("Fails if status is not approved", func(t *testing.T) {
		ID, phone, code := uuid.NewString(), faker.E164PhoneNumber(), "948372"
		c.TwilioService.EXPECT().CheckVerificationCode(gomock.Any(), &twilio.CheckVerificationCodeArgs{
			PhoneNumber: phone,
			Code:        code,
		}).Return(&twilio.Verification{
			Status:      "pending",
			PhoneNumber: phone,
			Sid:         "",
		}, nil).Times(1)
		_, err := client.CheckPhoneVerificationCode(
			context.Background(),
			&backendv1.CheckPhoneVerificationCodeRequest{
				To:           phone,
				Code:         code,
				OnboardingId: ID,
			},
		)

		assert.EqualError(t, err, "rpc error: code = InvalidArgument desc = Some fields are incorrect.")
	})

	s.Run("Successfully validates input", func(t *testing.T) {
		ID, phone := uuid.NewString(), faker.Phonenumber()
		c.OnboardingService.EXPECT().UpdateOnboarding(gomock.Any(), &onboarding.UpdateOnboardingArgs{
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

func TestInitiateUnitOnboarding(s *testing.T) {
	ctrl := gomock.NewController(s)
	c := NewTestContainer(s, ctrl)
	_, _, client := startTestServer(s, c)
	deviceFingerprints := make([]string, 1)
	deviceFingerprints = append(deviceFingerprints, "Some randon fingerprint")
	userId := uuid.NewString()
	Ssn := faker.CCNumber()
	DateOfBirth := faker.Date()
	Street := faker.Name()
	Street2 := faker.Name()
	City := faker.Name()
	State := faker.Name()
	PostalCode := faker.Name()
	IpAddress := faker.IPv4()

	s.Run("Calls the onboarding service to initiate the the onboarding workflow", func(t *testing.T) {
		email := faker.Email()
		args := &onboarding.InitiateUnitCustomerOnboardingArgs{
			IdentityID:         userId,
			Ssn:                Ssn,
			DateOfBirth:        DateOfBirth,
			Street:             Street,
			Street2:            Street2,
			City:               City,
			State:              State,
			PostalCode:         PostalCode,
			IpAddress:          IpAddress,
			DeviceFingerprints: deviceFingerprints,
		}
		c.OnboardingService.EXPECT().InitiateUnitCustomerOnboarding(gomock.Any(), args).Return(nil).Times(1)

		resp, err := client.InitiateUnitOnboarding(
			_user.ActingAsContext(t, context.Background(), &_user.User{
				ID:    userId,
				Email: email,
			}),
			&backendv1.InitiateUnitOnboardingRequest{
				Ssn:                Ssn,
				DateOfBirth:        DateOfBirth,
				Street:             Street,
				Street2:            Street2,
				City:               City,
				State:              State,
				PostalCode:         PostalCode,
				Ip:                 IpAddress,
				DeviceFingerprints: deviceFingerprints,
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, userId, resp.GetIdentityId())
	})
}
