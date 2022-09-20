package graph

import (
	"context"
	"testing"

	user_mock "gitlab.com/fynbos/backend/user/client/mock"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/graph/generated"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	_user "gitlab.com/fynbos/backend/user"
)

func TestUserOnboarding(s *testing.T) {
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
		Scenario: user creates an account at Fynbos

		Given a users preliminary information
		When they signup
		Then we create an identity and Fynbos account that is not yet backed by a provider
		And Fynbos account is created with CreditsMustNotExceedDebits flag
	*/
	s.Run("user creates an account at Fynbos", func(t *testing.T) {
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}

		id, err := NewIdentity(container, &identity.CreateArgs{
			ID:           user.ID,
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			MobileNumber: faker.E164PhoneNumber(),
			Email:        user.Email,
			Country:      "US",
		})
		if err != nil {
			t.Fatal(err)
		}

		response, err := createAccount(container, user, &generated.CreateAccountInput{
			FirstName:    id.FirstName,
			LastName:     id.LastName,
			MobileNumber: id.MobileNumber,
			Country:      id.Country,
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "200", response.Code)
		assert.Equal(t, true, response.Success)
		assert.Equal(t, "Created account.", response.Message)
		assert.NotNil(t, response.Account)
		assert.Equal(t, "$ 0.00", response.Account.Balance)

		acc, err := container.AccountService.GetByIdentityID(ctx, user.ID)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, acc.CreditsMustNotExceedDebits)
		assert.Equal(t, false, acc.DebitsMustNotExceedCredits)
	})

	/*
		Scenario: user tries to create duplicate account

		Given a user has already created an account
		When they try to signup again
		Then we return their original account
	*/
	s.Run("user tries to create multiple accounts", func(t *testing.T) {
		user := &_user.User{
			Email: faker.Email(),
			ID:    uuid.NewString(),
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
		assert.NotNil(t, acc)

		response, err := createAccount(container, user, &generated.CreateAccountInput{
			FirstName:    id.FirstName,
			LastName:     id.LastName,
			MobileNumber: id.MobileNumber,
			Country:      id.Country,
		})
		if err != nil {
			t.Fatal("User cannot create a duplicate account, should return original account.")
		}

		assert.Equal(t, "200", response.Code)
		assert.NotNil(t, response.Account)
		assert.Equal(t, acc.ID, response.Account.ID)
	})

	/*
		Scenario: unauthenticated request to create an account is rejected
	*/
	s.Run("unauthenticated request to create an account is rejected", func(t *testing.T) {
		input := generateAccountInput()

		response, err := createAccount(container, nil, input)

		assert.Error(t, err)
		assert.Nil(t, response)
	})

	/*
		Scenario: user verifies their account

		Given a user has an identity and Fynbos account
		When the user verifies their account
		Then we create a customer at the relevant provider
	*/
	s.Run("user verifies their account", func(t *testing.T) {
		user := &_user.User{
			Email: faker.Email(),
			ID:    uuid.NewString(),
		}
		_, err := NewIdentity(container, &identity.CreateArgs{
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

		response, err := verifyAccount(container, user, generateVerifyAccountInput())
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "200", response.Code)
		assert.Equal(t, true, response.Success)
		assert.Equal(t, "Verified.", response.Message)
	})

	// Scenario: unauthenticated request to verify account is rejected
	s.Run("unauthenticated request to verify account is rejected", func(t *testing.T) {
		input := generateVerifyAccountInput()

		response, err := verifyAccount(container, nil, input)

		assert.Error(t, err)
		assert.Nil(t, response)
	})

	// Scenario: onboard request creates and verifies account TODO: Dummy for now
	s.Run("onboard request creates and verifies account", func(t *testing.T) {
		user := &_user.User{
			Email: faker.Email(),
			ID:    uuid.NewString(),
		}
		input := generateVerifyAccountInput()

		response, err := onboardAccount(container, user, input)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "200", response.Code)
		assert.Equal(t, true, response.Success)
	})

	s.Run("can initiate onboarding which returns formUrl for unit", func(t *testing.T) {
		user := &_user.User{
			Email: faker.Email(),
			ID:    uuid.NewString(),
		}

		response, err := initiateOnboarding(container, user)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "200", response.Code)
		assert.Equal(t, true, response.Success)
		assert.Equal(t, "Initiated onboarding.", response.Message)
		assert.Equal(t, "https://application-form.sh/DXB4GXQMBGY377CD5KQ3OWX4XJEF4Z3DQPKTMDGF77CFQM7M55WOQR5C2C3D5N2NYP52AOCSVZX6JLLGSHRLI3DXZ45R43QPDIBWUAI7KL6I7ESUPTB7C7EFURQKMZZSINKSXYQ2N63L7TFPCQVQIW6TVQQLXUYJQP6FY", response.ProviderOnboarding.FormURL)
	})
}

