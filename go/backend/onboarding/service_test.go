package onboarding

import (
	context "context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	accounts_mock "gitlab.com/fynbos/backend/accounts/client/mock"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/noop"
	"gitlab.com/fynbos/backend/providers/unit"
	test_utils "gitlab.com/fynbos/backend/utils"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"
)

func TestInitiatesOnboarding(t *testing.T) {
	ctrl := gomock.NewController(t)
	tp := &mocks.Client{}
	os, err := NewService(&ServiceArgs{
		Db:   &sqlx.DB{},
		As:   accounts_mock.NewMockClient(ctrl),
		Is:   identity.NewMockService(ctrl),
		Noop: noop.NewMockService(ctrl),
		Tp:   tp,
	})
	if err != nil {
		t.Fatal(err)
	}

	identityID := uuid.NewString()
	deviceFingerprints := make([]string, 1)
	deviceFingerprints = append(deviceFingerprints, "Some randon fingerprint")

	args := &InitiateUnitCustomerOnboardingArgs{
		IdentityID:         identityID,
		Ssn:                faker.CCNumber(),
		DateOfBirth:        faker.Date(),
		Street:             faker.Name(),
		Street2:            faker.Name(),
		City:               faker.Name(),
		State:              faker.Name(),
		PostalCode:         faker.Name(),
		IpAddress:          faker.IPv4(),
		DeviceFingerprints: deviceFingerprints,
	}
	tp.On(
		"ExecuteWorkflow",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		unit.UnitOnboardCustomerState{
			CustomerID: "",
			Type:       "",
			IdentityID: args.IdentityID,
			AccountID:  "",
			ApplicationArgs: unit.CreateApplicationArgs{
				Ssn:                args.Ssn,
				DateOfBirth:        args.DateOfBirth,
				Street:             args.Street,
				Street2:            args.Street2,
				City:               args.City,
				State:              args.State,
				PostalCode:         args.PostalCode,
				IpAddress:          args.IpAddress,
				UserID:             args.IdentityID,
				DeviceFingerprints: args.DeviceFingerprints,
			},
		},
	).Return(
		func(ctx context.Context, opts client.StartWorkflowOptions, workflow interface{}, args ...interface{}) client.WorkflowRun {
			testWorkflowID := opts.ID
			testRunID := "test-runid"

			mockWorkflowRun := &mocks.WorkflowRun{}
			mockWorkflowRun.On("GetID").Return(testWorkflowID)
			mockWorkflowRun.On("GetRunID").Return(testRunID)
			mockWorkflowRun.On("Get", mock.Anything, mock.Anything).Return(nil)
			return mockWorkflowRun
		}, nil,
	).Times(1)

	err = os.InitiateUnitCustomerOnboarding(context.Background(), args)

	assert.NoError(t, err)
}

