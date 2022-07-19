package graph

import (
	"context"
	"testing"

	"gitlab.com/fynbos/backend/deposits/ops"

	"gitlab.com/fynbos/backend/identity"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"

	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/graph/generated"
	"gitlab.com/fynbos/backend/onboarding"
	_user "gitlab.com/fynbos/backend/user"
)

func TestUserDeposits(s *testing.T) {
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
		Scenario: user initiates deposit from a USD bank account
		Given a verified user
		And the user's verified usd bank account
		When the user initiates a deposit from the verified account
		Then a transaction is created and returned
	*/
	s.Run("user initiates deposit from a verified USD bank account", func(t *testing.T) {
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

		verifyFundingSource := true
		fsArgs := &fundingsources.CreateBankAccountArgs{
			IdentityID:    user.ID,
			AccountID:     acc.ID,
			Name:          faker.Name(),
			AccountNumber: faker.CCNumber(),
			RoutingNumber: faker.CCNumber(),
			Institution:   faker.Name(),
			Type:          "cheque",
		}
		fundingSource, err := NewBankAccount(
			container,
			user,
			fsArgs,
			verifyFundingSource,
		)
		if err != nil {
			t.Fatal(err)
		}

		response, err := initiateDeposit(container, user, &generated.DepositInput{
			FundingSourceID: fundingSource.ID,
			Amount:          "10000", // 100 dollars
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "200", response.Code)
		assert.Equal(t, true, response.Success)
		assert.Equal(t, "Deposit initiated.", response.Message)
		assert.Equal(t, "10000", response.Deposit.Amount) // TODO: where do we pretty print?
		assert.Equal(t, ops.Created.String(), response.Deposit.State)
	})

	/*
		Scenario: user initiates deposit from an unverified USD bank account
		Given a verified user
		And the user's unverified usd bank account
		When the user initiates a deposit from the unverified account
		Then an error is returned saying that the bank account is unverified
	*/
	s.Run("user cannot initiate deposit from unverified bank account", func(t *testing.T) {
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

		verifyFundingSource := false
		fundingSource, err := NewBankAccount(
			container,
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
			verifyFundingSource,
		)
		if err != nil {
			t.Fatal(err)
		}

		response, err := initiateDeposit(container, user, &generated.DepositInput{
			FundingSourceID: fundingSource.ID,
			Amount:          "10000", // 100 dollars
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "403", response.Code)
		assert.Equal(t, false, response.Success)
		assert.Equal(t, "Deposit failed: Funding source is not verified.", response.Message)
		assert.Nil(t, response.Deposit)
	})

	// TODO: does the user need to be verified to allow deposits?
}

func initiateDeposit(container *TestContainer, user *_user.User, input *generated.DepositInput) (*generated.DepositMutationResponse, error) {
	req := graphql.NewRequest(`
			    mutation ($input: DepositInput!) {
			        initiateDeposit (input: $input) {
			            code
			            success
			            message
			            deposit {
			            	id
			            	amount
			            	timestamp
			            	state
			            }
			        }
			    }
			`)
	req.Var("input", input)
	err := _user.ActingAs(req, user)
	if err != nil {
		return nil, err
	}
	var data map[string]generated.DepositMutationResponse
	if err := container.Client.Run(container.Ctx, req, &data); err != nil {
		return nil, err
	}

	ret := data["initiateDeposit"]

	return &ret, nil
}
