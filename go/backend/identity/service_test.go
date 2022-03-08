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
	defer func() {
		if err := logger.Sync(); err != nil {
			s.Fatal(err)
		}
	}()

	ctrl := gomock.NewController(s)
	defer ctrl.Finish()

	cs := _country.NewService()
	is, err := NewService(ServiceArgs{
		CountryService: cs,
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
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, args.ID, identity.ID)
		assert.Equal(t, args.Email, identity.Email)
		assert.Equal(t, args.FirstName, identity.FirstName)
		assert.Equal(t, args.LastName, identity.LastName)
		assert.Equal(t, args.MobileNumber, identity.MobileNumber)
		assert.Equal(t, country, identity.Country)

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

	// TODO: this is going to be moved to account
	// s.Run("verify identity", func(t *testing.T) {
	// 	args := generateCreateArgs()
	// 	var identity *Identity
	// 	err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
	// 		_identity, err := is.Create(ctx, tx, *args)
	// 		if err != nil {
	// 			t.Fatal(err)
	// 		}

	// 		identity = _identity
	// 		return nil
	// 	})
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}

	// 	t.Run("creates customer at provider and updates identity", func(tt *testing.T) {
	// 		customerID := uuid.NewString()
	// 		args := generateVerifyArgs(withIdentityID(identity.ID))

	// 		var id *Identity
	// 		err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
	// 			_id, err := is.Verify(ctx, tx, args)
	// 			if err != nil {
	// 				return err
	// 			}
	// 			id = _id

	// 			return nil
	// 		})
	// 		if err != nil {
	// 			tt.Fatal(err)
	// 		}

	// 		assert.Equal(tt, customerID, id.ProviderID)
	// 		assert.Equal(tt, identity.ID, id.ID)
	// 		assert.Equal(tt, identity.Email, id.Email)
	// 		assert.Equal(tt, identity.FirstName, id.FirstName)
	// 		assert.Equal(tt, identity.LastName, id.LastName)
	// 		assert.Equal(tt, identity.MobileNumber, id.MobileNumber)
	// 		assert.Equal(tt, identity.Country, id.Country)
	// 		assert.Equal(tt, args.DateOfBirth, id.DateOfBirth)
	// 		assert.Equal(tt, args.Address, id.Address)
	// 		assert.Equal(tt, args.State, id.State)
	// 		assert.Equal(tt, args.City, id.City)
	// 		assert.Equal(tt, args.PostalCode, id.PostalCode)
	// 		assert.Equal(tt, args.TaxIDNumber, id.TaxIDNumber)
	// 		assert.Equal(tt, identity.Provider, id.Provider)
	// 	})

	// 	t.Run("validates arguments", func(tt *testing.T) {
	// 		type Scenario struct {
	// 			Name          string
	// 			Args          *VerifyArgs
	// 			ExpectedError string
	// 		}
	// 		scenarios := []Scenario{
	// 			{
	// 				Name:          "IdentityID is required to verify identity to verify identity",
	// 				Args:          generateVerifyArgs(withIdentityID("")),
	// 				ExpectedError: "Key: 'VerifyArgs.IdentityID' Error:Field validation for 'IdentityID' failed on the 'required' tag",
	// 			},
	// 			{
	// 				Name:          "DateOfBirth must be valid date to verify identity",
	// 				Args:          generateVerifyArgs(withDateOfBirth("")),
	// 				ExpectedError: "Key: 'VerifyArgs.DateOfBirth' Error:Field validation for 'DateOfBirth' failed on the 'datetime' tag",
	// 			},
	// 			{
	// 				Name:          "At least 1 address is required to verify identity",
	// 				Args:          generateVerifyArgs(withAddress([]string{})),
	// 				ExpectedError: "Key: 'VerifyArgs.Address' Error:Field validation for 'Address' failed on the 'min' tag",
	// 			},
	// 			{
	// 				Name:          "State is required to verify identity",
	// 				Args:          generateVerifyArgs(withState("")),
	// 				ExpectedError: "Key: 'VerifyArgs.State' Error:Field validation for 'State' failed on the 'required' tag",
	// 			},
	// 			{
	// 				Name:          "City is required to verify identity",
	// 				Args:          generateVerifyArgs(withCity("")),
	// 				ExpectedError: "Key: 'VerifyArgs.City' Error:Field validation for 'City' failed on the 'required' tag",
	// 			},
	// 			{
	// 				Name:          "PostalCode is required to verify identity",
	// 				Args:          generateVerifyArgs(withPostalCode("")),
	// 				ExpectedError: "Key: 'VerifyArgs.PostalCode' Error:Field validation for 'PostalCode' failed on the 'required' tag",
	// 			},
	// 			{
	// 				Name:          "TaxIDNumber is required to verify identity",
	// 				Args:          generateVerifyArgs(withTaxIDNumber("")),
	// 				ExpectedError: "Key: 'VerifyArgs.TaxIDNumber' Error:Field validation for 'TaxIDNumber' failed on the 'required' tag",
	// 			},
	// 		}

	// 		for _, scenario := range scenarios {
	// 			var identity *Identity
	// 			err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
	// 				_identity, err := is.Verify(ctx, tx, scenario.Args)
	// 				if err != nil {
	// 					return err
	// 				}

	// 				identity = _identity
	// 				return nil
	// 			})
	// 			if err == nil {
	// 				tt.Fatal(scenario.Name)
	// 			}

	// 			assert.Equal(tt, scenario.ExpectedError, err.Error())
	// 			assert.Nil(tt, identity)
	// 		}
	// 	})

	// 	t.Run("does not record data if customer is not created at provider", func(tt *testing.T) {
	// 		args := generateVerifyArgs(withIdentityID(identity.ID))

	// 		var id *Identity
	// 		err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
	// 			_id, err := is.Verify(ctx, tx, args)
	// 			if err != nil {
	// 				return err
	// 			}
	// 			id = _id

	// 			return nil
	// 		})
	// 		if err == nil {
	// 			tt.Fatal("Should have returned error when provider call failed.")
	// 		}

	// 		assert.Nil(tt, id)
	// 	})

	// 	t.Run("fails if identity does not exist", func(tt *testing.T) {
	// 		var id *Identity
	// 		args := generateVerifyArgs(withIdentityID(uuid.NewString()))

	// 		err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
	// 			_id, err := is.Verify(ctx, tx, args)
	// 			if err != nil {
	// 				return err
	// 			}
	// 			id = _id

	// 			return nil
	// 		})
	// 		if err == nil {
	// 			tt.Fatal("Should have failed with no identity found.")
	// 		}

	// 		assert.Nil(tt, id)
	// 		assert.Equal(tt, "Not found.", err.Error())
	// 	})
	// })
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

