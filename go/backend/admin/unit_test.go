package admin_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/fundingsources"
	_grpc "gitlab.com/fynbos/backend/grpc"
	"gitlab.com/fynbos/backend/healthcheck"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/rafiki"
	"gitlab.com/fynbos/backend/providers/unit"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/proto/backend/v1"
)

func TestGetUnitCustomerByAccountID(t *testing.T) {
	// note https://github.com/golang/mock/issues/139
	// we run the grpc server in a goroutine so any errors from the mock will be swallowed.
	// TODO: see if there is a test server for grpc
	ctrl := gomock.NewController(t)
	up := unit.NewMockService(ctrl)
	is := identity.NewMockService(ctrl)
	as := accounts.NewMockService(ctrl)
	us := auth.NewMockService()
	fs := fundingsources.NewMockService(ctrl)
	hs, err := healthcheck.NewService()
	onboardingService := onboarding.NewMockService(ctrl)
	if err != nil {
		t.Fatal(err)
	}
	svr, err := _grpc.NewServer(&_grpc.ServerArgs{
		HealthCheckService:   hs,
		IdentityService:      is,
		AccountsService:      as,
		AdminAuthService:     us,
		UnitProvider:         up,
		UserService:          user.NewMockService(),
		FundingSourceService: fs,
		OnboardingService:    onboardingService,
		MxProvider:           mx.NewMockService(ctrl),
		RafikiProvider:       rafiki.NewMockService(ctrl),
	})
	if err != nil {
		t.Fatal(err)
	}
	client, cleanup := NewAdminClient(t, svr)
	t.Cleanup(func() {
		cleanup()
		ctrl.Finish()
	})

	ctx := context.Background()
	accountID := uuid.NewString()
	customerID := uuid.NewString()
	customerType := "individual"
	up.EXPECT().GetCustomerByAccountID(gomock.Any(), accountID).Return(
		&unit.Customer{
			ID:        customerID,
			AccountID: accountID,
			Type:      customerType,
		},
		nil,
	).Times(1)

	resp, err := client.GetUnitCustomerByAccountID(
		auth.ActingAs(ctx, "don@fynbos.dev"),
		&backend.GetUnitCustomerByAccountRequest{
			AccountId: accountID,
		})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, accountID, resp.GetAccountId())
	assert.Equal(t, customerID, resp.GetId())
	assert.Equal(t, customerType, resp.GetType())
}
