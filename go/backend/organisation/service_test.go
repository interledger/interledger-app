package organisation

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/osohq/go-oso"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/user"
	test_utils "gitlab.com/fynbos/backend/utils"
)

// Here to avoid circular deps
func newAuthzProvider() (*oso.Oso, error) {
	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok {
		return nil, errors.New("Could not get directory path for utils/testing.")
	}

	o, err := oso.NewOso()
	if err != nil {
		return nil, err
	}

	o.RegisterClass(Organisation{}, nil)
	o.RegisterClass(user.User{}, nil)

	err = o.LoadFiles([]string{filepath.Join(filepath.Dir(moduleDir), "../authorization/main.polar")})
	if err != nil {
		return nil, err
	}

	return &o, nil
}

func TestOrganisationService(s *testing.T) {
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

	authz, err := newAuthzProvider()
	if err != nil {
		s.Fatal(err)
	}

	organisations, err := NewService(db, authz)
	if err != nil {
		s.Fatal(err)
	}

	s.Run("user can create an organisation", func(t *testing.T) {
		t.Cleanup(func() {
			test_utils.TruncateDb(ctx, db)
		})
		user := user.User{
			ID: uuid.New().String(),
		}

		org, err := organisations.Create("My first organisation.", user)
		if err != nil {
			t.Fatal(err)
		}

		orgFromDb := Organisation{}
		err = db.Get(&orgFromDb, "SELECT * FROM organisations WHERE id=$1 LIMIT 1", org.ID)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "My first organisation.", org.Name)
		assert.Equal(t, user.ID, org.OwnerID) // ensure the owner is set.
		assert.Equal(t, org, &orgFromDb)
	})

	s.Run("owner can get their organisation", func(t *testing.T) {
		t.Cleanup(func() {
			test_utils.TruncateDb(ctx, db)
		})
		owner := user.User{
			ID: uuid.New().String(),
		}
		createdOrg, err := organisations.Create("test", owner)

		org, err := organisations.Get(createdOrg.ID, owner)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "test", org.Name)
	})

	s.Run("user can only get organisations they own", func(t *testing.T) {
		t.Cleanup(func() {
			test_utils.TruncateDb(ctx, db)
		})
		me := user.User{
			ID: uuid.New().String(),
		}
		otherUser := user.User{
			ID: uuid.New().String(),
		}
		otherOrg, err := organisations.Create("test", otherUser)

		org, err := organisations.Get(otherOrg.ID, me)
		if err == nil {
			t.Fatal("Expected a NotFound error to be thrown.")
		}

		assert.Nil(t, org)
	})
}
