package ops_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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

func TestEventWebhookReturnsOKImmediatelyForValidPayload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
			"incomingAmount": {"value": "1000", "assetCode": "EUR", "assetScale": 2},
			"receivedAmount": {"value": "0", "assetCode": "EUR", "assetScale": 2},
			"completed": false
		}
	}`
	req := httptest.NewRequest(http.MethodPost, "/rafiki", strings.NewReader(body))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "webhook should return 200 OK for a valid payload")
}

func TestEventWebhookReturnsOKImmediatelyForInvalidOrMalformedPayloads(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"incoming_payment.completed invalid receivedAmount", `{"id":"evt_1","type":"incoming_payment.completed","data":{"id":"ip_1","walletAddressId":"wa_1","createdAt":"2024-01-15T10:00:00Z","expiresAt":"2024-01-16T10:00:00Z","receivedAmount":{"value":"not_a_number","assetCode":"EUR","assetScale":2},"completed":true}}`},
		{"outgoing_payment.completed invalid sentAmount", `{"id":"evt_1","type":"outgoing_payment.completed","data":{"id":"op_1","walletAddressId":"wa_1","state":"COMPLETED","receiver":"https://example.com/ip/ip_1","debitAmount":{"value":"100","assetCode":"EUR","assetScale":2},"receiveAmount":{"value":"100","assetCode":"EUR","assetScale":2},"sentAmount":{"value":"invalid","assetCode":"EUR","assetScale":2},"createdAt":"2024-01-15T10:00:00Z","updatedAt":"2024-01-15T10:01:00Z"}}`},
		{"outgoing_payment.created invalid debitAmount", `{"id":"evt_1","type":"outgoing_payment.created","data":{"id":"op_1","walletAddressId":"wa_1","state":"FUNDING","receiver":"https://example.com/ip/ip_1","debitAmount":{"value":"invalid","assetCode":"EUR","assetScale":2},"receiveAmount":{"value":"1000","assetCode":"EUR","assetScale":2},"sentAmount":{"value":"0","assetCode":"EUR","assetScale":2},"createdAt":"2024-01-15T10:00:00Z","updatedAt":"2024-01-15T10:00:00Z"}}`},
		{"incoming_payment.partial_payment_received invalid receivedAmount", `{"id":"evt_1","type":"incoming_payment.partial_payment_received","data":{"id":"ip_1","walletAddressId":"wa_1","createdAt":"2024-01-15T10:00:00Z","expiresAt":"2024-01-16T10:00:00Z","receivedAmount":{"value":"not_a_number","assetCode":"EUR","assetScale":2},"completed":false}}`},
		{"incoming_payment.expired invalid receivedAmount", `{"id":"evt_1","type":"incoming_payment.expired","data":{"id":"ip_1","walletAddressId":"wa_1","createdAt":"2024-01-15T10:00:00Z","expiresAt":"2024-01-16T10:00:00Z","receivedAmount":{"value":"not_a_number","assetCode":"EUR","assetScale":2},"completed":false}}`},
		{"malformed data (not an object)", `{"id":"evt_1","type":"incoming_payment.created","data":"not an object"}`},
		{"outgoing_payment.created valid shape", `{"id":"evt_1","type":"outgoing_payment.created","data":{"id":"op_1","walletAddressId":"wa_1","state":"FUNDING","receiver":"https://example.com/ip/ip_1","debitAmount":{"value":"100","assetCode":"EUR","assetScale":2},"receiveAmount":{"value":"100","assetCode":"EUR","assetScale":2},"sentAmount":{"value":"0","assetCode":"EUR","assetScale":2},"createdAt":"2024-01-15T10:00:00Z","updatedAt":"2024-01-15T10:00:00Z"}}`},
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

			assert.Equal(t, http.StatusOK, rr.Code, "webhook should return 200 OK immediately even if async processing may fail")
		})
	}
}
