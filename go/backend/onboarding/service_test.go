package onboarding

import (
	context "context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	accounts "gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/noop"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"
)

func TestInitiatesOnboarding(t *testing.T) {
	ctrl := gomock.NewController(t)
	tp := &mocks.Client{}
	os, err := NewService(&ServiceArgs{
		Db:   &sqlx.DB{},
		As:   accounts.NewMockService(ctrl),
		Is:   identity.NewMockService(ctrl),
		Noop: noop.NewMockService(ctrl),
		Tp:   tp,
	})
	if err != nil {
		t.Fatal(err)
	}

	args := &InitiateUnitCustomerOnboardingArgs{
		IdentityID:   uuid.NewString(),
		Country:      "US",
		CustomerID:   uuid.NewString(),
		CustomerType: "individual",
	}
	tp.On(
		"ExecuteWorkflow",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		&OnboardUnitCustomerArgs{
			CustomerID: args.CustomerID,
			Type:       args.CustomerType,
			IdentityID: args.IdentityID,
			Country:    args.Country,
		},
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

	err = os.InitiateUnitCustomerOnboarding(context.Background(), args)

	assert.NoError(t, err)
}
