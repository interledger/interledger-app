package webhooks_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/providers/unit"
	unit_mock "gitlab.com/fynbos/backend/providers/unit/client/mock"
	"gitlab.com/fynbos/backend/providers/unit/external"
	"gitlab.com/fynbos/backend/providers/unit/webhooks"
)

func TestWebhook(t *testing.T) {
	t.Parallel()
	// ctx := context.Background()
	ctrl := gomock.NewController(t)
	client := unit_mock.NewMockClient(ctrl)
	wh := webhooks.MakeHttpHandler(client)
	svr := httptest.NewServer(wh)
	t.Cleanup(func() {
		svr.Close()
	})

	t.Run("Returns 200", func(st *testing.T) {
		client.EXPECT().VerifyWebhook(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		client.EXPECT().HandleEvent(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)
		payload := marshalBody(t, NewCustomerCreatedEvent(), NewCustomerCreatedEvent())

		resp, err := http.Post(svr.URL, "application/json", payload)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Returns 401 if webhook fails verification", func(st *testing.T) {
		client.EXPECT().VerifyWebhook(gomock.Any(), gomock.Any(), gomock.Any()).Return(unit.ErrUnauthorized).Times(1)
		payload := marshalBody(t, NewCustomerCreatedEvent(), NewCustomerCreatedEvent())

		resp, err := http.Post(svr.URL, "application/json", payload)
		if err != nil {
			t.Fatal(err)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 401, resp.StatusCode)
		assert.Equal(t, "Signature didn't match.\n", string(body))
	})

	t.Run("Returns 500 if marshalling payload fails", func(st *testing.T) {
		client.EXPECT().VerifyWebhook(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

		resp, err := http.Post(svr.URL, "application/json", bytes.NewBuffer([]byte("")))
		if err != nil {
			t.Fatal(err)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 500, resp.StatusCode)
		assert.Equal(t, "Failed to parse payload\n", string(body))
	})

	t.Run("Tries to handle all events even if first one fails", func(st *testing.T) {
		client.EXPECT().VerifyWebhook(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		client.EXPECT().HandleEvent(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		payload := marshalBody(t, "", NewCustomerCreatedEvent())

		resp, err := http.Post(svr.URL, "application/json", payload)
		if err != nil {
			t.Fatal(err)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 500, resp.StatusCode)
		assert.Equal(t, "Failed to parse payload\n", string(body))
	})
}

func marshalBody(t *testing.T, events ...interface{}) *bytes.Buffer {
	body := struct {
		Data []json.RawMessage `json:"data"`
	}{}
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
