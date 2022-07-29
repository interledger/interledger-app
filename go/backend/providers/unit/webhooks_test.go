package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	http "net/http"
	"net/http/httptest"
	"testing"

	identity_mock "gitlab.com/fynbos/backend/identity/client/mock"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	accounts_mock "gitlab.com/fynbos/backend/accounts/client/mock"
	"gitlab.com/fynbos/backend/providers/unit/external"
	test_utils "gitlab.com/fynbos/backend/utils"
	"go.temporal.io/sdk/mocks"
	"go.uber.org/zap"
)

func TestWebhook(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db := test_utils.MigrateCockroachDB(t, ctx)
	ctrl := gomock.NewController(t)
	providerMock := NewMockService(ctrl)
	temporalMockClient := &mocks.Client{}

	wh, err := NewWebhook(&WebhookArgs{
		Db: db,
		Up: providerMock,
		Tp: temporalMockClient,
	})
	if err != nil {
		t.Fatal(err)
	}

	svr := httptest.NewServer(wh.MakeHttpHandler())

	t.Cleanup(func() {
		svr.Close()
	})

	scenarios := []struct {
		Name                   string
		VerifyError            error
		Payload                *bytes.Buffer
		ExpectedHttpStatusCode int
		ResponseMessage        string
		MockCallTimes          int
	}{
		{
			Name:                   "Returns 200",
			VerifyError:            nil,
			Payload:                marshalBody(t, NewCustomerCreatedEvent(), NewCustomerCreatedEvent()),
			ExpectedHttpStatusCode: 200,
			MockCallTimes:          2,
		},
		{
			Name:                   "Returns 500 if webhook fails verification",
			VerifyError:            errors.New("test"),
			Payload:                marshalBody(t, NewCustomerCreatedEvent(), NewCustomerCreatedEvent()),
			ExpectedHttpStatusCode: 401,
			ResponseMessage:        "Signature didn't match.\n",
			MockCallTimes:          0,
		},
		{
			Name:                   "Returns 500 if marshalling payload fails",
			VerifyError:            nil,
			Payload:                bytes.NewBuffer([]byte("")),
			ExpectedHttpStatusCode: 500,
			ResponseMessage:        "Failed to parse payload\n",
			MockCallTimes:          0,
		},
		{
			Name:                   "Tries to handle all events even if first one fails",
			VerifyError:            nil,
			Payload:                marshalBody(t, "", NewCustomerCreatedEvent()),
			ExpectedHttpStatusCode: 500,
			ResponseMessage:        "Failed to parse payload\n",
			MockCallTimes:          1,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			temporalMockClient.On("SignalWorkflow", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), "onboard-unit-customer-created", mock.Anything).Return(nil).Times(scenario.MockCallTimes)
			providerMock.EXPECT().VerifyWebhook(context.Background(), gomock.Any(), gomock.Any()).Return(scenario.VerifyError)
			resp, err := http.Post(svr.URL, "application/json", scenario.Payload)
			if err != nil {
				t.Fatal(err)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, scenario.ExpectedHttpStatusCode, resp.StatusCode)
			assert.Equal(t, scenario.ResponseMessage, string(body))
		})
	}
}

func TestHandleCreatedCustomerEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	identityMock := identity_mock.NewMockClient(ctrl)
	temporalMockClient := &mocks.Client{}

	db := test_utils.MigrateCockroachDB(t, ctx)

	provider, err := NewService(ServiceArgs{
		BaseURL:         "localhost:8080",
		Token:           "token",
		WebhookToken:    "webhooktoken",
		Db:              db,
		IdentityService: identityMock,
		AccountClient:   accounts_mock.NewMockClient(ctrl),
		Logger:          zap.NewNop(),
	})
	if err != nil {
		t.Fatal(err)
	}

	wh, err := NewWebhook(&WebhookArgs{
		Up: provider,
		Db: db,
		Tp: temporalMockClient,
	})
	if err != nil {
		t.Fatal(err)
	}

	scenarios := []struct {
		Name            string
		OnboardingError error
	}{
		{
			Name:            "Succeeds if unit onboarding is initiated.",
			OnboardingError: nil,
		},
	}
	for _, scenario := range scenarios {
		customerCreatedEvent := NewCustomerCreatedEvent()

		temporalMockClient.On("SignalWorkflow", mock.Anything, "unit_onboarding_"+customerCreatedEvent.Attributes.Tags.FynbosUserId, mock.AnythingOfType("string"), "onboard-unit-customer-created", mock.Anything).Return(scenario.OnboardingError).Times(1)

		rawEvent := marshalEvent(t, customerCreatedEvent)
		err = wh.HandleEvent(context.Background(), Event{ID: customerCreatedEvent.ID, Type: EventType(customerCreatedEvent.Type)}, rawEvent)

		if scenario.OnboardingError != nil {
			assert.ErrorIs(t, err, ErrInternal, scenario.Name)
		} else {
			assert.NoError(t, err, scenario.Name)
		}
	}
}

func TestHandleApplicationDeniedEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	identityMock := identity_mock.NewMockClient(ctrl)
	temporalMockClient := &mocks.Client{}

	db := test_utils.MigrateCockroachDB(t, ctx)

	provider, err := NewService(ServiceArgs{
		BaseURL:         "localhost:8080",
		Token:           "token",
		WebhookToken:    "webhooktoken",
		Db:              db,
		IdentityService: identityMock,
		AccountClient:   accounts_mock.NewMockClient(ctrl),
		Logger:          zap.NewNop(),
	})
	if err != nil {
		t.Fatal(err)
	}

	wh, err := NewWebhook(&WebhookArgs{
		Up: provider,
		Db: db,
		Tp: temporalMockClient,
	})
	if err != nil {
		t.Fatal(err)
	}

	scenarios := []struct {
		Name            string
		OnboardingError error
	}{
		{
			Name:            "Succeeds if unit onboarding is initiated.",
			OnboardingError: nil,
		},
	}
	for _, scenario := range scenarios {
		applicationDeniedEvent := NewApplicationDeniedEvent()

		temporalMockClient.On("SignalWorkflow", mock.Anything, "unit_onboarding_"+applicationDeniedEvent.Attributes.Tags.FynbosUserId, mock.AnythingOfType("string"), "onboard-unit-application-denied", mock.Anything).Return(scenario.OnboardingError).Times(1)

		rawEvent := marshalEvent(t, applicationDeniedEvent)
		err = wh.HandleEvent(context.Background(), Event{ID: applicationDeniedEvent.ID, Type: EventType(applicationDeniedEvent.Type)}, rawEvent)

		if scenario.OnboardingError != nil {
			assert.ErrorIs(t, err, ErrInternal, scenario.Name)
		} else {
			assert.NoError(t, err, scenario.Name)
		}
	}
}

func TestHandlePaymentEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	identityMock := identity_mock.NewMockClient(ctrl)
	temporalMockClient := &mocks.Client{}

	db := test_utils.MigrateCockroachDB(t, ctx)

	provider, err := NewService(ServiceArgs{
		BaseURL:         "localhost:8080",
		Token:           "token",
		WebhookToken:    "webhooktoken",
		Db:              db,
		IdentityService: identityMock,
		AccountClient:   accounts_mock.NewMockClient(ctrl),
		Logger:          zap.NewNop(),
	})
	if err != nil {
		t.Fatal(err)
	}

	wh, err := NewWebhook(&WebhookArgs{
		Up: provider,
		Db: db,
		Tp: temporalMockClient,
	})
	if err != nil {
		t.Fatal(err)
	}

	depositID := uuid.NewString()
	paymentEvent := external.AchPayment{
		Type: string(PAYMENT_SENT),
		ID:   uuid.NewString(),
		Attributes: external.AchPaymentAttributes{
			Tags: external.DepositTags{
				DepositID: depositID,
			},
		},
	}

	cases := []EventType{
		PAYMENT_CANCELED,
		PAYMENT_CLEARING,
		PAYMENT_CREATED,
		PAYMENT_REJECTED,
		PAYMENT_RETURNED,
		PAYMENT_SENT,
		PAYMENT_PENDING_REVIEW,
	}

	for _, eventType := range cases {
		t.Run(string(eventType), func(st *testing.T) {
			temporalMockClient.On(
				"SignalWorkflow",
				mock.Anything,
				"deposit_"+depositID,
				mock.AnythingOfType("string"),
				"unit-user-ach-deposit",
				string(eventType),
			).Return(nil).Times(1)
			paymentEvent.ID = uuid.NewString()
			paymentEvent.Type = string(eventType)
			rawEvent := marshalEvent(t, paymentEvent)

			err = wh.HandleEvent(context.Background(), Event{ID: paymentEvent.ID, Type: eventType}, rawEvent)

			assert.NoError(t, err)
		})
	}
}

func TestDontFailForUnknownEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	identityMock := identity_mock.NewMockClient(ctrl)
	temporalMockClient := &mocks.Client{}
	db := test_utils.MigrateCockroachDB(t, ctx)
	provider, err := NewService(ServiceArgs{
		BaseURL:         "localhost:8080",
		Token:           "token",
		WebhookToken:    "webhooktoken",
		Db:              db,
		IdentityService: identityMock,
		AccountClient:   accounts_mock.NewMockClient(ctrl),
		Logger:          zap.NewNop(),
	})
	if err != nil {
		t.Fatal(err)
	}
	wh, err := NewWebhook(&WebhookArgs{
		Db: db,
		Up: provider,
		Tp: temporalMockClient,
	})
	if err != nil {
		t.Fatal(err)
	}

	customerCreatedEvent := NewCustomerCreatedEvent()

	rawEvent := marshalBody(t, customerCreatedEvent)
	err = wh.HandleEvent(context.Background(), Event{ID: customerCreatedEvent.ID, Type: EventType("unknown")}, rawEvent.Bytes())
	assert.NoError(t, err)
}

func TestStoreEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	identityMock := identity_mock.NewMockClient(ctrl)
	temporalMockClient := &mocks.Client{}
	db := test_utils.MigrateCockroachDB(t, ctx)
	provider, err := NewService(ServiceArgs{
		BaseURL:         "localhost:8080",
		Token:           "token",
		WebhookToken:    "webhooktoken",
		Db:              db,
		IdentityService: identityMock,
		AccountClient:   accounts_mock.NewMockClient(ctrl),
		Logger:          zap.NewNop(),
	})
	if err != nil {
		t.Fatal(err)
	}
	wh, err := NewWebhook(&WebhookArgs{
		Db: db,
		Up: provider,
		Tp: temporalMockClient,
	})
	if err != nil {
		t.Fatal(err)
	}

	customerCreatedEvent := NewCustomerCreatedEvent()
	rawEvent := marshalEvent(t, customerCreatedEvent)
	testEvent := Event{ID: customerCreatedEvent.ID, Type: EventType(customerCreatedEvent.Type)}

	storedEvent, err := wh.StoreEvent(ctx, testEvent, rawEvent)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, testEvent.ID, storedEvent.ID)
	assert.Equal(t, testEvent.Type, EventType(storedEvent.Type))
	assert.JSONEq(t, string(rawEvent), storedEvent.RawEvent.String())
}

func TestStoreDuplicateEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	identityMock := identity_mock.NewMockClient(ctrl)
	temporalMockClient := &mocks.Client{}
	db := test_utils.MigrateCockroachDB(t, ctx)
	provider, err := NewService(ServiceArgs{
		BaseURL:         "localhost:8080",
		Token:           "token",
		WebhookToken:    "webhooktoken",
		Db:              db,
		IdentityService: identityMock,
		AccountClient:   accounts_mock.NewMockClient(ctrl),
		Logger:          zap.NewNop(),
	})
	if err != nil {
		t.Fatal(err)
	}
	wh, err := NewWebhook(&WebhookArgs{
		Db: db,
		Up: provider,
		Tp: temporalMockClient,
	})
	if err != nil {
		t.Fatal(err)
	}

	customerCreatedEvent := NewCustomerCreatedEvent()
	rawEvent := marshalEvent(t, customerCreatedEvent)
	testEvent := Event{ID: customerCreatedEvent.ID, Type: EventType(customerCreatedEvent.Type)}

	_, err = wh.StoreEvent(ctx, testEvent, rawEvent)
	if err != nil {
		t.Fatal(err)
	}

	_, err = wh.StoreEvent(ctx, testEvent, rawEvent)

	assert.ErrorIs(t, err, ErrDuplicateEvent)
}

func marshalBody(t *testing.T, events ...interface{}) *bytes.Buffer {
	body := ResponseBody{}
	for _, event := range events {
		rawEvent, err := json.Marshal(event)
		if err != nil {
			t.Fatal(err)
		}
		body.Data = append(body.Data, rawEvent)
	}

	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	return bytes.NewBuffer(raw)
}

func marshalEvent(t *testing.T, event interface{}) json.RawMessage {
	raw, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}

	return raw
}

func NewCustomerCreatedEvent() CustomerCreatedEvent {
	return CustomerCreatedEvent{
		ID:   uuid.NewString(),
		Type: "customer.created",
		Attributes: EventAttributes{
			CreatedAt: "2020-07-29T12:53:05.882Z",
			Tags: Tags{
				FynbosUserId: uuid.NewString(),
			},
		},
		Relationships: EventRelationships{
			Customer: JsonCustomer{
				Data: Data{
					ID:   "52",
					Type: "individualCustomer",
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
}

func NewApplicationDeniedEvent() ApplicationDeniedEvent {
	return ApplicationDeniedEvent{
		ID:   uuid.NewString(),
		Type: "application.denied",
		Attributes: EventAttributes{
			CreatedAt: "2020-07-29T12:53:05.882Z",
			Tags: Tags{
				FynbosUserId: uuid.NewString(),
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
}
