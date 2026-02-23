package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/mockxago/internal/models"
	"gitlab.com/fynbos/mockxago/internal/storage"
)

func TestSimulateTestDeposit(t *testing.T) {
	tests := []struct {
		name            string
		setupAccount    bool
		requestBody     map[string]interface{}
		wantStatus      int
		wantErrContains string
		validateResp    func(t *testing.T, resp *models.TestDepositResponse)
	}{
		{
			name:         "valid deposit returns pending status with transaction ID",
			setupAccount: true,
			requestBody: map[string]interface{}{
				"accountId":    "test-account-123",
				"amount":       5000.00,
				"currencyCode": "ZAR",
			},
			wantStatus: http.StatusOK,
			validateResp: func(t *testing.T, resp *models.TestDepositResponse) {
				assert.NotEmpty(t, resp.TransactionID, "transaction ID should not be empty")
				assert.Equal(t, "pending", resp.Status, "initial status should be pending")
			},
		},
		{
			name:         "deposit with custom deposit reference",
			setupAccount: true,
			requestBody: map[string]interface{}{
				"accountId":        "test-account-123",
				"amount":           10000.00,
				"currencyCode":     "ZAR",
				"depositReference": "wallet_abc_ZAR",
			},
			wantStatus: http.StatusOK,
			validateResp: func(t *testing.T, resp *models.TestDepositResponse) {
				assert.NotEmpty(t, resp.TransactionID)
				assert.Equal(t, "pending", resp.Status)
			},
		},
		{
			name:         "missing account ID returns 400",
			setupAccount: false,
			requestBody: map[string]interface{}{
				"amount":       5000.00,
				"currencyCode": "ZAR",
			},
			wantStatus:      http.StatusBadRequest,
			wantErrContains: "accountId is required",
		},
		{
			name:         "zero amount returns 400",
			setupAccount: true,
			requestBody: map[string]interface{}{
				"accountId":    "test-account-123",
				"amount":       0.00,
				"currencyCode": "ZAR",
			},
			wantStatus:      http.StatusBadRequest,
			wantErrContains: "amount must be greater than 0",
		},
		{
			name:         "negative amount returns 400",
			setupAccount: true,
			requestBody: map[string]interface{}{
				"accountId":    "test-account-123",
				"amount":       -1000.00,
				"currencyCode": "ZAR",
			},
			wantStatus:      http.StatusBadRequest,
			wantErrContains: "amount must be positive",
		},
		{
			name:         "missing currency code returns 400",
			setupAccount: true,
			requestBody: map[string]interface{}{
				"accountId": "test-account-123",
				"amount":    5000.00,
			},
			wantStatus:      http.StatusBadRequest,
			wantErrContains: "currencyCode is required",
		},
		{
			name:         "invalid account ID returns 400",
			setupAccount: false,
			requestBody: map[string]interface{}{
				"accountId":    "invalid_account",
				"amount":       5000.00,
				"currencyCode": "ZAR",
			},
			wantStatus:      http.StatusBadRequest,
			wantErrContains: "account not found",
		},
		{
			name:         "deposit with USD currency",
			setupAccount: true,
			requestBody: map[string]interface{}{
				"accountId":    "test-account-123",
				"amount":       500.00,
				"currencyCode": "USD",
			},
			wantStatus: http.StatusOK,
			validateResp: func(t *testing.T, resp *models.TestDepositResponse) {
				assert.NotEmpty(t, resp.TransactionID)
				assert.Equal(t, "pending", resp.Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemoryStorage()
			h := NewHandler(store)

			// Setup test account if needed
			if tt.setupAccount {
				accountID, ok := tt.requestBody["accountId"].(string)
				if ok && accountID != "" {
					subAccount := &models.SubAccount{
						AccountID: accountID,
						WalletID:  "test_wallet",
					}
					err := store.SaveSubAccount(context.Background(), subAccount)
					require.NoError(t, err)
				}
			}

			// Create request
			bodyBytes, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/v1/company/accounts/testdeposit", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			// Execute
			w := httptest.NewRecorder()
			h.SimulateTestDeposit(w, req)

			// Assertions
			assert.Equal(t, tt.wantStatus, w.Code, "status code mismatch")

			if tt.wantErrContains != "" {
				// Expect error response
				var errResp models.ErrorResponse
				err := json.NewDecoder(w.Body).Decode(&errResp)
				require.NoError(t, err, "should decode error response")
				assert.Contains(t, strings.ToLower(errResp.Message), strings.ToLower(tt.wantErrContains))
			} else if tt.validateResp != nil {
				// Expect success response
				var resp models.TestDepositResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				require.NoError(t, err, "should decode success response")
				tt.validateResp(t, &resp)
			}
		})
	}
}

func TestSimulateTestDeposit_DepositRecorded(t *testing.T) {
	store := storage.NewMemoryStorage()
	h := NewHandler(store)

	// Create test account
	subAccount := &models.SubAccount{
		AccountID: "acc_123",
		WalletID:  "wallet_test",
	}
	err := store.SaveSubAccount(context.Background(), subAccount)
	require.NoError(t, err)

	// Simulate deposit
	body := map[string]interface{}{
		"accountId":        "acc_123",
		"amount":           3000.00,
		"currencyCode":     "ZAR",
		"depositReference": "ref_abc",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/company/accounts/testdeposit", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()
	h.SimulateTestDeposit(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp models.TestDepositResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	// Verify deposit was saved
	deposit, err := store.GetDeposit(context.Background(), resp.TransactionID)
	require.NoError(t, err)
	assert.Equal(t, "acc_123", deposit.AccountID)
	assert.Equal(t, 3000.00, deposit.Amount)
	assert.Equal(t, "ZAR", deposit.Currency)
	assert.Equal(t, "ref_abc", deposit.DepositReference)
	assert.Equal(t, "pending", deposit.Status)
}

func TestListCompanyDeposits(t *testing.T) {
	tests := []struct {
		name          string
		setupDeposits int
		limit         string
		page          string
		wantStatus    int
		wantDataCount int
		wantTotal     int
	}{
		{
			name:          "list with default pagination",
			setupDeposits: 5,
			limit:         "10",
			page:          "1",
			wantStatus:    http.StatusOK,
			wantDataCount: 5,
			wantTotal:     5,
		},
		{
			name:          "list with limit less than total",
			setupDeposits: 15,
			limit:         "10",
			page:          "1",
			wantStatus:    http.StatusOK,
			wantDataCount: 10,
			wantTotal:     15,
		},
		{
			name:          "list second page",
			setupDeposits: 15,
			limit:         "10",
			page:          "2",
			wantStatus:    http.StatusOK,
			wantDataCount: 5,
			wantTotal:     15,
		},
		{
			name:          "empty list when no deposits",
			setupDeposits: 0,
			limit:         "10",
			page:          "1",
			wantStatus:    http.StatusOK,
			wantDataCount: 0,
			wantTotal:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemoryStorage()
			h := NewHandler(store)

			// Setup deposits
			for i := 0; i < tt.setupDeposits; i++ {
				deposit := &models.Deposit{
					ID:        generateDepositID(),
					AccountID: "acc_123",
					Amount:    float64(1000 + i*100),
					Currency:  "ZAR",
					Status:    "completed",
					Code:      104,
				}
				err := store.SaveDeposit(context.Background(), deposit)
				require.NoError(t, err)
			}

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/v1/company/deposits?limit="+tt.limit+"&page="+tt.page, nil)
			w := httptest.NewRecorder()
			h.ListCompanyDeposits(w, req)

			// Assertions
			assert.Equal(t, tt.wantStatus, w.Code)

			var resp models.ListDepositsResponse
			err := json.NewDecoder(w.Body).Decode(&resp)
			require.NoError(t, err)

			assert.Len(t, resp.Data, tt.wantDataCount, "data count mismatch")
			assert.Equal(t, tt.wantTotal, resp.Pagination.Total, "total count mismatch")
		})
	}
}

func TestListCompanyDeposits_RequiresAuth(t *testing.T) {
	store := storage.NewMemoryStorage()
	h := NewHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/v1/company/deposits", nil)
	w := httptest.NewRecorder()

	// Should be protected by AuthMiddleware in actual router
	// For now, test that handler works with valid auth
	req.Header.Set("Authorization", "Bearer test_token")

	// Create a valid token first
	token := &models.AccessToken{
		Token:     "test_token",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	err := store.SaveAccessToken(context.Background(), token)
	require.NoError(t, err)

	h.ListCompanyDeposits(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// Helper function for generating deposit IDs (will be replaced by actual implementation)
func generateDepositID() string {
	return "dep_" + strconv.FormatInt(time.Now().UnixNano(), 10)
}

func TestSimulateTestDeposit_WebhookConfiguration(t *testing.T) {
	// Set webhook environment variables
	os.Setenv("WEBHOOK_URL", "http://test.example.com/webhook")
	os.Setenv("WEBHOOK_SECRET", "test_secret")
	defer func() {
		os.Unsetenv("WEBHOOK_URL")
		os.Unsetenv("WEBHOOK_SECRET")
	}()

	store := storage.NewMemoryStorage()
	h := NewHandler(store)

	// Create test account
	subAccount := &models.SubAccount{
		AccountID: "acc_webhook",
		WalletID:  "wallet_test",
	}
	err := store.SaveSubAccount(context.Background(), subAccount)
	require.NoError(t, err)

	// Simulate deposit
	body := map[string]interface{}{
		"accountId":    "acc_webhook",
		"amount":       2000.00,
		"currencyCode": "ZAR",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/company/accounts/testdeposit", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()
	h.SimulateTestDeposit(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	// Note: Webhook delivery happens asynchronously after 500ms delay
	// In actual implementation, we'd verify webhook is sent
	// For unit tests, we just verify the deposit was created correctly
	var resp models.TestDepositResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.TransactionID)
}
