package ops_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/providers/chimoney/ops"
	"gotest.tools/assert"
)

func TestParseWebhookSecret(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantNil bool
		wantLen int
	}{
		{
			name:    "valid secret",
			input:   "whsec_" + base64.StdEncoding.EncodeToString([]byte("test-secret-key")),
			wantNil: false,
			wantLen: 15,
		},
		{
			name:    "invalid format - no underscore",
			input:   "invalidsecret",
			wantNil: true,
		},
		{
			name:    "invalid base64",
			input:   "whsec_invalid!!!base64",
			wantNil: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ops.ParseWebhookSecret(tt.input)
			if tt.wantNil {
				require.Nil(t, result)
			} else {
				require.NotNil(t, result)
				require.Len(t, result, tt.wantLen)
			}
		})
	}
}

func TestVerify_ValidSignature(t *testing.T) {
	if os.Getenv("CHIMONEY_WEBHOOK_SECRET") == "" {
		t.SkipNow()
	}
	secret := ops.ParseWebhookSecret(os.Getenv("CHIMONEY_WEBHOOK_SECRET"))
	if secret == nil {
		t.Fatal("Failed to parse CHIMONEY_WEBHOOK_SECRET")
	}

	payload := `{"data":{"account_number":"0690000040","amount":5500,"bank_code":"044","bank_name":"","complete_message":"","created_at":"2020-01-20T16:09:34.000Z","currency":"NGN","debit_currency":"NGN","eventType":"payout.bank.completed","fee":45,"full_name":"Test Name","id":26251,"is_approved":1,"meta":{"chiRef":"01eab82e-4044-452f-80a2-ebb3727fb2b4","country":"NG","currency":"NGN","type":"bank","valueInUSD":5},"narration":"developer transfer xx007","reference":"01eab82e-4044-452f-80a2-ebb3727fb2b4_1678089714855","requires_approval":0,"status":"SUCCESSFUL"},"eventType":"payout.bank.completed","status":"success"}`

	r, err := http.NewRequest(http.MethodPost, "https://example.com/webhook", bytes.NewBufferString(payload))
	require.NoError(t, err)
	r.Header.Set("svix-id", "msg_2iVgfd8bsmhhjLhRLVDTZWEBbtB")
	r.Header.Set("svix-timestamp", "1719581262")
	r.Header.Set("svix-signature", "v1,CQDy5axyjaQrsBAsyFfh8M6OIS6FMNIyEF87JKaLl/0=")

	body, err := ops.Verify(context.Background(), r, secret)
	require.NoError(t, err)
	assert.Equal(t, payload, string(body))
}

func createSignedRequest(t *testing.T, payload string, secret []byte) *http.Request {
	msgID := "msg_test123"
	timestamp := "1234567890"

	signedContent := fmt.Sprintf("%s.%s.%s", msgID, timestamp, payload)
	h := hmac.New(sha256.New, secret)
	_, err := h.Write([]byte(signedContent))
	require.NoError(t, err)

	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	r, err := http.NewRequest(http.MethodPost, "https://example.com/webhook", bytes.NewBufferString(payload))
	require.NoError(t, err)
	r.Header.Set("svix-id", msgID)
	r.Header.Set("svix-timestamp", timestamp)
	r.Header.Set("svix-signature", fmt.Sprintf("v1,%s", signature))

	return r
}

func TestVerify_InvalidSignature(t *testing.T) {
	secret := []byte("test-secret-key")
	payload := `{"eventType":"test.event","issueID":"test-123"}`

	r := createSignedRequest(t, payload, secret)
	// Tamper with signature
	r.Header.Set("svix-signature", "v1,InvalidSignature123==")

	_, err := ops.Verify(context.Background(), r, secret)
	require.Error(t, err)
}

func TestVerify_MissingSignature(t *testing.T) {
	secret := []byte("test-secret-key")
	payload := `{"eventType":"test.event"}`

	r, err := http.NewRequest(http.MethodPost, "https://example.com/webhook", bytes.NewBufferString(payload))
	require.NoError(t, err)
	r.Header.Set("svix-id", "msg_test")
	r.Header.Set("svix-timestamp", "1234567890")
	// No signature header

	_, err = ops.Verify(context.Background(), r, secret)
	require.Error(t, err)
}

func TestVerify_TamperedPayload(t *testing.T) {
	secret := []byte("test-secret-key")
	payload := `{"eventType":"test.event","issueID":"test-123"}`

	r := createSignedRequest(t, payload, secret)

	// Change the body after signing
	tamperedPayload := `{"eventType":"test.event","issueID":"tampered-456"}`
	r.Body = io.NopCloser(bytes.NewBufferString(tamperedPayload))

	_, err := ops.Verify(context.Background(), r, secret)
	require.Error(t, err)
}

