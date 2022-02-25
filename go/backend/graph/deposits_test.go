package graph

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"gitlab.com/fynbos/backend/graph/generated"
	_user "gitlab.com/fynbos/backend/user"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
)

func TestUserDeposits(s *testing.T) {
	ctx := context.Background()
	container, err := NewTestContainer(ctx, s)
	if err != nil {
		s.Fatal(err)
	}

	s.Cleanup(func() {
		container.Cleanup(ctx)
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
		_, err := NewIdentity(container, user, generateIdentityInput())
		if err != nil {
			t.Fatal(err)
		}
		verifyFundingSource := true
		fsArgs := generateLinkUsdBankAccountInput()
		fundingSource, err := NewLinkedUsdBankAccount(
			container,
			user,
			fsArgs,
			verifyFundingSource,
		)
		if err != nil {
			t.Fatal(err)
		}
		container.MockPacioliClient.EXPECT().GetAccount(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, args *pacioliv1.GetAccountRequest, opts ...grpc.CallOption) (*pacioliv1.Account, error) {
				return &pacioliv1.Account{
					Id: args.Id,
				}, nil
			}).Times(3)
		container.MockPacioliClient.EXPECT().CreateTransfers(gomock.Any(), gomock.Any()).Return(
			&pacioliv1.CreateTransfersResponse{
				Errors: []*pacioliv1.EventError{},
			}, nil).Times(1)

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
		assert.Equal(t, "10000", response.Transaction.Amount)     // TODO: where do we pretty print?
		assert.NotEqual(t, 0, response.Transaction.Timestamp)     // TODO: format of timestamp
		assert.Equal(t, "completed", response.Transaction.Status) // TODO: status definitions
		assert.Equal(t, "Deposit from "+fsArgs.Name, response.Transaction.Description)
		assert.Equal(t, generated.TransactionTypeDeposit, response.Transaction.Type)
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
		_, err := NewIdentity(container, user, generateIdentityInput())
		if err != nil {
			t.Fatal(err)
		}
		verifyFundingSource := false
		fsArgs := generateLinkUsdBankAccountInput()
		fundingSource, err := NewLinkedUsdBankAccount(
			container,
			user,
			fsArgs,
			verifyFundingSource,
		)
		if err != nil {
			t.Fatal(err)
		}

		container.MockPacioliClient.EXPECT().GetAccount(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, args *pacioliv1.GetAccountRequest, opts ...grpc.CallOption) (*pacioliv1.Account, error) {
				return &pacioliv1.Account{
					Id: args.Id,
				}, nil
			}).Times(1)
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
		assert.Nil(t, response.Transaction)
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
			            transaction {
			            	id
			            	type
			            	description
			            	amount
			            	timestamp
			            	status
			            }
			        }
			    }
			`)
	req.Var("input", input)
	_user.ActingAs(req, user)
	var data map[string]generated.DepositMutationResponse
	if err := container.Client.Run(container.Ctx, req, &data); err != nil {
		return nil, err
	}

	ret := data["initiateDeposit"]

	return &ret, nil
}