func TestGetOnboarding(s *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(s)
	tp := &mocks.Client{}
	db := test_utils.MigrateCockroachDB(s, ctx)

	onboardingService, err := NewService(&ServiceArgs{
		Db:   db,
		As:   accounts_mock.NewMockClient(ctrl),
		Is:   identity.NewMockService(ctrl),
		Noop: noop.NewMockService(ctrl),
		Tp:   tp,
	})
	if err != nil {
		s.Fatal(err)
	}

	s.Run("fails appropriately if tries to fetch an onboarding if there is none", func(t *testing.T) {
		_, err := onboardingService.GetOnboarding(ctx, &GetOnboardingArgs{
			Id: uuid.NewString(),
		})

		assert.NotNil(t, err)
		assert.Equal(t, err, ErrNotFound)
	})

	s.Run("Finds an initialised, empty onboarding flow", func(t *testing.T) {
		var newOnboarding Onboarding
		err = db.Get(&newOnboarding, `INSERT INTO onboarding (id) VALUES ($1) RETURNING *;
			`, uuid.NewString())
		if err != nil {
			s.Fatal(err)
		}

		onboarding, err := onboardingService.GetOnboarding(ctx, &GetOnboardingArgs{
			Id: newOnboarding.ID,
		})
		if err != nil {
			s.Fatal(err)
		}

		assert.NotNil(t, onboarding)
		assert.Equal(t, newOnboarding.ID, onboarding.ID)
		assert.Equal(t, newOnboarding.FirstName, onboarding.FirstName)
		assert.Equal(t, newOnboarding.LastName, onboarding.LastName)
		assert.Equal(t, newOnboarding.Country, onboarding.Country)
		assert.Equal(t, newOnboarding.Email, onboarding.Email)
		assert.Equal(t, newOnboarding.Phone, onboarding.Phone)
		assert.Equal(t, newOnboarding.PhoneVerified, onboarding.PhoneVerified)
		assert.Equal(t, newOnboarding.ServiceAgreement, onboarding.ServiceAgreement)
	})

	s.Run("Can get half empty onboarding flow", func(t *testing.T) {
		id, name, surname, country, email := uuid.NewString(), faker.FirstName(), faker.LastName(), "US", faker.Email()
		var newOnboarding Onboarding
		err = db.Get(&newOnboarding, `INSERT INTO onboarding (id, first_name, last_name, country_of_residence, email) VALUES ($1, $2, $3, $4, $5) RETURNING *;
			`, id, name, surname, country, email)
		if err != nil {
			s.Fatal(err)
		}

		onboarding, err := onboardingService.GetOnboarding(ctx, &GetOnboardingArgs{
			Id: newOnboarding.ID,
		})
		if err != nil {
			s.Fatal(err)
		}

		assert.NotNil(t, onboarding)
		assert.Equal(t, newOnboarding.ID, onboarding.ID)
		assert.Equal(t, id, onboarding.ID)
		assert.Equal(t, newOnboarding.FirstName, onboarding.FirstName)
		assert.Equal(t, name, onboarding.FirstName)
		assert.Equal(t, newOnboarding.LastName, onboarding.LastName)
		assert.Equal(t, surname, onboarding.LastName)
		assert.Equal(t, newOnboarding.Country, onboarding.Country)
		assert.Equal(t, country, onboarding.Country)
		assert.Equal(t, newOnboarding.Email, onboarding.Email)
		assert.Equal(t, email, onboarding.Email)
		assert.Equal(t, newOnboarding.Phone, onboarding.Phone)
		assert.Equal(t, onboarding.Phone, "")
		assert.Equal(t, newOnboarding.PhoneVerified, onboarding.PhoneVerified)
		assert.Equal(t, newOnboarding.ServiceAgreement, onboarding.ServiceAgreement)
	})
}

