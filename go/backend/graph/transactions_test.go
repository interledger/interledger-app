package graph

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	account_transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/graph/generated"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	_user "gitlab.com/fynbos/backend/user"
)

func TestTransactions(s *testing.T) {
	s.Skip("being deprecated")
	ctx := context.Background()
	c, err := NewTestContainer(ctx, s)
	if err != nil {
		s.Fatal(err)
	}

	s.Cleanup(func() {
		err = c.Cleanup(ctx)
		if err != nil {
			s.Fatal(err)
		}
	})

	user := &_user.User{
		ID:    uuid.NewString(),
		Email: faker.Email(),
	}
	id, err := NewIdentity(c, &identity.CreateArgs{
		ID:           user.ID,
		FirstName:    faker.FirstName(),
		LastName:     faker.LastName(),
		MobileNumber: faker.E164PhoneNumber(),
		Email:        user.Email,
		Country:      "US",
	})
	if err != nil {
		s.Fatal(err)
	}

	acc, err := NewVerifiedAccount(
		c,
		&onboarding.CreateAccountArgs{
			IdentityID: id.ID,
			Country:    id.Country,
		},
		&onboarding.VerifyAccountArgs{
			DateOfBirth: faker.Date(),
			Address:     []string{faker.Name()},
			State:       faker.Name(),
			City:        faker.Name(),
			PostalCode:  faker.CCNumber(),
			TaxIDNumber: faker.CCNumber(),
		},
	)
	if err != nil {
		s.Fatal(err)
	}

	otherUser := &_user.User{
		ID:    uuid.NewString(),
		Email: faker.Email(),
	}
	otherUserId, err := NewIdentity(c, &identity.CreateArgs{
		ID:           otherUser.ID,
		FirstName:    faker.FirstName(),
		LastName:     faker.LastName(),
		MobileNumber: faker.E164PhoneNumber(),
		Email:        otherUser.Email,
		Country:      "US",
	})
	if err != nil {
		s.Fatal(err)
	}

	_, err = NewVerifiedAccount(
		c,
		&onboarding.CreateAccountArgs{
			IdentityID: otherUser.ID,
			Country:    otherUserId.Country,
		},
		&onboarding.VerifyAccountArgs{
			DateOfBirth: faker.Date(),
			Address:     []string{faker.Name()},
			State:       faker.Name(),
			City:        faker.Name(),
			PostalCode:  faker.CCNumber(),
			TaxIDNumber: faker.CCNumber(),
		},
	)
	if err != nil {
		s.Fatal(err)
	}

	_, err = NewBankAccount(
		c,
		user,
		&fundingsources.CreateBankAccountArgs{
			IdentityID:    user.ID,
			AccountID:     acc.ID,
			Name:          faker.Name(),
			AccountNumber: faker.CCNumber(),
			RoutingNumber: faker.CCNumber(),
			Institution:   faker.Name(),
			Type:          "cheque",
		},
		true,
	)
	if err != nil {
		s.Fatal(err)
	}
	totalTransactions := 10
	i := 0
	transactionsAsc := make([]*account_transactions.AccountTransaction, totalTransactions)
	for i < totalTransactions {
		if i%2 == 0 {
			transactionsAsc[i], err = NewDeposit(c, &account_transactions.CreateTransactionArgs{
				AccountID:   acc.ID,
				Description: "Test transaction",
				Type:        "deposit",
				NetAmount:   1000,
				LedgerTransfers: []account_transactions.CreateLedgerTransferArgs{
					{
						LedgerID:        c.NoopService.GetLedgerID(),
						CreditAccountID: c.NoopService.GetEquityAccountID(),
						DebitAccountID:  acc.LedgerAccountID,
						Amount:          1000,
						Code:            1,
						Flags:           account_transactions.LedgerTransferFlags{},
					},
				},
			})
		} else {
			transactionsAsc[i], err = NewOutgoingPayment(
				c, &account_transactions.CreateTransactionArgs{
					AccountID:   acc.ID,
					Description: "Sent to $test.fynbos.dev/alice",
					Type:        "outgoingPayment",
					NetAmount:   100,
					LedgerTransfers: []account_transactions.CreateLedgerTransferArgs{
						{
							LedgerID:        c.NoopService.GetLedgerID(),
							DebitAccountID:  c.NoopService.GetEquityAccountID(),
							CreditAccountID: acc.LedgerAccountID,
							Amount:          100,
							Code:            1,
							Flags:           account_transactions.LedgerTransferFlags{},
						},
					},
				},
			)
		}
		if err != nil {
			s.Fatal(err)
		}
		i++
	}

	/*
		Scenario: user has transactions and can paginate through them.
		Given a user
		And transactions on the users account
		When the user requests a page of transactions
		Then the page should be returned according to https://relay.dev/graphql/connections.html
	*/
	s.Run("user can get a page of account transactions", func(t *testing.T) {
		type scenario struct {
			Name                 string
			After                string
			First                int
			ExpectedTransactions []*account_transactions.AccountTransaction
			ExpectedHasNextPage  bool
			ExpectedStartCursor  string
			ExpectedEndCursor    string
		}
		scenarios := []scenario{
			{
				Name:                 "Can get first 5 transactions",
				First:                5,
				ExpectedHasNextPage:  true,
				ExpectedTransactions: transactionsAsc[totalTransactions-5:],

				// query will return in DESC order so expect the 5 latest transactions
				ExpectedStartCursor: transactionsAsc[totalTransactions-1].ID,
				ExpectedEndCursor:   transactionsAsc[totalTransactions-5].ID,
			},
			{
				Name:                 "Can get all transactions",
				First:                totalTransactions,
				ExpectedTransactions: transactionsAsc,
				ExpectedHasNextPage:  false,
				ExpectedStartCursor:  transactionsAsc[totalTransactions-1].ID,
				ExpectedEndCursor:    transactionsAsc[0].ID,
			},
			{
				Name:                 "Can get transactions after a specified one",
				After:                transactionsAsc[2].ID,
				ExpectedHasNextPage:  false,
				First:                5,
				ExpectedTransactions: transactionsAsc[:2],
				// query will return in DESC order so expect the first 2 that were created
				ExpectedStartCursor: transactionsAsc[1].ID,
				ExpectedEndCursor:   transactionsAsc[0].ID,
			},
		}

		for _, scenario := range scenarios {
			response, err := getTransactions(c, user, &generated.Pagination{
				First: scenario.First,
				After: scenario.After,
			})
			if err != nil {
				t.Fatal(err)
			}

			assert.Len(t, response.Edges, len(scenario.ExpectedTransactions))
			for i, edge := range response.Edges {
				// query will return it in DESC order.
				trx := scenario.ExpectedTransactions[len(scenario.ExpectedTransactions)-1-i]
				assert.Equal(t, trx.ID, edge.Node.ID, scenario.Name)
				assert.Equal(t, trx.ID, edge.Cursor, scenario.Name)
				assert.Equal(
					t,
					fmt.Sprintf("$ %.2f", float64(trx.NetAmount)/float64(100)),
					edge.Node.Amount,
					scenario.Name,
				)
				assert.Equal(t, trx.State.String(), edge.Node.Status, scenario.Name)
				assert.Equal(t, trx.Description, edge.Node.Description, scenario.Name)
				assert.Equal(t, trx.CreatedAt, edge.Node.Timestamp, scenario.Name)
				assert.Equal(
					t,
					generated.TransactionType(strings.ToUpper(trx.Type)),
					edge.Node.Type,
					scenario.Name,
				)
			}
			assert.Equal(t, scenario.ExpectedStartCursor, response.PageInfo.StartCursor, scenario.Name)
			assert.Equal(t, scenario.ExpectedEndCursor, response.PageInfo.EndCursor, scenario.Name)
			assert.Equal(t, scenario.ExpectedHasNextPage, response.PageInfo.HasNextPage, scenario.Name)
		}
	})

	s.Run("unauthenticated request to fetch a page of transactions is forbidden", func(t *testing.T) {
		response, err := getTransactions(c, nil, &generated.Pagination{})
		if err == nil {
			t.Fatal("Unauthenticated requests must be forbidden")
		}

		assert.Nil(t, response)
		assert.Error(t, err)
	})

	s.Run("user can only get their own transactions", func(t *testing.T) {
		response, err := getTransactions(c, otherUser, &generated.Pagination{
			First: 5,
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, response.Edges, 0)
	})

	/*
		Scenario: authenticated user can get a detailed view of a transaction
		Given an authenticated user
		And the user has transactions on their account
		When the user gets the transaction by its id
		Then the transaction is returned
	*/
	s.Run("user can get a detailed view of a transaction", func(t *testing.T) {
		transaction, err := getTransaction(c, user, transactionsAsc[0].ID)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, transactionsAsc[0].ID, transaction.ID)
		assert.Equal(t, transactionsAsc[0].Description, transaction.Description)
		assert.Equal(
			t,
			generated.TransactionType(strings.ToUpper(transactionsAsc[0].Type)),
			transaction.Type,
		)
		assert.Equal(t, transactionsAsc[0].CreatedAt, transaction.Timestamp)
		assert.Equal(t, transactionsAsc[0].State.String(), transaction.Status)
		assert.Equal(
			t,
			fmt.Sprintf("$ %.2f", float64(transactionsAsc[0].NetAmount)/float64(100)),
			transaction.Amount,
		)
	})

	/*
		Scenario: authenticated user can only get a detailed view of a transaction that is on
		 their own account.
		Given an authenticated user
		AND a transaction exists for another user's account
		When the authenticated user gets the transaction
		Then the graphql not found error is returned
	*/
	s.Run("user can only get a detailed view of their own transaction", func(t *testing.T) {
		transaction, err := getTransaction(c, otherUser, transactionsAsc[0].ID)
		if err == nil {
			t.Fatal("user must only be able to get their own transaction.")
		}

		assert.Nil(t, transaction)
		assert.Equal(t, "graphql: Not found.", err.Error())
	})

	/*
		Scenario: authenticated user can not look up a non-existent transaction
		Given an authenticated user
		AND the user has transactions on their account
		When the user gets a transaction that does not exist
		Then the graphql not found error is returned
	*/
	s.Run(
		"not found is returned to user when they try to get a non-existent transaction",
		func(t *testing.T) {
			transaction, err := getTransaction(c, user, uuid.NewString())
			if err == nil {
				t.Fatal("not found error must be returned when user looks up non-existent transaction.")
			}

			assert.Nil(t, transaction)
			assert.Equal(t, "graphql: Not found.", err.Error())
		},
	)

	/*
		Scenario: unauthenticated requests to get a detailed view of a transaction
		Given a transaction
		When an unauthenticated request is made to get it
		Then the graphql not found error is returned
	*/
	s.Run("unauthenticated request to get a transaction is forbidden", func(t *testing.T) {
		transaction, err := getTransaction(c, nil, transactionsAsc[0].ID)
		if err == nil {
			t.Fatal("unauthenticated request to get a transaction is forbidden.")
		}

		assert.Nil(t, transaction)
		assert.Equal(t, "graphql: Forbidden.", err.Error())
	})

	/*
		Scenario: malicious authenticated user tries to inject sql to look up a transaction
		Given an authenticated user
		When the user tries to make the WHERE clause true
		Then the graphql internal error is returned
	*/
	s.Run("user can not run an sql injection attack to retrieve a transaction", func(t *testing.T) {
		transaction, err := getTransaction(c, user, "1=1")
		if err == nil {
			t.Fatal("user must not be able to run an sql injection attack to get a transaction.")
		}

		assert.Nil(t, transaction)
		assert.Equal(t, "graphql: Unable to process request.", err.Error())
	})
}

