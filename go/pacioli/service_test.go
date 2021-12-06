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

		t.Run("account category code must be unique for tenant", func(tt *testing.T) {
			otherTenant, err := ps.CreateTenant(faker.Name())
			if err != nil {
				tt.Fatal(err)
			}
			code := uint16(102)
			_, err = ps.CreateAccountCategory(me.ID, AccountCategoryArgs{
				Name: faker.Name(),
				Type: "ASSET",
				Code: code,
			})
			if err != nil {
				tt.Fatal(err)
			}
			_, err = ps.CreateAccountCategory(otherTenant.ID, AccountCategoryArgs{
				Name: faker.Name(),
				Type: "ASSET",
				Code: code,
			})
			if err != nil {
				tt.Fatal(err)
			}

			_, err = ps.CreateAccountCategory(me.ID, AccountCategoryArgs{
				Name: faker.Name(),
				Type: "ASSET",
				Code: code,
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
				Code:        203,
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
				Code:        204,
			})
			if err != nil {
				tt.Fatal(err)
			}

			// tenant can get own account category
			category, err := ps.GetAccountCategoryByCode(me.ID, myCategory.Code)
			if err != nil {
				tt.Fatal(err)
			}
			assert.Equal(tt, myCategory.ID, category.ID)

			category, err = ps.GetAccountCategoryByCode(me.ID, otherTenantCategory.Code)
			if err == nil {
				tt.Fatal("Tenant must only be allowed to get their own account category.")
			}
			assert.Nil(tt, category)
			assert.Equal(tt, "Account category not found.", err.Error())
		})
	})

	s.Run("transaction type", func(t *testing.T) {
		me, err := ps.CreateTenant(faker.Name())
		myEquityAccountCategory, err := ps.CreateAccountCategory(me.ID, AccountCategoryArgs{
			Name:        "Equity",
			Type:        "ASSET",
			Description: "My Equity account",
			Code:        1,
		})
		if err != nil {
			t.Fatal(err)
		}
		myAccountHolderAccountCategory, err := ps.CreateAccountCategory(me.ID, AccountCategoryArgs{
			Name:        "Account Holder Funds",
			Type:        "LIABILITY",
			Description: "My account holder funds",
			Code:        2,
		})
		if err != nil {
			t.Fatal(err)
		}

		t.Run("tenant must exist to create transaction type", func(tt *testing.T) {
			tenantID := uuid.NewString()

			transaction, err := ps.CreateTransactionType(tenantID, TransactionTypeArgs{
				Name:                      faker.Name(),
				Description:               "Deposit by account holder.",
				CreditAccountCategoryCode: myAccountHolderAccountCategory.Code,
				DebitAccountCategoryCode:  myEquityAccountCategory.Code,
			})
			if err == nil {
				tt.Fatal("Tenant must exist to create transaction type.")
			}

			assert.Nil(tt, transaction)
			assert.Equal(tt, "Tenant not found.", err.Error())
		})

		t.Run("transaction type name must be unique per tenant", func(tt *testing.T) {
			transactionName := faker.Name()
			transaction, err := ps.CreateTransactionType(me.ID, TransactionTypeArgs{
				Name:                      transactionName,
				Description:               "Deposit by account holder.",
				CreditAccountCategoryCode: myAccountHolderAccountCategory.Code,
				DebitAccountCategoryCode:  myEquityAccountCategory.Code,
			})
			if err != nil {
				tt.Fatal(err)
			}
			assert.Equal(tt, transactionName, transaction.Name)
			assert.Equal(tt, "Deposit by account holder.", transaction.Description)
			assert.Equal(tt, myAccountHolderAccountCategory.Code, transaction.CreditAccountCategoryCode)
			assert.Equal(tt, myEquityAccountCategory.Code, transaction.DebitAccountCategoryCode)

			duplicateTransaction, err := ps.CreateTransactionType(me.ID, TransactionTypeArgs{
				Name:                      transactionName,
				Description:               "Deposit by account holder.",
				CreditAccountCategoryCode: myAccountHolderAccountCategory.Code,
				DebitAccountCategoryCode:  myEquityAccountCategory.Code,
			})
			if err == nil {
				tt.Fatal("Transaction type must be unique per tenant.")
			}
			assert.Nil(tt, duplicateTransaction)
		})

		t.Run("account category must belong to tenant", func(tt *testing.T) {
			transactionName := faker.Name()
			otherTenant, err := ps.CreateTenant(faker.Name())
			if err != nil {
				tt.Fatal(err)
			}
			otherCategory, err := ps.CreateAccountCategory(otherTenant.ID, AccountCategoryArgs{
				Name:        faker.Name(),
				Type:        "ASSET",
				Description: "other tenant's account category",
				Code:        301,
			})
			if err != nil {
				tt.Fatal(err)
			}

			transaction, err := ps.CreateTransactionType(me.ID, TransactionTypeArgs{
				Name:                      transactionName,
				Description:               "Deposit by account holder.",
				CreditAccountCategoryCode: myAccountHolderAccountCategory.Code,
				DebitAccountCategoryCode:  otherCategory.Code,
			})
			if err == nil {
				tt.Fatal("Tenants can only create transaction type using their own account categories.")
			}

			assert.Nil(tt, transaction)
			assert.Equal(tt, "Account category not found.", err.Error())
		})

		t.Run("account category codes must be different", func(tt *testing.T) {
			transactionName := faker.Name()

			transaction, err := ps.CreateTransactionType(me.ID, TransactionTypeArgs{
				Name:                      transactionName,
				Description:               "Deposit by account holder.",
				CreditAccountCategoryCode: myAccountHolderAccountCategory.Code,
				DebitAccountCategoryCode:  myAccountHolderAccountCategory.Code,
			})
			if err == nil {
				tt.Fatal("Should fail if account code categories aren't different.")
			}

			assert.Nil(tt, transaction)
			assert.Equal(tt, "Account category codes must be different.", err.Error())
		})

		s.Run("tenant can only get their own transaction type", func(tt *testing.T) {
			myTransactionType, err := ps.CreateTransactionType(me.ID, TransactionTypeArgs{
				Name:                      faker.Name(),
				Description:               "Deposit by account holder.",
				CreditAccountCategoryCode: myAccountHolderAccountCategory.Code,
				DebitAccountCategoryCode:  myEquityAccountCategory.Code,
			})
			if err != nil {
				tt.Fatal(err)
			}
			otherTenant, err := ps.CreateTenant(faker.Name())
			if err != nil {
				tt.Fatal(err)
			}

			transactionType, err := ps.GetTransactionType(otherTenant.ID, myTransactionType.ID)
			if err == nil {
				tt.Fatal("Tenants must only be able to get their own transaciton types.")
			}

			assert.Nil(tt, transactionType)
			assert.Equal(tt, "Transaction type not found.", err.Error())
		})
	})

	s.Run("ledger", func(t *testing.T) {
		t.Cleanup(func() {
			test_utils.TruncateDb(ctx, db)
		})
		me, err := ps.CreateTenant(faker.Name())
		if err != nil {
			t.Fatal(err)
		}

		t.Run("tenant must exist to create ledger", func(tt *testing.T) {
			tenantID := uuid.NewString()
			ledger, err := ps.CreateLedger(tenantID, "my first ledger")
			if err == nil {
				tt.Fatal("Tenant must exist to create ledger.")
			}

			assert.Nil(tt, ledger)
			assert.Equal(tt, "Tenant not found.", err.Error())
		})

		t.Run("tenant can create a ledger", func(tt *testing.T) {
			ledger, err := ps.CreateLedger(me.ID, "my first ledger")
			if err != nil {
				t.Fatal(err)
			}

			assert.NotNil(tt, ledger.ID)
			assert.Equal(tt, "my first ledger", ledger.Name)
			assert.Equal(tt, me.ID, ledger.TenantID)
		})

		t.Run("tenants can only get their own ledgers", func(tt *testing.T) {
			otherTenant, err := ps.CreateTenant(faker.Name())
			if err != nil {
				tt.Fatal(err)
			}
			ledger1, err := ps.CreateLedger(me.ID, "my first ledger")
			if err != nil {
				t.Fatal(err)
			}
			ledger2, err := ps.CreateLedger(otherTenant.ID, "other tenant's first ledger")
			if err != nil {
				t.Fatal(err)
			}

			myLedger, err := ps.GetLedger(me.ID, ledger1.ID)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "my first ledger", myLedger.Name)

			otherLedger, err := ps.GetLedger(me.ID, ledger2.ID)
			assert.Nil(tt, otherLedger)
			assert.Equal(tt, "Ledger not found.", err.Error())
		})
	})
}