func createAccount(container *TestContainer, user *_user.User, input *generated.CreateAccountInput) (*generated.CreateAccountMutationResponse, error) {
	req := graphql.NewRequest(`
			    mutation ($input: CreateAccountInput!) {
			        createAccount (input: $input) {
			            code
			            success
			            message
			            account {
			            	id
			            	balance
			            }
			        }
			    }
			`)
	req.Var("input", input)
	if err := user_mock.ActingAs(req, user); err != nil {
		return nil, err
	}

	var response map[string]generated.CreateAccountMutationResponse
	if err := container.Client.Run(container.Ctx, req, &response); err != nil {
		return nil, err
	}

	ret := response["createAccount"]

	return &ret, nil
}

func generateAccountInput(opts ...func(*generated.CreateAccountInput)) *generated.CreateAccountInput {
	args := &generated.CreateAccountInput{
		FirstName:    faker.Name(),
		LastName:     faker.LastName(),
		MobileNumber: faker.E164PhoneNumber(),
		Country:      "US",
	}

	for _, opt := range opts {
		opt(args)
	}

	return args
}

func verifyAccount(container *TestContainer, user *_user.User, input *generated.VerifyAccountInput) (*generated.VerifyAccountMutationResponse, error) {
	req := graphql.NewRequest(`
			    mutation ($input: VerifyAccountInput!) {
			        verifyAccount (input: $input) {
			            code
			            success
			            message
			            account {
			            	id
			            	balance
			            }
			        }
			    }
			`)
	req.Var("input", input)
	if err := user_mock.ActingAs(req, user); err != nil {
		return nil, err
	}

	var response map[string]generated.VerifyAccountMutationResponse
	if err := container.Client.Run(container.Ctx, req, &response); err != nil {
		return nil, err
	}

	ret := response["verifyAccount"]

	return &ret, nil
}

func initiateOnboarding(container *TestContainer, user *_user.User) (*generated.InitiateOnboardingMutationResponse, error) {
	req := graphql.NewRequest(`
			    mutation {
						initiateOnboarding {
			            code
			            success
			            message
  								providerOnboarding {
										id
										formUrl
									}
			        }
			    }
			`)
	if err := user_mock.ActingAs(req, user); err != nil {
		return nil, err
	}

	var response map[string]generated.InitiateOnboardingMutationResponse
	if err := container.Client.Run(container.Ctx, req, &response); err != nil {
		return nil, err
	}

	ret := response["initiateOnboarding"]

	return &ret, nil
}

func onboardAccount(container *TestContainer, user *_user.User, input *generated.VerifyAccountInput) (*generated.OnboardingMutationResponse, error) {
	req := graphql.NewRequest(`
			    mutation {
			        onboardAccount {
			            code
			            success
			            message
			            account {
			            	id
			            	balance
			            }
			        }
			    }
			`)
	req.Var("input", input)
	if err := user_mock.ActingAs(req, user); err != nil {
		return nil, err
	}

	var response map[string]generated.OnboardingMutationResponse
	if err := container.Client.Run(container.Ctx, req, &response); err != nil {
		return nil, err
	}

	ret := response["onboardAccount"]

	return &ret, nil
}

func generateVerifyAccountInput(opts ...func(*generated.VerifyAccountInput)) *generated.VerifyAccountInput {
	args := &generated.VerifyAccountInput{
		DateOfBirth: faker.Date(),
		Address:     []string{faker.Name()},
		State:       faker.FirstName(),
		City:        faker.FirstName(),
		PostalCode:  faker.Currency(),
		TaxIDNumber: faker.CCNumber(),
	}

	for _, opt := range opts {
		opt(args)
	}

	return args
}
