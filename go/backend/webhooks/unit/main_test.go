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

func TestWebhook(t *testing.T) {
	ctrl := gomock.NewController(t)
	os := onboarding.NewMockService(ctrl)
	provider := unit.NewMockService(ctrl)
	wh, err := NewWebhook(&WebhookArgs{
		Up: provider,
		Os: os,
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
	}{
		{
			Name:                   "Returns 200",
			VerifyError:            nil,
			Payload:                marshalBody(t, customerCreatedEvent, customerCreatedEvent),
			ExpectedHttpStatusCode: 200,
		},
		{
			Name:                   "Returns 500 if webhook fails verification",
			VerifyError:            errors.New("test"),
			Payload:                marshalBody(t, customerCreatedEvent, customerCreatedEvent),
			ExpectedHttpStatusCode: 401,
			ResponseMessage:        "Signature didn't match.\n",
		},
		{
			Name:                   "Returns 500 if marshalling payload fails",
			VerifyError:            nil,
			Payload:                bytes.NewBuffer([]byte("")),
			ExpectedHttpStatusCode: 500,
			ResponseMessage:        "Failed to parse payload\n",
		},
	}

	for _, scenario := range scenarios {
		if scenario.ExpectedHttpStatusCode == 200 {
			os.EXPECT().InitiateUnitCustomerOnboarding(context.Background(), gomock.Any()).Return(nil).Times(2)
		}
		provider.EXPECT().VerifyWebhook(context.Background(), gomock.Any()).Return(scenario.VerifyError)
		resp, err := http.Post(svr.URL, "application/json", scenario.Payload)
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
}

func TestHandleCreatedCustomerEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	os := onboarding.NewMockService(ctrl)
	db := &sqlx.DB{}
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
		Os: os,
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
		os.EXPECT().InitiateUnitCustomerOnboarding(gomock.Any(), &onboarding.InitiateUnitCustomerOnboardingArgs{
			IdentityID: customerCreatedEvent.Attributes.Tags[unit.ApplicationFormUserIDTag],
			Country:    "US",
		}).Return(scenario.OnboardingError).Times(1)

		rawEvent := marshalEvent(t, customerCreatedEvent)
		err = wh.HandleEvent(context.Background(), CUSTOMER_CREATED, rawEvent)

		if scenario.OnboardingError != nil {
			assert.ErrorIs(t, err, ErrInternal, scenario.Name)
		} else {
			assert.NoError(t, err, scenario.Name)
		}
	}
}

func TestDontFailForUnknownEvent(t *testing.T) {
	// don't fail as Unit may add new event types.
	ctrl := gomock.NewController(t)
	os := onboarding.NewMockService(ctrl)
	db := &sqlx.DB{}
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
		Os: os,
	})
	if err != nil {
		t.Fatal(err)
	}

	rawEvent := marshalBody(t, customerCreatedEvent)
	err = wh.HandleEvent(context.Background(), EventType("unknown"), rawEvent.Bytes())
	assert.NoError(t, err)
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
