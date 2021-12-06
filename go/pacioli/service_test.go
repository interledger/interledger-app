package pacioli

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
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

	s.Run("account category", func(t *testing.T) {
		t.Cleanup(func() {
			test_utils.TruncateDb(ctx, db)
		})
		me, err := ps.CreateTenant("me")
		if err != nil {
			t.Fatal(err)
		}

		t.Run("tenant can create an account category", func(tt *testing.T) {
			categoryName := faker.Name()
			categoryDescription := faker.Sentence()

			category, err := ps.CreateAccountCategory(me.ID, AccountCategoryArgs{
				Name:        categoryName,
				Type:        "ASSET",
				Code:        101,
				Description: categoryDescription,
			})
			if err != nil {
				tt.Fatal(err)
			}

			assert.NotNil(tt, category.ID)
			assert.Equal(tt, me.ID, category.TenantID)
			assert.Equal(tt, categoryName, category.Name)
			assert.Equal(tt, categoryDescription, category.Description)
			assert.Equal(tt, "ASSET", category.Type)
			assert.Equal(tt, uint16(101), category.Code)
		})

		t.Run("account category name must be unique for tenant", func(tt *testing.T) {
			otherTenant, err := ps.CreateTenant(faker.Name())
			if err != nil {
				tt.Fatal(err)
			}
			categoryName := faker.Name()
			_, err = ps.CreateAccountCategory(me.ID, AccountCategoryArgs{
				Name: categoryName,
				Type: "ASSET",
				Code: 101,
			})
			if err != nil {
				tt.Fatal(err)
			}
			_, err = ps.CreateAccountCategory(otherTenant.ID, AccountCategoryArgs{
				Name: categoryName,
				Type: "ASSET",
				Code: 101,
			})
			if err != nil {
				tt.Fatal(err)
			}

			_, err = ps.CreateAccountCategory(me.ID, AccountCategoryArgs{
				Name: categoryName,
				Type: "ASSET",
				Code: 101,
			})
			if err == nil {
				tt.Fatal("Expected duplicate account category error.")
			}

			assert.Equal(tt, "Duplicate account category.", err.Error())
		})

		t.Run("tenant must exist to create an account category", func(tt *testing.T) {
			tenantID := uuid.NewString()
			category, err := ps.CreateAccountCategory(tenantID, AccountCategoryArgs{
				Name: faker.Name(),
				Type: "ASSET",
				Code: 101,
			})
			if err == nil {
				tt.Fatal(err)
			}

			assert.Nil(tt, category)
			assert.Equal(tt, "Tenant not found.", err.Error())
		})

		t.Run("validates account category arguments", func(tt *testing.T) {
			type scenario struct {
				Args AccountCategoryArgs
				Name string
			}
			table := []scenario{
				{
					Name: "Type must be one of ASSET | LIABILITY | EQUITY.",
					Args: AccountCategoryArgs{
						Name: "test",
					},
				},
				{
					Name: "Name is required.",
					Args: AccountCategoryArgs{
						Type: "ASSET",
					},
				},
			}

			for _, scenario := range table {
				category, err := ps.CreateAccountCategory(me.ID, scenario.Args)
				if err == nil {
					tt.Fatal("Expected error: " + scenario.Name)
				}

				assert.Nil(tt, category)
				assert.Equal(tt, scenario.Name, err.Error())
			}
		})

		t.Run("tenant can only get their own account category", func(tt *testing.T) {
			myCategory, err := ps.CreateAccountCategory(me.ID, AccountCategoryArgs{
				Name:        "Equity",
				Type:        "ASSET",
				Description: "My Equity account",
				Code:        1,
			})
			if err != nil {
				t.Fatal(err)
			}
			otherTenant, err := ps.CreateTenant(faker.Name())
			if err != nil {
				tt.Fatal(err)
			}
			otherTenantCategory, err := ps.CreateAccountCategory(otherTenant.ID, AccountCategoryArgs{
				Name:        "Equity",
				Type:        "ASSET",
				Description: "Other Tenant's Equity account",
				Code:        1,
			})
			if err != nil {
				tt.Fatal(err)
			}

			// tenant can get own account category
			category, err := ps.GetAccountCategory(me.ID, myCategory.ID)
			if err != nil {
				tt.Fatal(err)
			}
			assert.Equal(tt, myCategory.ID, category.ID)

			category, err = ps.GetAccountCategory(me.ID, otherTenantCategory.ID)
			if err == nil {
				tt.Fatal("Tenant must only be allowed to get their own account category.")
			}
			assert.Nil(tt, category)
			assert.Equal(tt, "Account category not found.", err.Error())
		})
	})

}
