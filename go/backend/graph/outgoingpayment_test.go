package graph

import (
	"context"
	"errors"
	"testing"

	account_transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/payments"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/graph/generated"
	"gitlab.com/fynbos/backend/onboarding"
	_user "gitlab.com/fynbos/backend/user"
)

func TestUserOutgoingPayment(s *testing.T) {
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

		acc, err := NewVerifiedAccount(
			container,
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
			t.Fatal(err)
		}

		_, err = NewDeposit(container, &account_transactions.CreateTransactionArgs{
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
		assert.Equal(t, "10000", response.OutgoingPayment.Amount)
		assert.Equal(t, "$test.fynbos.test/alice", response.OutgoingPayment.Destination)
		assert.Equal(t, payments.Created.String(), response.OutgoingPayment.State)
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
		acc, err := NewVerifiedAccount(
			container,
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
			t.Fatal(err)
		}

		assert.True(t, acc.IsVerified())
		assert.Equal(t, int64(0), acc.AvailableBalance)

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
		assert.Nil(t, response.OutgoingPayment)
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
		acc, err := NewAccount(
			container,
			&onboarding.CreateAccountArgs{
				IdentityID: id.ID,
				Country:    id.Country,
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, acc.IsVerified())

		_, err = NewDeposit(container, &account_transactions.CreateTransactionArgs{
			AccountID:   acc.ID,
			Description: "Test transaction",
			Type:        "deposit",
			NetAmount:   1000,
			LedgerTransfers: []account_transactions.CreateLedgerTransferArgs{
				{
					LedgerID:        container.NoopService.GetLedgerID(),
					CreditAccountID: container.NoopService.GetEquityAccountID(),
					DebitAccountID:  acc.LedgerAccountID,
					Amount:          1000,
					Code:            1,
					Flags:           account_transactions.LedgerTransferFlags{},
				},
			},
		})
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
		assert.Equal(t, "Outgoing payment failed: Account unverified.", response.Message)
		assert.Nil(t, response.OutgoingPayment)
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
			            outgoingPayment {
			            	id
			            	destination
			            	state
										amount
			            	timestamp
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
