package graph

import (
	"context"
	"fmt"
	account_transactions "gitlab.com/fynbos/backend/accounttransactions"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"

	"gitlab.com/fynbos/backend/graph/generated"
	"gitlab.com/fynbos/backend/onboarding"
	_user "gitlab.com/fynbos/backend/user"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
)

func TestUserAccount(s *testing.T) {
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
		acc, err := NewAccount(container, &onboarding.CreateAccountArgs{
			IdentityID:   user.ID,
			Email:        user.Email,
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			MobileNumber: faker.E164PhoneNumber(),
			Country:      "US",
		})
		if err != nil {
			t.Fatal(err)
		}
		transferResponse, err := container.PacioliClient.CreateTransfers(ctx, &pacioliv1.CreateTransfersRequest{
			Transfers: []*pacioliv1.Transfer{
				{
					Id:              uuid.NewString(),
					DebitAccountId:  acc.LedgerAccountID,
					CreditAccountId: container.NoopService.GetEquityAccountID(),
					Amount:          200,
				},
				{
					Id:              uuid.NewString(),
					DebitAccountId:  container.NoopService.GetEquityAccountID(),
					CreditAccountId: acc.LedgerAccountID,
					Amount:          100,
				},
				{
					Id:              uuid.NewString(),
					DebitAccountId:  container.NoopService.GetEquityAccountID(),
					CreditAccountId: acc.LedgerAccountID,
					Amount:          20,

					Flags: &pacioliv1.TransferFlags{
						TwoPhaseCommit: true,
					},
					Timeout: uint64(10 * time.Millisecond),
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(transferResponse.GetErrors()) != 0 {
			fmt.Println()
			fmt.Printf("%+v", transferResponse.GetErrors())
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
		acc, err := NewAccount(container, &onboarding.CreateAccountArgs{
			IdentityID:   user.ID,
			Email:        user.Email,
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			MobileNumber: faker.E164PhoneNumber(),
			Country:      "US",
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
					// Code: uint16,
					Flags: account_transactions.LedgerTransferFlags{},
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
