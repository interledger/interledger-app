package mx

import (
	"context"
	"errors"
	"testing"

	identity_mock "gitlab.com/fynbos/backend/identity/client/mock"

	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/fynbos/backend/accounts"
	accounts_mock "gitlab.com/fynbos/backend/accounts/client/mock"
	"gitlab.com/fynbos/backend/identity"
	external "gitlab.com/fynbos/backend/providers/mx/external"
	test_utils "gitlab.com/fynbos/backend/utils"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"
)

func TestCreateAndGetAccount(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	db := test_utils.MigrateCockroachDB(t, ctx)
	mockExternalClient := external.NewMockMx(ctrl)
	mx, err := NewService(&ServiceArgs{
		ExternalClient:  mockExternalClient,
		Db:              db,
		AccountsService: accounts_mock.NewMockClient(ctrl),
		IdentityService: identity_mock.NewMockClient(ctrl),
		Temporal:        &mocks.Client{},
	})
	if err != nil {
		t.Fatal(err)
	}
	args := &CreateAccountArgs{
		Guid:            uuid.NewString(),
		UserGuid:        uuid.NewString(),
		MemberGuid:      uuid.NewString(),
		AccountID:       uuid.NewString(),
		FundingsourceID: uuid.NewString(),
	}

	mxAccount, err := mx.CreateAccount(ctx, args)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, args.Guid, mxAccount.Guid)
	assert.Equal(t, args.AccountID, mxAccount.AccountID)
	assert.Equal(t, args.UserGuid, mxAccount.UserGuid)
	assert.Equal(t, args.MemberGuid, mxAccount.MemberGuid)
	assert.Equal(t, args.FundingsourceID, mxAccount.FundingsourceID)

	freshMxFs, err := mx.GetAccount(ctx, args.Guid)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, args.Guid, freshMxFs.Guid)
	assert.Equal(t, args.AccountID, freshMxFs.AccountID)
	assert.Equal(t, args.UserGuid, freshMxFs.UserGuid)
	assert.Equal(t, args.MemberGuid, freshMxFs.MemberGuid)
	assert.Equal(t, args.FundingsourceID, freshMxFs.FundingsourceID)

	noAcc, err := mx.GetAccount(ctx, uuid.NewString())
	assert.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, noAcc)

	// idempotency
	idempotent, err := mx.CreateAccount(ctx, args)
	assert.ErrorIs(t, err, ErrDuplicate)
	assert.Nil(t, idempotent)
}

func TestGetMemberStatus(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	db := test_utils.MigrateCockroachDB(t, ctx)
	mockExternalClient := external.NewMockMx(ctrl)
	mx, err := NewService(&ServiceArgs{
		ExternalClient:  mockExternalClient,
		Db:              db,
		AccountsService: accounts_mock.NewMockClient(ctrl),
		IdentityService: identity_mock.NewMockClient(ctrl),
		Temporal:        &mocks.Client{},
	})
	if err != nil {
		t.Fatal(err)
	}

	mxAccount, err := mx.CreateAccount(ctx, &CreateAccountArgs{
		Guid:            uuid.NewString(),
		UserGuid:        uuid.NewString(),
		MemberGuid:      uuid.NewString(),
		AccountID:       uuid.NewString(),
		FundingsourceID: uuid.NewString(),
	})
	if err != nil {
		t.Fatal(err)
	}
	mockExternalClient.EXPECT().GetMemberStatus(ctx, mxAccount.UserGuid, mxAccount.MemberGuid).
		Return(
			&external.Member{
				Guid:              mxAccount.MemberGuid,
				UserGuid:          mxAccount.UserGuid,
				IsBeingAggregated: false,
			},
			nil,
		).Times(1)

	member, err := mx.GetMemberStatus(ctx, mxAccount.Guid)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, mxAccount.UserGuid, member.UserGuid)
	assert.Equal(t, mxAccount.MemberGuid, member.Guid)
	assert.Equal(t, false, member.IsBeingAggregated)
}

