package mx

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	test_utils "gitlab.com/fynbos/backend/utils"
)

var runIntegration = flag.Bool("integration", false, "Bool to run integration tests.")

func TestIntegration(t *testing.T) {
	if !*runIntegration {
		t.Skip()
	}
	username := os.Getenv("MX_USERNAME")
	password := os.Getenv("MX_PASSWORD")
	mx, err := NewService(&ServiceArgs{
		BaseUrl:  "https://int-api.mx.com",
		Username: username,
		Password: password,
		Db:       &sqlx.DB{},
	})
	if err != nil {
		t.Fatal(err)
	}

	userGuid, err := mx.CreateUser(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEqual(t, "", userGuid)

	url, err := mx.GetWidgetUrl(context.Background(), userGuid)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEqual(t, "", url)
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	mockMxServer := NewMockServer()
	mx, err := NewService(&ServiceArgs{
		BaseUrl:  mockMxServer.URL,
		Username: "test",
		Password: "test",
		Db:       &sqlx.DB{},
	})
	if err != nil {
		t.Fatal(err)
	}

	userGuid, err := mx.CreateUser(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_ = uuid.MustParse(userGuid)
}

func TestGetWidgetUrl(t *testing.T) {
	ctx := context.Background()
	mockMxServer := NewMockServer()
	mx, err := NewService(&ServiceArgs{
		BaseUrl:  mockMxServer.URL,
		Username: "test",
		Password: "test",
		Db:       &sqlx.DB{},
	})
	if err != nil {
		t.Fatal(err)
	}

	url, err := mx.GetWidgetUrl(ctx, uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}

	assert.NotEqual(t, "", url)
}

func TestCreateAndGetAccount(t *testing.T) {
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)
	mockMxServer := NewMockServer()
	mx, err := NewService(&ServiceArgs{
		BaseUrl:  mockMxServer.URL,
		Username: "test",
		Password: "test",
		Db:       db,
	})
	if err != nil {
		t.Fatal(err)
	}
	args := &CreateAccountArgs{
		ID:            uuid.NewString(),
		AccountID:     uuid.NewString(),
		MxUserGuid:    uuid.NewString(),
		MxMemberGuid:  uuid.NewString(),
		MxAccountGuid: uuid.NewString(),
	}

	mxAccount, err := mx.CreateAccount(ctx, args)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, args.ID, mxAccount.ID)
	assert.Equal(t, args.AccountID, mxAccount.AccountID)
	assert.Equal(t, args.MxUserGuid, mxAccount.MxUserGuid)
	assert.Equal(t, args.MxMemberGuid, mxAccount.MxMemberGuid)

	freshMxFs, err := mx.GetAccount(ctx, args.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, args.ID, freshMxFs.ID)
	assert.Equal(t, args.AccountID, freshMxFs.AccountID)
	assert.Equal(t, args.MxUserGuid, freshMxFs.MxUserGuid)
	assert.Equal(t, args.MxMemberGuid, freshMxFs.MxMemberGuid)

	noAcc, err := mx.GetAccount(ctx, uuid.NewString())
	assert.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, noAcc)

	// idempotency
	idempotent, err := mx.CreateAccount(ctx, args)
	assert.ErrorIs(t, err, ErrDuplicate)
	assert.Nil(t, idempotent)
}

func TestGetMemberStatus(t *testing.T) {
	ctx := context.Background()
	mockMxServer := NewMockServer()
	db := test_utils.MigrateCockroachDB(t, ctx)
	mx, err := NewService(&ServiceArgs{
		BaseUrl:  mockMxServer.URL,
		Username: "test",
		Password: "test",
		Db:       db,
	})
	if err != nil {
		t.Fatal(err)
	}

	mxAccount, err := mx.CreateAccount(ctx, &CreateAccountArgs{
		ID:            uuid.NewString(),
		AccountID:     uuid.NewString(),
		MxUserGuid:    uuid.NewString(),
		MxMemberGuid:  uuid.NewString(),
		MxAccountGuid: uuid.NewString(),
	})
	if err != nil {
		t.Fatal(err)
	}

	member, err := mx.GetMemberStatus(ctx, mxAccount.ID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, mxAccount.MxUserGuid, member.UserGuid)
	assert.Equal(t, mxAccount.MxMemberGuid, member.Guid)
	assert.Equal(t, false, member.IsBeingAggregated)
}

func TestStartIdentityAggregation(t *testing.T) {
	ctx := context.Background()
	mockMxServer := NewMockServer()
	db := test_utils.MigrateCockroachDB(t, ctx)
	mx, err := NewService(&ServiceArgs{
		BaseUrl:  mockMxServer.URL,
		Username: "test",
		Password: "test",
		Db:       db,
	})
	if err != nil {
		t.Fatal(err)
	}

	mxAccount, err := mx.CreateAccount(ctx, &CreateAccountArgs{
		ID:            uuid.NewString(),
		AccountID:     uuid.NewString(),
		MxUserGuid:    uuid.NewString(),
		MxMemberGuid:  uuid.NewString(),
		MxAccountGuid: uuid.NewString(),
	})
	if err != nil {
		t.Fatal(err)
	}

	member, err := mx.StartIdentityAggregation(ctx, mxAccount.ID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, mxAccount.MxUserGuid, member.UserGuid)
	assert.Equal(t, mxAccount.MxMemberGuid, member.Guid)
	assert.Equal(t, true, member.IsBeingAggregated)
}

