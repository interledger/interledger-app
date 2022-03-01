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

func TestUserOutgoingPayment(s *testing.T) {
	ctx := context.Background()
	container, err := NewTestContainer(ctx, s)
	if err != nil {
		s.Fatal(err)
	}

	s.Cleanup(func() {
		container.Cleanup(ctx)
	})

	/*
		Scenario: user initiates an outgoing payment to an ilp address

		Given a verified user
		And the user's account has sufficient balance
		When the user initiates an outgoing payment
		Then noop provider creates the transaction
		Then a successful response is returned along with the newly created transaction
	*/
	s.Run("user initiates an outgoing payment to an ilp address", func(t *testing.T) {
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}
		_, err := NewVerifiedAccount(
			container,
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
			t.Fatal(err)
		}
		// TODO: assert account verification state
		err = FundUsersAccount(container, user, "10000") // 100 dollars
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
				Errors: []*pacioliv1.EventError{},
			}, nil).Times(1)
		response, err := InitiateOutgoingPayment(container, user, &generated.OutgoingPaymentInput{
			Amount: "10000",
			To:     "$test.fynbos.test/alice",
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "200", response.Code)
		assert.Equal(t, true, response.Success)
		assert.Equal(t, "Outgoing payment initiated.", response.Message)
		assert.Equal(t, generated.TransactionTypeSent, response.Transaction.Type)
		assert.Equal(t, "10000", response.Transaction.Amount)
		assert.Equal(t, "Sent to $test.fynbos.test/alice", response.Transaction.Description)
		assert.Equal(t, "completed", response.Transaction.Status)
	})

	/*
		Scenario: user has insufficient balance to initiate outgoing payment

		Given a verified user
		And the user does not have sufficient balance
		When the user initiates an outgoing payment
		Then an error is returned saying there is insufficient balance
	*/
	s.Run("user has insufficient balance to initiate outgoing payment", func(t *testing.T) {
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}
		_, err := NewVerifiedAccount(
			container,
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
			t.Fatal(err)
		}
		// TODO: assert account verification state

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
		response, err := InitiateOutgoingPayment(container, user, &generated.OutgoingPaymentInput{
			Amount: "10000",
			To:     "$test.fynbos.test/alice",
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "500", response.Code)
		assert.Equal(t, false, response.Success)
		assert.Equal(t, "Outgoing payment failed: Insufficient balance.", response.Message)
		assert.Nil(t, response.Transaction)
	})

	/*
		TODO: refactor so that account has verification
		Noop provider does not allow outgoing payments unless user's identity is verified
		Scenario: user tries to initiate outgoing payment without verifying their identity

		Given a verified user
		And the user has sufficient balance
		And the user's identity has not been verified
		When the user initiates an outgoing payment
		Then an error is returned saying there is insufficient balance
	*/
	s.Run("user tries to initiate outgoing payment without verifying their identity", func(t *testing.T) {
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}
		_, err := NewAccount(
			container,
			&onboarding.CreateAccountArgs{
				IdentityID:   user.ID,
				FirstName:    faker.FirstName(),
				LastName:     faker.LastName(),
				MobileNumber: faker.E164PhoneNumber(),
				Email:        user.Email,
				Country:      "US",
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		// TODO: assert account verification state
		err = FundUsersAccount(container, user, "10000") // 100 dollars
		if err != nil {
			t.Fatal(err)
		}

		response, err := InitiateOutgoingPayment(container, user, &generated.OutgoingPaymentInput{
			Amount: "10000",
			To:     "$test.fynbos.test/alice",
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "403", response.Code)
		assert.Equal(t, false, response.Success)
		assert.Equal(t, "Outgoing payment failed: Unverified identity.", response.Message)
		assert.Nil(t, response.Transaction)
	})
}

// Helper function to create an identity for the user as well fund their account.
// A funding source is created and verified to fund the account.
func FundUsersAccount(container *TestContainer, user *_user.User, amount string) error {
	verifyFundingSource := true
	bankAccountArgs := generateLinkUsdBankAccountInput()
	bankAccount, err := NewLinkedUsdBankAccount(
		container,
		user,
		bankAccountArgs,
		verifyFundingSource,
	)
	if err != nil {
		return err
	}

	deposit, err := NewDeposit(container, user, &generated.DepositInput{
		FundingSourceID: bankAccount.ID,
		Amount:          amount,
	})
	if err != nil {
		return err
	}
	if deposit.Status != "completed" && deposit.Type != generated.TransactionTypeDeposit {
		return errors.New("Test expected a successful deposit.")
	}

	return nil

}

func InitiateOutgoingPayment(container *TestContainer, user *_user.User, input *generated.OutgoingPaymentInput) (*generated.OutgoingPaymentMutationResponse, error) {
	req := graphql.NewRequest(`
			    mutation ($input: OutgoingPaymentInput!) {
			        initiateOutgoingPayment (input: $input) {
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
	var data map[string]generated.OutgoingPaymentMutationResponse
	if err := container.Client.Run(container.Ctx, req, &data); err != nil {
		return nil, err
	}

	response := data["initiateOutgoingPayment"]

	if response.Code != "200" && response.Success {
		return nil, errors.New("Failed to initiate outgoing payment.")
	}

	return &response, nil
}