func TestStartIdentityAggregation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	db := test_utils.MigrateCockroachDB(t, ctx)
	mockExternalClient := external.NewMockMx(ctrl)
	mx, err := NewService(&ServiceArgs{
		ExternalClient:  mockExternalClient,
		Db:              db,
		AccountsService: accounts_mock.NewMockClient(ctrl),
		IdentityService: identity_mock.NewMockClient(ctrl),
		Temporal:        &mocks.Client{},
	})
	if err != nil {
		t.Fatal(err)
	}

	mxAccount, err := mx.CreateAccount(ctx, &CreateAccountArgs{
		Guid:            uuid.NewString(),
		UserGuid:        uuid.NewString(),
		MemberGuid:      uuid.NewString(),
		AccountID:       uuid.NewString(),
		FundingsourceID: uuid.NewString(),
	})
	if err != nil {
		t.Fatal(err)
	}
	mockExternalClient.EXPECT().AggregateIdentity(ctx, mxAccount.UserGuid, mxAccount.MemberGuid).
		Return(
			&external.Member{
				Guid:              mxAccount.MemberGuid,
				UserGuid:          mxAccount.UserGuid,
				IsBeingAggregated: true,
			},
			nil,
		).Times(1)

	member, err := mx.StartIdentityAggregation(ctx, mxAccount.Guid)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, mxAccount.UserGuid, member.UserGuid)
	assert.Equal(t, mxAccount.MemberGuid, member.Guid)
	assert.Equal(t, true, member.IsBeingAggregated)
}

