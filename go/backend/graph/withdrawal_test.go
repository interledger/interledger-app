package graph

import (
	"context"
	"testing"

	account_transactions "gitlab.com/fynbos/backend/accounttransactions"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	"gitlab.com/fynbos/backend/withdrawals"

	"gitlab.com/fynbos/backend/identity"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/graph/generated"
	"gitlab.com/fynbos/backend/onboarding"
	_user "gitlab.com/fynbos/backend/user"
)

func TestUserWithdrawals(s *testing.T) {
	s.Skip("being deprecated")
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
		assert.Equal(t, "10000", response.Withdrawal.Amount) // TODO: where do we pretty print?
		assert.NotEqual(t, 0, response.Withdrawal.Timestamp) // TODO: format of timestamp
		assert.Equal(t, withdrawals.Created.String(), response.Withdrawal.State)
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

		response, err := initiateWithdrawal(container, user, &generated.WithdrawalInput{
			FundingSourceID: fundingSource.ID,
			Amount:          "10000", // 100 dollars
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "422", response.Code)
		assert.Equal(t, false, response.Success)
		assert.Equal(t, "Withdrawal failed: Insufficient balance.", response.Message)
		assert.Nil(t, response.Withdrawal)
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
		assert.Nil(t, response.Withdrawal)
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
		aliceId, err := NewIdentity(container, &identity.CreateArgs{
			ID:           alice.ID,
			FirstName:    faker.FirstName(),
			LastName:     faker.FirstName(),
			MobileNumber: faker.E164PhoneNumber(),
			Email:        alice.Email,
			Country:      "US",
		})
		if err != nil {
			t.Fatal(err)
		}

		bobId, err := NewIdentity(container, &identity.CreateArgs{
			ID:           bob.ID,
			FirstName:    faker.FirstName(),
			LastName:     faker.FirstName(),
			MobileNumber: faker.E164PhoneNumber(),
			Email:        bob.Email,
			Country:      "US",
		})
		if err != nil {
			t.Fatal(err)
		}

		aliceAcc, err := NewAccount(container, &onboarding.CreateAccountArgs{
			IdentityID: aliceId.ID,
			Country:    aliceId.Country,
		})
		if err != nil {
			t.Fatal(err)
		}

		bobAcc, err := NewAccount(container, &onboarding.CreateAccountArgs{
			IdentityID: bobId.ID,
			Country:    bobId.Country,
		})
		if err != nil {
			t.Fatal(err)
		}

		verifyFundingSource := true
		_, err = NewBankAccount(
			container,
			alice,
			&fundingsources.CreateBankAccountArgs{
				IdentityID:    alice.ID,
				AccountID:     aliceAcc.ID,
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
		_, err = NewDeposit(container, &account_transactions.CreateTransactionArgs{
			AccountID:   aliceAcc.ID,
			Description: "Test transaction",
			Type:        "deposit",
			NetAmount:   1000,
			LedgerTransfers: []account_transactions.CreateLedgerTransferArgs{
				{
					LedgerID:        container.NoopService.GetLedgerID(),
					CreditAccountID: container.NoopService.GetEquityAccountID(),
					DebitAccountID:  aliceAcc.LedgerAccountID,
					Amount:          10000,
					Code:            1,
					Flags:           account_transactions.LedgerTransferFlags{},
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		bobBankAccount, err := NewBankAccount(
			container,
			bob,
			&fundingsources.CreateBankAccountArgs{
				IdentityID:    bob.ID,
				AccountID:     bobAcc.ID,
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

		response, err := initiateWithdrawal(container, alice, &generated.WithdrawalInput{
			FundingSourceID: bobBankAccount.ID,
			Amount:          "10000", // 100 dollars
		})
		if err == nil || response != nil {
			t.Fatal()
		}

		assert.Contains(t, err.Error(), "Unable to process request")
	})
}

func initiateWithdrawal(container *TestContainer, user *_user.User, input *generated.WithdrawalInput) (*generated.WithdrawalMutationResponse, error) {
	req := graphql.NewRequest(`
			    mutation ($input: WithdrawalInput!) {
			        initiateWithdrawal (input: $input) {
			            code
			            success
			            message
			            withdrawal {
			            	id
			            	amount
			            	timestamp
			            	state
			            }
			        }
			    }
			`)
	req.Var("input", input)
	err := user_mock.ActingAs(req, user)
	if err != nil {
		return nil, err
	}
	var data map[string]generated.WithdrawalMutationResponse
	if err := container.Client.Run(container.Ctx, req, &data); err != nil {
		return nil, err
	}

	ret := data["initiateWithdrawal"]

	return &ret, nil
}
