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
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/providers/unit"
)

var customerCreatedEvent = CustomerCreatedEvent{
	ID:   "1",
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

func TestWebhooks(t *testing.T) {
	ctx := context.Background()

	c, err := NewTestContainer(ctx, t)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		err := c.Cleanup(ctx)
		if err != nil {
			return
		}
	})

	t.Run("Test webhooks", func(t *testing.T) {
		t.Cleanup(func() {
			c.SvrMockWh.Close()
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
				Payload:                marshalBody(t, customerCreatedEvent, customerCreatedEvent),
				ExpectedHttpStatusCode: 200,
				MockCallTimes:          2,
			},
			{
				Name:                   "Returns 500 if webhook fails verification",
				VerifyError:            errors.New("test"),
				Payload:                marshalBody(t, customerCreatedEvent, customerCreatedEvent),
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
				Payload:                marshalBody(t, "", customerCreatedEvent),
				ExpectedHttpStatusCode: 500,
				ResponseMessage:        "Failed to parse payload\n",
				MockCallTimes:          1,
			},
		}

		for _, scenario := range scenarios {
			c.OsMock.EXPECT().InitiateUnitCustomerOnboarding(context.Background(), gomock.Any()).Return(nil).Times(scenario.MockCallTimes)
			c.ProviderMock.EXPECT().VerifyWebhook(context.Background(), gomock.Any(), gomock.Any()).Return(scenario.VerifyError)
			resp, err := http.Post(c.SvrMockWh.URL, "application/json", scenario.Payload)
			if err != nil {
				t.Fatal(err)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, scenario.ExpectedHttpStatusCode, resp.StatusCode, scenario.Name)
			assert.Equal(t, scenario.ResponseMessage, string(body), scenario.Name)
		}
	})

	t.Run("Handle created customer event", func(t *testing.T) {
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
			c.OsMock.EXPECT().InitiateUnitCustomerOnboarding(gomock.Any(), &onboarding.InitiateUnitCustomerOnboardingArgs{
				IdentityID:   customerCreatedEvent.Attributes.Tags[unit.ApplicationFormUserIDTag],
				Country:      "US",
				CustomerID:   customerCreatedEvent.Relationships.Customer.Data.ID,
				CustomerType: customerCreatedEvent.Relationships.Customer.Data.Type,
			}).Return(scenario.OnboardingError).Times(1)

			rawEvent := marshalEvent(t, customerCreatedEvent)
			err = c.Wh.HandleEvent(context.Background(), CUSTOMER_CREATED, rawEvent)

			if scenario.OnboardingError != nil {
				assert.ErrorIs(t, err, ErrInternal, scenario.Name)
			} else {
				assert.NoError(t, err, scenario.Name)
			}
		}
	})

	t.Run("Don't fail for unknown event", func(t *testing.T) {
		rawEvent := marshalBody(t, customerCreatedEvent)
		err = c.Wh.HandleEvent(context.Background(), EventType("unknown"), rawEvent.Bytes())
		assert.NoError(t, err)
	})
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

type TestContainer struct {
	Ctx 	context.Context
	Ctrl *gomock.Controller
	OsMock *onboarding.MockService
	ProviderMock *unit.MockService
	Provider unit.Service
	Wh Webhook
	WhMockProvider Webhook
	Db *sqlx.DB
	SvrMockWh *httptest.Server
}

func NewTestContainer(ctx context.Context, t *testing.T) (*TestContainer, error) {
	container := &TestContainer{}

	container.Ctx = ctx

	ctrl := gomock.NewController(t)
	container.Ctrl = ctrl

	osMock := onboarding.NewMockService(ctrl)
	container.OsMock = osMock

	providerMock := unit.NewMockService(ctrl)
	container.ProviderMock = providerMock

	whMockProvider, err := NewWebhook(&WebhookArgs{
		Up: providerMock,
		Os: osMock,
	})
	if err != nil {
		return nil, err
	}
	container.WhMockProvider = whMockProvider

	wh, err := NewWebhook(&WebhookArgs{
		Up: providerMock,
		Os: osMock,
	})
	if err != nil {
		return nil, err
	}
	container.Wh = wh

	db := &sqlx.DB{}
	container.Db = db

	svrMockWh := httptest.NewServer(whMockProvider.MakeHttpHandler())
	container.SvrMockWh = svrMockWh

	return container, nil
}

func (c *TestContainer) Cleanup(ctx context.Context) error {
	c.SvrMockWh.Close()

	return nil
}