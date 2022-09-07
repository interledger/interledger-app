package workflows_test

import (
	context "context"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gitlab.com/fynbos/backend/providers/unit"
	"gitlab.com/fynbos/backend/providers/unit/activities"
	"gitlab.com/fynbos/backend/providers/unit/external"
	"gitlab.com/fynbos/backend/providers/unit/workflows"
	"go.temporal.io/sdk/testsuite"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env               *testsuite.TestWorkflowEnvironment
	onbordingActivity *activities.Activity
}

func (s *UnitTestSuite) SetupSuite() {
	s.onbordingActivity = &activities.Activity{}
}

func (s *UnitTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	s.env.RegisterActivity(s.onbordingActivity)
}

func (s *UnitTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

type TestApplication struct {
	Data struct {
		Type       string
		ID         string
		Attributes struct {
			Status string
			Tags   struct {
				FynbosUserId string
			}
			Archived bool
		}
	}
}

func (s *UnitTestSuite) Test_UnitOnboardCustomerWorkflow_ImmediatelyApproved() {
	identityID := uuid.NewString()
	applicationArgs := &unit.CreateApplicationArgs{
		Ssn:    faker.Phonenumber(),
		UserID: identityID,
	}
	CustomerID, ApplicationType, AccountID := uuid.NewString(), "individualCustomer", uuid.NewString()
	DepositAccountID := uuid.NewString()
	workflowState := workflows.UnitOnboardCustomerState{
		CustomerID:      "",
		Type:            "",
		IdentityID:      identityID,
		AccountID:       "",
		ApplicationArgs: *applicationArgs,
	}
	s.env.OnActivity(s.onbordingActivity.UnitCreateApplication, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args *unit.CreateApplicationArgs) (*unit.Application, error) {
			s.Equal(workflowState.IdentityID, args.UserID)
			return &unit.Application{
				Type:         ApplicationType,
				ID:           "1234",
				Status:       "Approved",
				FynbosUserId: workflowState.IdentityID,
				Archived:     false,
				CustomerID:   CustomerID,
			}, nil
		},
	)
	s.env.OnActivity(s.onbordingActivity.UnitCreateCustomer, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args *activities.UnitCreateCustomerArgs) error {
			s.Equal(args.Type, ApplicationType)
			s.Equal(args.CustomerID, CustomerID)
			s.Equal(args.IdentityID, identityID)
			return nil
		},
	)
	s.env.OnActivity(s.onbordingActivity.UnitCreateDepositAccount, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, customerID string) (*unit.DepositAccount, error) {
			s.Equal(customerID, CustomerID)
			return &unit.DepositAccount{
				ID:         DepositAccountID,
				CustomerID: CustomerID,
			}, nil
		},
	)
	s.env.OnActivity(s.onbordingActivity.UnitCreateAccount, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args *activities.UnitCreateAccountArgs) (string, error) {
			s.Equal(workflowState.IdentityID, identityID)
			return AccountID, nil
		},
	)
	s.env.ExecuteWorkflow(workflows.UnitOnboardCustomerWorkflow, workflowState)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_UnitOnboardCustomerWorkflow_PendingWithApprovedSignal() {
	identityID := uuid.NewString()
	applicationArgs := &unit.CreateApplicationArgs{
		Ssn:    faker.Phonenumber(),
		UserID: identityID,
	}
	CustomerID, CustomerType, AccountID := uuid.NewString(), "individualCustomer", uuid.NewString()
	DepositAccountID := uuid.NewString()
	workflowState := workflows.UnitOnboardCustomerState{
		CustomerID:      "",
		Type:            "",
		IdentityID:      identityID,
		AccountID:       "",
		ApplicationArgs: *applicationArgs,
	}
	s.env.OnActivity(s.onbordingActivity.UnitCreateApplication, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args *unit.CreateApplicationArgs) (*unit.Application, error) {
			s.Equal(workflowState.IdentityID, args.UserID)
			return &unit.Application{
				Type:         CustomerType,
				ID:           "1234",
				Status:       "Pending",
				FynbosUserId: workflowState.IdentityID,
				Archived:     false,
			}, nil
		},
	)

	s.env.RegisterDelayedCallback(func() {
		event := external.CustomerCreatedEvent{
			ID:   uuid.NewString(),
			Type: "customer.created",
			Attributes: external.EventAttributes{
				CreatedAt: "2020-07-29T12:53:05.882Z",
				Tags: external.ApplicationTags{
					FynbosUserID: identityID,
				},
			},
			Relationships: external.EventRelationships{
				Customer: external.JsonCustomer{
					Data: external.Data{
						ID:   CustomerID,
						Type: CustomerType,
					},
				},
				Application: external.JsonApplication{
					Data: external.Data{
						ID:   "52",
						Type: "individualApplication",
					},
				},
			},
		}
		s.env.SignalWorkflow("onboard-unit-customer-created", event)
	}, time.Millisecond*1)

	s.env.OnActivity(s.onbordingActivity.UnitCreateCustomer, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args *activities.UnitCreateCustomerArgs) error {
			s.Equal(args.Type, CustomerType)
			s.Equal(args.CustomerID, CustomerID)
			s.Equal(args.IdentityID, identityID)
			return nil
		},
	)
	s.env.OnActivity(s.onbordingActivity.UnitCreateDepositAccount, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, customerID string) (*unit.DepositAccount, error) {
			s.Equal(customerID, CustomerID)
			return &unit.DepositAccount{
				ID:         DepositAccountID,
				CustomerID: CustomerID,
			}, nil
		},
	)
	s.env.OnActivity(s.onbordingActivity.UnitCreateAccount, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args *activities.UnitCreateAccountArgs) (string, error) {
			s.Equal(workflowState.IdentityID, identityID)
			return AccountID, nil
		},
	)
	s.env.ExecuteWorkflow(workflows.UnitOnboardCustomerWorkflow, workflowState)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_UnitOnboardCustomerWorkflow_ImmediatelyDenied() {
	identityID := uuid.NewString()
	applicationArgs := &unit.CreateApplicationArgs{
		Ssn:    faker.Phonenumber(),
		UserID: identityID,
	}
	ApplicationType := "IndividualApplication"

	workflowState := workflows.UnitOnboardCustomerState{
		CustomerID:      "",
		Type:            "",
		IdentityID:      identityID,
		AccountID:       "",
		ApplicationArgs: *applicationArgs,
	}
	s.env.OnActivity(s.onbordingActivity.UnitCreateApplication, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args *unit.CreateApplicationArgs) (*unit.Application, error) {
			s.Equal(workflowState.IdentityID, args.UserID)
			return &unit.Application{
				Type:         ApplicationType,
				ID:           "1234",
				Status:       "Denied",
				FynbosUserId: workflowState.IdentityID,
				Archived:     false,
			}, nil
		},
	)

	s.env.ExecuteWorkflow(workflows.UnitOnboardCustomerWorkflow, workflowState)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_UnitOnboardCustomerWorkflow_PendingWithDeniedSignal() {
	identityID := uuid.NewString()
	applicationArgs := &unit.CreateApplicationArgs{
		Ssn:    faker.Phonenumber(),
		UserID: identityID,
	}
	CustomerType := "individualCustomer"

	workflowState := workflows.UnitOnboardCustomerState{
		CustomerID:      "",
		Type:            "",
		IdentityID:      identityID,
		AccountID:       "",
		ApplicationArgs: *applicationArgs,
	}
	s.env.OnActivity(s.onbordingActivity.UnitCreateApplication, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args *unit.CreateApplicationArgs) (*unit.Application, error) {
			s.Equal(workflowState.IdentityID, args.UserID)
			return &unit.Application{
				Type:         CustomerType,
				ID:           "1234",
				Status:       "Pending",
				FynbosUserId: workflowState.IdentityID,
				Archived:     false,
			}, nil
		},
	)

	s.env.RegisterDelayedCallback(func() {
		event := external.ApplicationDeniedEvent{
			ID:   uuid.NewString(),
			Type: "application.denied",
			Attributes: external.EventAttributes{
				CreatedAt: "2020-07-29T12:53:05.882Z",
				Tags: external.ApplicationTags{
					FynbosUserID: identityID,
				},
			},
			Relationships: external.EventRelationships{
				Application: external.JsonApplication{
					Data: external.Data{
						ID:   "52",
						Type: "individualApplication",
					},
				},
			},
		}
		s.env.SignalWorkflow("onboard-unit-application-denied", event)
	}, time.Millisecond*1)

	s.env.ExecuteWorkflow(workflows.UnitOnboardCustomerWorkflow, workflowState)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}
