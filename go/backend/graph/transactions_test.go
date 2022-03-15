package graph

import (
	"context"
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
	i := 0
	transactions := make([]*account_transactions.AccountTransaction, 10)
	for i < 10 {
		if i%2 == 0 {
			transactions[i], err = NewDeposit(c, &deposits.InitiateDepositArgs{
				AccountID:       acc.ID,
				IdentityID:      user.ID,
				FundingSourceID: bankAccount.ID,
				Amount:          1000,
			})
		} else {
			transactions[i], err = c.Ps.InitiateOutgoingPayment(ctx, &payments.InitiateOutgoingPaymentArgs{
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

	s.Run("user can get a page of account transactions", func(t *testing.T) {
		response, err := getTransactions(c, user, &generated.Pagination{
			First: 10,
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, response.Edges, 1)
		assert.Equal(t, "aaaaa", response.Edges[0].Node.ID)
		assert.Equal(t, "aaaaa", response.Edges[0].Cursor)
		assert.False(t, response.PageInfo.HasNextPage)
		assert.Equal(t, "aaaaa", response.PageInfo.StartCursor)
		assert.Equal(t, "aaaaa", response.PageInfo.EndCursor)
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
