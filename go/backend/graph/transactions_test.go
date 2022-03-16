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
	"gitlab.com/fynbos/backend/deposits"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/graph/generated"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/payments"
	_user "gitlab.com/fynbos/backend/user"
)

func TestTransactions(s *testing.T) {
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
	acc, err := NewVerifiedAccount(
		c,
		&onboarding.CreateAccountArgs{
			IdentityID:   user.ID,
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			MobileNumber: faker.E164PhoneNumber(),
			Email:        user.Email,
			Country:      "US",
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
	bankAccount, err := NewBankAccount(
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
			transactionsAsc[i], err = NewDeposit(c, &deposits.InitiateDepositArgs{
				AccountID:       acc.ID,
				IdentityID:      user.ID,
				FundingSourceID: bankAccount.ID,
				Amount:          1000,
			})
		} else {
			transactionsAsc[i], err = c.Ps.InitiateOutgoingPayment(ctx, &payments.InitiateOutgoingPaymentArgs{
				IdentityID: user.ID,
				AccountID:  acc.ID,
				Amount:     100,
				To:         "$test.fynbos.dev/alice",
			})
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
				trx := scenario.ExpectedTransactions[len(scenario.ExpectedTransactions)-1-i] // query will return it in DESC order.
				assert.Equal(t, trx.ID, edge.Node.ID, scenario.Name)
				assert.Equal(t, trx.ID, edge.Cursor, scenario.Name)
				assert.Equal(t, fmt.Sprintf("$ %.2f", float64(trx.NetAmount)/float64(100)), edge.Node.Amount, scenario.Name)
				assert.Equal(t, trx.State, edge.Node.Status, scenario.Name)
				assert.Equal(t, trx.Description, edge.Node.Description, scenario.Name)
				assert.Equal(t, trx.CreatedAt, edge.Node.Timestamp, scenario.Name)
				assert.Equal(t, generated.TransactionType(strings.ToUpper(trx.Type)), edge.Node.Type, scenario.Name)
			}
			assert.Equal(t, scenario.ExpectedStartCursor, response.PageInfo.StartCursor, scenario.Name)
			assert.Equal(t, scenario.ExpectedEndCursor, response.PageInfo.EndCursor, scenario.Name)
			assert.Equal(t, scenario.ExpectedHasNextPage, response.PageInfo.HasNextPage, scenario.Name)
		}
	})

	s.Run("unauthenticated request is forbidden", func(t *testing.T) {
		response, err := getTransactions(c, nil, &generated.Pagination{})
		if err == nil {
			t.Fatal("Unauthenticated requests must be forbidden")
		}

		assert.Nil(t, response)
		assert.Error(t, err)
	})

	s.Run("user can only get their own transactions", func(t *testing.T) {
		otherUser := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}
		_, err = NewVerifiedAccount(
			c,
			&onboarding.CreateAccountArgs{
				IdentityID:   otherUser.ID,
				FirstName:    faker.FirstName(),
				LastName:     faker.LastName(),
				MobileNumber: faker.E164PhoneNumber(),
				Email:        user.Email,
				Country:      "US",
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
			t.Fatal(err)
		}

		response, err := getTransactions(c, otherUser, &generated.Pagination{
			First: 5,
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, response.Edges, 0)
	})
}

func getTransactions(container *TestContainer, user *_user.User, input *generated.Pagination) (*generated.TransactionsConnection, error) {
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
