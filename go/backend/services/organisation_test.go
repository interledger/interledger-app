package services

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/db/utils"
)

func TestOrganisationService(t *testing.T) {
	ctx := context.Background()
	crdb, err := utils.SetupTestCockroachDB(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer crdb.Container.Terminate(ctx)

	// the tests are run in serial. We use a global connection for
	// each of the tests.
	db, err := sqlx.Connect("postgres", crdb.URI)
	defer db.Close()

	organisations, err := NewOrganisationsService(db)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("can get an organisation", func(tt *testing.T) {
		tt.Cleanup(func() {
			utils.TruncateDb(ctx, db)
		})

		var id string
		row, err := db.Query("INSERT INTO organisations (name) VALUES ('test') RETURNING id")
		if err != nil {
			t.Fatal(err)
		}
		if row.Next() {
			err = row.Scan(&id)
			if err != nil {
				t.Fatal(err)
			}
		}

		org, err := organisations.Get(id)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "test", org.Name)
	})

	t.Run("can create an organisation", func(tt *testing.T) {
		tt.Cleanup(func() {
			utils.TruncateDb(ctx, db)
		})

		org, err := organisations.Create("My first organisation.")
		if err != nil {
			t.Fatal(err)
		}

		freshOrg, err := organisations.Get(org.ID)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "My first organisation.", freshOrg.Name)
		assert.Equal(t, org.ID, freshOrg.ID)
	})

}
