package workflows_test

import (
	context "context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gitlab.com/fynbos/backend/providers/unit/activities"
	"gitlab.com/fynbos/backend/providers/unit/external"
	"gitlab.com/fynbos/backend/providers/unit/workflows"
	"go.temporal.io/sdk/testsuite"
)

type UnitEventsTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env            *testsuite.TestWorkflowEnvironment
	eventsActivity *activities.Activity
}

func (s *UnitEventsTestSuite) SetupSuite() {
	s.eventsActivity = &activities.Activity{}
}

func (s *UnitEventsTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	s.env.RegisterActivity(s.eventsActivity)
}

func (s *UnitEventsTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *UnitEventsTestSuite) TestHandlesAllEvents() {
	var rawEvents []json.RawMessage
	applicationDeniedEvent := NewApplicationDeniedEvent()
	customerCreatedEvent := NewCustomerCreatedEvent()
	rawEvents = append(rawEvents, marshalEvent(s.T(), applicationDeniedEvent))
	rawEvents = append(rawEvents, marshalEvent(s.T(), customerCreatedEvent))
	s.env.OnActivity(s.eventsActivity.UnitStoreEvents, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args []json.RawMessage) error {
			s.Len(args, 2)
			return nil
		},
	)

	s.env.OnActivity(s.eventsActivity.UnitNotifyApplicationDenied, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, event external.ApplicationDeniedEvent) error {
			s.Equal(applicationDeniedEvent.ID, event.ID)
			return nil
		},
	)

	s.env.OnActivity(s.eventsActivity.UnitNotifyCustomerCreated, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, event external.CustomerCreatedEvent) error {
			s.Equal(customerCreatedEvent.ID, event.ID)
			return nil
		},
	)
	s.env.ExecuteWorkflow(workflows.UnitHandleEventsWorkflow, rawEvents)
	s.True(s.env.IsWorkflowCompleted())
}

func (s *UnitEventsTestSuite) TestProceedsIfOneEventFails() {
	var rawEvents []json.RawMessage
	applicationDeniedEvent := NewApplicationDeniedEvent()
	customerCreatedEvent := NewCustomerCreatedEvent()
	rawEvents = append(rawEvents, marshalEvent(s.T(), applicationDeniedEvent))
	rawEvents = append(rawEvents, marshalEvent(s.T(), customerCreatedEvent))
	s.env.OnActivity(s.eventsActivity.UnitStoreEvents, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args []json.RawMessage) error {
			s.Len(args, 2)
			return nil
		},
	)

	s.env.OnActivity(s.eventsActivity.UnitNotifyApplicationDenied, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, event external.ApplicationDeniedEvent) error {
			s.Equal(applicationDeniedEvent.ID, event.ID)
			return errors.New("test failure")
		},
	)

	s.env.OnActivity(s.eventsActivity.UnitNotifyCustomerCreated, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, event external.CustomerCreatedEvent) error {
			s.Equal(customerCreatedEvent.ID, event.ID)
			return nil
		},
	)

	s.env.ExecuteWorkflow(workflows.UnitHandleEventsWorkflow, rawEvents)
	s.True(s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	s.Require().Error(err)
	s.Contains(err.Error(), "Failed to handle all unit events. (1/2 handled)")
}

func TestUnitEventsTestSuite(t *testing.T) {
	suite.Run(t, new(UnitEventsTestSuite))
}

func marshalEvent(t *testing.T, event interface{}) json.RawMessage {
	raw, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}

	return raw
}

func NewCustomerCreatedEvent() external.CustomerCreatedEvent {
	return external.CustomerCreatedEvent{
		ID:   uuid.NewString(),
		Type: "customer.created",
		Attributes: external.EventAttributes{
			CreatedAt: "2020-07-29T12:53:05.882Z",
			Tags: external.ApplicationTags{
				FynbosUserID: uuid.NewString(),
			},
		},
		Relationships: external.EventRelationships{
			Customer: external.JsonCustomer{
				Data: external.Data{
					ID:   "52",
					Type: "individualCustomer",
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
}

func NewApplicationDeniedEvent() external.ApplicationDeniedEvent {
	return external.ApplicationDeniedEvent{
		ID:   uuid.NewString(),
		Type: "application.denied",
		Attributes: external.EventAttributes{
			CreatedAt: "2020-07-29T12:53:05.882Z",
			Tags: external.ApplicationTags{
				FynbosUserID: uuid.NewString(),
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
}