// func generateVerifyArgs(opts ...func(*VerifyArgs)) *VerifyArgs {
// 	ret := &VerifyArgs{
// 		IdentityID:  uuid.NewString(),
// 		DateOfBirth: faker.Date(),
// 		Address:     []string{faker.Name()},
// 		State:       faker.FirstName(),
// 		City:        faker.FirstName(),
// 		PostalCode:  faker.LastName(),
// 		TaxIDNumber: faker.CCNumber(),
// 	}
// 	for _, opt := range opts {
// 		opt(ret)
// 	}

// 	return ret
// }

// func withIdentityID(id string) func(*VerifyArgs) {
// 	return func(args *VerifyArgs) {
// 		args.IdentityID = id
// 	}
// }

// func withDateOfBirth(dob string) func(*VerifyArgs) {
// 	return func(args *VerifyArgs) {
// 		args.DateOfBirth = dob
// 	}
// }

// func withAddress(address []string) func(*VerifyArgs) {
// 	return func(args *VerifyArgs) {
// 		args.Address = address
// 	}
// }

// func withState(state string) func(*VerifyArgs) {
// 	return func(args *VerifyArgs) {
// 		args.State = state
// 	}
// }

// func withCity(city string) func(*VerifyArgs) {
// 	return func(args *VerifyArgs) {
// 		args.City = city
// 	}
// }

// func withPostalCode(code string) func(*VerifyArgs) {
// 	return func(args *VerifyArgs) {
// 		args.PostalCode = code
// 	}
// }

// func withTaxIDNumber(tax string) func(*VerifyArgs) {
// 	return func(args *VerifyArgs) {
// 		args.TaxIDNumber = tax
// 	}
// }
