package ops_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	temporal_mock "gitlab.com/fynbos/backend/temporal/mock"
	"go.temporal.io/sdk/client"

	"gitlab.com/fynbos/backend/rafiki/ops"
)

func TestEventWebhookUnsupportedType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	b := ops.NewTestBackends()
	handler := ops.EventWebhook(b)

	body := `{"id":"evt_123","type":"unknown.event","data":{}}`
	req := httptest.NewRequest(http.MethodPost, "/rafiki", strings.NewReader(body))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "unsupported webhook types should return 200 OK")
}

func TestEventWebhookInvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	b := ops.NewTestBackends()
	handler := ops.EventWebhook(b)

	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/rafiki", strings.NewReader(body))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "invalid JSON should return 500")
}

func TestEventWebhookOptions(t *testing.T) {
	b := ops.NewTestBackends()
	handler := ops.EventWebhook(b)

	req := httptest.NewRequest(http.MethodOptions, "/rafiki", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "OPTIONS should return 200")
}

func TestEventWebhookIncomingPaymentCreated_NilDB(t *testing.T) {
	b := ops.NewTestBackends()
	handler := ops.EventWebhook(b)

	body := `{
		"id": "evt_incoming_created_123",
		"type": "incoming_payment.created",
		"data": {
			"id": "ip_123",
			"walletAddressId": "wa_123",
			"createdAt": "2024-01-15T10:00:00Z",
			"expiresAt": "2024-01-16T10:00:00Z",
			"receivedAmount": {"value": "0", "assetCode": "EUR", "assetScale": 2},
			"completed": false
		}
	}`
	req := httptest.NewRequest(http.MethodPost, "/rafiki", strings.NewReader(body))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code, "incoming_payment.created returns 400 if provider lookup fails")
}

func TestEventWebhookWorkflowTypes_StartsWorkflow(t *testing.T) {
	tests := []struct {
		name     string
		hookType string
		body     string
	}{
		{
			"incoming_payment.completed",
			"incoming_payment.completed",
			`{"id":"evt_1","type":"incoming_payment.completed","data":{"id":"ip_1","walletAddressId":"wa_1","createdAt":"2024-01-15T10:00:00Z","expiresAt":"2024-01-16T10:00:00Z","receivedAmount":{"value":"5000","assetCode":"EUR","assetScale":2},"completed":true}}`,
		},
		{
			"incoming_payment.expired",
			"incoming_payment.expired",
			`{"id":"evt_2","type":"incoming_payment.expired","data":{"id":"ip_2","walletAddressId":"wa_2","createdAt":"2024-01-15T10:00:00Z","expiresAt":"2024-01-16T10:00:00Z","receivedAmount":{"value":"1000","assetCode":"EUR","assetScale":2},"completed":false}}`,
		},
		{
			"outgoing_payment.created",
			"outgoing_payment.created",
			`{"id":"evt_3","type":"outgoing_payment.created","data":{"id":"op_1","walletAddressId":"wa_1","state":"FUNDING","receiver":"https://example.com/ip/ip_1","debitAmount":{"value":"100","assetCode":"EUR","assetScale":2},"receiveAmount":{"value":"100","assetCode":"EUR","assetScale":2},"sentAmount":{"value":"0","assetCode":"EUR","assetScale":2},"createdAt":"2024-01-15T10:00:00Z","updatedAt":"2024-01-15T10:00:00Z"}}`,
		},
		{
			"outgoing_payment.completed",
			"outgoing_payment.completed",
			`{"id":"evt_4","type":"outgoing_payment.completed","data":{"id":"op_2","walletAddressId":"wa_2","state":"COMPLETED","receiver":"https://example.com/ip/ip_2","debitAmount":{"value":"100","assetCode":"EUR","assetScale":2},"receiveAmount":{"value":"100","assetCode":"EUR","assetScale":2},"sentAmount":{"value":"100","assetCode":"EUR","assetScale":2},"createdAt":"2024-01-15T10:00:00Z","updatedAt":"2024-01-15T10:01:00Z"}}`,
		},
		{
			"outgoing_payment.failed",
			"outgoing_payment.failed",
			`{"id":"evt_5","type":"outgoing_payment.failed","data":{"id":"op_3","walletAddressId":"wa_3","state":"FAILED","receiver":"https://example.com/ip/ip_3","debitAmount":{"value":"200","assetCode":"EUR","assetScale":2},"receiveAmount":{"value":"200","assetCode":"EUR","assetScale":2},"sentAmount":{"value":"0","assetCode":"EUR","assetScale":2},"createdAt":"2024-01-15T10:00:00Z","updatedAt":"2024-01-15T10:02:00Z"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tp := temporal_mock.NewMockClient(ctrl)
			tp.EXPECT().ExecuteWorkflow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&fakeWorkflowRun{}, nil)

			b := ops.NewTestBackends(
				func(tb *ops.TestBackends) { tb.SetTemporal(tp) },
			)
			handler := ops.EventWebhook(b)

			req := httptest.NewRequest(http.MethodPost, "/rafiki", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code, "webhook should return 200 when workflow starts successfully")
		})
	}
}

func TestEventWebhookWorkflowTypes_UnmarshalFails(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			"incoming_payment.completed bad data",
			`{"id":"evt_1","type":"incoming_payment.completed","data":"not an object"}`,
		},
		{
			"outgoing_payment.created bad data",
			`{"id":"evt_2","type":"outgoing_payment.created","data":"not an object"}`,
		},
		{
			"outgoing_payment.completed bad data",
			`{"id":"evt_3","type":"outgoing_payment.completed","data":"not an object"}`,
		},
		{
			"outgoing_payment.failed bad data",
			`{"id":"evt_4","type":"outgoing_payment.failed","data":"not an object"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			b := ops.NewTestBackends()
			handler := ops.EventWebhook(b)

			req := httptest.NewRequest(http.MethodPost, "/rafiki", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code, "bad data should return 400")
		})
	}
}

// fakeWorkflowRun satisfies client.WorkflowRun for tests.
type fakeWorkflowRun struct{}

func (f *fakeWorkflowRun) GetID() string    { return "wf_test" }
func (f *fakeWorkflowRun) GetRunID() string { return "run_test" }
func (f *fakeWorkflowRun) Get(_ context.Context, _ interface{}) error {
	return nil
}
func (f *fakeWorkflowRun) GetWithOptions(_ context.Context, _ interface{}, _ client.WorkflowRunGetOptions) error {
	return nil
}
