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

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/providers/unit"
	test_utils "gitlab.com/fynbos/backend/utils"
)

func TestWebhook(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db := test_utils.MigrateCockroachDB(t, ctx)
	ctrl := gomock.NewController(t)
	osMock := onboarding.NewMockService(ctrl)
	providerMock := unit.NewMockService(ctrl)

	wh, err := NewWebhook(&WebhookArgs{
		Db: db,
		Os: osMock,
		Up: providerMock,
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
			osMock.EXPECT().InitiateUnitCustomerOnboarding(context.Background(), gomock.Any()).Return(nil).Times(scenario.MockCallTimes)
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
	osMock := onboarding.NewMockService(ctrl)
	
	db := test_utils.MigrateCockroachDB(t, ctx)

	provider, err := unit.NewService(unit.ServiceArgs{
		BaseURL:      "localhost:8080",
		Token:        "token",
		WebhookToken: "webhooktoken",
		Db:           db,
	})
	if err != nil {
		t.Fatal(err)
	}

	wh, err := NewWebhook(&WebhookArgs{
		Up: provider,
		Os: osMock,
		Db: db,
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
		{
			Name:            "Returns ErrInternal if unit onboarding fails to initiate",
			OnboardingError: onboarding.ErrInternal,
		},
	}
	for _, scenario := range scenarios {
		customerCreatedEvent := NewCustomerCreatedEvent()

		osMock.EXPECT().InitiateUnitCustomerOnboarding(gomock.Any(), &onboarding.InitiateUnitCustomerOnboardingArgs{
			IdentityID:   customerCreatedEvent.Attributes.Tags[unit.ApplicationFormUserIDTag],
			Country:      "US",
			CustomerID:   customerCreatedEvent.Relationships.Customer.Data.ID,
			CustomerType: customerCreatedEvent.Relationships.Customer.Data.Type,
		}).Return(scenario.OnboardingError).Times(1)

		rawEvent := marshalEvent(t, customerCreatedEvent)
		err = wh.HandleEvent(context.Background(), Event{ ID: customerCreatedEvent.ID, Type: EventType(customerCreatedEvent.Type)}, rawEvent)

		if scenario.OnboardingError != nil {
			assert.ErrorIs(t, err, ErrInternal, scenario.Name)
		} else {
			assert.NoError(t, err, scenario.Name)
		}
	}
}

func TestDontFailForUnknownEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	osMock := onboarding.NewMockService(ctrl)
	db := test_utils.MigrateCockroachDB(t, ctx)
	provider, err := unit.NewService(unit.ServiceArgs{
		BaseURL:      "localhost:8080",
		Token:        "token",
		WebhookToken: "webhooktoken",
		Db:           db,
	})
	if err != nil {
		t.Fatal(err)
	}
	wh, err := NewWebhook(&WebhookArgs{
		Db: db,
		Os: osMock,
		Up: provider,
	})
	if err != nil {
		t.Fatal(err)
	}

	customerCreatedEvent := NewCustomerCreatedEvent()

	rawEvent := marshalBody(t, customerCreatedEvent)
	err = wh.HandleEvent(context.Background(), Event{ ID: customerCreatedEvent.ID, Type: EventType("unknown")}, rawEvent.Bytes())
	assert.NoError(t, err)
}

func TestStoreEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	osMock := onboarding.NewMockService(ctrl)
	db := test_utils.MigrateCockroachDB(t, ctx)
	provider, err := unit.NewService(unit.ServiceArgs{
		BaseURL:      "localhost:8080",
		Token:        "token",
		WebhookToken: "webhooktoken",
		Db:           db,
	})
	if err != nil {
		t.Fatal(err)
	}
	wh, err := NewWebhook(&WebhookArgs{
		Db: db,
		Os: osMock,
		Up: provider,
	})
	if err != nil {
		t.Fatal(err)
	}

	customerCreatedEvent := NewCustomerCreatedEvent()
	rawEvent := marshalEvent(t, customerCreatedEvent)
	testEvent := Event{ ID: customerCreatedEvent.ID, Type: EventType(customerCreatedEvent.Type) }

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
	osMock := onboarding.NewMockService(ctrl)
	db := test_utils.MigrateCockroachDB(t, ctx)
	provider, err := unit.NewService(unit.ServiceArgs{
		BaseURL:      "localhost:8080",
		Token:        "token",
		WebhookToken: "webhooktoken",
		Db:           db,
	})
	if err != nil {
		t.Fatal(err)
	}
	wh, err := NewWebhook(&WebhookArgs{
		Db: db,
		Os: osMock,
		Up: provider,
	})
	if err != nil {
		t.Fatal(err)
	}

	customerCreatedEvent := NewCustomerCreatedEvent()
	rawEvent := marshalEvent(t, customerCreatedEvent)
	testEvent := Event{ ID: customerCreatedEvent.ID, Type: EventType(customerCreatedEvent.Type) }

	_, err = wh.StoreEvent(ctx, testEvent, rawEvent)
	if err != nil {
		t.Fatal(err)
	}

	_, err = wh.StoreEvent(ctx, testEvent, rawEvent)

	assert.ErrorIs(t, err, ErrDuplicateEvent)
}

func marshalBody(t *testing.T, events ...interface{}) *bytes.Buffer {
	body := Body{}
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
		Attributes: CustomerCreatedAttributes{
			CreatedAt: "2020-07-29T12:53:05.882Z",
			Tags: map[string]string{
				unit.ApplicationFormUserIDTag: uuid.NewString(),
			},
		},
		Relationships: CustomerCreatedRelationships{
			Customer: Customer{
				Data: Data{
					ID:   "52",
					Type: "individualCustomer",
				},
			},
			Application: Application{
				Data: Data{
					ID:   "52",
					Type: "individualApplication",
				},
			},
		},
	}
}
