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
	"gitlab.com/fynbos/backend/user"
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

	is, err := NewService()
	if err != nil {
		s.Fatal(err)
	}
	is = NewLoggingService(is, logger)

	s.Run("can create an identity", func(t *testing.T) {
		user := &user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}
		name := faker.Name()

		var identity *Identity
		err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
			_identity, err := is.Create(ctx, tx, CreateArgs{
				User:      user,
				Country:   "USA",
				LegalName: name,
			})
			if err != nil {
				t.Fatal(err)
			}

			identity = _identity
			return nil
		})
		assert.Equal(t, user.ID, identity.ID)
		assert.Equal(t, user.Email, identity.Email)
		assert.Equal(t, name, identity.LegalName)
		assert.Equal(t, "USA", identity.Country)

		var fetchedIdentity *Identity
		err = crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
			_fetchedIdentity, err := is.Get(ctx, tx, user.ID)
			if err != nil {
				return err
			}

			fetchedIdentity = _fetchedIdentity
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, user.ID, fetchedIdentity.ID)
		assert.Equal(t, user.Email, fetchedIdentity.Email)
		assert.Equal(t, name, fetchedIdentity.LegalName)
		assert.Equal(t, "USA", fetchedIdentity.Country)
	})

	s.Run("enforces 1-1 mapping between user and identity", func(t *testing.T) {
		usr := &user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}
		err := crdbsqlx.ExecuteTx(ctx, db, nil, func(tx *sqlx.Tx) error {
			_, err := is.Create(ctx, tx, CreateArgs{
				User:      usr,
				Country:   "USA",
				LegalName: faker.Name(),
			})
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
			_duplicate, err := is.Create(ctx, tx, CreateArgs{
				User:      usr,
				Country:   "USA",
				LegalName: faker.Name(),
			})
			if err != nil {
				return err
			}

			duplicate = _duplicate
			return nil
		})

		assert.Nil(t, duplicate)
		assert.EqualError(t, err, "Identity exists.")
	})

	s.Run("all fields are required to create an identity", func(t *testing.T) {
		userId := uuid.NewString()
		email := faker.Email()
		name := faker.Name()
		type Scenario struct {
			Name          string
			Args          CreateArgs
			ExpectedError string
		}
		scenarios := []Scenario{
			{
				Name: "User is required to create identity",
				Args: CreateArgs{
					Country:   "USA",
					LegalName: name,
				},
				ExpectedError: "User is required.",
			},
			{
				Name: "User Email is required to create identity",
				Args: CreateArgs{
					User: &user.User{
						ID: userId,
					},
					Country:   "USA",
					LegalName: name,
				},
				ExpectedError: "User Email is required.",
			},
			{
				Name: "User ID is required to create identity",
				Args: CreateArgs{
					User: &user.User{
						Email: email,
					},
					Country:   "USA",
					LegalName: name,
				},
				ExpectedError: "User ID is required.",
			},
			{
				Name: "LegalName is required to create identity",
				Args: CreateArgs{
					User: &user.User{
						ID:    userId,
						Email: email,
					},
					Country: "USA",
				},
				ExpectedError: "LegalName is required.",
			},
			{
				Name: "Country is required to create identity",
				Args: CreateArgs{
					User: &user.User{
						ID:    userId,
						Email: email,
					},
					LegalName: name,
				},
				ExpectedError: "Country is required.",
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
