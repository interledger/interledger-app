package graph

import (
	"context"
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
	"google.golang.org/grpc"
)

func TestUserOnboarding(s *testing.T) {
	ctx := context.Background()
	container, err := NewTestContainer(ctx, s)
	if err != nil {
		s.Fatal(err)
	}

	s.Cleanup(func() {
		container.Cleanup(ctx)
	})

	/*
		Scenario: user creates an account at Fynbos

		Given a users preliminary information
		When they signup
		Then we create an identity and Fynbos account that is not yet backed by a provider
	*/
	s.Run("user creates an account at Fynbos", func(t *testing.T) {
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}
		input := generateAccountInput()
		ledgerAccountID := uuid.NewString()
		container.MockPacioliClient.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Return(&pacioliv1.Account{
			Id: ledgerAccountID,
		}, nil).Times(1)

		response, err := createAccount(container, user, input)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "200", response.Code)
		assert.Equal(t, true, response.Success)
		assert.Equal(t, "Created account.", response.Message)
		assert.NotNil(t, response.Account)
		assert.Equal(t, "$ 0.00", response.Account.Balance)
	})

	/*
		Scenario: user tries to create duplicate account

		Given a user has already created an account
		When they try to signup again
		Then we fail and say that they're trying to create a duplicate account
	*/
	s.Run("user tries to create a duplicate account", func(t *testing.T) {
		user := &_user.User{
			Email: faker.Email(),
			ID:    uuid.NewString(),
		}
		input := generateAccountInput(withCountry("US"))
		args := &onboarding.CreateAccountArgs{
			IdentityID:   user.ID,
			FirstName:    input.FirstName,
			LastName:     input.LastName,
			MobileNumber: input.MobileNumber,
			Email:        user.Email,
			Country:      "US",
		}
		acc, err := NewAccount(container, args)
		if err != nil {
			t.Fatal(err)
		}
		assert.NotNil(t, acc)

		response, err := createAccount(container, user, input)
		if err == nil {
			t.Fatal("user cannot create a duplicate account.")
		}

		assert.EqualError(t, err, "graphql: Unable to process request.")
		assert.Nil(t, response)
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
		acc, err := NewAccount(container, &onboarding.CreateAccountArgs{
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
		assert.False(t, acc.IsVerified())
		container.MockPacioliClient.EXPECT().GetAccount(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, args *pacioliv1.GetAccountRequest, opts ...grpc.CallOption) (*pacioliv1.Account, error) {
				return &pacioliv1.Account{
					Id: args.Id,
				}, nil
			}).Times(4)

		response, err := verifyAccount(container, user, generateVerifyAccountInput())
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "200", response.Code)
		assert.Equal(t, true, response.Success)
		assert.Equal(t, "Verified.", response.Message)
		freshAcc, err := container.AccountService.GetByIdentityID(ctx, user.ID)
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, freshAcc.IsVerified())
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
		ledgerAccountID := uuid.NewString()
		container.MockPacioliClient.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Return(&pacioliv1.Account{
			Id: ledgerAccountID,
		}, nil).Times(1)
		container.MockPacioliClient.EXPECT().GetAccount(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, args *pacioliv1.GetAccountRequest, opts ...grpc.CallOption) (*pacioliv1.Account, error) {
				return &pacioliv1.Account{
					Id: args.Id,
				}, nil
			}).Times(2)

		response, err := onboardAccount(container, user, input)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "200", response.Code)
		assert.Equal(t, true, response.Success)
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
	_user.ActingAs(req, user)

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

func withFirstName(name string) func(*generated.CreateAccountInput) {
	return func(args *generated.CreateAccountInput) {
		args.FirstName = name
	}
}

func withLastName(name string) func(*generated.CreateAccountInput) {
	return func(args *generated.CreateAccountInput) {
		args.LastName = name
	}
}

func withMobileNumber(number string) func(*generated.CreateAccountInput) {
	return func(args *generated.CreateAccountInput) {
		args.MobileNumber = number
	}
}

func withCountry(country string) func(*generated.CreateAccountInput) {
	return func(args *generated.CreateAccountInput) {
		args.Country = country
	}
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
	_user.ActingAs(req, user)

	var response map[string]generated.VerifyAccountMutationResponse
	if err := container.Client.Run(container.Ctx, req, &response); err != nil {
		return nil, err
	}

	ret := response["verifyAccount"]

	return &ret, nil
}

func onboardAccount(container *TestContainer, user *_user.User, input *generated.VerifyAccountInput) (*generated.CreateAccountMutationResponse, error) {
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
	_user.ActingAs(req, user)

	var response map[string]generated.CreateAccountMutationResponse
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

func withDateOfBirth(dob string) func(*generated.VerifyAccountInput) {
	return func(args *generated.VerifyAccountInput) {
		args.DateOfBirth = dob
	}
}

func withAddress(address []string) func(*generated.VerifyAccountInput) {
	return func(args *generated.VerifyAccountInput) {
		args.Address = address
	}
}

func withState(state string) func(*generated.VerifyAccountInput) {
	return func(args *generated.VerifyAccountInput) {
		args.State = state
	}
}

func withCity(city string) func(*generated.VerifyAccountInput) {
	return func(args *generated.VerifyAccountInput) {
		args.City = city
	}
}

func withPostalCode(code string) func(*generated.VerifyAccountInput) {
	return func(args *generated.VerifyAccountInput) {
		args.PostalCode = code
	}
}

func withTaxIDNumber(tax string) func(*generated.VerifyAccountInput) {
	return func(args *generated.VerifyAccountInput) {
		args.TaxIDNumber = tax
	}
}
