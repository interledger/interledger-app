package graph

import (
	"context"
	"errors"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbsqlx"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"

	_account "gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/graph/generated"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/identity/noop"
	_user "gitlab.com/fynbos/backend/user"
	test_utils "gitlab.com/fynbos/backend/utils"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
)

func TestGraphql(s *testing.T) {
	ctx := context.Background()
	container, err := NewTestContainer(ctx, s)
	if err != nil {
		s.Fatal(err)
	}

	s.Cleanup(func() {
		container.Cleanup(ctx)
	})

	s.Run("create identity", func(t *testing.T) {
		t.Run("requires authenticated user", func(tt *testing.T) {
			req := createIdentityRequest(generateIdentityInput())
			_user.ActingAs(req, nil)

			var respData map[string]generated.CreateIdentityMutationResponse
			err = container.Client.Run(container.Ctx, req, &respData)

			assert.Error(tt, err)
		})

		t.Run("creates identity and account", func(tt *testing.T) {
			tt.Cleanup(func() {
				test_utils.TruncateDb(ctx, container.Db)
			})
			user := &_user.User{
				ID:    uuid.New().String(),
				Email: faker.Email(),
			}
			input := generateIdentityInput()
			req := createIdentityRequest(input)
			_user.ActingAs(req, user)
			ledgerAccountID := uuid.NewString()
			container.MockPacioliClient.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Return(&pacioliv1.Account{
				Id: ledgerAccountID,
			}, nil).Times(1)

			var respData map[string]generated.CreateIdentityMutationResponse
			if err := container.Client.Run(container.Ctx, req, &respData); err != nil {
				tt.Fatal(err)
			}

			response := respData["createIdentity"]
			assert.Equal(tt, "200", response.Code)
			assert.Equal(tt, "Created account holder.", response.Message)
			assert.Equal(tt, true, response.Success)
			assert.Equal(tt, user.ID, response.Identity.ID)
			assert.Equal(tt, input.FirstName, response.Identity.FirstName)
			assert.Equal(tt, input.LastName, response.Identity.LastName)
			assert.Equal(tt, input.MobileNumber, response.Identity.MobileNumber)
			assert.Equal(tt, user.Email, response.Identity.Email)
			assert.Equal(tt, input.Country, response.Identity.Country)
			assert.Equal(tt, "", response.Identity.DateOfBirth)
			assert.Equal(tt, []string{}, response.Identity.Address)
			assert.Equal(tt, "", response.Identity.City)
			assert.Equal(tt, "", response.Identity.State)
			assert.Equal(tt, "", response.Identity.PostalCode)
			assert.Equal(tt, "", response.Identity.TaxIDNumber)

			var userIdentity *identity.Identity
			err = crdbsqlx.ExecuteTx(ctx, container.Db, nil, func(tx *sqlx.Tx) error {
				_identity, err := container.IdentityService.Get(ctx, tx, user.ID)
				if err != nil {
					return err
				}

				userIdentity = _identity
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(tt, user.ID, userIdentity.ID)
			assert.Equal(tt, input.FirstName, userIdentity.FirstName)
			assert.Equal(tt, input.LastName, userIdentity.LastName)
			assert.Equal(tt, input.MobileNumber, userIdentity.MobileNumber)
			assert.Equal(tt, user.Email, userIdentity.Email)
			assert.Equal(tt, input.Country, userIdentity.Country)
			assert.Equal(tt, "", userIdentity.DateOfBirth)
			assert.Equal(tt, []string{}, userIdentity.Address)
			assert.Equal(tt, "", userIdentity.City)
			assert.Equal(tt, "", userIdentity.State)
			assert.Equal(tt, "", userIdentity.PostalCode)
			assert.Equal(tt, "", userIdentity.TaxIDNumber)

			container.MockPacioliClient.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Return(&pacioliv1.Account{
				Id:              ledgerAccountID,
				DebitsReserved:  1, // return non-zero to make sure default values aren't used.
				DebitsAccepted:  2,
				CreditsAccepted: 3,
				CreditsReserved: 4,
			}, nil).Times(1)
			var account *_account.Account
			err = crdbsqlx.ExecuteTx(ctx, container.Db, nil, func(tx *sqlx.Tx) error {
				_acc, err := container.AccountService.GetByIdentityID(ctx, tx, user.ID)
				if err != nil {
					return err
				}

				account = _acc
				return nil
			})
			if err != nil {
				tt.Fatal(err)
			}
			assert.Equal(tt, userIdentity.ID, account.IdentityID)
			assert.Equal(tt, ledgerAccountID, account.LedgerAccountID)
			assert.Equal(tt, uint64(1), account.DebitsReserved)
			assert.Equal(tt, uint64(2), account.DebitsAccepted)
			assert.Equal(tt, uint64(3), account.CreditsAccepted)
			assert.Equal(tt, uint64(4), account.CreditsReserved)
		})

		t.Run("user can only create 1 identity", func(tt *testing.T) {
			tt.Cleanup(func() {
				test_utils.TruncateDb(ctx, container.Db)
			})
			container.MockPacioliClient.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Return(&pacioliv1.Account{
				Id: uuid.NewString(),
			}, nil).Times(1)
			user := &_user.User{
				ID:    uuid.NewString(),
				Email: faker.Email(),
			}
			input := generateIdentityInput()
			req := createIdentityRequest(input)
			_user.ActingAs(req, user)
			var respData map[string]generated.CreateIdentityMutationResponse
			if err := container.Client.Run(container.Ctx, req, &respData); err != nil {
				tt.Fatal(err)
			}
			response := respData["createIdentity"]
			assert.Equal(tt, true, response.Success)

			additionalInput := generateIdentityInput(withCountry("ZA"))
			req.Var("input", additionalInput)
			err = container.Client.Run(container.Ctx, req, &respData)
			assert.EqualError(tt, err, "graphql: Unable to process request.")
		})
	})

	s.Run("get identity", func(t *testing.T) {
		t.Run("requires authenticated user", func(tt *testing.T) {
			req := getIdentityRequest()
			_user.ActingAs(req, nil)

			var respData map[string]identity.Identity
			err := container.Client.Run(container.Ctx, req, &respData)

			assert.Error(tt, err)
		})

		t.Run("returns not found if there is no identity", func(tt *testing.T) {
			tt.Cleanup(func() {
				test_utils.TruncateDb(ctx, container.Db)
			})
			user := &_user.User{
				ID:    uuid.New().String(),
				Email: faker.Name(),
			}
			req := getIdentityRequest()
			_user.ActingAs(req, user)

			var respData map[string]identity.Identity
			err := container.Client.Run(container.Ctx, req, &respData)

			assert.EqualError(tt, err, "graphql: Not found.")
		})

		t.Run("user can get their identity", func(tt *testing.T) {
			tt.Cleanup(func() {
				test_utils.TruncateDb(ctx, container.Db)
			})
			user := &_user.User{
				ID:    uuid.New().String(),
				Email: faker.Email(),
			}
			input := generateIdentityInput()
			req := createIdentityRequest(input)
			_user.ActingAs(req, user)
			ledgerAccountID := uuid.NewString()
			container.MockPacioliClient.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Return(&pacioliv1.Account{
				Id: ledgerAccountID,
			}, nil).Times(1)

			var respData map[string]generated.CreateIdentityMutationResponse
			if err := container.Client.Run(container.Ctx, req, &respData); err != nil {
				tt.Fatal(err)
			}

			getReq := getIdentityRequest()
			_user.ActingAs(getReq, user)

			var getResp map[string]identity.Identity
			if err := container.Client.Run(container.Ctx, getReq, &getResp); err != nil {
				tt.Fatal(err)
			}

			response := getResp["identity"]
			assert.Equal(tt, user.ID, response.ID)
			assert.Equal(tt, input.FirstName, response.FirstName)
			assert.Equal(tt, input.LastName, response.LastName)
			assert.Equal(tt, input.MobileNumber, response.MobileNumber)
			assert.Equal(tt, user.Email, response.Email)
			assert.Equal(tt, input.Country, response.Country)
			assert.Equal(tt, "", response.DateOfBirth)
			assert.Equal(tt, []string{}, response.Address)
			assert.Equal(tt, "", response.City)
			assert.Equal(tt, "", response.State)
			assert.Equal(tt, "", response.PostalCode)
			assert.Equal(tt, "", response.TaxIDNumber)
		})
	})

	s.Run("verify", func(t *testing.T) {
		t.Run("fails if there is no authenticated user", func(tt *testing.T) {
			req := verifyRequest(generateVerificationInput())
			_user.ActingAs(req, nil)

			var respData map[string]identity.Identity
			err := container.Client.Run(container.Ctx, req, &respData)

			assert.Error(tt, err)
		})

		t.Run("returns not found if there is no identity", func(tt *testing.T) {
			user := &_user.User{
				ID:    uuid.New().String(),
				Email: faker.Email(),
			}
			req := verifyRequest(generateVerificationInput())
			_user.ActingAs(req, user)

			var respData map[string]identity.Identity
			err := container.Client.Run(container.Ctx, req, &respData)

			assert.EqualError(tt, err, "graphql: Not found.")
		})

		t.Run("fails if call to provider fails", func(tt *testing.T) {
			tt.Cleanup(func() {
				test_utils.TruncateDb(ctx, container.Db)
			})
			user := &_user.User{
				ID:    uuid.New().String(),
				Email: faker.Email(),
			}
			input := generateIdentityInput()
			req := createIdentityRequest(input)
			_user.ActingAs(req, user)
			container.MockPacioliClient.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Return(&pacioliv1.Account{
				Id: uuid.NewString(),
			}, nil).Times(1)
			container.NoopProvider.EXPECT().CreateCustomer(gomock.Any()).Return(nil, errors.New("Request failed.")).Times(1)

			var respData map[string]generated.CreateIdentityMutationResponse
			if err := container.Client.Run(container.Ctx, req, &respData); err != nil {
				tt.Fatal(err)
			}

			verifyReq := verifyRequest(generateVerificationInput())
			_user.ActingAs(verifyReq, user)
			var verifyResp map[string]generated.VerifyMutationResponse
			err = container.Client.Run(container.Ctx, verifyReq, &verifyResp)

			assert.EqualError(tt, err, "graphql: Unable to process request.")
		})

		t.Run("creates customer and stores kyc data", func(tt *testing.T) {
			tt.Cleanup(func() {
				test_utils.TruncateDb(ctx, container.Db)
			})
			user := &_user.User{
				ID:    uuid.New().String(),
				Email: faker.Email(),
			}
			input := generateIdentityInput()
			req := createIdentityRequest(input)
			_user.ActingAs(req, user)
			container.MockPacioliClient.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Return(&pacioliv1.Account{
				Id: uuid.NewString(),
			}, nil).Times(1)
			customerID := uuid.NewString()
			container.NoopProvider.EXPECT().CreateCustomer(gomock.Any()).Return(&noop.Customer{
				ID:     customerID,
				Status: noop.Verified,
			}, nil).Times(1)

			var respData map[string]generated.CreateIdentityMutationResponse
			if err := container.Client.Run(container.Ctx, req, &respData); err != nil {
				tt.Fatal(err)
			}

			verifyReq := verifyRequest(generateVerificationInput())
			_user.ActingAs(verifyReq, user)
			var verifyResp map[string]generated.VerifyMutationResponse
			if err := container.Client.Run(container.Ctx, verifyReq, &verifyResp); err != nil {
				tt.Fatal(err)
			}

			response := verifyResp["verify"]
			assert.Equal(tt, "200", response.Code)
			assert.Equal(tt, true, response.Success)
			assert.Equal(tt, "Verified.", response.Message)
			assert.Equal(tt, noop.Verified, response.Identity.VerificationState)
		})
	})

	s.Run("authenticated user is required to link usd bank account", func(t *testing.T) {
		args := generateLinkUsdBankAccountInput()
		req := linkUsdBankAccountRequest(args)
		_user.ActingAs(req, nil)

		var respData map[string]generated.LinkFundingSourceMutationResponse
		err := container.Client.Run(container.Ctx, req, &respData)

		assert.Error(t, err)
	})

	s.Run("user can link and verify usd bank account", func(t *testing.T) {
		t.Cleanup(func() {
			test_utils.TruncateDb(ctx, container.Db)
		})
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}
		input := generateIdentityInput()
		identityReq := createIdentityRequest(input)
		_user.ActingAs(identityReq, user)
		ledgerAccountID := uuid.NewString()
		container.MockPacioliClient.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Return(&pacioliv1.Account{
			Id: ledgerAccountID,
		}, nil).Times(1)
		var identityData map[string]generated.CreateIdentityMutationResponse
		if err := container.Client.Run(container.Ctx, identityReq, &identityData); err != nil {
			t.Fatal(err)
		}

		args := generateLinkUsdBankAccountInput()
		req := linkUsdBankAccountRequest(args)
		err := _user.ActingAs(req, user)
		if err != nil {
			t.Fatal(err)
		}

		var respData map[string]generated.LinkFundingSourceMutationResponse
		if err := container.Client.Run(container.Ctx, req, &respData); err != nil {
			t.Fatal(err)
		}

		response := respData["linkUsdBankAccount"]
		assert.Equal(t, "200", response.Code)
		assert.Equal(t, true, response.Success)
		assert.Equal(t, args.Name, response.FundingSource.Name)
		assert.Equal(t, "noop", response.FundingSource.Type)
		assert.Equal(t, args.Type, response.FundingSource.SubType)
		assert.NotEqual(t, args.AccountNumber, response.FundingSource.Mask)
		assert.NotEqual(t, args.RoutingNumber, response.FundingSource.Mask)
		assert.Equal(t, "pending", response.FundingSource.VerificationStatus)

		var fundingsource *fundingsources.FundingSource
		err = crdbsqlx.ExecuteTx(ctx, container.Db, nil, func(tx *sqlx.Tx) error {
			_fs, err := container.FundingSourceService.Get(ctx, tx, response.FundingSource.ID)
			if err != nil {
				return err
			}
			fundingsource = _fs

			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, args.Name, fundingsource.Name)
		assert.Equal(t, "noop", fundingsource.Type)
		assert.Equal(t, args.Type, fundingsource.SubType)
		assert.NotEqual(t, "", fundingsource.TypeID)
		assert.NotEqual(t, args.AccountNumber, fundingsource.Mask)
		assert.NotEqual(t, args.RoutingNumber, fundingsource.Mask)
		assert.Equal(t, "pending", fundingsource.VerificationState)

		// verify it
		verifyReq := verifyUsdBankAccount(generateVerifyUsdBankAccountInput(withFundingSourceID(fundingsource.ID)))
		_user.ActingAs(verifyReq, user)
		var verifyData map[string]generated.VerifyUsdBankAccountMutationResponse
		if err := container.Client.Run(container.Ctx, verifyReq, &verifyData); err != nil {
			t.Fatal(err)
		}
		verifyResponse := verifyData["verifyUsdBankAccount"]
		assert.Equal(t, "200", verifyResponse.Code)
		assert.Equal(t, "Verified account.", verifyResponse.Message)
		assert.Equal(t, true, verifyResponse.Success)
		assert.Equal(t, "verified", verifyResponse.FundingSource.VerificationStatus)
	})

	s.Run("account", func(t *testing.T) {
		t.Run("fails if there is no authenticated user", func(tt *testing.T) {
			req := verifyRequest(generateVerificationInput())
			_user.ActingAs(req, nil)

			var respData map[string]identity.Identity
			err := container.Client.Run(ctx, req, &respData)

			assert.Error(tt, err)
		})

		t.Run("user can get their account", func(tt *testing.T) {
			t.Cleanup(func() {
				test_utils.TruncateDb(ctx, container.Db)
			})
			user := &_user.User{
				ID:    uuid.NewString(),
				Email: faker.Email(),
			}
			input := generateIdentityInput()
			identityReq := createIdentityRequest(input)
			_user.ActingAs(identityReq, user)
			ledgerAccountID := uuid.NewString()
			container.MockPacioliClient.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Return(&pacioliv1.Account{
				Id: ledgerAccountID,
			}, nil).Times(1)
			container.MockPacioliClient.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Return(&pacioliv1.Account{
				Id:              ledgerAccountID,
				CreditsAccepted: 200,
				DebitsAccepted:  80,
			}, nil).Times(1)
			var identityData map[string]generated.CreateIdentityMutationResponse
			if err := container.Client.Run(ctx, identityReq, &identityData); err != nil {
				t.Fatal(err)
			}

			req := getAccountRequest()
			err := _user.ActingAs(req, user)
			if err != nil {
				t.Fatal(err)
			}

			var respData map[string]generated.Account
			if err := container.Client.Run(ctx, req, &respData); err != nil {
				t.Fatal(err)
			}

			response := respData["account"]
			assert.Equal(t, "120", response.Balance)
		})
	})
}

func createIdentityRequest(input *generated.CreateIdentityInput) *graphql.Request {
	req := graphql.NewRequest(`
			    mutation ($input: CreateIdentityInput!) {
			        createIdentity (input: $input) {
			            code
			            success
			            message
			            identity {
			            	id
			            	firstName
			            	lastName
			            	mobileNumber
			            	email
			            	dateOfBirth
			            	address
			            	city
			            	state
			            	postalCode
			            	country
			            	taxIdNumber
			            }
			        }
			    }
			`)
	req.Var("input", input)

	return req
}

func getIdentityRequest() *graphql.Request {
	return graphql.NewRequest(`
			    query {
			        identity {
			            id
		            	firstName
		            	lastName
		            	mobileNumber
		            	email
		            	dateOfBirth
		            	address
		            	city
		            	state
		            	postalCode
		            	country
		            	taxIdNumber
			        }
			    }
			`)
}

func getAccountRequest() *graphql.Request {
	return graphql.NewRequest(`
			    query {
			        account {
			            id
			            balance
			        }
			    }
			`)
}

// TODO: auto generate helpers.
// Factory to generate create args
func generateIdentityInput(opts ...func(*generated.CreateIdentityInput)) *generated.CreateIdentityInput {
	args := &generated.CreateIdentityInput{
		FirstName:    faker.Name(),
		LastName:     faker.LastName(),
		MobileNumber: faker.Phonenumber(),
		Country:      "US",
	}

	for _, opt := range opts {
		opt(args)
	}

	return args
}

func withFirstName(name string) func(*generated.CreateIdentityInput) {
	return func(args *generated.CreateIdentityInput) {
		args.FirstName = name
	}
}

func withLastName(name string) func(*generated.CreateIdentityInput) {
	return func(args *generated.CreateIdentityInput) {
		args.LastName = name
	}
}

func withMobileNumber(number string) func(*generated.CreateIdentityInput) {
	return func(args *generated.CreateIdentityInput) {
		args.MobileNumber = number
	}
}

func withCountry(country string) func(*generated.CreateIdentityInput) {
	return func(args *generated.CreateIdentityInput) {
		args.Country = country
	}
}

func verifyRequest(input *generated.VerificationInput) *graphql.Request {
	req := graphql.NewRequest(`
			    mutation ($input: VerificationInput!) {
			        verify (input: $input) {
			            code
			            success
			            message
			            identity{
			            	id
			            	firstName
			            	lastName
			            	mobileNumber
			            	email
			            	dateOfBirth
			            	address
			            	city
			            	state
			            	postalCode
			            	country
			            	taxIdNumber
			            	verificationState
			            }
			        }
			    }
			`)
	req.Var("input", input)
	return req
}

func generateVerificationInput(opts ...func(*generated.VerificationInput)) *generated.VerificationInput {
	args := &generated.VerificationInput{
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

func withDateOfBirth(dob string) func(generated.VerificationInput) {
	return func(args generated.VerificationInput) {
		args.DateOfBirth = dob
	}
}

func withAddress(address []string) func(generated.VerificationInput) {
	return func(args generated.VerificationInput) {
		args.Address = address
	}
}

func withState(state string) func(generated.VerificationInput) {
	return func(args generated.VerificationInput) {
		args.State = state
	}
}

func withCity(city string) func(generated.VerificationInput) {
	return func(args generated.VerificationInput) {
		args.City = city
	}
}

func withPostalCode(code string) func(generated.VerificationInput) {
	return func(args generated.VerificationInput) {
		args.PostalCode = code
	}
}

func withTaxIDNumber(tax string) func(generated.VerificationInput) {
	return func(args generated.VerificationInput) {
		args.TaxIDNumber = tax
	}
}

func linkUsdBankAccountRequest(input *generated.LinkUsdBankAccountInput) *graphql.Request {
	req := graphql.NewRequest(`
			    mutation ($input: LinkUsdBankAccountInput!) {
			        linkUsdBankAccount (input: $input) {
			            code
			            success
			            message
			            fundingSource {
				            id
			            	name
			            	verificationStatus
			            	mask
			            	type
			            	subType
			            }
			        }
			    }
			`)
	req.Var("input", input)

	return req
}

func generateLinkUsdBankAccountInput(opts ...func(*generated.LinkUsdBankAccountInput)) *generated.LinkUsdBankAccountInput {
	args := &generated.LinkUsdBankAccountInput{
		Name:          faker.FirstName(),
		AccountNumber: faker.CCNumber(),
		RoutingNumber: faker.CCNumber(),
		Institution:   faker.Name(),
		Type:          "cheque",
	}

	for _, opt := range opts {
		opt(args)
	}

	return args
}

func verifyUsdBankAccount(input *generated.VerifyUsdBankAccountInput) *graphql.Request {
	req := graphql.NewRequest(`
		mutation ($input: VerifyUsdBankAccountInput!){
			verifyUsdBankAccount (input: $input) {
				code
	            success
	            message
	            fundingSource {
		            id
	            	name
	            	verificationStatus
	            	mask
	            	type
	            	subType
	            }
			}
		}`)
	req.Var("input", input)

	return req
}

func generateVerifyUsdBankAccountInput(opts ...func(*generated.VerifyUsdBankAccountInput)) *generated.VerifyUsdBankAccountInput {
	args := &generated.VerifyUsdBankAccountInput{}
	for _, opt := range opts {
		opt(args)
	}

	return args
}

func withFundingSourceID(id string) func(*generated.VerifyUsdBankAccountInput) {
	return func(args *generated.VerifyUsdBankAccountInput) {
		args.FundingSourceID = id
	}
}
