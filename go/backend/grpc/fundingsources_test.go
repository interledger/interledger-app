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
}

type TestContainerOption func(*TestContainer)

func NewTestContainer(t *testing.T, ctrl *gomock.Controller, opts ...TestContainerOption) *TestContainer {
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
				c.FundingsourceService.EXPECT().GetMxConnectWidget(gomock.Any(), accountID, userID).Return(widgetUrl, nil).Times(1)
			},
		},
		{
			Name:          "Returns internal error if account not found.",
			ExpectedError: "rpc error: code = Internal desc = Unable to get account.",
			RunBefore: func() {
				c.AccountService.EXPECT().GetByIdentityID(gomock.Any(), userID).Return(nil, accounts.ErrNotFound).Times(1)
			},
		},
		{
			Name:          "Returns internal error if the connect widget url cannot be generated.",
			ExpectedError: "rpc error: code = Internal desc = Unable to get widget.",
			RunBefore: func() {
				c.AccountService.EXPECT().GetByIdentityID(gomock.Any(), userID).Return(&accounts.Account{
					ID:         accountID,
					IdentityID: userID,
				}, nil).Times(1)
				c.FundingsourceService.EXPECT().GetMxConnectWidget(gomock.Any(), accountID, userID).Return("", fundingsources.ErrInternal).Times(1)
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

func TestCreateMxBankAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	accountID := ""
	user := &user.User{}
	mxUserGuid := ""
	mxMemberGuid := ""
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
				user.ID = uuid.NewString()
				c.AccountService.EXPECT().GetByIdentityID(gomock.Any(), user.ID).Return(
					&accounts.Account{
						ID:         accountID,
						IdentityID: user.ID,
					},
					nil,
				).Times(1)
				c.FundingsourceService.EXPECT().CreateMxBankAccount(gomock.Any(), &fundingsources.CreateMxBankAccountArgs{
					IdentityID:   user.ID,
					AccountID:    accountID,
					MxUserGuid:   mxUserGuid,
					MxMemberGuid: mxMemberGuid,
					Name:         fundingSourceName,
				}).Return(
					&fundingsources.FundingSource{
						ID:        uuid.NewString(),
						AccountID: accountID,
						Name:      fundingSourceName,
					},
					nil,
				).Times(1)
			},
		},
		{
			Name:          "Returns ErrInternal if account not found",
			ExpectedError: "rpc error: code = Internal desc = Unable to get account.",
			RunBefore: func() {
				accountID = uuid.NewString()
				user.ID = uuid.NewString()
				c.AccountService.EXPECT().GetByIdentityID(gomock.Any(), user.ID).Return(
					nil,
					accounts.ErrNotFound,
				).Times(1)
				c.FundingsourceService.EXPECT().CreateMxBankAccount(gomock.Any(), gomock.Any()).Times(0)
			},
		},
		{
			Name:          "Returns ErrInternal if cannot create mx bank account",
			ExpectedError: "rpc error: code = Internal desc = Unable to create bank account.",
			RunBefore: func() {
				accountID = uuid.NewString()
				user.ID = uuid.NewString()
				c.AccountService.EXPECT().GetByIdentityID(gomock.Any(), user.ID).Return(
					&accounts.Account{
						ID:         accountID,
						IdentityID: user.ID,
					},
					nil,
				).Times(1)
				c.FundingsourceService.EXPECT().CreateMxBankAccount(gomock.Any(), gomock.Any()).Return(
					nil,
					fundingsources.ErrInternal,
				).Times(1)
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(st *testing.T) {
			scenario.RunBefore()

			fs, err := client.CreateBankAccount(
				_user.ActingAsContext(t, context.Background(), user),
				&backendv1.CreateBankAccountRequest{
					UserGuid:   mxUserGuid,
					MemberGuid: mxMemberGuid,
					Name:       fundingSourceName,
				},
			)

			if scenario.ExpectedError == "" {
				assert.NoError(t, err, scenario.Name)
				assert.Equal(t, fundingSourceName, fundingSourceName, scenario.Name)
			} else {
				assert.Equal(t, scenario.ExpectedError, err.Error(), scenario.Name)
				assert.Nil(t, fs, scenario.Name)
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
