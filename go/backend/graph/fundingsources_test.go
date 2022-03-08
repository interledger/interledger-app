package graph

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/graph/generated"
	"gitlab.com/fynbos/backend/onboarding"
	_user "gitlab.com/fynbos/backend/user"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
)

func TestUserFundingSources(s *testing.T) {
	ctx := context.Background()
	container, err := NewTestContainer(ctx, s)
	if err != nil {
		s.Fatal(err)
	}

	s.Cleanup(func() {
		container.Cleanup(ctx)
	})

	/*
		Scenario: user needs to fetch their funding sources
		Given a verified user
		And the user's verified usd bank account
		Should return the users bank account
	*/
	s.Run("user gets their funding sources with a verified USD bank account", func(t *testing.T) {
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
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

		container.MockPacioliClient.EXPECT().GetAccount(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, args *pacioliv1.GetAccountRequest, opts ...grpc.CallOption) (*pacioliv1.Account, error) {
				return &pacioliv1.Account{
					Id: args.Id,
				}, nil
			}).Times(1)
		response, err := getFundingSourcesByUserId(container, user)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 1, len(response))
		assert.Equal(t, fundingSource.ID, response[0].ID)
		assert.Equal(t, fundingSource.Name, response[0].Name)
		assert.Equal(t, fundingSource.VerificationState, response[0].VerificationStatus)
		assert.Equal(t, fundingSource.Mask, response[0].Mask)
		assert.Equal(t, fundingSource.Type, response[0].Type)
		assert.Equal(t, fundingSource.SubType, response[0].SubType)
	})

	/*
		Scenario: invalid user can't fetch funding sources
		Given a verified user that has funding sources
		And a malformed user
		Should return no funding sources for the malformed user
	*/
	s.Run("invalid user can't fetch funding sources", func(t *testing.T) {
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}
		imposter := &_user.User{
			ID:    "",
			Email: faker.Email(),
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
		verifyFundingSource := true
		_, err = NewBankAccount(
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

		response, err := getFundingSourcesByUserId(container, imposter)

		assert.Error(t, err)
		assert.Equal(t, 0, len(response))
	})
}

func getFundingSourcesByUserId(container *TestContainer, user *_user.User) ([]*generated.FundingSource, error) {
	req := graphql.NewRequest(`
			    query GetUserFundingSources {
						fundingSources {
							id
							name
							verificationStatus
							mask
							type
							subType
			        }
			    }
			`)
	err := _user.ActingAs(req, user)
	if err != nil {
		return nil, err
	}
	var data map[string][]*generated.FundingSource
	if err := container.Client.Run(container.Ctx, req, &data); err != nil {
		return nil, err
	}
	ret := data["fundingSources"]

	return ret, nil
}
