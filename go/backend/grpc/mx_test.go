package grpc

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/accounts"
	accounts_mock "gitlab.com/fynbos/backend/accounts/client/mock"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/agreements"
	"gitlab.com/fynbos/backend/deposits"
	"gitlab.com/fynbos/backend/fundingsources"
	funding_mock "gitlab.com/fynbos/backend/fundingsources/client/mock"
	"gitlab.com/fynbos/backend/healthcheck"
	identity_mock "gitlab.com/fynbos/backend/identity/client/mock"
	onboarding_mock "gitlab.com/fynbos/backend/onboarding/client/mock"
	payments_mock "gitlab.com/fynbos/backend/payments/client/mock"
	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/rafiki"
	"gitlab.com/fynbos/backend/providers/unit"
	support_mock "gitlab.com/fynbos/backend/supporttickets/client/mock"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/user"
	_user "gitlab.com/fynbos/backend/user"
	test_utils "gitlab.com/fynbos/backend/utils"
	waitlist_mock "gitlab.com/fynbos/backend/waitlist/client/mock"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
	"go.temporal.io/sdk/mocks"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TestContainer struct {
	HealthService        healthcheck.Service
	AccountService       *accounts_mock.MockClient
	AgreementsService    *agreements.MockService
	IdentityService      *identity_mock.MockClient
	AdminAuthService     auth.Service
	UserService          user.Service
	FundingsourceService *funding_mock.MockClient
	TwilioService        *twilio.MockService
	OnboardingService    *onboarding_mock.MockClient
	UnitProvider         *unit.MockService
	MxProvider           *mx.MockService
	RafikiProvider       *rafiki.MockService
	DepositService       *deposits.MockService
	WaitlistClient       *waitlist_mock.MockClient
	Temporal             *mocks.Client
	TicketClient         *support_mock.MockClient
	PaymentsClient       *payments_mock.MockClient
}

type TestContainerOption func(*TestContainer)

