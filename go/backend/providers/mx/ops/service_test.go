package ops_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"gitlab.com/fynbos/backend/providers/mx/workflows"

	"github.com/bxcodec/faker/v3"
	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/fynbos/backend/accounts"
	accounts_mock "gitlab.com/fynbos/backend/accounts/client/mock"
	"gitlab.com/fynbos/backend/identity"
	identity_mock "gitlab.com/fynbos/backend/identity/client/mock"
	"gitlab.com/fynbos/backend/providers/mx"
	external "gitlab.com/fynbos/backend/providers/mx/external"
	"gitlab.com/fynbos/backend/providers/mx/ops"
	"gitlab.com/fynbos/backend/twilio"
	test_utils "gitlab.com/fynbos/backend/utils"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"
)

type testBackends struct {
	val       *validator.Validate
	extClient *external.MockMx
	db        *sqlx.DB
	acc       *accounts_mock.MockClient
	ident     *identity_mock.MockClient
	temporal  *mocks.Client
	twil      *twilio.MockService
}

func (t testBackends) Validator() *validator.Validate {
	return t.val
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func (t testBackends) Accounts() accounts.Client {
	return t.acc
}

func (t testBackends) Identity() identity.Client {
	return t.ident
}

func (t testBackends) Temporal() client.Client {
	return t.temporal
}

func (t testBackends) Twilio() twilio.Service {
	return t.twil
}

func (t testBackends) MXExternal() external.Mx {
	return t.extClient
}

func getTestBackends(t *testing.T) testBackends {
	ctrl := gomock.NewController(t)

	return testBackends{
		val:       validator.New(),
		extClient: external.NewMockMx(ctrl),
		db:        test_utils.MigrateCockroachDB(t, context.Background()),
		acc:       accounts_mock.NewMockClient(ctrl),
		ident:     identity_mock.NewMockClient(ctrl),
		temporal:  &mocks.Client{},
		twil:      twilio.NewMockService(ctrl),
	}
}

func TestCreateAndGetAccount(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b := getTestBackends(t)
	args := mx.CreateAccountArgs{
		Guid:            uuid.NewString(),
		UserGuid:        uuid.NewString(),
		MemberGuid:      uuid.NewString(),
		AccountID:       uuid.NewString(),
		FundingsourceID: uuid.NewString(),
	}

	mxAccount, err := ops.CreateAccount(ctx, b, args)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, args.Guid, mxAccount.Guid)
	assert.Equal(t, args.AccountID, mxAccount.AccountID)
	assert.Equal(t, args.UserGuid, mxAccount.UserGuid)
	assert.Equal(t, args.MemberGuid, mxAccount.MemberGuid)
	assert.Equal(t, args.FundingsourceID, mxAccount.FundingsourceID)

	freshMxFs, err := ops.GetAccount(ctx, b, args.Guid)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, args.Guid, freshMxFs.Guid)
	assert.Equal(t, args.AccountID, freshMxFs.AccountID)
	assert.Equal(t, args.UserGuid, freshMxFs.UserGuid)
	assert.Equal(t, args.MemberGuid, freshMxFs.MemberGuid)
	assert.Equal(t, args.FundingsourceID, freshMxFs.FundingsourceID)

	noAcc, err := ops.GetAccount(ctx, b, uuid.NewString())
	assert.ErrorIs(t, err, mx.ErrNotFound)
	assert.Nil(t, noAcc)

	// idempotency
	idempotent, err := ops.CreateAccount(ctx, b, args)
	assert.ErrorIs(t, err, mx.ErrDuplicate)
	assert.Nil(t, idempotent)
}

func TestGetMemberStatus(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b := getTestBackends(t)

	userGuid := uuid.NewString()
	memberGuid := uuid.NewString()
	b.extClient.EXPECT().GetMemberStatus(ctx, userGuid, memberGuid).
		Return(
			&external.Member{
				Guid:              memberGuid,
				UserGuid:          userGuid,
				IsBeingAggregated: false,
			},
			nil,
		).Times(1)

	member, err := ops.GetMemberStatus(ctx, b, userGuid, memberGuid)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, userGuid, member.UserGuid)
	assert.Equal(t, memberGuid, member.Guid)
	assert.Equal(t, false, member.IsBeingAggregated)
}

func TestStartIdentityAggregation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b := getTestBackends(t)

	userGuid := uuid.NewString()
	memberGuid := uuid.NewString()
	b.extClient.EXPECT().AggregateIdentity(ctx, userGuid, memberGuid).
		Return(
			&external.Member{
				Guid:              memberGuid,
				UserGuid:          userGuid,
				IsBeingAggregated: true,
			},
			nil,
		).Times(1)

	member, err := ops.StartIdentityAggregation(ctx, b, userGuid, memberGuid)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, userGuid, member.UserGuid)
	assert.Equal(t, memberGuid, member.Guid)
	assert.Equal(t, true, member.IsBeingAggregated)
}

