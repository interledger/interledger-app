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
	test_utils "gitlab.com/fynbos/backend/utils"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestGetBankAccountWidget(t *testing.T) {
	ctrl := gomock.NewController(t)
	health, err := healthcheck.NewService()
	accountsService := accounts.NewMockService(ctrl)
	identityService := identity.NewMockService(ctrl)
	adminAuthService := auth.NewMockService()
	userService := user.NewMockService()
	fundingsourceService := fundingsources.NewMockService(ctrl)
	if err != nil {
		t.Fatal(err)
	}
	_, _, client := startTestServer(t, &ServerArgs{
		HealthCheckService:   health,
		IdentityService:      identityService,
		AccountsService:      accountsService,
		AdminAuthService:     adminAuthService,
		UserService:          userService,
		UnitProvider:         unit.NewMockService(ctrl),
		FundingSourceService: fundingsourceService,
		OnboardingService:    onboarding.NewMockService(ctrl),
	})

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
				accountsService.EXPECT().GetByIdentityID(gomock.Any(), userID).Return(&accounts.Account{
					ID:         accountID,
					IdentityID: userID,
				}, nil).Times(1)
				fundingsourceService.EXPECT().GetMxConnectWidget(gomock.Any(), accountID, userID).Return(widgetUrl, nil).Times(1)
			},
		},
		{
			Name:          "Returns internal error if account not found.",
			ExpectedError: "rpc error: code = Internal desc = Unable to get account.",
			RunBefore: func() {
				accountsService.EXPECT().GetByIdentityID(gomock.Any(), userID).Return(nil, accounts.ErrNotFound).Times(1)
			},
		},
		{
			Name:          "Returns internal error if the connect widget url cannot be generated.",
			ExpectedError: "rpc error: code = Internal desc = Unable to get widget.",
			RunBefore: func() {
				accountsService.EXPECT().GetByIdentityID(gomock.Any(), userID).Return(&accounts.Account{
					ID:         accountID,
					IdentityID: userID,
				}, nil).Times(1)
				fundingsourceService.EXPECT().GetMxConnectWidget(gomock.Any(), accountID, userID).Return("", fundingsources.ErrInternal).Times(1)
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.RunBefore()

		resp, err := client.GetBankAccountWidget(
			user.ActingAsContext(t, context.Background(), &user.User{
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

func startTestServer(
	t *testing.T,
	args *ServerArgs,
) (*grpc.Server, backendv1.BackendAdminServiceClient, backendv1.BackendServiceClient) {
	server, err := NewServer(args)
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
