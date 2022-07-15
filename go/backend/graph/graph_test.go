package graph

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"

	"gitlab.com/fynbos/backend/graph/generated"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	_user "gitlab.com/fynbos/backend/user"
	test_utils "gitlab.com/fynbos/backend/utils"
)

func TestGraphql(s *testing.T) {
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

	s.Run("get identity", func(t *testing.T) {
		t.Run("requires authenticated user", func(tt *testing.T) {
			req := getIdentityRequest()
			err := _user.ActingAs(req, nil)
			if err != nil {
				return
			}

			var respData map[string]identity.Identity
			err = container.Client.Run(container.Ctx, req, &respData)

			assert.Error(tt, err)
		})

		t.Run("returns not found if there is no identity", func(tt *testing.T) {
			tt.Cleanup(func() {
				err := test_utils.TruncateDb(ctx, container.Db)
				if err != nil {
					tt.Fatal(err)
				}
			})
			user := &_user.User{
				ID:    uuid.New().String(),
				Email: faker.Email(),
			}
			req := getIdentityRequest()
			err = _user.ActingAs(req, user)
			if err != nil {
				tt.Fatal(err)
			}

			var respData map[string]identity.Identity
			err := container.Client.Run(container.Ctx, req, &respData)

			assert.EqualError(tt, err, "graphql: Not found.")
		})

		t.Run("user can get their identity", func(tt *testing.T) {
			tt.Cleanup(func() {
				err := test_utils.TruncateDb(ctx, container.Db)
				if err != nil {
					tt.Fatal(err)
				}
			})
			user := &_user.User{
				ID:    uuid.New().String(),
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

			getReq := getIdentityRequest()
			err = _user.ActingAs(getReq, user)
			if err != nil {
				return
			}

			var getResp map[string]identity.Identity
			if err := container.Client.Run(container.Ctx, getReq, &getResp); err != nil {
				tt.Fatal(err)
			}

			response := getResp["identity"]
			assert.Equal(tt, id.ID, response.ID)
			assert.Equal(tt, id.FirstName, response.FirstName)
			assert.Equal(tt, id.LastName, response.LastName)
			assert.Equal(tt, id.MobileNumber, response.MobileNumber)
			assert.Equal(tt, id.Email, response.Email)
			assert.Equal(tt, id.Country, response.Country)
		})
	})

	s.Run("authenticated user is required to link usd bank account", func(t *testing.T) {
		args := generateLinkUsdBankAccountInput()
		req := linkUsdBankAccountRequest(args)
		err := _user.ActingAs(req, nil)
		if err != nil {
			return
		}

		var respData map[string]generated.LinkFundingSourceMutationResponse
		err = container.Client.Run(container.Ctx, req, &respData)

		assert.Error(t, err)
	})

	s.Run("user can link and verify usd bank account", func(t *testing.T) {
		t.Cleanup(func() {
			err = test_utils.TruncateDb(ctx, container.Db)
			if err != nil {
				t.Fatal(err)
			}
		})
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

		_, err = NewAccount(container, &onboarding.CreateAccountArgs{
			IdentityID: id.ID,
			Country:    id.Country,
		})
		if err != nil {
			t.Fatal(err)
		}

		args := generateLinkUsdBankAccountInput()
		req := linkUsdBankAccountRequest(args)
		err = _user.ActingAs(req, user)
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
		assert.Equal(t, "required", response.FundingSource.VerificationStatus)

		fs, err := container.FundingSourceService.Get(ctx, response.FundingSource.ID)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, args.Name, fs.Name)
		assert.Equal(t, "noop", fs.Type)
		assert.Equal(t, args.Type, fs.SubType)
		assert.NotEqual(t, args.AccountNumber, fs.Mask)
		assert.NotEqual(t, args.RoutingNumber, fs.Mask)
		assert.Equal(t, "required", fs.VerificationState)

		// verify it
		verifyReq := verifyUsdBankAccount(generateVerifyUsdBankAccountInput(withFundingSourceID(fs.ID)))
		err = _user.ActingAs(verifyReq, user)
		if err != nil {
			return
		}
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
		            	country
			        }
			    }
			`)
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
