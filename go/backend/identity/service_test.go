package identity

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
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

	is, err := NewService(db)
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

		identity, err := is.Create(CreateArgs{
			User:      user,
			Country:   "USA",
			LegalName: name,
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, user.ID, identity.ID)
		assert.Equal(t, user.Email, identity.Email)
		assert.Equal(t, name, identity.LegalName)
		assert.Equal(t, "USA", identity.Country)

		fetchedIdentity, err := is.Get(user.ID)
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
		_, err := is.Create(CreateArgs{
			User:      usr,
			Country:   "USA",
			LegalName: faker.Name(),
		})
		if err != nil {
			t.Fatal(err)
		}

		duplicate, err := is.Create(CreateArgs{
			User:      usr,
			Country:   "USA",
			LegalName: faker.Name(),
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
			identity, err := is.Create(scenario.Args)
			if err == nil {
				t.Fatal(scenario.Name)
			}

			assert.Equal(t, scenario.ExpectedError, err.Error())
			assert.Nil(t, identity)
		}
	})

	s.Run("id is required to get identity", func(t *testing.T) {
		identity, err := is.Get("")
		if err == nil {
			t.Fatal("User is supposed to be required to get identity.")
		}

		assert.Nil(t, identity)
		assert.Equal(t, "ID is required.", err.Error())
	})

	s.Run("returns not found if there is no identity", func(t *testing.T) {
		identity, err := is.Get(uuid.NewString())
		if err == nil {
			t.Fatal("Should return not found.")
		}

		assert.Nil(t, identity)
		assert.Equal(t, "Not found.", err.Error())
	})
}
