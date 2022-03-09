package graph

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbsqlx"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/graph/generated"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
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

	container.MockPacioliClient.EXPECT().ConfigureAccounts(gomock.Any(), gomock.Any()).Return(
		&pacioliv1.ConfigureAccountsResponse{}, nil,
	).AnyTimes()
	container.MockPacioliClient.EXPECT().GetAccounts(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, args *pacioliv1.GetAccountsRequest, opts ...grpc.CallOption) (*pacioliv1.GetAccountsResponse, error) {
			return &pacioliv1.GetAccountsResponse{
				Accounts: []*pacioliv1.Account{{Id: args.GetIds()[0]}},
			}, nil
		}).AnyTimes()

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
				Email: faker.Name(),
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
			input := generateAccountInput()
			_, err := NewAccount(container, &onboarding.CreateAccountArgs{
				IdentityID:   user.ID,
				Email:        user.Email,
				FirstName:    input.FirstName,
				LastName:     input.LastName,
				MobileNumber: input.MobileNumber,
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
			assert.Equal(tt, user.ID, response.ID)
			assert.Equal(tt, input.FirstName, response.FirstName)
			assert.Equal(tt, input.LastName, response.LastName)
			assert.Equal(tt, input.MobileNumber, response.MobileNumber)
			assert.Equal(tt, user.Email, response.Email)
			assert.Equal(tt, input.Country, response.Country)
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
		assert.Equal(t, "required", fundingsource.VerificationState)

		// verify it
		verifyReq := verifyUsdBankAccount(generateVerifyUsdBankAccountInput(withFundingSourceID(fundingsource.ID)))
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
