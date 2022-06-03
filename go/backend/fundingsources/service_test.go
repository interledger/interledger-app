package fundingsources

import (
	"context"
	"crypto/sha256"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	_mx "gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/noop"
	_unit "gitlab.com/fynbos/backend/providers/unit"
	"gitlab.com/fynbos/backend/user"
	_user "gitlab.com/fynbos/backend/user"
	test_utils "gitlab.com/fynbos/backend/utils"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"
)

func TestFundingSources(s *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(s)
	c, err := NewTestContainer(ctx, s, ctrl)
	if err != nil {
		s.Fatal(err)
	}

	s.Run("validates create bank account arguments", func(t *testing.T) {
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}

		id, err := c.Is.Create(c.Ctx, &identity.CreateArgs{
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
		_, err = NewAccount(c, &onboarding.CreateAccountArgs{
			IdentityID: id.ID,
			Country:    id.Country,
		})
		if err != nil {
			t.Fatal(err)
		}
		type Scenario struct {
			Name                 string
			Args                 *CreateBankAccountArgs
			ExpectedErrorMessage string
			ExpectedError        error
		}
		scenarios := []Scenario{
			{
				Name:                 "IdentityID is required to create bank account",
				Args:                 generateCreateBankAccountArgs(withIdentityID("")),
				ExpectedErrorMessage: "Key: 'CreateBankAccountArgs.IdentityID' Error:Field validation for 'IdentityID' failed on the 'required' tag",
				ExpectedError:        ErrInvalidArgument,
			},
			{
				Name:                 "IdentityID must exist to create bank account",
				Args:                 generateCreateBankAccountArgs(withIdentityID(uuid.NewString())),
				ExpectedErrorMessage: "not found.",
				ExpectedError:        ErrInternal,
			},
			{
				Name:                 "AccountID is required to create bank account",
				Args:                 generateCreateBankAccountArgs(withAccountID("")),
				ExpectedErrorMessage: "Key: 'CreateBankAccountArgs.AccountID' Error:Field validation for 'AccountID' failed on the 'required' tag",
				ExpectedError:        ErrInvalidArgument,
			},
			{
				Name:                 "AccountID must exist to create bank account",
				Args:                 generateCreateBankAccountArgs(withIdentityID(id.ID), withAccountID(uuid.NewString())),
				ExpectedErrorMessage: "not found.",
				ExpectedError:        ErrInternal,
			},
			{
				Name:                 "Name is required to create bank account",
				Args:                 generateCreateBankAccountArgs(withName("")),
				ExpectedErrorMessage: "Key: 'CreateBankAccountArgs.Name' Error:Field validation for 'Name' failed on the 'required' tag",
				ExpectedError:        ErrInvalidArgument,
			},
			{
				Name:                 "Institution is required to create bank account",
				Args:                 generateCreateBankAccountArgs(withInstitution("")),
				ExpectedErrorMessage: "Key: 'CreateBankAccountArgs.Institution' Error:Field validation for 'Institution' failed on the 'required' tag",
				ExpectedError:        ErrInvalidArgument,
			},
			{
				Name:                 "AccountNumber is required to create bank account",
				Args:                 generateCreateBankAccountArgs(withAccountNumber("")),
				ExpectedErrorMessage: "Key: 'CreateBankAccountArgs.AccountNumber' Error:Field validation for 'AccountNumber' failed on the 'required' tag",
				ExpectedError:        ErrInvalidArgument,
			},
			{
				Name:                 "RoutingNumber is required to create bank account",
				Args:                 generateCreateBankAccountArgs(withRoutingNumber("")),
				ExpectedErrorMessage: "Key: 'CreateBankAccountArgs.RoutingNumber' Error:Field validation for 'RoutingNumber' failed on the 'required' tag",
				ExpectedError:        ErrInvalidArgument,
			},
			{
				Name:                 "Type must be one of noop required to create bank account",
				Args:                 generateCreateBankAccountArgs(withType("")),
				ExpectedErrorMessage: "Key: 'CreateBankAccountArgs.Type' Error:Field validation for 'Type' failed on the 'required' tag",
				ExpectedError:        ErrInvalidArgument,
			},
		}

		for _, scenario := range scenarios {
			fs, err := c.Fs.CreateBankAccount(c.Ctx, scenario.Args)
			if err == nil {
				t.Fatal(scenario.Name)
			}

			assert.ErrorIs(t, err, scenario.ExpectedError)
			assert.Contains(t, err.Error(), scenario.ExpectedErrorMessage)
			assert.Nil(t, fs)
		}
	})

	s.Run("creates bank account", func(t *testing.T) {
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}

		id, err := c.Is.Create(c.Ctx, &identity.CreateArgs{
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
		acc, err := NewAccount(c, &onboarding.CreateAccountArgs{
			IdentityID: id.ID,
			Country:    id.Country,
		})
		if err != nil {
			t.Fatal(err)
		}
		args := generateCreateBankAccountArgs(
			withIdentityID(id.ID),
			withAccountID(acc.ID),
		)
		fs, err := c.Fs.CreateBankAccount(ctx, args)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, args.Name, fs.Name)
		assert.NotEqual(t, args.AccountNumber, fs.Mask)
		assert.Equal(t, "noop", fs.Type)
		assert.Equal(t, args.Type, fs.SubType)
		assert.Equal(t, acc.ID, fs.AccountID)
		assert.Equal(t, "required", fs.VerificationState)
	})

	s.Run("verifies bank account", func(t *testing.T) {
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}

		id, err := c.Is.Create(c.Ctx, &identity.CreateArgs{
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
		acc, err := NewAccount(c, &onboarding.CreateAccountArgs{
			IdentityID: id.ID,
			Country:    id.Country,
		})
		if err != nil {
			t.Fatal(err)
		}
		args := generateCreateBankAccountArgs(
			withIdentityID(id.ID),
			withAccountID(acc.ID),
		)
		bankAccount, err := c.Fs.CreateBankAccount(ctx, args)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "required", bankAccount.VerificationState)

		fs, err := c.Fs.Verify(ctx, &VerifyArgs{
			IdentityID:      id.ID,
			FundingSourceID: bankAccount.ID,
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "verified", fs.VerificationState)
	})

	s.Run("returns not found if funding source does not belong to user", func(t *testing.T) {
		otherUser := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}

		otherUserID, err := c.Is.Create(c.Ctx, &identity.CreateArgs{
			ID:           otherUser.ID,
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			MobileNumber: faker.E164PhoneNumber(),
			Email:        otherUser.Email,
			Country:      "US",
		})
		if err != nil {
			t.Fatal(err)
		}

		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}

		myID, err := c.Is.Create(c.Ctx, &identity.CreateArgs{
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
		myAcc, err := NewAccount(c, &onboarding.CreateAccountArgs{
			IdentityID: myID.ID,
			Country:    myID.Country,
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = NewAccount(c, &onboarding.CreateAccountArgs{
			IdentityID: otherUserID.ID,
			Country:    otherUserID.Country,
		})
		if err != nil {
			t.Fatal(err)
		}
		args := generateCreateBankAccountArgs(
			withIdentityID(myID.ID),
			withAccountID(myAcc.ID),
		)
		bankAccount, err := c.Fs.CreateBankAccount(ctx, args)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "required", bankAccount.VerificationState)

		fs, err := c.Fs.Verify(ctx, &VerifyArgs{
			IdentityID:      otherUserID.ID,
			FundingSourceID: bankAccount.ID,
		})
		if err == nil {
			t.Fatal("User must only be able to verify their own funding sources.")
		}

		assert.ErrorIs(t, err, ErrUnauthorized)
		assert.Contains(t, err.Error(), "unauthorized.")
		assert.Nil(t, fs)
	})

	s.Run("returns not found if funding source does not exist", func(t *testing.T) {
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}

		id, err := c.Is.Create(c.Ctx, &identity.CreateArgs{
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
		_, err = NewAccount(c, &onboarding.CreateAccountArgs{
			IdentityID: id.ID,
			Country:    id.Country,
		})
		if err != nil {
			t.Fatal(err)
		}

		fs, err := c.Fs.Verify(ctx, &VerifyArgs{
			IdentityID:      id.ID,
			FundingSourceID: uuid.NewString(),
		})
		if err == nil {
			t.Fatal("Must only be able to verify funding sources that exist.")
		}

		assert.Nil(t, fs)
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Contains(t, err.Error(), "not found.")
	})

	s.Run("get user's funding sources", func(t *testing.T) {
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}

		id, err := c.Is.Create(c.Ctx, &identity.CreateArgs{
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
		acc, err := NewAccount(c, &onboarding.CreateAccountArgs{
			IdentityID: id.ID,
			Country:    id.Country,
		})
		if err != nil {
			t.Fatal(err)
		}
		fundingsource, err := c.Fs.CreateBankAccount(
			ctx,
			generateCreateBankAccountArgs(withIdentityID(id.ID), withAccountID(acc.ID)),
		)
		if err != nil {
			t.Fatal(err)
		}
		fundingsource1, err := c.Fs.CreateBankAccount(
			ctx,
			generateCreateBankAccountArgs(withIdentityID(id.ID), withAccountID(acc.ID)),
		)
		if err != nil {
			t.Fatal(err)
		}

		t.Run("returns a list of all the users funding sources", func(tt *testing.T) {
			fs, err := c.Fs.GetByAccountId(ctx, acc.ID)
			if err != nil {
				tt.Fatal(err)
			}

			fundingSourcesIDs := []string{fs[1].ID, fs[0].ID}

			assert.Equal(tt, 2, len(fs))
			assert.Contains(tt, fundingSourcesIDs, fundingsource.ID)
			assert.Contains(tt, fundingSourcesIDs, fundingsource1.ID)
		})

		t.Run("returns an empty list if a user has no funding sources", func(tt *testing.T) {
			user := &_user.User{
				ID:    uuid.NewString(),
				Email: faker.Email(),
			}

			id, err := c.Is.Create(c.Ctx, &identity.CreateArgs{
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
			otherAcc, err := NewAccount(c, &onboarding.CreateAccountArgs{
				IdentityID: id.ID,
				Country:    id.Country,
			})
			if err != nil {
				t.Fatal(err)
			}

			fs, err := c.Fs.GetByAccountId(ctx, otherAcc.ID)
			if err != nil {
				tt.Fatal(err)
			}

			assert.Equal(tt, 0, len(fs))
		})
	})
}

func TestGetMxConnectWidget(t *testing.T) {
	// TODO: refactor test container
	ctrl := gomock.NewController(t)
	as := accounts.NewMockService(ctrl)
	is := identity.NewMockService(ctrl)
	nos := noop.NewMockService(ctrl)
	mx := _mx.NewMockService(ctrl)
	tp := &mocks.Client{}
	unit := _unit.NewMockService(ctrl)
	fs, err := NewService(&ServiceArgs{
		Is:   is,
		As:   as,
		Db:   &sqlx.DB{},
		Noop: nos,
		Mx:   mx,
		Tp:   tp,
		Unit: unit,
	})
	if err != nil {
		t.Fatal(err)
	}

	accountID := uuid.NewString()
	user := user.User{
		ID: uuid.NewString(),
	}
	widgetUrl := "test"

	scenarios := []struct {
		Name          string
		ExpectedError error
		AccountID     string
		IdentityID    string
		RunBefore     func()
	}{
		{
			Name:          "Returns unauthorized if identity does not match that on account.",
			ExpectedError: ErrUnauthorized,
			AccountID:     accountID,
			IdentityID:    uuid.NewString(),
			RunBefore: func() {
				as.EXPECT().Get(gomock.Any(), accountID).Return(&accounts.Account{
					ID:         accountID,
					IdentityID: user.ID,
				}, nil).Times(1)
			},
		},
		{
			Name:          "Creates mx user and generates widget url.",
			ExpectedError: nil,
			AccountID:     accountID,
			IdentityID:    user.ID,
			RunBefore: func() {
				as.EXPECT().Get(gomock.Any(), accountID).Return(&accounts.Account{
					ID:         accountID,
					IdentityID: user.ID,
				}, nil).Times(1)

				mxUserGuid := uuid.NewString()
				mx.EXPECT().CreateUser(gomock.Any()).Return(mxUserGuid, nil).Times(1)
				mx.EXPECT().GetWidgetUrl(gomock.Any(), mxUserGuid).Return(widgetUrl, nil)
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.RunBefore()

		url, err := fs.GetMxConnectWidget(context.Background(), scenario.AccountID, scenario.IdentityID)

		if scenario.ExpectedError == nil {
			assert.NoError(t, err)
			assert.Equal(t, widgetUrl, url)
		} else {
			assert.ErrorIs(t, err, scenario.ExpectedError, scenario.Name)
			assert.Equal(t, "", url, scenario.Name)
		}
	}
}

func TestCreateMxBankAccount(t *testing.T) {
	// TODO: refactor test container
	db, dbCleanup := test_utils.MigrateCockroachDB(t, context.Background())
	t.Cleanup(func() {
		dbCleanup()
	})
	ctrl := gomock.NewController(t)
	as := accounts.NewMockService(ctrl)
	is := identity.NewMockService(ctrl)
	nos := noop.NewMockService(ctrl)
	mx := _mx.NewMockService(ctrl)
	tp := &mocks.Client{}
	unit := _unit.NewMockService(ctrl)
	fs, err := NewService(&ServiceArgs{
		Is:   is,
		As:   as,
		Db:   db,
		Noop: nos,
		Mx:   mx,
		Tp:   tp,
		Unit: unit,
	})
	if err != nil {
		t.Fatal(err)
	}

	userID := uuid.NewString()
	accountID := uuid.NewString()
	mxUserGuid := uuid.NewString()
	mxMemberGuid := uuid.NewString()
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
				as.EXPECT().Get(gomock.Any(), accountID).Return(
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
				as.EXPECT().Get(gomock.Any(), accountID).Return(
					&accounts.Account{
						ID:         accountID,
						IdentityID: userID,
					},
					nil,
				).Times(2)
				is.EXPECT().Get(gomock.Any(), userID).Return(&identity.Identity{
					ID: userID,
				}, nil).Times(1)
				as.EXPECT().CanCreateFundingSource(gomock.Any(), userID).Return(true).Times(1)
				tp.On(
					"ExecuteWorkflow",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.MatchedBy(func(args *CreateMxBankAccountWorkflowArgs) bool {
						_, err := uuid.Parse(args.FundingSourceID)
						if err != nil {
							return false
						}
						return args.MxMemberGuid == mxMemberGuid && args.MxUserGuid == mxUserGuid && args.IdentityID == userID
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
		t.Run(scenario.Name, func(st *testing.T) {
			scenario.RunBefore()
			fundingSource, err := fs.CreateMxBankAccount(
				context.Background(),
				&CreateMxBankAccountArgs{
					IdentityID:   userID,
					AccountID:    accountID,
					MxUserGuid:   mxUserGuid,
					MxMemberGuid: mxMemberGuid,
					Name:         fundingSourceName,
				},
			)

			if scenario.ExpectedError == nil {
				assert.NoError(st, err, scenario.Name)
				assert.Equal(st, "processing", fundingSource.VerificationState, scenario.Name)
			} else {
				assert.ErrorIs(st, err, ErrUnauthorized, scenario.Name)
				assert.Nil(st, fundingSource, scenario.Name)
			}
		})
	}
}

func TestVerifyMxBankAccount(t *testing.T) {
	// TODO: refactor once we settle on a container pattern
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	as := accounts.NewMockService(ctrl)
	is := identity.NewMockService(ctrl)
	nos := noop.NewMockService(ctrl)
	mx := _mx.NewMockService(ctrl)
	tp := &mocks.Client{}
	db, dbCleanup := test_utils.MigrateCockroachDB(t, ctx)
	t.Cleanup(func() {
		dbCleanup()
	})
	fs, err := NewService(&ServiceArgs{
		Is:   is,
		As:   as,
		Db:   db,
		Noop: nos,
		Mx:   mx,
		Tp:   tp,
	})
	if err != nil {
		t.Fatal(err)
	}

	identityID := ""
	accountID := ""
	mxUserGuid := ""
	mxMemberGuid := ""
	mxAccountGuid := ""
	fundingSourceID := ""
	as.EXPECT().CanCreateFundingSource(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
	as.EXPECT().CanVerifyFundingSource(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
	scenarios := []struct {
		Name          string
		ExpectedError error
		RunBefore     func()
	}{
		{
			Name:          "Returns verified funding source.",
			ExpectedError: nil,
			RunBefore: func() {
				email := faker.Email()
				identityID = uuid.NewString()
				accountID = uuid.NewString()
				mxUserGuid = uuid.NewString()
				mxMemberGuid = uuid.NewString()
				mxAccountGuid = uuid.NewString()
				fundingSourceID = uuid.NewString()
				as.EXPECT().Get(ctx, accountID).Return(
					&accounts.Account{
						ID:         accountID,
						IdentityID: identityID,
					},
					nil,
				).Times(3)

				is.EXPECT().Get(ctx, identityID).Return(
					&identity.Identity{
						ID:    identityID,
						Email: email,
					},
					nil,
				).Times(3)

				fundingsource, err := fs.Create(ctx, &CreateArgs{
					IdentityID:        identityID,
					AccountID:         accountID,
					Name:              "test",
					VerificationState: string(PROCESSING),
					Type:              "mx",
					SubType:           "bank",
				})
				if err != nil {
					t.Fatal(err)
				}
				fundingSourceID = fundingsource.ID

				mx.EXPECT().GetMxFundingSource(ctx, fundingSourceID).Return(
					&_mx.MxFundingSource{
						ID:              fundingsource.ID,
						MxAccountGuidID: mxAccountGuid,
						MxUserGuid:      mxUserGuid,
						MxMemberGuid:    mxMemberGuid,
						AccountID:       accountID,
					},
					nil,
				).Times(1)

				mx.EXPECT().GetAccountOwner(ctx, fundingSourceID).Return(
					&_mx.AccountOwner{
						AccountGuid: mxAccountGuid,
						OwnerName:   faker.Name(),
						Country:     "US",
						Email:       email,
						Phone:       faker.E164PhoneNumber(),
					},
					nil,
				).Times(1)
			},
		},
		{
			Name:          "Returns ErrUnauthorized if user email does not match that returned from mx.",
			ExpectedError: ErrUnauthorized,
			RunBefore: func() {
				identityID = uuid.NewString()
				accountID = uuid.NewString()
				mxUserGuid = uuid.NewString()
				mxMemberGuid = uuid.NewString()
				mxAccountGuid = uuid.NewString()
				fundingSourceID = uuid.NewString()
				as.EXPECT().Get(ctx, accountID).Return(
					&accounts.Account{
						ID:         accountID,
						IdentityID: identityID,
					},
					nil,
				).Times(2)

				is.EXPECT().Get(ctx, identityID).Return(
					&identity.Identity{
						ID:    identityID,
						Email: faker.Email(),
					},
					nil,
				).Times(2)

				fundingsource, err := fs.Create(ctx, &CreateArgs{
					IdentityID:        identityID,
					AccountID:         accountID,
					Name:              "test",
					VerificationState: string(PROCESSING),
					Type:              "mx",
					SubType:           "bank",
				})
				if err != nil {
					t.Fatal(err)
				}
				fundingSourceID = fundingsource.ID

				mx.EXPECT().GetMxFundingSource(ctx, fundingSourceID).Return(
					&_mx.MxFundingSource{
						ID:              fundingsource.ID,
						MxAccountGuidID: mxAccountGuid,
						MxUserGuid:      mxUserGuid,
						MxMemberGuid:    mxMemberGuid,
						AccountID:       accountID,
					},
					nil,
				).Times(1)

				mx.EXPECT().GetAccountOwner(ctx, fundingSourceID).Return(
					&_mx.AccountOwner{
						AccountGuid: mxAccountGuid,
						OwnerName:   faker.Name(),
						Country:     "US",
						Email:       faker.Email(),
						Phone:       faker.E164PhoneNumber(),
					},
					nil,
				).Times(1)
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(st *testing.T) {
			scenario.RunBefore()

			fundingsource, err := fs.VerifyMxBankAccount(ctx, identityID, fundingSourceID)

			if scenario.ExpectedError == nil {
				assert.NoError(st, err)
				assert.Equal(st, string(VERIFIED), fundingsource.VerificationState, scenario.Name)
			} else {
				assert.ErrorIs(st, err, scenario.ExpectedError, scenario.Name)
				assert.Nil(st, fundingsource, scenario.Name)
			}
		})
	}
}

func TestCreateUnitCounterPartyFromMxAccount(t *testing.T) {
	// TODO: refactor once we settle on a container pattern
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	as := accounts.NewMockService(ctrl)
	is := identity.NewMockService(ctrl)
	nos := noop.NewMockService(ctrl)
	mx := _mx.NewMockService(ctrl)
	tp := &mocks.Client{}
	unit := _unit.NewMockService(ctrl)
	db, dbCleanup := test_utils.MigrateCockroachDB(t, ctx)
	t.Cleanup(func() {
		dbCleanup()
	})
	fs, err := NewService(&ServiceArgs{
		Is:   is,
		As:   as,
		Db:   db,
		Noop: nos,
		Mx:   mx,
		Tp:   tp,
		Unit: unit,
	})
	if err != nil {
		t.Fatal(err)
	}

	userID := uuid.NewString()
	accountID := uuid.NewString()
	fundingsourceID := ""
	unitCounterPartyID := ""
	unitCustomerID := ""

	as.EXPECT().Get(gomock.Any(), accountID).Return(
		&accounts.Account{
			ID:         accountID,
			IdentityID: userID,
		},
		nil,
	).AnyTimes()
	is.EXPECT().Get(gomock.Any(), userID).Return(&identity.Identity{
		ID: userID,
	}, nil).AnyTimes()
	as.EXPECT().CanCreateFundingSource(gomock.Any(), userID).Return(true).AnyTimes()

	scenarios := []struct {
		Name          string
		ExpectedError error
		RunBefore     func()
	}{
		{
			Name:          "Creates unit counter party",
			ExpectedError: nil,
			RunBefore: func() {
				mxFs, err := fs.Create(ctx, &CreateArgs{
					IdentityID:        userID,
					AccountID:         accountID,
					Name:              "test",
					Mask:              "",
					VerificationState: string(PROCESSING),
					Type:              "mx",
					SubType:           "bank",
				})
				if err != nil {
					t.Fatal(err)
				}
				fundingsourceID = mxFs.ID

				mxAccount := &_mx.MxAccount{
					Guid:              uuid.NewString(),
					UserGuid:          uuid.NewString(),
					MemberGuid:        uuid.NewString(),
					AccountNumber:     "123",
					InstitutionNumber: "321",
					RoutingNumber:     "abc",
					TransitNumber:     "cba",
					CurrencyCode:      "780",
					Type:              "SAVINGS",
					AvailableBalance:  500.00,
					Balance:           500.00,
				}
				mx.EXPECT().GetMxAccount(gomock.Any(), fundingsourceID).Return(mxAccount, nil).Times(1)
				unit.EXPECT().GetCustomerByAccountID(gomock.Any(), accountID).Return(
					&_unit.Customer{
						ID:        unitCustomerID,
						AccountID: accountID,
					},
					nil,
				).Times(1)
				key := sha256.Sum256([]byte(fundingsourceID))
				unitCounterPartyID = uuid.NewString()
				unit.EXPECT().CreateCounterParty(gomock.Any(), &_unit.CreateCounterPartyArgs{
					Name:           mxFs.Name,
					RoutingNumber:  mxAccount.RoutingNumber,
					AccountNumber:  mxAccount.AccountNumber,
					AccountType:    mxAccount.Type,
					Type:           "person",
					IdempotencyKey: string(key[0:]),
				}).Return(
					&_unit.CounterParty{
						Type:          "achCounterparty",
						ID:            unitCounterPartyID,
						Attributes:    _unit.CounterPartyAttributes{},
						Relationships: _unit.CounterPartyRelationships{},
					},
					nil,
				).Times(1)
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(st *testing.T) {
			scenario.RunBefore()

			createdCp, err := fs.CreateUnitCounterPartyFromMxAccount(ctx, fundingsourceID)

			if scenario.ExpectedError == nil {
				assert.NoError(st, err, scenario.Name)

				unitCp, err := fs.GetUnitCounterParty(ctx, fundingsourceID)

				assert.NoError(st, err, scenario.Name)
				assert.Equal(st, unitCounterPartyID, unitCp.UnitCounterpartyID, scenario.Name)
			} else {
				assert.ErrorIs(st, err, scenario.ExpectedError)
				assert.Nil(st, createdCp, scenario.Name)
			}
		})
	}
}

func generateCreateBankAccountArgs(opts ...func(*CreateBankAccountArgs)) *CreateBankAccountArgs {
	args := &CreateBankAccountArgs{
		IdentityID:    uuid.NewString(),
		AccountID:     uuid.NewString(),
		Name:          faker.Name(),
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

func withAccountID(id string) func(args *CreateBankAccountArgs) {
	return func(args *CreateBankAccountArgs) {
		args.AccountID = id
	}
}

func withIdentityID(id string) func(args *CreateBankAccountArgs) {
	return func(args *CreateBankAccountArgs) {
		args.IdentityID = id
	}
}

func withName(name string) func(args *CreateBankAccountArgs) {
	return func(args *CreateBankAccountArgs) {
		args.Name = name
	}
}

func withType(_type string) func(args *CreateBankAccountArgs) {
	return func(args *CreateBankAccountArgs) {
		args.Type = _type
	}
}

func withAccountNumber(num string) func(args *CreateBankAccountArgs) {
	return func(args *CreateBankAccountArgs) {
		args.AccountNumber = num
	}
}

func withRoutingNumber(num string) func(args *CreateBankAccountArgs) {
	return func(args *CreateBankAccountArgs) {
		args.RoutingNumber = num
	}
}

func withInstitution(name string) func(args *CreateBankAccountArgs) {
	return func(args *CreateBankAccountArgs) {
		args.Institution = name
	}
}
