package webhooks_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/providers/unit"
	unit_mock "gitlab.com/fynbos/backend/providers/unit/client/mock"
	"gitlab.com/fynbos/backend/providers/unit/external"
	"gitlab.com/fynbos/backend/providers/unit/webhooks"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"
)

type backends struct {
	unit     *unit_mock.MockClient
	temporal *mocks.Client
}

func NewTestBackends(t *testing.T) backends {
	ctrl := gomock.NewController(t)

	return backends{
		unit:     unit_mock.NewMockClient(ctrl),
		temporal: &mocks.Client{},
	}
}

func (b backends) Temporal() client.Client {
	return b.temporal
}

func (b backends) Unit() unit.Client {
	return b.unit
}

func TestWebhook(t *testing.T) {
	t.Parallel()
	b := NewTestBackends(t)
	wh := webhooks.MakeHttpHandler(b)
	svr := httptest.NewServer(wh)
	t.Cleanup(func() {
		svr.Close()
	})

	t.Run("Returns 200", func(st *testing.T) {
		b.unit.EXPECT().VerifyWebhook(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		b.temporal.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
			func(ctx context.Context, opts client.StartWorkflowOptions, workflow interface{}, args ...interface{}) client.WorkflowRun {
				testWorkflowID := opts.ID
				testRunID := "test-runid"
				require.Len(st, args, 1)    // we sent in the array of events
				require.Len(st, args[0], 2) // the array has 2 entries

				mockWorkflowRun := &mocks.WorkflowRun{}
				mockWorkflowRun.On("GetID").Return(testWorkflowID)
				mockWorkflowRun.On("GetRunID").Return(testRunID)
				mockWorkflowRun.On("Get", mock.Anything, mock.Anything).Return(nil)
				return mockWorkflowRun
			}, nil,
		).Times(1)
		payload := marshalBody(t, NewCustomerCreatedEvent(), NewCustomerCreatedEvent())

		resp, err := http.Post(svr.URL, "application/json", payload)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Returns 401 if webhook fails verification", func(st *testing.T) {
		b.unit.EXPECT().VerifyWebhook(gomock.Any(), gomock.Any(), gomock.Any()).Return(unit.ErrUnauthorized).Times(1)
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
		b.unit.EXPECT().VerifyWebhook(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

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
