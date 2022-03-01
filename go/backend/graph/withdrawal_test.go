package graph

import (
	"context"
	"errors"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/graph/generated"
	"gitlab.com/fynbos/backend/onboarding"
	_user "gitlab.com/fynbos/backend/user"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
	tb_types "gitlab.com/fynbos/tigerbeetle_go/pkg/types"
	"google.golang.org/grpc"
)

func TestUserWithdrawals(s *testing.T) {
	ctx := context.Background()
	container, err := NewTestContainer(ctx, s)
	if err != nil {
		s.Fatal(err)
	}

	s.Cleanup(func() {
		container.Cleanup(ctx)
	})

	/*
		Scenario: user initiates withdrawal to a verified bank account
		Given a verified user
		And the user's verified usd bank account
		And the user has sufficient balance
		When the user initiates a withdrawal to the verified account
		Then a transaction is created and returned
	*/
	s.Run("user initiates withdrawal to a verified bank account", func(t *testing.T) {
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}
		_, err := NewAccount(container, &onboarding.CreateAccountArgs{
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
		deposit, err := NewDeposit(container, user, &generated.DepositInput{
			FundingSourceID: fundingSource.ID,
			Amount:          "10000",
		})
		if err != nil {
			t.Fatal(err)
		}
		if deposit.Type != generated.TransactionTypeDeposit && deposit.Status != "completed" {
			t.Fatal("Test expects deposit status to be 'completed'.")
		}

		container.MockPacioliClient.EXPECT().GetAccount(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, args *pacioliv1.GetAccountRequest, opts ...grpc.CallOption) (*pacioliv1.Account, error) {
				return &pacioliv1.Account{
					Id: args.Id,
				}, nil
			}).Times(2)
		container.MockPacioliClient.EXPECT().CreateTransfers(gomock.Any(), gomock.Any()).Return(
			&pacioliv1.CreateTransfersResponse{
				Errors: []*pacioliv1.EventError{},
			}, nil).Times(1)

		response, err := initiateWithdrawal(container, user, &generated.WithdrawalInput{
			FundingSourceID: fundingSource.ID,
			Amount:          "10000", // 100 dollars
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "200", response.Code)
		assert.Equal(t, true, response.Success)
		assert.Equal(t, "Withdrawal initiated.", response.Message)
		assert.Equal(t, "10000", response.Transaction.Amount)     // TODO: where do we pretty print?
		assert.NotEqual(t, 0, response.Transaction.Timestamp)     // TODO: format of timestamp
		assert.Equal(t, "completed", response.Transaction.Status) // TODO: status definitions
		assert.Equal(t, "Withdrawal to "+fsArgs.Name, response.Transaction.Description)
		assert.Equal(t, generated.TransactionTypeWithdrawal, response.Transaction.Type)
	})

	/*
		Scenario: user has insufficient balance to initiate withdrawal
		Given a verified user
		And the user's verified usd bank account
		And the user does not have a sufficient balance
		When the user initiates a withdrawal to the verified account
		Then an error is returned saying that there is insufficient balance
	*/
	s.Run("user has insufficient balance to initiate withdrawal", func(t *testing.T) {
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}
		_, err := NewAccount(container, &onboarding.CreateAccountArgs{
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
			}).Times(2)
		container.MockPacioliClient.EXPECT().CreateTransfers(gomock.Any(), gomock.Any()).Return(
			&pacioliv1.CreateTransfersResponse{
				Errors: []*pacioliv1.EventError{
					{
						Index: 0,
						Code:  tb_types.TransferExceedsCredits,
					},
				},
			}, nil).Times(1)

		response, err := initiateWithdrawal(container, user, &generated.WithdrawalInput{
			FundingSourceID: fundingSource.ID,
			Amount:          "10000", // 100 dollars
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "500", response.Code)
		assert.Equal(t, false, response.Success)
		assert.Equal(t, "Withdrawal failed: Insufficient balance.", response.Message)
		assert.Nil(t, response.Transaction)
	})

	/*
		Scenario: user tries to initiate withdrawal to unverified bank account
		Given a verified user
		And the user's unverified usd bank account
		And there is sufficient funds in the user's account to withdraw
		When the user initiates a withdrawal to the unverified account
		Then an error is returned saying that the bank account is unverified
	*/
	s.Run("user tries to initiate withdrawal to unverified bank account", func(t *testing.T) {
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}
		_, err := NewAccount(container, &onboarding.CreateAccountArgs{
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

		response, err := initiateWithdrawal(container, user, &generated.WithdrawalInput{
			FundingSourceID: fundingSource.ID,
			Amount:          "10000", // 100 dollars
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "403", response.Code)
		assert.Equal(t, false, response.Success)
		assert.Equal(t, "Withdrawal failed: Destination is unverified.", response.Message)
		assert.Nil(t, response.Transaction)
	})

	/*
		Scenario: alice tries to withdraw to bob's bank account
		Given alice is a verified user
		And bob's verified bank account
		And alice has sufficient funds in her account to withdraw
		When alice initiates a withdrawal to bob's bank account
		Then an error is returned saying that the bank account is not found

		oso recommends returning a not found rather than an unauthorized/forbidden.
	*/
	s.Run("alice tries to withdraw to bob's bank account", func(t *testing.T) {
		alice := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}
		bob := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}
		_, err := NewAccount(container, &onboarding.CreateAccountArgs{
			IdentityID:   alice.ID,
			Email:        alice.Email,
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			MobileNumber: faker.E164PhoneNumber(),
			Country:      "US",
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = NewAccount(container, &onboarding.CreateAccountArgs{
			IdentityID:   bob.ID,
			Email:        bob.Email,
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			MobileNumber: faker.E164PhoneNumber(),
			Country:      "US",
		})
		if err != nil {
			t.Fatal(err)
		}
		verifyFundingSource := true
		aliceBankArgs := generateLinkUsdBankAccountInput()
		aliceBankAccount, err := NewLinkedUsdBankAccount(
			container,
			alice,
			aliceBankArgs,
			verifyFundingSource,
		)
		if err != nil {
			t.Fatal(err)
		}
		deposit, err := NewDeposit(container, alice, &generated.DepositInput{
			FundingSourceID: aliceBankAccount.ID,
			Amount:          "10000",
		})
		if err != nil {
			t.Fatal(err)
		}
		if deposit.Status != "completed" && deposit.Type != generated.TransactionTypeDeposit {
			t.Fatal("Test expects a successful deposit into alice's bank account.")
		}
		bobBankArgs := generateLinkUsdBankAccountInput()
		bobBankAccount, err := NewLinkedUsdBankAccount(
			container,
			bob,
			bobBankArgs,
			verifyFundingSource,
		)
		if err != nil {
			t.Fatal(err)
		}

		response, err := initiateWithdrawal(container, alice, &generated.WithdrawalInput{
			FundingSourceID: bobBankAccount.ID,
			Amount:          "10000", // 100 dollars
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "404", response.Code)
		assert.Equal(t, false, response.Success)
		assert.Equal(t, "Withdrawal failed: Destination not found.", response.Message)
		assert.Nil(t, response.Transaction)
	})
}

func initiateWithdrawal(container *TestContainer, user *_user.User, input *generated.WithdrawalInput) (*generated.WithdrawalMutationResponse, error) {
	req := graphql.NewRequest(`
			    mutation ($input: WithdrawalInput!) {
			        initiateWithdrawal (input: $input) {
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
	var data map[string]generated.WithdrawalMutationResponse
	if err := container.Client.Run(container.Ctx, req, &data); err != nil {
		return nil, err
	}

	ret := data["initiateWithdrawal"]

	return &ret, nil
}

func NewDeposit(container *TestContainer, user *_user.User, input *generated.DepositInput) (*generated.Transaction, error) {
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

	response, err := initiateDeposit(container, user, input)
	if err != nil {
		return nil, err
	}

	if response.Code != "200" && response.Success {
		return nil, errors.New("Failed to initiate deposit.")
	}

	return response.Transaction, nil
}