func TestGetAccountOwner(t *testing.T) {
	ctx := context.Background()
	mxAccountGuid := uuid.NewString()
	accountOwners := []AccountOwner{
		{
			AccountGuid: mxAccountGuid,
			OwnerName:   faker.Name(),
			Country:     "US",
			Email:       faker.Email(),
			Phone:       faker.E164PhoneNumber(),
		},
	}
	mockMxServer := NewMockServer(func(s *MockServerState) {
		s.AccountOwners = accountOwners
	})
	db := test_utils.MigrateCockroachDB(t, ctx)
	mx, err := NewService(&ServiceArgs{
		BaseUrl:  mockMxServer.URL,
		Username: "test",
		Password: "test",
		Db:       db,
	})
	if err != nil {
		t.Fatal(err)
	}

	mxFundingSourceID := ""
	scenarios := []struct {
		Name          string
		ExpectedError error
		RunBefore     func()
	}{
		{
			Name:          "Returns account owner details",
			ExpectedError: nil,
			RunBefore: func() {
				mxAccount, err := mx.CreateAccount(ctx, &CreateAccountArgs{
					ID:            uuid.NewString(),
					AccountID:     uuid.NewString(),
					MxUserGuid:    uuid.NewString(),
					MxMemberGuid:  uuid.NewString(),
					MxAccountGuid: mxAccountGuid,
				})
				if err != nil {
					t.Fatal(err)
				}
				mxFundingSourceID = mxAccount.ID
			},
		},
		{
			Name:          "Returns ErrNotFound if mx account guid is not found.",
			ExpectedError: ErrNotFound,
			RunBefore: func() {
				mxAccount, err := mx.CreateAccount(ctx, &CreateAccountArgs{
					ID:            uuid.NewString(),
					AccountID:     uuid.NewString(),
					MxUserGuid:    uuid.NewString(),
					MxMemberGuid:  uuid.NewString(),
					MxAccountGuid: uuid.NewString(),
				})
				if err != nil {
					t.Fatal(err)
				}
				mxFundingSourceID = mxAccount.ID
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(st *testing.T) {
			scenario.RunBefore()

			accountOwner, err := mx.GetAccountOwner(ctx, mxFundingSourceID)

			if scenario.ExpectedError == nil {
				assert.NoError(st, err, scenario.Name)
				owner := accountOwners[0]
				assert.Equal(st, mxAccountGuid, accountOwner.AccountGuid, scenario.Name)
				assert.Equal(st, owner.Country, accountOwner.Country, scenario.Name)
				assert.Equal(st, owner.OwnerName, accountOwner.OwnerName, scenario.Name)
				assert.Equal(st, owner.Email, accountOwner.Email, scenario.Name)
				assert.Equal(st, owner.Phone, accountOwner.Phone, scenario.Name)
			} else {
				assert.Nil(st, accountOwner, scenario.Name)
				assert.ErrorIs(st, err, scenario.ExpectedError, scenario.Name)
			}
		})
	}

}

func TestGetMxAccount(t *testing.T) {
	// TODO: refactor container setup
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)
	mxAccountGuid := uuid.NewString()
	mxUserGuid := uuid.NewString()
	mxMemberGuid := uuid.NewString()
	expectedMxAccount := MxAccount{
		Guid:              mxAccountGuid,
		UserGuid:          mxUserGuid,
		MemberGuid:        mxMemberGuid,
		AccountNumber:     "123",
		InstitutionNumber: "321",
		RoutingNumber:     "68899990000000",
		TransitNumber:     "123",
		CurrencyCode:      "780",
		Type:              "SAVINGS",
		AvailableBalance:  500.00,
		Balance:           500.00,
	}
	mockMxServer := NewMockServer(func(s *MockServerState) {
		s.MxAccount = expectedMxAccount
	})
	mx, err := NewService(&ServiceArgs{
		BaseUrl:  mockMxServer.URL,
		Username: "test",
		Password: "test",
		Db:       db,
	})
	if err != nil {
		t.Fatal(err)
	}

	mxFundingSourceID := ""
	scenarios := []struct {
		Name          string
		ExpectedError error
		RunBefore     func()
	}{
		{
			Name:          "Returns account numbers",
			ExpectedError: nil,
			RunBefore: func() {
				mxAccount, err := mx.CreateAccount(ctx, &CreateAccountArgs{
					ID:            uuid.NewString(),
					AccountID:     uuid.NewString(),
					MxUserGuid:    mxUserGuid,
					MxMemberGuid:  mxMemberGuid,
					MxAccountGuid: mxAccountGuid,
				})
				if err != nil {
					t.Fatal(err)
				}
				mxFundingSourceID = mxAccount.ID
			},
		},
		{
			Name:          "Returns ErrInternal if mx account guid is not found on mx.",
			ExpectedError: ErrInternal,
			RunBefore: func() {
				mxAccount, err := mx.CreateAccount(ctx, &CreateAccountArgs{
					ID:            uuid.NewString(),
					AccountID:     uuid.NewString(),
					MxUserGuid:    uuid.NewString(),
					MxMemberGuid:  uuid.NewString(),
					MxAccountGuid: uuid.NewString(),
				})
				if err != nil {
					t.Fatal(err)
				}
				mxFundingSourceID = mxAccount.ID
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(st *testing.T) {
			scenario.RunBefore()

			mxAccount, err := mx.ReadAccount(ctx, mxFundingSourceID)

			if scenario.ExpectedError == nil {
				assert.NoError(st, err, scenario.Name)
				assert.Equal(st, mxAccountGuid, mxAccount.Guid, scenario.Name)
				assert.Equal(st, expectedMxAccount.InstitutionNumber, mxAccount.InstitutionNumber, scenario.Name)
				assert.Equal(st, expectedMxAccount.AvailableBalance, mxAccount.AvailableBalance, scenario.Name)
				assert.Equal(st, expectedMxAccount.RoutingNumber, mxAccount.RoutingNumber, scenario.Name)
				assert.Equal(st, expectedMxAccount.TransitNumber, mxAccount.TransitNumber, scenario.Name)
			} else {
				assert.Nil(st, mxAccount, scenario.Name)
				assert.ErrorIs(st, err, scenario.ExpectedError, scenario.Name)
			}
		})
	}

}

func TestGetMxUserByAccountID(t *testing.T) {
	// TODO: refactor container setup
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)
	mx, err := NewService(&ServiceArgs{
		BaseUrl:  "localhost:8080",
		Username: "test",
		Password: "test",
		Db:       db,
	})
	if err != nil {
		t.Fatal(err)
	}

	accountID := ""
	expectedMxUserGuid := ""
	scenarios := []struct {
		Name          string
		ExpectedError error
		RunBefore     func(*testing.T)
	}{
		{
			Name:          "Returns mx user guid",
			ExpectedError: nil,
			RunBefore: func(rbt *testing.T) {
				accountID = uuid.NewString()
				expectedMxUserGuid = uuid.NewString()
				_, err = mx.CreateAccount(ctx, &CreateAccountArgs{
					ID:            uuid.NewString(),
					AccountID:     accountID,
					MxUserGuid:    expectedMxUserGuid,
					MxMemberGuid:  uuid.NewString(),
					MxAccountGuid: uuid.NewString(),
				})
				if err != nil {
					rbt.Fatal(err)
				}
				_, err = mx.CreateAccount(ctx, &CreateAccountArgs{
					ID:            uuid.NewString(),
					AccountID:     accountID,
					MxUserGuid:    expectedMxUserGuid,
					MxMemberGuid:  uuid.NewString(),
					MxAccountGuid: uuid.NewString(),
				})
				if err != nil {
					rbt.Fatal(err)
				}
			},
		},
		{
			Name:          "Returns ErrInternal if there is more than one mx user for the account",
			ExpectedError: ErrInternal,
			RunBefore: func(rbt *testing.T) {
				accountID = uuid.NewString()
				expectedMxUserGuid = uuid.NewString()
				_, err = mx.CreateAccount(ctx, &CreateAccountArgs{
					ID:            uuid.NewString(),
					AccountID:     accountID,
					MxUserGuid:    expectedMxUserGuid,
					MxMemberGuid:  uuid.NewString(),
					MxAccountGuid: uuid.NewString(),
				})
				if err != nil {
					rbt.Fatal(err)
				}
				_, err = mx.CreateAccount(ctx, &CreateAccountArgs{
					ID:            uuid.NewString(),
					AccountID:     accountID,
					MxUserGuid:    uuid.NewString(),
					MxMemberGuid:  uuid.NewString(),
					MxAccountGuid: uuid.NewString(),
				})
				if err != nil {
					rbt.Fatal(err)
				}
			},
		},
		{
			Name:          "Returns ErrNotFound if there is no mx user for the account.",
			ExpectedError: ErrNotFound,
			RunBefore: func(rbt *testing.T) {
				accountID = uuid.NewString()
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(st *testing.T) {
			scenario.RunBefore(st)

			mxUserGuid, err := mx.GetMxUserByAccountID(ctx, accountID)

			if scenario.ExpectedError == nil {
				assert.NoError(st, err, scenario.Name)
				assert.Equal(st, expectedMxUserGuid, mxUserGuid)
			} else {
				assert.ErrorIs(st, err, scenario.ExpectedError, scenario.Name)
				assert.Equal(st, "", mxUserGuid, scenario.Name)
			}
		})
	}
}
