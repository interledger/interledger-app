package ops_test

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbsqlx"
	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/country"
	country_client "gitlab.com/fynbos/backend/country/client"
	"gitlab.com/fynbos/backend/identity"
	identity_client "gitlab.com/fynbos/backend/identity/client"
	test_utils "gitlab.com/fynbos/backend/utils"
	"go.uber.org/zap"
)

func TestIdentityService(s *testing.T) {
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(s, ctx)

	logger, err := zap.NewDevelopment()
	if err != nil {
		s.Fatal(err)
	}

	b := &testBackends{val: validator.New(), db: db}

	ctrl := gomock.NewController(s)
	defer ctrl.Finish()

	cs := country_client.New(b)
	b.counties = cs

	is := identity_client.New(b, logger)

	s.Run("can create an identityModel", func(t *testing.T) {
		args := generateCreateArgs()
		var identity *identity.Identity
		var country string
		err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
			c, err := cs.GetByAlpha2(ctx, args.Country)
			if err != nil {
				t.Fatal(err)
			}
			country = c.Alpha_2
			_identity, err := is.Create(ctx, args)
			if err != nil {
				t.Fatal(err)
			}

			identity = _identity
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, args.ID, identity.ID)
		assert.Equal(t, args.Email, identity.Email)
		assert.Equal(t, args.FirstName, identity.FirstName)
		assert.Equal(t, args.LastName, identity.LastName)
		assert.Equal(t, args.MobileNumber, identity.MobileNumber)
		assert.Equal(t, country, identity.Country)

		fetchedIdentity, err := is.Get(ctx, args.ID)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, args.ID, fetchedIdentity.ID)
		assert.Equal(t, args.Email, fetchedIdentity.Email)
		assert.Equal(t, args.FirstName, fetchedIdentity.FirstName)
		assert.Equal(t, args.LastName, fetchedIdentity.LastName)
		assert.Equal(t, args.MobileNumber, fetchedIdentity.MobileNumber)
		assert.Equal(t, country, fetchedIdentity.Country)
	})

	s.Run("enforces 1-1 mapping between user and identityModel", func(t *testing.T) {
		args := generateCreateArgs()
		err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
			_, err := is.Create(ctx, args)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			t.Fatal(err)
		}

		var duplicate *identity.Identity
		err = crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
			_duplicate, err := is.Create(ctx, args)
			if err != nil {
				return err
			}

			duplicate = _duplicate
			return nil
		})

		assert.Nil(t, duplicate)
		assert.ErrorIs(t, err, identity.ErrDuplicate)
		assert.Contains(t, err.Error(), "duplicate.")
	})

	s.Run("validates create args", func(t *testing.T) {
		type Scenario struct {
			Name                 string
			Args                 identity.CreateArgs
			ExpectedErrorMessage string
		}
		scenarios := []Scenario{
			{
				Name:                 "ID is required to create identityModel",
				Args:                 *generateCreateArgs(withID("")),
				ExpectedErrorMessage: "Key: 'CreateArgs.ID' Error:Field validation for 'ID' failed on the 'required' tag",
			},
			{
				Name:                 "FirstName is required to create identityModel",
				Args:                 *generateCreateArgs(withFirstName("")),
				ExpectedErrorMessage: "Key: 'CreateArgs.FirstName' Error:Field validation for 'FirstName' failed on the 'required' tag",
			},
			{
				Name:                 "LastName is required to create identityModel",
				Args:                 *generateCreateArgs(withLastName("")),
				ExpectedErrorMessage: "Key: 'CreateArgs.LastName' Error:Field validation for 'LastName' failed on the 'required' tag",
			},
			{
				Name:                 "Email must be in email format to create identityModel",
				Args:                 *generateCreateArgs(withEmail("test")),
				ExpectedErrorMessage: "Key: 'CreateArgs.Email' Error:Field validation for 'Email' failed on the 'email' tag",
			},
			{
				Name:                 "Country must be valid iso3166 alpha2 code to create identityModel",
				Args:                 *generateCreateArgs(withCountry("AA")),
				ExpectedErrorMessage: "Key: 'CreateArgs.Country' Error:Field validation for 'Country' failed on the 'iso3166_1_alpha2' tag",
			},
			{
				Name:                 "MobileNumber is required to create identityModel",
				Args:                 *generateCreateArgs(withMobileNumber("")),
				ExpectedErrorMessage: "Key: 'CreateArgs.MobileNumber' Error:Field validation for 'MobileNumber' failed on the 'required' tag",
			},
		}

		for _, scenario := range scenarios {
			var id *identity.Identity
			err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
				_identity, err := is.Create(ctx, &scenario.Args)
				if err != nil {
					return err
				}

				id = _identity
				return nil
			})
			if err == nil {
				t.Fatal(scenario.Name)
			}

			assert.ErrorIs(t, err, identity.ErrInvalidArgument)
			assert.Contains(t, err.Error(), scenario.ExpectedErrorMessage)
			assert.Nil(t, id)
		}
	})

	s.Run("id is required to get identityModel", func(t *testing.T) {
		identity, err := is.Get(ctx, "")
		if err == nil {
			t.Fatal("User is supposed to be required to get identityModel.")
		}

		assert.Nil(t, identity)
		assert.Contains(t, err.Error(), "ID is required.")
	})

	s.Run("returns not found if there is no identityModel", func(t *testing.T) {
		id, err := is.Get(ctx, uuid.NewString())
		if err == nil {
			t.Fatal("Should return not found.")
		}

		assert.Nil(t, id)
		assert.ErrorIs(t, err, identity.ErrNotFound)
		assert.Contains(t, err.Error(), "not found.")
	})

	s.Run("can get by email", func(t *testing.T) {
		userID := uuid.NewString()
		email := faker.Email()
		var id *identity.Identity
		err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
			_identity, err := is.Create(ctx, generateCreateArgs(withEmail(email), withID(userID)))
			if err != nil {
				return err
			}

			id = _identity
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, userID, id.ID)
		assert.Equal(t, email, id.Email)
	})
}

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Countries() country.Client
}

type testBackends struct {
	db       *sqlx.DB
	val      *validator.Validate
	counties country.Client
}

func (t testBackends) Validator() *validator.Validate {
	return t.val
}

func (t testBackends) Countries() country.Client {
	return t.counties
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

// TODO: auto generate helpers.
// Factory to generate create args
func generateCreateArgs(opts ...func(*identity.CreateArgs)) *identity.CreateArgs {
	args := &identity.CreateArgs{
		ID:           uuid.NewString(),
		Email:        faker.Email(),
		FirstName:    faker.Name(),
		LastName:     faker.LastName(),
		MobileNumber: faker.Phonenumber(),
		Country:      "US",
	}

	for _, opt := range opts {
		opt(args)
	}

	return args
}

func withID(id string) func(*identity.CreateArgs) {
	return func(args *identity.CreateArgs) {
		args.ID = id
	}
}

func withEmail(email string) func(*identity.CreateArgs) {
	return func(args *identity.CreateArgs) {
		args.Email = email
	}
}

func withFirstName(name string) func(*identity.CreateArgs) {
	return func(args *identity.CreateArgs) {
		args.FirstName = name
	}
}

func withLastName(name string) func(*identity.CreateArgs) {
	return func(args *identity.CreateArgs) {
		args.LastName = name
	}
}

func withMobileNumber(number string) func(*identity.CreateArgs) {
	return func(args *identity.CreateArgs) {
		args.MobileNumber = number
	}
}

func withCountry(country string) func(*identity.CreateArgs) {
	return func(args *identity.CreateArgs) {
		args.Country = country
	}
}
