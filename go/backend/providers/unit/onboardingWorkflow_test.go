package unit

import (
	context "context"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env               *testsuite.TestWorkflowEnvironment
	onbordingActivity *Activity
}

func (s *UnitTestSuite) SetupSuite() {
	s.onbordingActivity = &Activity{}
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
	applicationArgs := &CreateApplicationArgs{
		Ssn:    faker.Phonenumber(),
		UserID: identityID,
	}
	CustomerID, ApplicationType, AccountID := uuid.NewString(), "individualCustomer", uuid.NewString()

	workflowState := &UnitOnboardCustomerState{
		CustomerID:      "",
		Type:            "",
		IdentityID:      identityID,
		AccountID:       "",
		ApplicationArgs: *applicationArgs,
	}
	s.env.OnActivity(s.onbordingActivity.UnitCreateApplication, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args *CreateApplicationArgs) (*Application, error) {
			s.Equal(workflowState.IdentityID, args.UserID)
			return &Application{
				Type:         ApplicationType,
				ID:           "1234",
				Status:       "Approved",
				FynbosUserId: workflowState.IdentityID,
				Archived:     false,
				CustomerID:   CustomerID,
			}, nil
		},
	)

	s.env.OnActivity(s.onbordingActivity.UnitCreateAccount, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, identityID string) (string, error) {
			s.Equal(workflowState.IdentityID, identityID)
			return AccountID, nil
		},
	)

	s.env.OnActivity(s.onbordingActivity.UnitMapCustomerToAccount, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args *UnitMapCustomerToAccountArgs) error {
			s.Equal(args.Type, ApplicationType)
			s.Equal(args.CustomerID, CustomerID)
			s.Equal(args.AccountID, AccountID)
			return nil
		},
	)
	s.env.ExecuteWorkflow(UnitOnboardCustomerWorkflow, workflowState)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_UnitOnboardCustomerWorkflow_PendingWithApprovedSignal() {
	identityID := uuid.NewString()
	applicationArgs := &CreateApplicationArgs{
		Ssn:    faker.Phonenumber(),
		UserID: identityID,
	}
	CustomerID, CustomerType, AccountID := uuid.NewString(), "individualCustomer", uuid.NewString()

	workflowState := &UnitOnboardCustomerState{
		CustomerID:      "",
		Type:            "",
		IdentityID:      identityID,
		AccountID:       "",
		ApplicationArgs: *applicationArgs,
	}
	s.env.OnActivity(s.onbordingActivity.UnitCreateApplication, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args *CreateApplicationArgs) (*Application, error) {
			s.Equal(workflowState.IdentityID, args.UserID)
			return &Application{
				Type:         CustomerType,
				ID:           "1234",
				Status:       "Pending",
				FynbosUserId: workflowState.IdentityID,
				Archived:     false,
			}, nil
		},
	)

	s.env.RegisterDelayedCallback(func() {
		event := CustomerCreatedEvent{
			ID:   uuid.NewString(),
			Type: "customer.created",
			Attributes: EventAttributes{
				CreatedAt: "2020-07-29T12:53:05.882Z",
				Tags: map[string]string{
					ApplicationUserIDTag: identityID,
				},
			},
			Relationships: EventRelationships{
				Customer: JsonCustomer{
					Data: Data{
						ID:   CustomerID,
						Type: CustomerType,
					},
				},
				Application: JsonApplication{
					Data: Data{
						ID:   "52",
						Type: "individualApplication",
					},
				},
			},
		}
		s.env.SignalWorkflow("onboard-unit-customer-created", event)
	}, time.Millisecond*1)

	s.env.OnActivity(s.onbordingActivity.UnitCreateAccount, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, identityID string) (string, error) {
			s.Equal(workflowState.IdentityID, identityID)
			return AccountID, nil
		},
	)

	s.env.OnActivity(s.onbordingActivity.UnitMapCustomerToAccount, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args *UnitMapCustomerToAccountArgs) error {
			s.Equal(args.Type, CustomerType)
			s.Equal(args.CustomerID, CustomerID)
			s.Equal(args.AccountID, AccountID)
			return nil
		},
	)
	s.env.ExecuteWorkflow(UnitOnboardCustomerWorkflow, workflowState)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_UnitOnboardCustomerWorkflow_ImmediatelyDenied() {
	identityID := uuid.NewString()
	applicationArgs := &CreateApplicationArgs{
		Ssn:    faker.Phonenumber(),
		UserID: identityID,
	}
	ApplicationType := "IndividualApplication"

	workflowState := &UnitOnboardCustomerState{
		CustomerID:      "",
		Type:            "",
		IdentityID:      identityID,
		AccountID:       "",
		ApplicationArgs: *applicationArgs,
	}
	s.env.OnActivity(s.onbordingActivity.UnitCreateApplication, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args *CreateApplicationArgs) (*Application, error) {
			s.Equal(workflowState.IdentityID, args.UserID)
			return &Application{
				Type:         ApplicationType,
				ID:           "1234",
				Status:       "Denied",
				FynbosUserId: workflowState.IdentityID,
				Archived:     false,
			}, nil
		},
	)

	s.env.ExecuteWorkflow(UnitOnboardCustomerWorkflow, workflowState)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_UnitOnboardCustomerWorkflow_PendingWithDeniedSignal() {
	identityID := uuid.NewString()
	applicationArgs := &CreateApplicationArgs{
		Ssn:    faker.Phonenumber(),
		UserID: identityID,
	}
	CustomerType := "individualCustomer"

	workflowState := &UnitOnboardCustomerState{
		CustomerID:      "",
		Type:            "",
		IdentityID:      identityID,
		AccountID:       "",
		ApplicationArgs: *applicationArgs,
	}
	s.env.OnActivity(s.onbordingActivity.UnitCreateApplication, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args *CreateApplicationArgs) (*Application, error) {
			s.Equal(workflowState.IdentityID, args.UserID)
			return &Application{
				Type:         CustomerType,
				ID:           "1234",
				Status:       "Pending",
				FynbosUserId: workflowState.IdentityID,
				Archived:     false,
			}, nil
		},
	)

	s.env.RegisterDelayedCallback(func() {
		event := ApplicationDeniedEvent{
			ID:   uuid.NewString(),
			Type: "application.denied",
			Attributes: EventAttributes{
				CreatedAt: "2020-07-29T12:53:05.882Z",
				Tags: map[string]string{
					ApplicationUserIDTag: identityID,
				},
			},
			Relationships: EventRelationships{
				Application: JsonApplication{
					Data: Data{
						ID:   "52",
						Type: "individualApplication",
					},
				},
			},
		}
		s.env.SignalWorkflow("onboard-unit-application-denied", event)
	}, time.Millisecond*1)

	s.env.ExecuteWorkflow(UnitOnboardCustomerWorkflow, workflowState)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}