func TestGetAccountOwner(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	db := test_utils.MigrateCockroachDB(t, ctx)
	mockExternalClient := external.NewMockMx(ctrl)
	mx, err := NewService(&ServiceArgs{
		ExternalClient:  mockExternalClient,
		Db:              db,
		AccountsService: accounts_mock.NewMockClient(ctrl),
		IdentityService: identity_mock.NewMockClient(ctrl),
		Temporal:        &mocks.Client{},
	})
	if err != nil {
		t.Fatal(err)
	}

	mxAccount, err := mx.CreateAccount(ctx, &CreateAccountArgs{
		Guid:            uuid.NewString(),
		UserGuid:        uuid.NewString(),
		MemberGuid:      uuid.NewString(),
		AccountID:       uuid.NewString(),
		FundingsourceID: uuid.NewString(),
	})
	if err != nil {
		t.Fatal(err)
	}

	scenarios := []struct {
		Name          string
		ExpectedError error
		AccountOwners []external.AccountOwner
	}{
		{
			Name:          "Returns account owner details",
			ExpectedError: nil,
			AccountOwners: []external.AccountOwner{
				{
					AccountGuid: mxAccount.Guid,
					OwnerName:   faker.Name(),
				},
			},
		},
		{
			Name:          "Returns ErrNotFound if mx account guid is not found.",
			ExpectedError: ErrNotFound,
			AccountOwners: []external.AccountOwner{
				{
					AccountGuid: uuid.NewString(),
					OwnerName:   faker.Name(),
				},
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(st *testing.T) {
			mockExternalClient.EXPECT().
				GetAccountOwners(ctx, mxAccount.UserGuid, mxAccount.MemberGuid).
				Return(scenario.AccountOwners, nil).Times(1)

			accountOwner, err := mx.GetAccountOwner(ctx, mxAccount.Guid)

			if scenario.ExpectedError == nil {
				assert.NoError(st, err, scenario.Name)
				owner := scenario.AccountOwners[0]
				assert.Equal(st, mxAccount.Guid, accountOwner.AccountGuid, scenario.Name)
				assert.Equal(st, owner.OwnerName, accountOwner.OwnerName, scenario.Name)
			} else {
				assert.Nil(st, accountOwner, scenario.Name)
				assert.ErrorIs(st, err, scenario.ExpectedError, scenario.Name)
			}
		})
	}

}

func TestReadAccount(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	db := test_utils.MigrateCockroachDB(t, ctx)
	mockExternalClient := external.NewMockMx(ctrl)
	mx, err := NewService(&ServiceArgs{
		ExternalClient:  mockExternalClient,
		Db:              db,
		AccountsService: accounts_mock.NewMockClient(ctrl),
		IdentityService: identity_mock.NewMockClient(ctrl),
		Temporal:        &mocks.Client{},
	})
	if err != nil {
		t.Fatal(err)
	}

	mxAccount, err := mx.CreateAccount(ctx, &CreateAccountArgs{
		Guid:            uuid.NewString(),
		UserGuid:        uuid.NewString(),
		MemberGuid:      uuid.NewString(),
		AccountID:       uuid.NewString(),
		FundingsourceID: uuid.NewString(),
	})
	if err != nil {
		t.Fatal(err)
	}
	scenarios := []struct {
		Name                string
		ExpectedError       error
		ExternalClientError error
		Account             *external.Account
	}{
		{
			Name:                "Returns account numbers",
			ExpectedError:       nil,
			ExternalClientError: nil,
			Account: &external.Account{
				AccountNumber:     "123",
				InstitutionNumber: "321",
				RoutingNumber:     "68899990000000",
				TransitNumber:     "123",
				CurrencyCode:      "780",
				Type:              "SAVINGS",
				AvailableBalance:  500.00,
				Balance:           500.00,
			},
		},
		{
			Name:                "Returns ErrInternal if mx account guid is not found on mx.",
			ExpectedError:       ErrInternal,
			ExternalClientError: errors.New("not found"),
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(st *testing.T) {
			mockExternalClient.EXPECT().ReadAccount(ctx, mxAccount.UserGuid, mxAccount.Guid).
				Return(scenario.Account, scenario.ExternalClientError).Times(1)

			mxAccount, err := mx.ReadAccount(ctx, mxAccount.Guid)

			if scenario.ExpectedError == nil {
				assert.NoError(st, err, scenario.Name)
				assert.Equal(st, mxAccount.Guid, mxAccount.Guid, scenario.Name)
				assert.Equal(st, scenario.Account.InstitutionNumber, mxAccount.InstitutionNumber, scenario.Name)
				assert.Equal(st, scenario.Account.RoutingNumber, mxAccount.RoutingNumber, scenario.Name)
				assert.Equal(st, scenario.Account.TransitNumber, mxAccount.TransitNumber, scenario.Name)
			} else {
				assert.Nil(st, mxAccount, scenario.Name)
				assert.ErrorIs(st, err, scenario.ExpectedError, scenario.Name)
			}
		})
	}

}

func TestGetMxUserByAccountID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	db := test_utils.MigrateCockroachDB(t, ctx)
	mockExternalClient := external.NewMockMx(ctrl)
	mx, err := NewService(&ServiceArgs{
		ExternalClient:  mockExternalClient,
		Db:              db,
		AccountsService: accounts_mock.NewMockClient(ctrl),
		IdentityService: identity_mock.NewMockClient(ctrl),
		Temporal:        &mocks.Client{},
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
					Guid:            uuid.NewString(),
					UserGuid:        expectedMxUserGuid,
					MemberGuid:      uuid.NewString(),
					AccountID:       accountID,
					FundingsourceID: uuid.NewString(),
				})
				if err != nil {
					rbt.Fatal(err)
				}
				_, err = mx.CreateAccount(ctx, &CreateAccountArgs{
					Guid:            uuid.NewString(),
					UserGuid:        expectedMxUserGuid,
					MemberGuid:      uuid.NewString(),
					AccountID:       accountID,
					FundingsourceID: uuid.NewString(),
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
					Guid:            uuid.NewString(),
					UserGuid:        expectedMxUserGuid,
					MemberGuid:      uuid.NewString(),
					AccountID:       accountID,
					FundingsourceID: uuid.NewString(),
				})
				if err != nil {
					rbt.Fatal(err)
				}
				_, err = mx.CreateAccount(ctx, &CreateAccountArgs{
					Guid:            uuid.NewString(),
					UserGuid:        uuid.NewString(),
					MemberGuid:      uuid.NewString(),
					AccountID:       accountID,
					FundingsourceID: uuid.NewString(),
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

func TestVerifyOwnership(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	db := test_utils.MigrateCockroachDB(t, ctx)
	accountService := accounts_mock.NewMockClient(ctrl)
	identityService := identity_mock.NewMockClient(ctrl)
	mockExternalClient := external.NewMockMx(ctrl)
	mx, err := NewService(&ServiceArgs{
		ExternalClient:  mockExternalClient,
		Db:              db,
		AccountsService: accountService,
		IdentityService: identityService,
		Temporal:        &mocks.Client{},
	})
	if err != nil {
		t.Fatal(err)
	}
	userID := uuid.NewString()
	mxAccount, err := mx.CreateAccount(ctx, &CreateAccountArgs{
		Guid:            uuid.NewString(),
		UserGuid:        uuid.NewString(),
		MemberGuid:      uuid.NewString(),
		AccountID:       uuid.NewString(),
		FundingsourceID: uuid.NewString(),
	})
	if err != nil {
		t.Fatal(err)
	}
	accountService.EXPECT().Get(gomock.Any(), mxAccount.AccountID).Return(
		&accounts.Account{
			ID:         mxAccount.Guid,
			IdentityID: userID,
		},
		nil,
	).AnyTimes()
	identityService.EXPECT().Get(ctx, userID).Return(
		&identity.Identity{
			ID:        userID,
			FirstName: "James",
			LastName:  "Bond",
		},
		nil,
	).AnyTimes()

	testcases := []struct {
		Name          string
		ExpectedError error
		AccountOwners []external.AccountOwner
	}{
		{
			Name:          "Verifies if account owner's name is the same as user's name",
			ExpectedError: nil,
			AccountOwners: []external.AccountOwner{
				{
					AccountGuid: mxAccount.Guid,
					OwnerName:   "James bond",
				},
			},
		},
		{
			Name:          "Returns ErrOwnershipCheckFailed if account owner's name does not match user's name",
			ExpectedError: ErrOwnershipCheckFailed,
			AccountOwners: []external.AccountOwner{
				{
					AccountGuid: mxAccount.Guid,
					OwnerName:   "James Blunt",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			mockExternalClient.EXPECT().GetAccountOwners(ctx, mxAccount.UserGuid, mxAccount.MemberGuid).
				Return(tc.AccountOwners, nil).Times(1)

			err = mx.VerifyOwnership(ctx, mxAccount.Guid)

			if tc.ExpectedError == nil {
				assert.NoError(t, err, tc.Name)
			} else {
				assert.ErrorIs(t, err, tc.ExpectedError, tc.Name)
			}
		})
	}
}

func TestGetMxConnectWidget(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	db := test_utils.MigrateCockroachDB(t, ctx)
	accountService := accounts_mock.NewMockClient(ctrl)
	identityService := identity_mock.NewMockClient(ctrl)
	mockExternalClient := external.NewMockMx(ctrl)
	mx, err := NewService(&ServiceArgs{
		ExternalClient:  mockExternalClient,
		Db:              db,
		AccountsService: accountService,
		IdentityService: identityService,
		Temporal:        &mocks.Client{},
	})
	if err != nil {
		t.Fatal(err)
	}

	accountID := uuid.NewString()
	userID := uuid.NewString()
	accountService.EXPECT().Get(gomock.Any(), accountID).Return(&accounts.Account{
		ID:         accountID,
		IdentityID: userID,
	}, nil).AnyTimes()
	mxUserGuid := uuid.NewString()
	mockExternalClient.EXPECT().CreateUser(ctx).Return(mxUserGuid, nil).Times(1)
	mockExternalClient.EXPECT().GetWidgetUrl(ctx, mxUserGuid).Return("localhost", nil)

	url, err := mx.GetConnectWidget(context.Background(), accountID, userID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "localhost", url)
}

func TestInitiateCreateAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	accountsService := accounts_mock.NewMockClient(ctrl)
	identityService := identity_mock.NewMockClient(ctrl)
	temporal := &mocks.Client{}
	mockExternalClient := external.NewMockMx(ctrl)
	mx, err := NewService(&ServiceArgs{
		ExternalClient:  mockExternalClient,
		Db:              &sqlx.DB{},
		AccountsService: accountsService,
		IdentityService: identityService,
		Temporal:        temporal,
	})
	if err != nil {
		t.Fatal(err)
	}

	userID := uuid.NewString()
	accountID := uuid.NewString()
	userGuid := uuid.NewString()
	memberGuid := uuid.NewString()
	fundingSourceName := "test"
	scenarios := []struct {
		Name          string
		ExpectedError error
		RunBefore     func()
	}{
		{
			Name:          "Returns ErrUnauthorized if identity does not own account",
			ExpectedError: ErrUnauthorized,
			RunBefore: func() {
				accountsService.EXPECT().Get(gomock.Any(), accountID).Return(
					&accounts.Account{
						ID:         accountID,
						IdentityID: uuid.NewString(),
					},
					nil,
				).Times(1)
			},
		},
		{
			Name:          "Creates funding source and initiates workflow.",
			ExpectedError: nil,
			RunBefore: func() {
				accountsService.EXPECT().Get(gomock.Any(), accountID).Return(
					&accounts.Account{
						ID:         accountID,
						IdentityID: userID,
					},
					nil,
				).Times(1)
				temporal.On(
					"ExecuteWorkflow",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.MatchedBy(func(args *CreateMxAccountWorkflowArgs) bool {
						_, err := uuid.Parse(args.ID)
						if err != nil {
							return false
						}
						return args.MemberGuid == memberGuid && args.UserGuid == userGuid && args.IdentityID == userID
					}),
				).Return(
					func(ctx context.Context, opts client.StartWorkflowOptions, workflow interface{}, args ...interface{}) client.WorkflowRun {
						testWorkflowID := opts.ID
						testRunID := "test-runid"

						mockWorkflowRun := &mocks.WorkflowRun{}
						mockWorkflowRun.On("GetID").Return(testWorkflowID)
						mockWorkflowRun.On("GetRunID").Return(testRunID)
						mockWorkflowRun.On("Get", mock.Anything, mock.Anything).Return(nil)
						return mockWorkflowRun
					}, nil,
				).Times(1)
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(tc *testing.T) {
			scenario.RunBefore()
			worflowUuid, err := mx.InitiateCreateAccount(
				context.Background(),
				&InitiateCreateAccountArgs{
					IdentityID:        userID,
					AccountID:         accountID,
					UserGuid:          userGuid,
					MemberGuid:        memberGuid,
					FundingsourceName: fundingSourceName,
				},
			)

			if scenario.ExpectedError == nil {
				assert.NoError(tc, err, scenario.Name)
				_, err = uuid.Parse(worflowUuid)
				if err != nil {
					tc.Fatal(err)
				}
			} else {
				assert.ErrorIs(tc, err, scenario.ExpectedError, scenario.Name)
			}
		})
	}
}

func TestGetAccountBalance(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	accountsService := accounts.NewMockService(ctrl)
	identityService := identity.NewMockService(ctrl)
	temporal := &mocks.Client{}
	mockExternalClient := external.NewMockMx(ctrl)
	mx, err := NewService(&ServiceArgs{
		ExternalClient:  mockExternalClient,
		Db:              test_utils.MigrateCockroachDB(t, context.Background()),
		AccountsService: accountsService,
		IdentityService: identityService,
		Temporal:        temporal,
	})
	if err != nil {
		t.Fatal(err)
	}
	mxAccount, err := mx.CreateAccount(context.Background(), &CreateAccountArgs{
		Guid:            uuid.NewString(),
		UserGuid:        uuid.NewString(),
		MemberGuid:      uuid.NewString(),
		AccountID:       uuid.NewString(),
		FundingsourceID: uuid.NewString(),
	})
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		Name            string
		FloatBalance    float64
		ExpectedBalance int64
	}{
		{
			Name:            "Parses positive float balance",
			FloatBalance:    117.19,
			ExpectedBalance: 11719,
		},
		{
			Name:            "Parses negative float balance",
			FloatBalance:    -129.13,
			ExpectedBalance: -12913,
		},
		{
			Name:            "Truncates after cents",
			FloatBalance:    117.9999999,
			ExpectedBalance: 11799,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(st *testing.T) {
			mockExternalClient.EXPECT().ReadAccount(gomock.Any(), mxAccount.UserGuid, mxAccount.Guid).
				Return(
					&external.Account{
						Guid:             mxAccount.Guid,
						UserGuid:         mxAccount.UserGuid,
						MemberGuid:       mxAccount.MemberGuid,
						CurrencyCode:     "USD",
						AvailableBalance: tc.FloatBalance,
					},
					nil,
				).Times(1)

			balance, err := mx.GetAccountBalance(context.Background(), mxAccount.Guid)
			if err != nil {
				st.Fatal(err)
			}

			assert.Equal(st, tc.ExpectedBalance, balance.Value, tc.Name)
			assert.Equal(st, "USD", balance.AssetCode, tc.Name)
			assert.Equal(st, uint8(2), balance.AssetScale, tc.Name)
		})
	}
}
