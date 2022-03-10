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

	"gitlab.com/fynbos/backend/deposits"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/graph/generated"
	"gitlab.com/fynbos/backend/onboarding"
	_user "gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/proto/pacioli/v1"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
)

func TestUserAccount(s *testing.T) {
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
	container.MockPacioliClient.EXPECT().CreateTransfers(gomock.Any(), gomock.Any()).Return(
		&pacioliv1.CreateTransfersResponse{
			Errors: []*pacioliv1.EventError{},
		}, nil).AnyTimes()
	/*
		Scenario: user queries their account balance
	*/
	s.Run("user can get their account balance", func(t *testing.T) {
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
		container.MockPacioliClient.EXPECT().GetAccounts(gomock.Any(), &pacioli.GetAccountsRequest{Ids: []string{acc.LedgerAccountID}}).DoAndReturn(
			func(_ context.Context, args *pacioliv1.GetAccountsRequest, opts ...grpc.CallOption) (*pacioliv1.GetAccountsResponse, error) {
				return &pacioliv1.GetAccountsResponse{
					Accounts: []*pacioliv1.Account{
						{
							Id:              args.GetIds()[0],
							CreditsAccepted: 100,
							CreditsReserved: 20,
							DebitsAccepted:  200,
							DebitsReserved:  0,
						},
					},
				}, nil
			}).Times(1)

		req := getAccountRequest()
		err = _user.ActingAs(req, user)
		if err != nil {
			t.Fatal(err)
		}
		var data map[string]generated.Account
		if err := container.Client.Run(container.Ctx, req, &data); err != nil {
			t.Fatal(err)
		}

		response := data["account"]
		assert.Equal(t, "$ 0.80", response.Balance)
		assert.NotNil(t, response.ID)
	})

	/*
		Scenario: user can get their recent transactions
	*/
	s.Run("user can get their recent transactions", func(t *testing.T) {
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
		container.MockPacioliClient.EXPECT().GetAccounts(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, args *pacioliv1.GetAccountsRequest, opts ...grpc.CallOption) (*pacioliv1.GetAccountsResponse, error) {
				return &pacioliv1.GetAccountsResponse{
					Accounts: []*pacioliv1.Account{
						{
							Id: args.GetIds()[0],
						},
					},
				}, nil
			}).Times(4)
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
			true,
		)
		if err != nil {
			t.Fatal(err)
		}
		deposit, err := NewDeposit(container, &deposits.InitiateDepositArgs{
			IdentityID:      user.ID,
			AccountID:       acc.ID,
			FundingSourceID: fundingSource.ID,
			Amount:          10000, // 100 dollars
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "completed", deposit.State)
		container.MockPacioliClient.EXPECT().GetAccounts(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, args *pacioliv1.GetAccountsRequest, opts ...grpc.CallOption) (*pacioliv1.GetAccountsResponse, error) {
				return &pacioliv1.GetAccountsResponse{
					Accounts: []*pacioliv1.Account{
						{
							Id:              args.GetIds()[0],
							CreditsAccepted: 0,
							CreditsReserved: 0,
							DebitsAccepted:  0,
							DebitsReserved:  0,
						},
					},
				}, nil
			}).Times(1)

		req := getAccountTransactionsRequest()
		err = _user.ActingAs(req, user)
		if err != nil {
			t.Fatal(err)
		}
		var data map[string]generated.Account
		if err := container.Client.Run(container.Ctx, req, &data); err != nil {
			t.Fatal(err)
		}

		rTrxs := data["account"].RecentTransactions
		assert.Len(t, rTrxs, 1)
		firstTrx := rTrxs[0]
		assert.Equal(t, deposit.ID, firstTrx.ID)
		assert.Equal(t, "$ 100.00", firstTrx.Amount)
		assert.Equal(t, generated.TransactionTypeDeposit, firstTrx.Type)
	})
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

func getAccountTransactionsRequest() *graphql.Request {
	return graphql.NewRequest(`
		query {
			account {
				recentTransactions {
					id
					amount
					type
				}
			}
		}
	`)
}
