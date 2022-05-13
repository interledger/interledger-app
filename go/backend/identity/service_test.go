package identity

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbsqlx"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	_country "gitlab.com/fynbos/backend/country"
	test_utils "gitlab.com/fynbos/backend/utils"
	"go.uber.org/zap"
)

func TestIdentityService(s *testing.T) {
	ctx := context.Background()
	crdb, err := test_utils.SetupTestCockroachDB(ctx)
	if err != nil {
		s.Fatal(err)
	}
	defer func() {
		if err := crdb.Container.Terminate(ctx); err != nil {
			s.Fatal(err)
		}
	}()

	// the tests are run in serial. We use a global connection for
	// each of the tests.
	db, err := sqlx.Connect("postgres", crdb.URI)
	if err != nil {
		s.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			s.Fatal(err)
		}
	}()

	logger, err := zap.NewDevelopment()
	if err != nil {
		s.Fatal(err)
	}

	ctrl := gomock.NewController(s)
	defer ctrl.Finish()

	cs := _country.NewService(db)
	is, err := NewService(ServiceArgs{
		CountryService: cs,
		Db:             db,
	})
	if err != nil {
		s.Fatal(err)
	}
	is = NewLoggingService(is, logger)

	s.Run("can create an identity", func(t *testing.T) {
		args := generateCreateArgs()
		var identity *Identity
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

	s.Run("enforces 1-1 mapping between user and identity", func(t *testing.T) {
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

		var duplicate *Identity
		err = crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
			_duplicate, err := is.Create(ctx, args)
			if err != nil {
				return err
			}

			duplicate = _duplicate
			return nil
		})

		assert.Nil(t, duplicate)
		assert.ErrorIs(t, err, ErrDuplicate)
		assert.Contains(t, err.Error(), "duplicate.")
	})

	s.Run("validates create args", func(t *testing.T) {
		type Scenario struct {
			Name                 string
			Args                 CreateArgs
			ExpectedErrorMessage string
		}
		scenarios := []Scenario{
			{
				Name:                 "ID is required to create identity",
				Args:                 *generateCreateArgs(withID("")),
				ExpectedErrorMessage: "Key: 'CreateArgs.ID' Error:Field validation for 'ID' failed on the 'required' tag",
			},
			{
				Name:                 "FirstName is required to create identity",
				Args:                 *generateCreateArgs(withFirstName("")),
				ExpectedErrorMessage: "Key: 'CreateArgs.FirstName' Error:Field validation for 'FirstName' failed on the 'required' tag",
			},
			{
				Name:                 "LastName is required to create identity",
				Args:                 *generateCreateArgs(withLastName("")),
				ExpectedErrorMessage: "Key: 'CreateArgs.LastName' Error:Field validation for 'LastName' failed on the 'required' tag",
			},
			{
				Name:                 "Email must be in email format to create identity",
				Args:                 *generateCreateArgs(withEmail("test")),
				ExpectedErrorMessage: "Key: 'CreateArgs.Email' Error:Field validation for 'Email' failed on the 'email' tag",
			},
			{
				Name:                 "Country must be valid iso3166 alpha2 code to create identity",
				Args:                 *generateCreateArgs(withCountry("AA")),
				ExpectedErrorMessage: "Key: 'CreateArgs.Country' Error:Field validation for 'Country' failed on the 'iso3166_1_alpha2' tag",
			},
			{
				Name:                 "MobileNumber is required to create identity",
				Args:                 *generateCreateArgs(withMobileNumber("")),
				ExpectedErrorMessage: "Key: 'CreateArgs.MobileNumber' Error:Field validation for 'MobileNumber' failed on the 'required' tag",
			},
		}

		for _, scenario := range scenarios {
			var identity *Identity
			err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
				_identity, err := is.Create(ctx, &scenario.Args)
				if err != nil {
					return err
				}

				identity = _identity
				return nil
			})
			if err == nil {
				t.Fatal(scenario.Name)
			}

			assert.ErrorIs(t, err, ErrInvalidArgument)
			assert.Contains(t, err.Error(), scenario.ExpectedErrorMessage)
			assert.Nil(t, identity)
		}
	})

	s.Run("id is required to get identity", func(t *testing.T) {
		identity, err := is.Get(ctx, "")
		if err == nil {
			t.Fatal("User is supposed to be required to get identity.")
		}

		assert.Nil(t, identity)
		assert.Contains(t, err.Error(), "ID is required.")
	})

	s.Run("returns not found if there is no identity", func(t *testing.T) {
		identity, err := is.Get(ctx, uuid.NewString())
		if err == nil {
			t.Fatal("Should return not found.")
		}

		assert.Nil(t, identity)
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Contains(t, err.Error(), "not found.")
	})

	s.Run("can get by email", func(t *testing.T) {
		userID := uuid.NewString()
		email := faker.Email()
		var identity *Identity
		err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
			_identity, err := is.Create(ctx, generateCreateArgs(withEmail(email), withID(userID)))
			if err != nil {
				return err
			}

			identity = _identity
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, userID, identity.ID)
		assert.Equal(t, email, identity.Email)
	})
}

// TODO: auto generate helpers.
// Factory to generate create args
func generateCreateArgs(opts ...func(*CreateArgs)) *CreateArgs {
	args := &CreateArgs{
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

func withID(id string) func(*CreateArgs) {
	return func(args *CreateArgs) {
		args.ID = id
	}
}

func withEmail(email string) func(*CreateArgs) {
	return func(args *CreateArgs) {
		args.Email = email
	}
}

func withFirstName(name string) func(*CreateArgs) {
	return func(args *CreateArgs) {
		args.FirstName = name
	}
}

func withLastName(name string) func(*CreateArgs) {
	return func(args *CreateArgs) {
		args.LastName = name
	}
}

func withMobileNumber(number string) func(*CreateArgs) {
	return func(args *CreateArgs) {
		args.MobileNumber = number
	}
}

func withCountry(country string) func(*CreateArgs) {
	return func(args *CreateArgs) {
		args.Country = country
	}
}