func getTransactions(
	container *TestContainer,
	user *_user.User,
	input *generated.Pagination,
) (*generated.TransactionsConnection, error) {
	req := graphql.NewRequest(`
			    query ($input: Pagination!) {
			        transactions (input: $input) {
			            pageInfo {
			            	hasNextPage
			            	startCursor
			            	endCursor
			            }
			            edges {
			            	node {
			            		id
			            		type
			            		description
			            		status
			            		timestamp
			            		amount
			            	}
			            	cursor
			            }
			        }
			    }
			`)
	req.Var("input", input)
	err := _user.ActingAs(req, user)
	if err != nil {
		return nil, err
	}
	var data map[string]generated.TransactionsConnection
	if err := container.Client.Run(container.Ctx, req, &data); err != nil {
		return nil, err
	}

	response := data["transactions"]

	return &response, nil
}

func getTransaction(container *TestContainer, user *_user.User, id string) (
	*generated.Transaction,
	error,
) {
	req := graphql.NewRequest(`
			    query ($id: String!) {
			        transaction (id: $id) {
			            id
	            		type
	            		description
	            		status
	            		timestamp
	            		amount
			        }
			    }
			`)
	req.Var("id", id)
	err := _user.ActingAs(req, user)
	if err != nil {
		return nil, err
	}
	var data map[string]generated.Transaction
	if err := container.Client.Run(container.Ctx, req, &data); err != nil {
		return nil, err
	}

	response := data["transaction"]

	return &response, nil
}