func NewTestContainer(t *testing.T, ctrl *gomock.Controller, opts ...TestContainerOption) *TestContainer {
	t.Cleanup(func() {
		ctrl.Finish()
	})
	hs, err := healthcheck.NewService()
	if err != nil {
		t.Fatal(err)
	}
	c := &TestContainer{
		HealthService:        hs,
		AccountService:       accounts_mock.NewMockClient(ctrl),
		AgreementsService:    agreements.NewMockService(ctrl),
		IdentityService:      identity_mock.NewMockClient(ctrl),
		AdminAuthService:     auth.NewMockService(),
		UserService:          user.NewMockService(),
		FundingsourceService: funding_mock.NewMockClient(ctrl),
		TwilioService:        twilio.NewMockService(ctrl),
		UnitProvider:         unit.NewMockService(ctrl),
		OnboardingService:    onboarding_mock.NewMockClient(ctrl),
		MxProvider:           mx.NewMockService(ctrl),
		RafikiProvider:       rafiki.NewMockService(ctrl),
		DepositService:       deposits.NewMockService(ctrl),
		WaitlistClient:       waitlist_mock.NewMockClient(ctrl),
		TicketClient:         support_mock.NewMockClient(ctrl),
		Temporal:             &mocks.Client{},
		PaymentsClient:       payments_mock.NewMockClient(ctrl),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func TestGetBankAccountWidget(t *testing.T) {
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	userID := uuid.NewString()
	accountID := uuid.NewString()
	widgetUrl := "test"
	scenarios := []struct {
		Name          string
		ExpectedError string
		RunBefore     func()
	}{
		{
			Name:          "Returns the mx connect widget url.",
			ExpectedError: "",
			RunBefore: func() {
				c.AccountService.EXPECT().GetByIdentityID(gomock.Any(), userID).Return(&accounts.Account{
					ID:         accountID,
					IdentityID: userID,
				}, nil).Times(1)
				c.MxProvider.EXPECT().GetConnectWidget(gomock.Any(), accountID, userID).Return(widgetUrl, nil).Times(1)
			},
		},
		{
			Name:          "Returns internal error if account not found.",
			ExpectedError: "rpc error: code = Internal desc = Internal server error: Unable to get account.",
			RunBefore: func() {
				c.AccountService.EXPECT().GetByIdentityID(gomock.Any(), userID).Return(nil, accounts.ErrNotFound).Times(1)
			},
		},
		{
			Name:          "Returns internal error if the connect widget url cannot be generated.",
			ExpectedError: "rpc error: code = Internal desc = Internal server error: Unable to get widget.",
			RunBefore: func() {
				c.AccountService.EXPECT().GetByIdentityID(gomock.Any(), userID).Return(&accounts.Account{
					ID:         accountID,
					IdentityID: userID,
				}, nil).Times(1)
				c.MxProvider.EXPECT().GetConnectWidget(gomock.Any(), accountID, userID).Return("", fundingsources.ErrInternal).Times(1)
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.RunBefore()

		resp, err := client.GetBankAccountWidget(
			_user.ActingAsContext(t, context.Background(), &_user.User{
				ID:    userID,
				Email: faker.Email(),
			}),
			&backendv1.GetBankAccountWidgetRequest{},
		)

		if scenario.ExpectedError == "" {
			assert.NoError(t, err)
			assert.Equal(t, widgetUrl, resp.GetUrl())
		} else {
			assert.Error(t, err)
			assert.Equal(t, scenario.ExpectedError, err.Error())
		}
	}
}

func TestAddBankAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	accountID := ""
	userID := ""
	mxUserGuid := ""
	mxMemberGuid := ""
	workflowUuid := ""
	fundingSourceName := "test"
	scenarios := []struct {
		Name          string
		ExpectedError string
		RunBefore     func()
	}{
		{
			Name:          "Creates mx bank account.",
			ExpectedError: "",
			RunBefore: func() {
				mxMemberGuid = uuid.NewString()
				mxUserGuid = uuid.NewString()
				accountID = uuid.NewString()
				userID = uuid.NewString()
				c.AccountService.EXPECT().GetByIdentityID(gomock.Any(), userID).Return(
					&accounts.Account{
						ID:         accountID,
						IdentityID: userID,
					},
					nil,
				).Times(1)
				workflowUuid = uuid.NewString()
				c.MxProvider.EXPECT().InitiateCreateAccount(gomock.Any(), &mx.InitiateCreateAccountArgs{
					IdentityID:        userID,
					AccountID:         accountID,
					UserGuid:          mxUserGuid,
					MemberGuid:        mxMemberGuid,
					FundingsourceName: fundingSourceName,
				}).Return(workflowUuid, nil).Times(1)
			},
		},
		{
			Name:          "Returns ErrInternal if account not found",
			ExpectedError: "rpc error: code = Internal desc = Internal server error: Unable to get account.",
			RunBefore: func() {
				accountID = uuid.NewString()
				userID = uuid.NewString()
				c.AccountService.EXPECT().GetByIdentityID(gomock.Any(), userID).Return(
					nil,
					accounts.ErrNotFound,
				).Times(1)
				c.MxProvider.EXPECT().InitiateCreateAccount(gomock.Any(), gomock.Any()).Times(0)
			},
		},
		{
			Name:          "Returns ErrInternal if cannot create mx bank account",
			ExpectedError: "rpc error: code = Internal desc = Internal server error: Unable to create bank account.",
			RunBefore: func() {
				accountID = uuid.NewString()
				userID = uuid.NewString()
				c.AccountService.EXPECT().GetByIdentityID(gomock.Any(), userID).Return(
					&accounts.Account{
						ID:         accountID,
						IdentityID: userID,
					},
					nil,
				).Times(1)
				c.MxProvider.EXPECT().InitiateCreateAccount(gomock.Any(), gomock.Any()).Return(
					"",
					mx.ErrInternal,
				).Times(1)
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(st *testing.T) {
			scenario.RunBefore()

			response, err := client.AddBankAccount(
				_user.ActingAsContext(t, context.Background(), &user.User{ID: userID}),
				&backendv1.AddBankAccountRequest{
					UserGuid:   mxUserGuid,
					MemberGuid: mxMemberGuid,
					Name:       fundingSourceName,
				},
			)

			if scenario.ExpectedError == "" {
				assert.NoError(t, err, scenario.Name)
				assert.Equal(t, workflowUuid, response.FundingsourceId, scenario.Name)
			} else {
				assert.Equal(t, scenario.ExpectedError, err.Error(), scenario.Name)
				assert.Nil(t, response, scenario.Name)
			}
		})
	}
}

func TestGetBankDetails(t *testing.T) {
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	t.Run("can get bank details", func(st *testing.T) {
		accountID := uuid.NewString()
		userID := uuid.NewString()
		fundingsourceID := uuid.NewString()
		mxAccountGuid := "acc_" + uuid.NewString()

		c.AccountService.EXPECT().GetByIdentityID(gomock.Any(), userID).Return(
			&accounts.Account{
				ID:         accountID,
				IdentityID: userID,
			},
			nil,
		)
		c.MxProvider.EXPECT().GetAccountByFundingsource(gomock.Any(), fundingsourceID).Return(
			&mx.Account{
				Guid:            mxAccountGuid,
				AccountID:       accountID,
				FundingsourceID: fundingsourceID,
			},
			nil,
		)
		c.MxProvider.EXPECT().ReadAccount(gomock.Any(), mxAccountGuid).Return(
			&mx.AccountDetails{
				Guid:              mxAccountGuid,
				AccountNumber:     "123456789",
				InstitutionNumber: "321",
				Type:              "Checking",
			},
			nil,
		)

		resp, err := client.GetBankAccountDetails(
			_user.ActingAsContext(st, context.Background(), &user.User{ID: userID}),
			&backendv1.GetBankAccountDetailsRequest{
				FundingsourceId: fundingsourceID,
			})
		if err != nil {
			st.Fatal(err)
		}

		assert.Equal(st, "6789", resp.GetMask())
		assert.Equal(st, "Checking", resp.GetType())
		assert.Equal(st, fundingsourceID, resp.GetFundingsourceId())
		assert.Equal(st, "321", resp.GetInstitution())
	})

	t.Run("user can only get their own account info", func(st *testing.T) {
		accountID := uuid.NewString()
		userID := uuid.NewString()
		fundingsourceID := uuid.NewString()
		mxAccountGuid := "acc_" + uuid.NewString()

		c.AccountService.EXPECT().GetByIdentityID(gomock.Any(), userID).Return(
			&accounts.Account{
				ID:         accountID,
				IdentityID: userID,
			},
			nil,
		)
		c.MxProvider.EXPECT().GetAccountByFundingsource(gomock.Any(), fundingsourceID).Return(
			&mx.Account{
				Guid:            mxAccountGuid,
				AccountID:       uuid.NewString(),
				FundingsourceID: fundingsourceID,
			},
			nil,
		)

		resp, err := client.GetBankAccountDetails(
			_user.ActingAsContext(st, context.Background(), &user.User{ID: userID}),
			&backendv1.GetBankAccountDetailsRequest{
				FundingsourceId: fundingsourceID,
			})
		if err == nil {
			st.Fatal("User must only be able to get own account details")
		}

		assert.Nil(st, resp)
		assert.Equal(st, "rpc error: code = PermissionDenied desc = Forbidden: Unauthorized.", err.Error())
	})
}

func startTestServer(
	t *testing.T,
	c *TestContainer,
) (*grpc.Server, backendv1.BackendAdminServiceClient, backendv1.BackendServiceClient) {
	server, err := NewServer(&ServerArgs{
		HealthCheckService:   c.HealthService,
		IdentityService:      c.IdentityService,
		AccountsService:      c.AccountService,
		AgreementsService:    c.AgreementsService,
		AdminAuthService:     c.AdminAuthService,
		UserService:          c.UserService,
		UnitProvider:         c.UnitProvider,
		FundingSourceService: c.FundingsourceService,
		TwilioService:        c.TwilioService,
		OnboardingService:    c.OnboardingService,
		MxProvider:           c.MxProvider,
		RafikiProvider:       c.RafikiProvider,
		DepositService:       c.DepositService,
		WaitlistClient:       c.WaitlistClient,
		Temporal:             c.Temporal,
		TicketClient:         c.TicketClient,
		PaymentsClient:       c.PaymentsClient,
	})
	if err != nil {
		t.Fatal(err)
	}

	port, err := test_utils.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}

	serverUrl := fmt.Sprintf("0.0.0.0:%d", port)
	listener, err := net.Listen("tcp", serverUrl)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		if err := server.Serve(listener); err != nil {
			panic(fmt.Errorf("Failed to start test grpc server. %s", err))
		}
	}()

	conn, err := grpc.Dial(
		serverUrl, grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	adminClient := backendv1.NewBackendAdminServiceClient(conn)
	backendClient := backendv1.NewBackendServiceClient(conn)

	t.Cleanup(func() {
		if err := conn.Close(); err != nil {
			t.Fatal(err)
		}
		server.Stop()
	})

	return server, adminClient, backendClient
}
