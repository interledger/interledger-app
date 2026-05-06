package ops_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linkedaccounts_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"
	"gitlab.com/fynbos/backend/providers/chimoney"
	"gitlab.com/fynbos/backend/providers/gatehub"
	rafiki_external "gitlab.com/fynbos/backend/rafiki/external"
	external_mock "gitlab.com/fynbos/backend/rafiki/external/mock"
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
	if os.Getenv("DB_URL") == "" {
		t.Skip("DB_URL not set; skipping test")
	}

	ctx := context.Background()
	conn := db.MigrateTestDB(t, ctx)

	for _, ppID := range []string{"wa_1", "wa_2", "wa_3", "wa_r1", "wa_r2", "wa_r3"} {
		_, err := conn.ExecContext(ctx,
			`INSERT INTO rafiki_payment_pointers (wallet_id, payment_pointer_id) VALUES ($1, $2) ON CONFLICT (payment_pointer_id) DO NOTHING`,
			uuid.NewString(), ppID)
		require.NoError(t, err)
	}

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
			`{"id":"evt_3","type":"outgoing_payment.created","data":{"id":"op_1","walletAddressId":"wa_1","state":"FUNDING","receiver":"https://example.com/incoming-payments/ip_1","debitAmount":{"value":"100","assetCode":"EUR","assetScale":2},"receiveAmount":{"value":"100","assetCode":"EUR","assetScale":2},"sentAmount":{"value":"0","assetCode":"EUR","assetScale":2},"createdAt":"2024-01-15T10:00:00Z","updatedAt":"2024-01-15T10:00:00Z"}}`,
		},
		{
			"outgoing_payment.completed",
			"outgoing_payment.completed",
			`{"id":"evt_4","type":"outgoing_payment.completed","data":{"id":"op_2","walletAddressId":"wa_2","state":"COMPLETED","receiver":"https://example.com/incoming-payments/ip_2","debitAmount":{"value":"100","assetCode":"EUR","assetScale":2},"receiveAmount":{"value":"100","assetCode":"EUR","assetScale":2},"sentAmount":{"value":"100","assetCode":"EUR","assetScale":2},"createdAt":"2024-01-15T10:00:00Z","updatedAt":"2024-01-15T10:01:00Z"}}`,
		},
		{
			"outgoing_payment.failed",
			"outgoing_payment.failed",
			`{"id":"evt_5","type":"outgoing_payment.failed","data":{"id":"op_3","walletAddressId":"wa_3","state":"FAILED","receiver":"https://example.com/incoming-payments/ip_3","debitAmount":{"value":"200","assetCode":"EUR","assetScale":2},"receiveAmount":{"value":"200","assetCode":"EUR","assetScale":2},"sentAmount":{"value":"0","assetCode":"EUR","assetScale":2},"createdAt":"2024-01-15T10:00:00Z","updatedAt":"2024-01-15T10:02:00Z"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tp := temporal_mock.NewMockClient(ctrl)
			tp.EXPECT().ExecuteWorkflow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&fakeWorkflowRun{}, nil)

			laMock := linkedaccounts_mock.NewMockClient(ctrl)
			laMock.EXPECT().ListBalances(gomock.Any(), gomock.Any()).AnyTimes().
				Return([]linkedaccounts.LinkedAccount{{
					ID:              "la_test",
					Provider:        gatehub.ProviderName,
					Type:            gatehub.AccTypeBalance,
					SendCurrency:    currency.EUR,
					ReceiveCurrency: currency.EUR,
				}}, nil)

			ext := external_mock.NewMockClient(ctrl)
			if strings.HasPrefix(tt.hookType, "outgoing_payment.") {
				ipID := "ip_1"
				receiverWA := "wa_r1"
				if tt.hookType == "outgoing_payment.completed" {
					ipID = "ip_2"
					receiverWA = "wa_r2"
				}
				if tt.hookType == "outgoing_payment.failed" {
					ipID = "ip_3"
					receiverWA = "wa_r3"
				}
				ext.EXPECT().GetIncomingPayment(gomock.Any(), ipID).
					Return(&rafiki_external.GetIncomingPaymentIncomingPayment{
						Id:              ipID,
						WalletAddressId: receiverWA,
					}, nil)
			}

			b := ops.NewTestBackends(
				func(tb *ops.TestBackends) {
					tb.SetTemporal(tp)
					tb.SetDB(conn)
					tb.SetLinkedAccounts(laMock)
					tb.SetExternal(ext)
				},
			)
			handler := ops.EventWebhook(b)

			req := httptest.NewRequest(http.MethodPost, "/rafiki", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code, "webhook should return 200 when workflow starts successfully")
		})
	}
}

func TestEventWebhookOutgoingCreated_MixedProviders_UsesOldFlow(t *testing.T) {
	if os.Getenv("DB_URL") == "" {
		t.Skip("DB_URL not set; skipping test")
	}

	ctx := context.Background()
	conn := db.MigrateTestDB(t, ctx)

	senderWalletID := uuid.NewString()
	receiverWalletID := uuid.NewString()
	_, err := conn.ExecContext(ctx,
		`INSERT INTO rafiki_payment_pointers (wallet_id, payment_pointer_id) VALUES ($1, $2) ON CONFLICT (payment_pointer_id) DO NOTHING`,
		senderWalletID, "wa_sender")
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx,
		`INSERT INTO rafiki_payment_pointers (wallet_id, payment_pointer_id) VALUES ($1, $2) ON CONFLICT (payment_pointer_id) DO NOTHING`,
		receiverWalletID, "wa_receiver")
	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// No ExecuteWorkflow expectation: mixed providers must not use new workflows.
	tp := temporal_mock.NewMockClient(ctrl)

	laMock := linkedaccounts_mock.NewMockClient(ctrl)
	laMock.EXPECT().ListBalances(gomock.Any(), senderWalletID).
		Return([]linkedaccounts.LinkedAccount{{
			ID:              "la_sender",
			Provider:        gatehub.ProviderName,
			Type:            gatehub.AccTypeBalance,
			SendCurrency:    currency.EUR,
			ReceiveCurrency: currency.EUR,
		}}, nil).Times(1)
	laMock.EXPECT().ListBalances(gomock.Any(), receiverWalletID).
		Return([]linkedaccounts.LinkedAccount{{
			ID:              "la_receiver",
			Provider:        chimoney.ProviderName,
			Type:            gatehub.AccTypeBalance,
			SendCurrency:    currency.EUR,
			ReceiveCurrency: currency.EUR,
		}}, nil).Times(1)

	ext := external_mock.NewMockClient(ctrl)
	ext.EXPECT().GetIncomingPayment(gomock.Any(), "ip_mixed").
		Return(&rafiki_external.GetIncomingPaymentIncomingPayment{
			Id:              "ip_mixed",
			WalletAddressId: "wa_receiver",
		}, nil).Times(1)

	b := ops.NewTestBackends(func(tb *ops.TestBackends) {
		tb.SetTemporal(tp)
		tb.SetDB(conn)
		tb.SetLinkedAccounts(laMock)
		tb.SetExternal(ext)
	})

	handler := ops.EventWebhook(b)
	body := `{"id":"evt_mixed","type":"outgoing_payment.created","data":{"id":"op_mixed","walletAddressId":"wa_sender","state":"FUNDING","receiver":"https://example.com/incoming-payments/ip_mixed","debitAmount":{"value":"0","assetCode":"EUR","assetScale":2},"receiveAmount":{"value":"0","assetCode":"EUR","assetScale":2},"sentAmount":{"value":"0","assetCode":"EUR","assetScale":2},"createdAt":"2024-01-15T10:00:00Z","updatedAt":"2024-01-15T10:00:00Z"}}`
	req := httptest.NewRequest(http.MethodPost, "/rafiki", strings.NewReader(body))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "mixed providers should use old flow path for outgoing_payment.created")
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
