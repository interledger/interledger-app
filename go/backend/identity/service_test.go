package identity

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbsqlx"
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
	defer crdb.Container.Terminate(ctx)

	// the tests are run in serial. We use a global connection for
	// each of the tests.
	db, err := sqlx.Connect("postgres", crdb.URI)
	defer db.Close()

	logger, err := zap.NewDevelopment()
	if err != nil {
		s.Fatal(err)
	}
	defer logger.Sync()

	cs := _country.NewService()
	is, err := NewService(cs)
	if err != nil {
		s.Fatal(err)
	}
	is = NewLoggingService(is, logger)

	s.Run("can create an identity", func(t *testing.T) {
		args := generateCreateArgs()
		var identity *Identity
		var country string
		err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
			c, err := cs.GetByAlpha2(ctx, tx, args.Country)
			if err != nil {
				t.Fatal(err)
			}
			country = c.Alpha_2
			_identity, err := is.Create(ctx, tx, *args)
			if err != nil {
				t.Fatal(err)
			}

			identity = _identity
			return nil
		})
		assert.Equal(t, args.ID, identity.ID)
		assert.Equal(t, args.Email, identity.Email)
		assert.Equal(t, args.FirstName, identity.FirstName)
		assert.Equal(t, args.LastName, identity.LastName)
		assert.Equal(t, args.MobileNumber, identity.MobileNumber)
		assert.Equal(t, country, identity.Country)
		assert.Equal(t, "", identity.DateOfBirth)
		assert.Equal(t, []string{}, identity.Address)
		assert.Equal(t, "", identity.State)
		assert.Equal(t, "", identity.City)
		assert.Equal(t, "", identity.PostalCode)
		assert.Equal(t, "", identity.TaxIDNumber)
		assert.Equal(t, "", identity.ProviderID)
		assert.Equal(t, "", identity.Provider)

		var fetchedIdentity *Identity
		err = crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
			_fetchedIdentity, err := is.Get(ctx, tx, args.ID)
			if err != nil {
				return err
			}

			fetchedIdentity = _fetchedIdentity
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, args.ID, fetchedIdentity.ID)
		assert.Equal(t, args.Email, fetchedIdentity.Email)
		assert.Equal(t, args.FirstName, fetchedIdentity.FirstName)
		assert.Equal(t, args.LastName, fetchedIdentity.LastName)
		assert.Equal(t, args.MobileNumber, fetchedIdentity.MobileNumber)
		assert.Equal(t, country, fetchedIdentity.Country)
		assert.Equal(t, "", fetchedIdentity.DateOfBirth)
		assert.Equal(t, []string{}, fetchedIdentity.Address)
		assert.Equal(t, "", fetchedIdentity.State)
		assert.Equal(t, "", fetchedIdentity.City)
		assert.Equal(t, "", fetchedIdentity.PostalCode)
		assert.Equal(t, "", fetchedIdentity.TaxIDNumber)
		assert.Equal(t, "", fetchedIdentity.ProviderID)
		assert.Equal(t, "", fetchedIdentity.Provider)
	})

	s.Run("enforces 1-1 mapping between user and identity", func(t *testing.T) {
		args := generateCreateArgs()
		err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
			_, err := is.Create(ctx, tx, *args)
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
			_duplicate, err := is.Create(ctx, tx, *args)
			if err != nil {
				return err
			}

			duplicate = _duplicate
			return nil
		})

		assert.Nil(t, duplicate)
		assert.EqualError(t, err, "Identity exists.")
	})

	s.Run("validates create args", func(t *testing.T) {
		type Scenario struct {
			Name          string
			Args          CreateArgs
			ExpectedError string
		}
		scenarios := []Scenario{
			{
				Name:          "ID is required to create identity",
				Args:          *generateCreateArgs(withID("")),
				ExpectedError: "Key: 'CreateArgs.ID' Error:Field validation for 'ID' failed on the 'required' tag",
			},
			{
				Name:          "FirstName is required to create identity",
				Args:          *generateCreateArgs(withFirstName("")),
				ExpectedError: "Key: 'CreateArgs.FirstName' Error:Field validation for 'FirstName' failed on the 'required' tag",
			},
			{
				Name:          "LastName is required to create identity",
				Args:          *generateCreateArgs(withLastName("")),
				ExpectedError: "Key: 'CreateArgs.LastName' Error:Field validation for 'LastName' failed on the 'required' tag",
			},
			{
				Name:          "Email must be in email format to create identity",
				Args:          *generateCreateArgs(withEmail("test")),
				ExpectedError: "Key: 'CreateArgs.Email' Error:Field validation for 'Email' failed on the 'email' tag",
			},
			{
				Name:          "Country must be valid iso3166 alpha2 code to create identity",
				Args:          *generateCreateArgs(withCountry("AA")),
				ExpectedError: "Key: 'CreateArgs.Country' Error:Field validation for 'Country' failed on the 'iso3166_1_alpha2' tag",
			},
			{
				Name:          "MobileNumber is required to create identity",
				Args:          *generateCreateArgs(withMobileNumber("")),
				ExpectedError: "Key: 'CreateArgs.MobileNumber' Error:Field validation for 'MobileNumber' failed on the 'required' tag",
			},
		}

		for _, scenario := range scenarios {
			var identity *Identity
			err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
				_identity, err := is.Create(ctx, tx, scenario.Args)
				if err != nil {
					return err
				}

				identity = _identity
				return nil
			})
			if err == nil {
				t.Fatal(scenario.Name)
			}

			assert.Equal(t, scenario.ExpectedError, err.Error())
			assert.Nil(t, identity)
		}
	})

	s.Run("id is required to get identity", func(t *testing.T) {
		var identity *Identity
		err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
			_identity, err := is.Get(ctx, tx, "")
			if err != nil {
				return err
			}

			identity = _identity
			return nil
		})
		if err == nil {
			t.Fatal("User is supposed to be required to get identity.")
		}

		assert.Nil(t, identity)
		assert.Equal(t, "ID is required.", err.Error())
	})

	s.Run("returns not found if there is no identity", func(t *testing.T) {
		var identity *Identity
		err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
			_identity, err := is.Get(ctx, tx, uuid.NewString())
			if err != nil {
				return err
			}

			identity = _identity
			return nil
		})
		if err == nil {
			t.Fatal("Should return not found.")
		}

		assert.Nil(t, identity)
		assert.Equal(t, "Not found.", err.Error())
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
