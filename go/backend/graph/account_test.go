package graph

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gitlab.com/fynbos/pacioli"

	account_transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/identity"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"

	"gitlab.com/fynbos/backend/graph/generated"
	"gitlab.com/fynbos/backend/onboarding"
	_user "gitlab.com/fynbos/backend/user"
)

func TestUserAccount(s *testing.T) {
	s.Skip("being deprecated")
	ctx := context.Background()
	container, err := NewTestContainer(ctx, s)
	if err != nil {
		s.Fatal(err)
	}

	s.Cleanup(func() {
		err = container.Cleanup(ctx)
		if err != nil {
			s.Fatal(err)
		}
	})

	/*
		Scenario: user queries their account balance
	*/
	s.Run("user can get their account balance", func(t *testing.T) {
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}

		ledgerID := uint32(1)

		id, err := NewIdentity(container, &identity.CreateArgs{
			ID:           user.ID,
			FirstName:    faker.FirstName(),
			LastName:     faker.FirstName(),
			MobileNumber: faker.E164PhoneNumber(),
			Email:        user.Email,
			Country:      "US",
		})
		if err != nil {
			t.Fatal(err)
		}

		acc, err := NewAccount(container, &onboarding.CreateAccountArgs{
			IdentityID: id.ID,
			Country:    id.Country,
		})
		if err != nil {
			t.Fatal(err)
		}

		transferErrs, err := container.PacioliClient.CreateTransfers(ctx, []pacioli.CreateTransferArgs{
			{
				ID:              uuid.NewString(),
				DebitAccountID:  acc.LedgerAccountID,
				CreditAccountID: container.NoopService.GetEquityAccountID(),
				Amount:          200,
				Code:            1,
				Ledger:          ledgerID,
			},
			{
				ID:              uuid.NewString(),
				DebitAccountID:  container.NoopService.GetEquityAccountID(),
				CreditAccountID: acc.LedgerAccountID,
				Amount:          100,
				Code:            1,
				Ledger:          ledgerID,
			},
			{
				ID:              uuid.NewString(),
				DebitAccountID:  container.NoopService.GetEquityAccountID(),
				CreditAccountID: acc.LedgerAccountID,
				Amount:          20,
				Code:            1,
				Flags: pacioli.TransferFlags{
					Pending: true,
				},
				Timeout: uint64(10 * time.Millisecond),
				Ledger:  ledgerID,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(transferErrs) != 0 {
			fmt.Println()
			fmt.Printf("%+v", transferErrs)
			fmt.Println()
			t.Fatal("Failed to create transfers in pacioli.")
		}

		req := getAccountRequest()
		err = _user.ActingAs(req, user)
		if err != nil {
			t.Fatal(err)
		}
		var data map[string]generated.Account
		if err := container.Client.Run(container.Ctx, req, &data); err != nil {
			t.Fatal(err)
		}

		response := data["account"]
		assert.Equal(t, "$ 0.80", response.Balance)
		assert.NotNil(t, response.ID)
	})

	/*
		Scenario: user can get their recent transactions
	*/
	s.Run("user can get their recent transactions", func(t *testing.T) {
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}
		id, err := NewIdentity(container, &identity.CreateArgs{
			ID:           user.ID,
			FirstName:    faker.FirstName(),
			LastName:     faker.FirstName(),
			MobileNumber: faker.E164PhoneNumber(),
			Email:        user.Email,
			Country:      "US",
		})
		if err != nil {
			t.Fatal(err)
		}
		acc, err := NewAccount(container, &onboarding.CreateAccountArgs{
			IdentityID: id.ID,
			Country:    id.Country,
		})
		if err != nil {
			t.Fatal(err)
		}
		deposit, err := NewDeposit(container, &account_transactions.CreateTransactionArgs{
			AccountID:   acc.ID,
			Description: "Test transaction",
			Type:        "deposit",
			NetAmount:   10000,
			LedgerTransfers: []account_transactions.CreateLedgerTransferArgs{
				{
					LedgerID:        container.NoopService.GetLedgerID(),
					CreditAccountID: container.NoopService.GetEquityAccountID(),
					DebitAccountID:  acc.LedgerAccountID,
					Amount:          10000,
					Code:            1,
					Flags:           account_transactions.LedgerTransferFlags{},
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		req := getAccountTransactionsRequest()
		err = _user.ActingAs(req, user)
		if err != nil {
			t.Fatal(err)
		}
		var data map[string]generated.Account
		if err := container.Client.Run(container.Ctx, req, &data); err != nil {
			t.Fatal(err)
		}

		rTrxs := data["account"].RecentTransactions
		assert.Len(t, rTrxs, 1)
		firstTrx := rTrxs[0]
		assert.Equal(t, deposit.ID, firstTrx.ID)
		assert.Equal(t, "$ 100.00", firstTrx.Amount)
		assert.Equal(t, generated.TransactionTypeDeposit, firstTrx.Type)
	})
}

func getAccountRequest() *graphql.Request {
	return graphql.NewRequest(`
			    query {
			        account {
			            id
			            balance
			        }
			    }
			`)
}

func getAccountTransactionsRequest() *graphql.Request {
	return graphql.NewRequest(`
		query {
			account {
				recentTransactions {
					id
					amount
					type
				}
			}
		}
	`)
}
