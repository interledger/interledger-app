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
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/healthcheck"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/rafiki"
	"gitlab.com/fynbos/backend/providers/unit"
	"gitlab.com/fynbos/backend/user"
	_user "gitlab.com/fynbos/backend/user"
	test_utils "gitlab.com/fynbos/backend/utils"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TestContainer struct {
	HealthService        healthcheck.Service
	AccountService       *accounts.MockService
	IdentityService      *identity.MockService
	AdminAuthService     auth.Service
	UserService          user.Service
	FundingsourceService *fundingsources.MockService
	OnboardingService    *onboarding.MockService
	UnitProvider         *unit.MockService
	MxProvider           *mx.MockService
	RafikiProvider       *rafiki.MockService
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
		AccountService:       accounts.NewMockService(ctrl),
		IdentityService:      identity.NewMockService(ctrl),
		AdminAuthService:     auth.NewMockService(),
		UserService:          user.NewMockService(),
		FundingsourceService: fundingsources.NewMockService(ctrl),
		UnitProvider:         unit.NewMockService(ctrl),
		OnboardingService:    onboarding.NewMockService(ctrl),
		MxProvider:           mx.NewMockService(ctrl),
		RafikiProvider:       rafiki.NewMockService(ctrl),
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

func TestInitiateCreateBankAccount(t *testing.T) {
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

			response, err := client.InitiateCreateBankAccount(
				_user.ActingAsContext(t, context.Background(), &user.User{ID: userID}),
				&backendv1.InitiateCreateBankAccountRequest{
					UserGuid:   mxUserGuid,
					MemberGuid: mxMemberGuid,
					Name:       fundingSourceName,
				},
			)

			if scenario.ExpectedError == "" {
				assert.NoError(t, err, scenario.Name)
				assert.Equal(t, workflowUuid, response.Reference, scenario.Name)
			} else {
				assert.Equal(t, scenario.ExpectedError, err.Error(), scenario.Name)
				assert.Nil(t, response, scenario.Name)
			}
		})
	}
}

func startTestServer(
	t *testing.T,
	c *TestContainer,
) (*grpc.Server, backendv1.BackendAdminServiceClient, backendv1.BackendServiceClient) {
	server, err := NewServer(&ServerArgs{
		HealthCheckService:   c.HealthService,
		IdentityService:      c.IdentityService,
		AccountsService:      c.AccountService,
		AdminAuthService:     c.AdminAuthService,
		UserService:          c.UserService,
		UnitProvider:         c.UnitProvider,
		FundingSourceService: c.FundingsourceService,
		OnboardingService:    c.OnboardingService,
		MxProvider:           c.MxProvider,
		RafikiProvider:       c.RafikiProvider,
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
