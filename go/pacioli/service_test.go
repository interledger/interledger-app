package pacioli

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	test_utils "gitlab.com/fynbos/pacioli/utils"
)

func TestPacioliService(s *testing.T) {
	ctx := context.Background()
	crdb, err := test_utils.SetupTestCockroachDB(ctx)
	if err != nil {
		s.Fatal(err)
	}

	db, err := sqlx.Connect("postgres", crdb.URI)
	if err != nil {
		s.Fatal(err)
	}
	defer db.Close()

	ps, err := NewPacioliService(db)
	if err != nil {
		s.Fatal(err)
	}

	s.Run("tenant", func(t *testing.T) {
		t.Cleanup(func() {
			test_utils.TruncateDb(ctx, db)
		})

		t.Run("tenant can sign up", func(tt *testing.T) {
			tenant, err := ps.CreateTenant("first tenant")
			if err != nil {
				tt.Fatal(err)
			}

			assert.Equal(t, "first tenant", tenant.Identifier)
		})

		t.Run("identifier is required when signing up", func(tt *testing.T) {
			tenant, err := ps.CreateTenant("")
			if err == nil {
				tt.Fatal("Identifier must be required to signup.")
			}

			assert.Nil(tt, tenant)
			assert.Equal(tt, "Identifier is required.", err.Error())
		})
	})

}
