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
	"github.com/stretchr/testify/assert"
)

func TestWebhook(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := NewMockService(ctrl)
	svr := httptest.NewServer(NewHttpHandler(provider))
	t.Cleanup(func() {
		svr.Close()
	})
	event := CustomerCreatedEvent{
		ID:   "1",
		Type: "customer.created",
		Attributes: CustomerCreatedAttributes{
			CreatedAt: "2020-07-29T12:53:05.882Z",
			Tags: map[string]string{
				"tag": "value",
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
			Payload:                marshal(t, []CustomerCreatedEvent{event, event}),
			ExpectedHttpStatusCode: 200,
		},
		{
			Name:                   "Returns 500 if webhook fails verification",
			VerifyError:            errors.New("test"),
			Payload:                marshal(t, []CustomerCreatedEvent{event, event}),
			ExpectedHttpStatusCode: 500,
			ResponseMessage:        "Signature didn't match.",
		},
		{
			Name:                   "Returns 500 if marshalling payload fails",
			VerifyError:            nil,
			Payload:                bytes.NewBuffer([]byte("")),
			ExpectedHttpStatusCode: 500,
			ResponseMessage:        "Failed to parse payload",
		},
	}

	for _, scenario := range scenarios {
		provider.EXPECT().VerifyWebhook(context.Background(), gomock.Any()).Return(scenario.VerifyError)
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
	}
}

func marshal(t *testing.T, data interface{}) *bytes.Buffer {
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	return bytes.NewBuffer(raw)
}
