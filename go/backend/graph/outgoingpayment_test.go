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
	"gitlab.com/fynbos/backend/deposits"
	"gitlab.com/fynbos/backend/fundingsources"
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

	container.MockPacioliClient.EXPECT().ConfigureAccounts(gomock.Any(), gomock.Any()).Return(
		&pacioliv1.ConfigureAccountsResponse{}, nil,
	).AnyTimes()
	container.MockPacioliClient.EXPECT().GetAccounts(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, args *pacioliv1.GetAccountsRequest, opts ...grpc.CallOption) (*pacioliv1.GetAccountsResponse, error) {
			return &pacioliv1.GetAccountsResponse{
				Accounts: []*pacioliv1.Account{{Id: args.GetIds()[0]}},
			}, nil
		}).AnyTimes()

	/*
		Scenario: user initiates an outgoing payment to an ilp address

		Given a verified user
		And the user's account has sufficient balance
		When the user initiates an outgoing payment
		Then noop provider creates the transaction
		Then a successful response is returned along with the newly created transaction
	*/
	s.Run("user initiates an outgoing payment to an ilp address", func(t *testing.T) {
		container.MockPacioliClient.EXPECT().CreateTransfers(gomock.Any(), gomock.Any()).Return(
			&pacioliv1.CreateTransfersResponse{
				Errors: []*pacioliv1.EventError{},
			}, nil).Times(2)
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}
		acc, err := NewVerifiedAccount(
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
		bankAccount, err := NewBankAccount(
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
			true,
		)
		if err != nil {
			t.Fatal(err)
		}
		deposit, err := NewDeposit(container, &deposits.InitiateDepositArgs{
			IdentityID:      user.ID,
			AccountID:       acc.ID,
			FundingSourceID: bankAccount.ID,
			Amount:          10000, // 100 dollars
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "completed", deposit.State)

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
		container.MockPacioliClient.EXPECT().CreateTransfers(gomock.Any(), gomock.Any()).Return(
			&pacioliv1.CreateTransfersResponse{
				Errors: []*pacioliv1.EventError{
					{Index: 0, Code: tb_types.TransferExceedsCredits},
				},
			}, nil).Times(1)
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}
		acc, err := NewVerifiedAccount(
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
		assert.True(t, acc.IsVerified())

		response, err := InitiateOutgoingPayment(container, user, &generated.OutgoingPaymentInput{
			Amount: "10000",
			To:     "$test.fynbos.test/alice",
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "422", response.Code)
		assert.Equal(t, false, response.Success)
		assert.Equal(t, "Outgoing payment failed: Insufficient balance.", response.Message)
		assert.Nil(t, response.Transaction)
	})

	/*
		Outgoing payments aren't allowed unless account is verified
		Scenario: user tries to initiate outgoing payment without verifying their identity

		Given a verified user
		And the user has sufficient balance
		And the account has not been verified
		When the user initiates an outgoing payment
		Then an error is returned saying there is insufficient balance
	*/
	s.Run("user tries to initiate outgoing payment from an unverified account", func(t *testing.T) {
		container.MockPacioliClient.EXPECT().CreateTransfers(gomock.Any(), gomock.Any()).Return(
			&pacioliv1.CreateTransfersResponse{
				Errors: []*pacioliv1.EventError{},
			}, nil).Times(1)
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}
		acc, err := NewAccount(
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
		assert.False(t, acc.IsVerified())
		bankAccount, err := NewBankAccount(
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
			true,
		)
		if err != nil {
			t.Fatal(err)
		}
		deposit, err := NewDeposit(container, &deposits.InitiateDepositArgs{
			IdentityID:      user.ID,
			AccountID:       acc.ID,
			FundingSourceID: bankAccount.ID,
			Amount:          10000, // 100 dollars
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "completed", deposit.State)

		response, err := InitiateOutgoingPayment(container, user, &generated.OutgoingPaymentInput{
			Amount: "10000",
			To:     "$test.fynbos.test/alice",
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "403", response.Code)
		assert.Equal(t, false, response.Success)
		assert.Equal(t, "Outgoing payment failed: Account unverified.", response.Message)
		assert.Nil(t, response.Transaction)
	})
}

func InitiateOutgoingPayment(
	container *TestContainer,
	user *_user.User,
	input *generated.OutgoingPaymentInput,
) (*generated.OutgoingPaymentMutationResponse, error) {
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
	err := _user.ActingAs(req, user)
	if err != nil {
		return nil, err
	}
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
