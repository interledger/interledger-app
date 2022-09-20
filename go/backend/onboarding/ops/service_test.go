package ops_test

import (
	context "context"
	"testing"

	"gitlab.com/fynbos/backend/onboarding/ops"

	"gitlab.com/fynbos/backend/onboarding"

	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/identity"

	"github.com/bxcodec/faker/v3"
	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	accounts_mock "gitlab.com/fynbos/backend/accounts/client/mock"
	identity_mock "gitlab.com/fynbos/backend/identity/client/mock"
	test_utils "gitlab.com/fynbos/backend/utils"
	temporal "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"
)

type backends struct {
	validator *validator.Validate
	db        *sqlx.DB
	accounts  accounts.Client
	identity  identity.Client
	temporal  temporal.Client
}

func (b backends) Validator() *validator.Validate {
	return b.validator
}

func (b backends) DB() *sqlx.DB {
	return b.db
}

func (b backends) Accounts() accounts.Client {
	return b.accounts
}

func (b backends) Identity() identity.Client {
	return b.identity
}

func (b backends) Temporal() temporal.Client {
	return b.temporal
}

func TestGetOnboarding(s *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(s)
	tp := &mocks.Client{}
	db := test_utils.MigrateCockroachDB(s, ctx)

	b := &backends{
		validator: validator.New(),
		db:        db,
		accounts:  accounts_mock.NewMockClient(ctrl),
		identity:  identity_mock.NewMockClient(ctrl),
		temporal:  tp,
	}

	s.Run("fails appropriately if tries to fetch an onboarding if there is none", func(t *testing.T) {
		_, err := ops.GetOnboarding(ctx, b, &onboarding.GetOnboardingArgs{
			Id: uuid.NewString(),
		})

		assert.NotNil(t, err)
		assert.Equal(t, err, onboarding.ErrNotFound)
	})

	s.Run("Finds an initialised, empty onboarding flow", func(t *testing.T) {
		var newOnboarding onboarding.Onboarding
		err := db.Get(&newOnboarding, `INSERT INTO onboarding (id) VALUES ($1) RETURNING *;
			`, uuid.NewString())
		if err != nil {
			s.Fatal(err)
		}

		onboarding, err := ops.GetOnboarding(ctx, b, &onboarding.GetOnboardingArgs{
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
		var newOnboarding onboarding.Onboarding
		err := db.Get(&newOnboarding, `INSERT INTO onboarding (id, first_name, last_name, country_of_residence, email) VALUES ($1, $2, $3, $4, $5) RETURNING *;
			`, id, name, surname, country, email)
		if err != nil {
			s.Fatal(err)
		}

		onboarding, err := ops.GetOnboarding(ctx, b, &onboarding.GetOnboardingArgs{
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

	b := &backends{
		validator: validator.New(),
		db:        db,
		accounts:  accounts_mock.NewMockClient(ctrl),
		identity:  identity_mock.NewMockClient(ctrl),
		temporal:  tp,
	}

	s.Run("can set new onboarding with empty args", func(t *testing.T) {
		ob, err := ops.UpdateOnboarding(ctx, b, &onboarding.UpdateOnboardingArgs{})
		if err != nil {
			s.Fatal(err)
		}

		newOnboarding, err := ops.GetOnboarding(ctx, b, &onboarding.GetOnboardingArgs{
			Id: ob.ID,
		})
		if err != nil {
			s.Fatal(err)
		}
		_, err = uuid.Parse(ob.ID)
		assert.Nil(t, err)
		assert.NotNil(t, ob)
		assert.Equal(t, newOnboarding.ID, ob.ID)
		assert.Equal(t, "", ob.FirstName)
		assert.Equal(t, "", ob.LastName)
		assert.Equal(t, "", ob.Country)
		assert.Equal(t, "", ob.Email)
		assert.Equal(t, "", ob.Phone)
		assert.Equal(t, false, ob.PhoneVerified)
		assert.Equal(t, false, ob.ServiceAgreement)
	})

	s.Run("can set new onboarding with half empty args", func(t *testing.T) {
		ob, err := ops.UpdateOnboarding(ctx, b, &onboarding.UpdateOnboardingArgs{
			FirstName: faker.FirstName(),
			LastName:  faker.LastName(),
			Country:   "US",
			Email:     faker.Email(),
		})
		if err != nil {
			s.Fatal(err)
		}

		newOnboarding, err := ops.GetOnboarding(ctx, b, &onboarding.GetOnboardingArgs{
			Id: ob.ID,
		})
		if err != nil {
			s.Fatal(err)
		}

		assert.NotNil(t, ob)
		assert.Equal(t, newOnboarding.ID, ob.ID)
		assert.Equal(t, newOnboarding.FirstName, ob.FirstName)
		assert.Equal(t, newOnboarding.LastName, ob.LastName)
		assert.Equal(t, newOnboarding.Country, ob.Country)
		assert.Equal(t, newOnboarding.Email, ob.Email)
		assert.Equal(t, newOnboarding.Phone, ob.Phone)
		assert.Equal(t, newOnboarding.PhoneVerified, ob.PhoneVerified)
		assert.Equal(t, newOnboarding.ServiceAgreement, ob.ServiceAgreement)
	})

	s.Run("can update existing onboarding, can replace and add new fields", func(t *testing.T) {
		ob, err := ops.UpdateOnboarding(ctx, b, &onboarding.UpdateOnboardingArgs{
			FirstName: faker.FirstName(),
			LastName:  faker.LastName(),
			Country:   "US",
			Email:     faker.Email(),
		})
		if err != nil {
			s.Fatal(err)
		}

		newOnboarding, err := ops.UpdateOnboarding(ctx, b, &onboarding.UpdateOnboardingArgs{
			Id:    ob.ID,
			Email: faker.Email(),
			Phone: faker.E164PhoneNumber(),
		})
		if err != nil {
			s.Fatal(err)
		}

		assert.NotNil(t, ob)
		assert.Equal(t, newOnboarding.ID, ob.ID)
		assert.Equal(t, newOnboarding.FirstName, ob.FirstName)
		assert.Equal(t, newOnboarding.LastName, ob.LastName)
		assert.Equal(t, newOnboarding.Country, ob.Country)
		assert.NotEqual(t, newOnboarding.Email, ob.Email)
		assert.NotEqual(t, ob.Email, "")
		assert.NotEqual(t, newOnboarding.Email, "")
		assert.Equal(t, ob.Phone, "")
		assert.NotEqual(t, newOnboarding.Phone, ob.Phone)
		assert.NotEqual(t, newOnboarding.Phone, "")
		assert.Equal(t, newOnboarding.PhoneVerified, ob.PhoneVerified)
		assert.Equal(t, newOnboarding.ServiceAgreement, ob.ServiceAgreement)
	})

	s.Run("can set phone number to verified", func(t *testing.T) {
		ob, err := ops.UpdateOnboarding(ctx, b, &onboarding.UpdateOnboardingArgs{
			FirstName: faker.FirstName(),
			LastName:  faker.LastName(),
			Country:   "US",
			Email:     faker.Email(),
			Phone:     faker.E164PhoneNumber(),
		})
		if err != nil {
			s.Fatal(err)
		}
		assert.False(t, ob.PhoneVerified)

		newOnboarding, err := ops.UpdateOnboarding(ctx, b, &onboarding.UpdateOnboardingArgs{
			Id:            ob.ID,
			PhoneVerified: true,
		})
		if err != nil {
			s.Fatal(err)
		}

		assert.True(t, newOnboarding.PhoneVerified)
	})
}