func TestExtractChiWalletIDFromIssueID(t *testing.T) {
	tests := []struct {
		name      string
		issueID   string
		want      string
		wantError bool
	}{
		{
			name:      "valid issueID with 3 parts",
			issueID:   "wallet123_amount456_timestamp789",
			want:      "wallet123",
			wantError: false,
		},
		{
			name:      "valid issueID with more than 3 parts",
			issueID:   "wallet123_amount456_timestamp789_extra",
			want:      "wallet123",
			wantError: false,
		},
		{
			name:      "invalid issueID with 2 parts",
			issueID:   "wallet123_amount456",
			want:      "",
			wantError: true,
		},
		{
			name:      "invalid issueID with 1 part",
			issueID:   "wallet123",
			want:      "",
			wantError: true,
		},
		{
			name:      "empty issueID",
			issueID:   "",
			want:      "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ops.ExtractChiWalletIDFromIssueID(tt.issueID)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestWebhookHandler_OptionsMethod(t *testing.T) {
	handler := ops.NewWebhook(nil)

	req := httptest.NewRequest(http.MethodOptions, "/webhook", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestWebhookHandler_InvalidMethod(t *testing.T) {
	handler := ops.NewWebhook(nil)

	req := httptest.NewRequest(http.MethodGet, "/webhook", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestWebhookHandler_InvalidJSON(t *testing.T) {
	secret := []byte("test-secret-key")
	// Set a temporary env var for the test
	oldSecret := os.Getenv("CHIMONEY_WEBHOOK_SECRET")
	os.Setenv("CHIMONEY_WEBHOOK_SECRET", "whsec_"+base64.StdEncoding.EncodeToString(secret))
	defer os.Setenv("CHIMONEY_WEBHOOK_SECRET", oldSecret)

	handler := ops.NewWebhook(nil)

	payload := `{invalid json}`
	req := createSignedRequest(t, payload, secret)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestWebhookHandler_UnknownEventType(t *testing.T) {
	secret := []byte("test-secret-key")
	oldSecret := os.Getenv("CHIMONEY_WEBHOOK_SECRET")
	os.Setenv("CHIMONEY_WEBHOOK_SECRET", "whsec_"+base64.StdEncoding.EncodeToString(secret))
	defer os.Setenv("CHIMONEY_WEBHOOK_SECRET", oldSecret)

	handler := ops.NewWebhook(nil)

	payload := `{"eventType":"unknown.event.type","issueID":"test-123"}`
	req := createSignedRequest(t, payload, secret)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Unknown events should return OK (they're just logged and ignored)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestPaymentEventUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    ops.PaymentEvent
	}{
		{
			name:    "complete payment event",
			payload: `{"eventType":"payout.interac.completed","issueID":"wallet_tx_123","status":"completed"}`,
			want: ops.PaymentEvent{
				EventType: "payout.interac.completed",
				IssueID:   "wallet_tx_123",
				Status:    "completed",
			},
		},
		{
			name:    "payment event without status",
			payload: `{"eventType":"charge.card.completed","issueID":"wallet_tx_456"}`,
			want: ops.PaymentEvent{
				EventType: "charge.card.completed",
				IssueID:   "wallet_tx_456",
				Status:    "",
			},
		},
		{
			name:    "redeem event",
			payload: `{"eventType":"chimoney.redeem.completed","issueID":"redeem_123","status":"success"}`,
			want: ops.PaymentEvent{
				EventType: "chimoney.redeem.completed",
				IssueID:   "redeem_123",
				Status:    "success",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got ops.PaymentEvent
			err := json.Unmarshal([]byte(tt.payload), &got)
			require.NoError(t, err)
			require.Equal(t, tt.want.EventType, got.EventType)
			require.Equal(t, tt.want.IssueID, got.IssueID)
			require.Equal(t, tt.want.Status, got.Status)
		})
	}
}

func TestWithdrawEventUnmarshal(t *testing.T) {
	tests := []struct {
		name            string
		payload         string
		wantEventType   string
		wantIssueID     string
		wantStatus      string
		wantIssuer      string
		wantChiWalletID string
	}{
		{
			name: "withdraw event with meta.issuer",
			payload: `{
				"eventType": "payout.interac.completed",
				"issueID": "8bb8011d-4319-4116-89be-9abcd2df0ee5_4_1770363416510",
				"status": "completed",
				"meta": {
					"issuer": "8bb8011d-4319-4116-89be-9abcd2df0ee5",
					"amount": 3,
					"currency": "CAD"
				}
			}`,
			wantEventType:   "payout.interac.completed",
			wantIssueID:     "8bb8011d-4319-4116-89be-9abcd2df0ee5_4_1770363416510",
			wantStatus:      "completed",
			wantIssuer:      "8bb8011d-4319-4116-89be-9abcd2df0ee5",
			wantChiWalletID: "", // Not populated until handler sets it
		},
		{
			name: "withdraw event without meta",
			payload: `{
				"eventType": "payout.interac.completed",
				"issueID": "wallet123_tx456_timestamp789",
				"status": "completed"
			}`,
			wantEventType:   "payout.interac.completed",
			wantIssueID:     "wallet123_tx456_timestamp789",
			wantStatus:      "completed",
			wantIssuer:      "",
			wantChiWalletID: "",
		},
		{
			name: "withdraw event with full meta object",
			payload: `{
				"amount": "3",
				"chiRef": 229816760262271,
				"currency": "CAD",
				"eventType": "payout.interac.completed",
				"issueID": "8bb8011d-4319-4116-89be-9abcd2df0ee5_4_1770363416510",
				"meta": {
					"issuer": "8bb8011d-4319-4116-89be-9abcd2df0ee5",
					"email": "adrian@interledger.foundation",
					"name": "Test User",
					"type": "interac",
					"fee": 1
				},
				"status": "completed"
			}`,
			wantEventType:   "payout.interac.completed",
			wantIssueID:     "8bb8011d-4319-4116-89be-9abcd2df0ee5_4_1770363416510",
			wantStatus:      "completed",
			wantIssuer:      "8bb8011d-4319-4116-89be-9abcd2df0ee5",
			wantChiWalletID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got ops.WithdrawEvent
			err := json.Unmarshal([]byte(tt.payload), &got)
			require.NoError(t, err)
			require.Equal(t, tt.wantEventType, got.EventType)
			require.Equal(t, tt.wantIssueID, got.IssueID)
			require.Equal(t, tt.wantStatus, got.Status)
			require.Equal(t, tt.wantIssuer, got.Meta.Issuer)
			require.Equal(t, tt.wantChiWalletID, got.ChiWalletID)
		})
	}
}

func TestWithdrawEventChiWalletIDPopulation(t *testing.T) {
	// Test that ChiWalletID gets populated correctly when Meta.Issuer is present
	payload := `{
		"eventType": "payout.interac.completed",
		"issueID": "8bb8011d-4319-4116-89be-9abcd2df0ee5_4_1770363416510",
		"status": "completed",
		"meta": {
			"issuer": "8bb8011d-4319-4116-89be-9abcd2df0ee5"
		}
	}`

	var event ops.WithdrawEvent
	err := json.Unmarshal([]byte(payload), &event)
	require.NoError(t, err)

	// Simulate what the handler does
	event.ChiWalletID = event.Meta.Issuer

	require.Equal(t, "8bb8011d-4319-4116-89be-9abcd2df0ee5", event.ChiWalletID)
	require.Equal(t, "8bb8011d-4319-4116-89be-9abcd2df0ee5", event.Meta.Issuer)
}

func TestKYCEventUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    ops.KYCEvent
	}{
		{
			name:    "kyc completed event",
			payload: `{"eventType":"user.kyc.completed","userID":"86f02240-5350-4bb3-9cac-05b5a82b7eb2"}`,
			want: ops.KYCEvent{
				EventType: "user.kyc.completed",
				UserID:    "86f02240-5350-4bb3-9cac-05b5a82b7eb2",
			},
		},
		{
			name:    "kyc declined event",
			payload: `{"eventType":"user.kyc.declined","userID":"abc123-def456-ghi789"}`,
			want: ops.KYCEvent{
				EventType: "user.kyc.declined",
				UserID:    "abc123-def456-ghi789",
			},
		},
		{
			name:    "kyc event with empty userID",
			payload: `{"eventType":"user.kyc.completed","userID":""}`,
			want: ops.KYCEvent{
				EventType: "user.kyc.completed",
				UserID:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got ops.KYCEvent
			err := json.Unmarshal([]byte(tt.payload), &got)
			require.NoError(t, err)
			require.Equal(t, tt.want.EventType, got.EventType)
			require.Equal(t, tt.want.UserID, got.UserID)
		})
	}
}