func TestUpdateOnboarding(s *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(s)
	tp := &mocks.Client{}
	db := test_utils.MigrateCockroachDB(s, ctx)

	onboardingService, err := NewService(&ServiceArgs{
		Db:   db,
		As:   accounts_mock.NewMockClient(ctrl),
		Is:   identity.NewMockService(ctrl),
		Noop: noop.NewMockService(ctrl),
		Tp:   tp,
	})
	if err != nil {
		s.Fatal(err)
	}

	s.Run("can set new onboarding with empty args", func(t *testing.T) {
		onboarding, err := onboardingService.UpdateOnboarding(ctx, &UpdateOnboardingArgs{})
		if err != nil {
			s.Fatal(err)
		}

		newOnboarding, err := onboardingService.GetOnboarding(ctx, &GetOnboardingArgs{
			Id: onboarding.ID,
		})
		if err != nil {
			s.Fatal(err)
		}
		_, err = uuid.Parse(onboarding.ID)
		assert.Nil(t, err)
		assert.NotNil(t, onboarding)
		assert.Equal(t, newOnboarding.ID, onboarding.ID)
		assert.Equal(t, "", onboarding.FirstName)
		assert.Equal(t, "", onboarding.LastName)
		assert.Equal(t, "", onboarding.Country)
		assert.Equal(t, "", onboarding.Email)
		assert.Equal(t, "", onboarding.Phone)
		assert.Equal(t, false, onboarding.PhoneVerified)
		assert.Equal(t, false, onboarding.ServiceAgreement)
	})

	s.Run("can set new onboarding with half empty args", func(t *testing.T) {
		onboarding, err := onboardingService.UpdateOnboarding(ctx, &UpdateOnboardingArgs{
			FirstName: faker.FirstName(),
			LastName:  faker.LastName(),
			Country:   "US",
			Email:     faker.Email(),
		})
		if err != nil {
			s.Fatal(err)
		}

		newOnboarding, err := onboardingService.GetOnboarding(ctx, &GetOnboardingArgs{
			Id: onboarding.ID,
		})
		if err != nil {
			s.Fatal(err)
		}

		assert.NotNil(t, onboarding)
		assert.Equal(t, newOnboarding.ID, onboarding.ID)
		assert.Equal(t, newOnboarding.FirstName, onboarding.FirstName)
		assert.Equal(t, newOnboarding.LastName, onboarding.LastName)
		assert.Equal(t, newOnboarding.Country, onboarding.Country)
		assert.Equal(t, newOnboarding.Email, onboarding.Email)
		assert.Equal(t, newOnboarding.Phone, onboarding.Phone)
		assert.Equal(t, newOnboarding.PhoneVerified, onboarding.PhoneVerified)
		assert.Equal(t, newOnboarding.ServiceAgreement, onboarding.ServiceAgreement)
	})

	s.Run("can update existing onboarding, can replace and add new fields", func(t *testing.T) {
		onboarding, err := onboardingService.UpdateOnboarding(ctx, &UpdateOnboardingArgs{
			FirstName: faker.FirstName(),
			LastName:  faker.LastName(),
			Country:   "US",
			Email:     faker.Email(),
		})
		if err != nil {
			s.Fatal(err)
		}

		newOnboarding, err := onboardingService.UpdateOnboarding(ctx, &UpdateOnboardingArgs{
			Id:    onboarding.ID,
			Email: faker.Email(),
			Phone: faker.E164PhoneNumber(),
		})
		if err != nil {
			s.Fatal(err)
		}

		assert.NotNil(t, onboarding)
		assert.Equal(t, newOnboarding.ID, onboarding.ID)
		assert.Equal(t, newOnboarding.FirstName, onboarding.FirstName)
		assert.Equal(t, newOnboarding.LastName, onboarding.LastName)
		assert.Equal(t, newOnboarding.Country, onboarding.Country)
		assert.NotEqual(t, newOnboarding.Email, onboarding.Email)
		assert.NotEqual(t, onboarding.Email, "")
		assert.NotEqual(t, newOnboarding.Email, "")
		assert.Equal(t, onboarding.Phone, "")
		assert.NotEqual(t, newOnboarding.Phone, onboarding.Phone)
		assert.NotEqual(t, newOnboarding.Phone, "")
		assert.Equal(t, newOnboarding.PhoneVerified, onboarding.PhoneVerified)
		assert.Equal(t, newOnboarding.ServiceAgreement, onboarding.ServiceAgreement)
	})

	s.Run("can set phone number to verified", func(t *testing.T) {
		onboarding, err := onboardingService.UpdateOnboarding(ctx, &UpdateOnboardingArgs{
			FirstName: faker.FirstName(),
			LastName:  faker.LastName(),
			Country:   "US",
			Email:     faker.Email(),
			Phone:     faker.E164PhoneNumber(),
		})
		if err != nil {
			s.Fatal(err)
		}
		assert.False(t, onboarding.PhoneVerified)

		newOnboarding, err := onboardingService.UpdateOnboarding(ctx, &UpdateOnboardingArgs{
			Id:            onboarding.ID,
			PhoneVerified: true,
		})
		if err != nil {
			s.Fatal(err)
		}

		assert.True(t, newOnboarding.PhoneVerified)
	})
}