func TestGetAccountOwner(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b := getTestBackends(t)

	mxAccount, err := ops.CreateAccount(ctx, b, mx.CreateAccountArgs{
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
			ExpectedError: mx.ErrNotFound,
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
			b.extClient.EXPECT().
				GetAccountOwners(ctx, mxAccount.UserGuid, mxAccount.MemberGuid).
				Return(scenario.AccountOwners, nil).Times(1)

			accountOwner, err := ops.GetAccountOwner(ctx, b, mx.GetAccountOwnerArgs{
				MxUserGuid:    mxAccount.UserGuid,
				MxMemberGuid:  mxAccount.MemberGuid,
				MxAccountGuid: mxAccount.Guid,
			})

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
	b := getTestBackends(t)

	mxAccount, err := ops.CreateAccount(ctx, b, mx.CreateAccountArgs{
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
				AccountNumber:    "123",
				InstitutionCode:  "321",
				RoutingNumber:    "68899990000000",
				TransitNumber:    "123",
				CurrencyCode:     "780",
				Type:             "SAVINGS",
				AvailableBalance: 500.00,
				Balance:          500.00,
			},
		},
		{
			Name:                "Returns ErrInternal if mx account guid is not found on mx.",
			ExpectedError:       mx.ErrInternal,
			ExternalClientError: errors.New("not found"),
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(st *testing.T) {
			b.extClient.EXPECT().ReadAccount(ctx, mxAccount.UserGuid, mxAccount.Guid).
				Return(scenario.Account, scenario.ExternalClientError).Times(1)

			mxAccount, err := ops.ReadAccount(ctx, b, mxAccount.Guid)

			if scenario.ExpectedError == nil {
				assert.NoError(st, err, scenario.Name)
				assert.Equal(st, mxAccount.Guid, mxAccount.Guid, scenario.Name)
				assert.Equal(st, scenario.Account.InstitutionCode, mxAccount.InstitutionCode, scenario.Name)
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
	b := getTestBackends(t)

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
				_, err := ops.CreateAccount(ctx, b, mx.CreateAccountArgs{
					Guid:            uuid.NewString(),
					UserGuid:        expectedMxUserGuid,
					MemberGuid:      uuid.NewString(),
					AccountID:       accountID,
					FundingsourceID: uuid.NewString(),
				})
				if err != nil {
					rbt.Fatal(err)
				}
				_, err = ops.CreateAccount(ctx, b, mx.CreateAccountArgs{
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
			ExpectedError: mx.ErrInternal,
			RunBefore: func(rbt *testing.T) {
				accountID = uuid.NewString()
				expectedMxUserGuid = uuid.NewString()
				_, err := ops.CreateAccount(ctx, b, mx.CreateAccountArgs{
					Guid:            uuid.NewString(),
					UserGuid:        expectedMxUserGuid,
					MemberGuid:      uuid.NewString(),
					AccountID:       accountID,
					FundingsourceID: uuid.NewString(),
				})
				if err != nil {
					rbt.Fatal(err)
				}
				_, err = ops.CreateAccount(ctx, b, mx.CreateAccountArgs{
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
			ExpectedError: mx.ErrNotFound,
			RunBefore: func(rbt *testing.T) {
				accountID = uuid.NewString()
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(st *testing.T) {
			scenario.RunBefore(st)

			mxUserGuid, err := ops.GetMxUserByAccountID(ctx, b, accountID)

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
	b := getTestBackends(t)

	userID := uuid.NewString()
	mxAccount, err := ops.CreateAccount(ctx, b, mx.CreateAccountArgs{
		Guid:            uuid.NewString(),
		UserGuid:        uuid.NewString(),
		MemberGuid:      uuid.NewString(),
		AccountID:       uuid.NewString(),
		FundingsourceID: uuid.NewString(),
	})
	if err != nil {
		t.Fatal(err)
	}
	b.acc.EXPECT().Get(gomock.Any(), mxAccount.AccountID).Return(
		&accounts.Account{
			ID:         mxAccount.Guid,
			IdentityID: userID,
		},
		nil,
	).AnyTimes()

	testcases := []struct {
		Name          string
		ExpectedError error
		AccountOwners []external.AccountOwner
		User          *identity.Identity
		FynbosEnv     string
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
			User: &identity.Identity{
				ID:        userID,
				FirstName: "James",
				LastName:  "Bond",
			},
			FynbosEnv: "prod",
		},
		{
			Name:          "Returns ErrOwnershipCheckFailed if account owner's name does not match user's name",
			ExpectedError: mx.ErrOwnershipCheckFailed,
			AccountOwners: []external.AccountOwner{
				{
					AccountGuid: mxAccount.Guid,
					OwnerName:   "James Blunt",
				},
			},
			User: &identity.Identity{
				ID:        userID,
				FirstName: "James",
				LastName:  "Bond",
			},
			FynbosEnv: "prod",
		},
		{
			Name:          "Auto verifies MX USER when not in prod",
			ExpectedError: nil,
			User: &identity.Identity{
				ID:        userID,
				FirstName: "mx",
				LastName:  "user",
			},
			// we still make a call to get the account owner details
			AccountOwners: []external.AccountOwner{
				{
					AccountGuid: mxAccount.Guid,
					OwnerName:   "James bond",
				},
			},
			FynbosEnv: "testing",
		},
		{
			Name:          "Does not auto verify MX USER when in prod",
			ExpectedError: mx.ErrOwnershipCheckFailed,
			AccountOwners: []external.AccountOwner{
				{
					AccountGuid: mxAccount.Guid,
					OwnerName:   "James bond",
				},
			},
			User: &identity.Identity{
				ID:        userID,
				FirstName: "mx",
				LastName:  "user",
			},
			FynbosEnv: "prod",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			os.Setenv("FYNBOS_ENV", tc.FynbosEnv)
			b.extClient.EXPECT().GetAccountOwners(ctx, mxAccount.UserGuid, mxAccount.MemberGuid).
				Return(tc.AccountOwners, nil).Times(1)
			b.ident.EXPECT().Get(ctx, userID).Return(tc.User, nil).Times(1)

			err = ops.VerifyOwnership(ctx, b, mx.VerifyOwnershipArgs{
				AccountID:     mxAccount.AccountID,
				MxUserGuid:    mxAccount.UserGuid,
				MxMemberGuid:  mxAccount.MemberGuid,
				MxAccountGuid: mxAccount.Guid,
			})

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
	b := getTestBackends(t)

	accountID := uuid.NewString()
	userID := uuid.NewString()
	b.acc.EXPECT().Get(gomock.Any(), accountID).Return(&accounts.Account{
		ID:         accountID,
		IdentityID: userID,
	}, nil).AnyTimes()
	mxUserGuid := uuid.NewString()
	b.extClient.EXPECT().CreateUser(ctx).Return(mxUserGuid, nil).Times(1)
	b.extClient.EXPECT().GetWidgetUrl(ctx, mxUserGuid).Return("localhost", nil)

	url, err := ops.GetConnectWidget(context.Background(), b, accountID, userID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "localhost", url)
}

func TestInitiateCreateAccount(t *testing.T) {

	b := getTestBackends(t)

	userID := uuid.NewString()
	accountID := uuid.NewString()
	userGuid := uuid.NewString()
	memberGuid := uuid.NewString()
	scenarios := []struct {
		Name          string
		ExpectedError error
		RunBefore     func()
	}{
		{
			Name:          "Returns ErrUnauthorized if identity does not own account",
			ExpectedError: mx.ErrUnauthorized,
			RunBefore: func() {
				b.acc.EXPECT().Get(gomock.Any(), accountID).Return(
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
				b.acc.EXPECT().Get(gomock.Any(), accountID).Return(
					&accounts.Account{
						ID:         accountID,
						IdentityID: userID,
					},
					nil,
				).Times(1)
				b.temporal.On(
					"ExecuteWorkflow",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.MatchedBy(func(args *workflows.CreateMxAccountWorkflowArgs) bool {
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
			worflowUuid, err := ops.InitiateCreateAccount(
				context.Background(),
				b,
				mx.InitiateCreateAccountArgs{
					IdentityID: userID,
					AccountID:  accountID,
					UserGuid:   userGuid,
					MemberGuid: memberGuid,
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
	b := getTestBackends(t)
	mxAccount, err := ops.CreateAccount(context.Background(), b, mx.CreateAccountArgs{
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
			b.extClient.EXPECT().ReadAccount(gomock.Any(), mxAccount.UserGuid, mxAccount.Guid).
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

			balance, err := ops.GetAccountBalance(context.Background(), b, mxAccount.Guid)
			if err != nil {
				st.Fatal(err)
			}

			assert.Equal(st, tc.ExpectedBalance, balance.Value, tc.Name)
			assert.Equal(st, "USD", balance.AssetCode, tc.Name)
			assert.Equal(st, uint8(2), balance.AssetScale, tc.Name)
		})
	}
}

func TestInitiateCreateFundingsource(t *testing.T) {
	t.Parallel()
	b := getTestBackends(t)
	userID := uuid.NewString()
	accountID := uuid.NewString()
	fundingsourceID := uuid.NewString()
	b.acc.EXPECT().Get(gomock.Any(), accountID).Return(
		&accounts.Account{
			ID:         accountID,
			IdentityID: userID,
		},
		nil,
	).AnyTimes()
	b.ident.EXPECT().Get(gomock.Any(), userID).Return(
		&identity.Identity{ID: userID, MobileNumber: "+275555555"},
		nil,
	).AnyTimes()

	t.Run("fails if mxAccount does not belong to account", func(st *testing.T) {
		mxAccount, err := ops.CreateAccount(context.Background(), b, mx.CreateAccountArgs{
			Guid:            uuid.NewString(),
			UserGuid:        uuid.NewString(),
			MemberGuid:      uuid.NewString(),
			AccountID:       uuid.NewString(),
			FundingsourceID: fundingsourceID,
		})
		if err != nil {
			st.Fatal(err)
		}

		err = ops.InitiateCreateFundingsource(context.Background(), b, mx.InitiateCreateFundingsourceArgs{
			AccountID:     accountID,
			Otp:           "1234",
			Name:          "test",
			MxAccountGuid: mxAccount.Guid,
		})
		assert.ErrorIs(st, err, mx.ErrUnauthorized)
	})

	t.Run("fails if otp is invalid", func(st *testing.T) {
		mxAccount, err := ops.CreateAccount(context.Background(), b, mx.CreateAccountArgs{
			Guid:            uuid.NewString(),
			UserGuid:        uuid.NewString(),
			MemberGuid:      uuid.NewString(),
			AccountID:       accountID,
			FundingsourceID: fundingsourceID,
		})
		if err != nil {
			st.Fatal(err)
		}
		b.twil.EXPECT().CheckVerificationCode(gomock.Any(), &twilio.CheckVerificationCodeArgs{
			PhoneNumber: "+275555555",
			Code:        "1234",
		}).Return(
			&twilio.Verification{
				Status: "denied",
			},
			nil,
		).Times(1)

		err = ops.InitiateCreateFundingsource(context.Background(), b, mx.InitiateCreateFundingsourceArgs{
			AccountID:     accountID,
			Otp:           "1234",
			Name:          "test",
			MxAccountGuid: mxAccount.Guid,
		})
		assert.Error(st, err)
	})

	t.Run("executes workflow", func(st *testing.T) {
		mxAccount, err := ops.CreateAccount(context.Background(), b, mx.CreateAccountArgs{
			Guid:            uuid.NewString(),
			UserGuid:        uuid.NewString(),
			MemberGuid:      uuid.NewString(),
			AccountID:       accountID,
			FundingsourceID: fundingsourceID,
		})
		if err != nil {
			st.Fatal(err)
		}
		b.twil.EXPECT().CheckVerificationCode(gomock.Any(), &twilio.CheckVerificationCodeArgs{
			PhoneNumber: "+275555555",
			Code:        "1234",
		}).Return(
			&twilio.Verification{
				Status: "approved",
			},
			nil,
		).Times(1)
		b.temporal.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
			func(ctx context.Context, opts client.StartWorkflowOptions, workflow interface{}, args ...interface{}) client.WorkflowRun {
				workflowArgs, ok := args[0].(*workflows.MxCreateFundingsourceWorkflowArgs)
				if !ok {
					st.Fatal("incorrect args to workflow.")
				}
				assert.Equal(st, accountID, workflowArgs.AccountID)
				assert.Equal(st, mxAccount.Guid, workflowArgs.MxAccountGuid)
				assert.Equal(st, "test", workflowArgs.Name)

				testWorkflowID := opts.ID
				testRunID := "test-runid"

				mockWorkflowRun := &mocks.WorkflowRun{}
				mockWorkflowRun.On("GetID").Return(testWorkflowID)
				mockWorkflowRun.On("GetRunID").Return(testRunID)
				mockWorkflowRun.On("Get", mock.Anything, mock.Anything).Return(nil)
				return mockWorkflowRun
			}, nil,
		).Times(1)

		err = ops.InitiateCreateFundingsource(context.Background(), b, mx.InitiateCreateFundingsourceArgs{
			AccountID:     accountID,
			Otp:           "1234",
			Name:          "test",
			MxAccountGuid: mxAccount.Guid,
		})
		if err != nil {
			st.Fatal(err)
		}
	})
}
