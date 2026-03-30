package external_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/chimoney/external"
	"gopkg.in/stretchr/testify.v1/require"
)

// mockChimoneyServer creates a test server that mocks Chimoney API responses
func mockChimoneyServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, external.Client) {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client := external.NewWithBaseURL(server.URL, "test-api-key", server.Client())
	return server, client
}

func TestPayoutStatus_Mock_Success(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"id":      "payout-123",
			"amount":  100.50,
			"fee":     2.50,
			"type":    "interac",
			"issueID": "issue-456",
			"status":  "SUCCESSFUL",
		},
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/payouts/status")

		// Verify headers
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.Equal(t, "application/json", r.Header.Get("Accept"))
		require.Equal(t, "test-api-key", r.Header.Get("X-API-KEY"))

		// Verify request body
		var req external.PayoutStatusRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		require.Equal(t, "test-chiref-123", req.Reference)

		// Send mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	resp, err := client.PayoutStatus(context.Background(), external.PayoutStatusRequest{
		Reference:           "test-chiref-123",
		ChiWallet:           "test-wallet",
		TurnOffNotification: true,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "payout-123", resp.ID)
	require.Equal(t, 100.50, resp.Amount)
	require.Equal(t, 2.50, resp.Fee)
	require.Equal(t, "interac", resp.Type)
	require.Equal(t, "issue-456", resp.IssueID)
	require.Equal(t, "SUCCESSFUL", resp.Status)
}

func TestPayoutStatus_Mock_NotFound(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "error",
		"error":  "Payout not found",
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/payouts/status")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	_, err := client.PayoutStatus(context.Background(), external.PayoutStatusRequest{
		Reference: "non-existent-ref",
		ChiWallet: "test-wallet",
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "Payout not found")
}

func TestPayoutStatus_Mock_InvalidJSON(t *testing.T) {
	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	})

	_, err := client.PayoutStatus(context.Background(), external.PayoutStatusRequest{
		Reference: "test-ref",
	})

	require.Error(t, err)
}

func TestPayoutStatus_Mock_NetworkError(t *testing.T) {
	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Close connection immediately to simulate network error
		hj, ok := w.(http.Hijacker)
		if ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}
	})

	_, err := client.PayoutStatus(context.Background(), external.PayoutStatusRequest{
		Reference: "test-ref",
	})

	require.Error(t, err)
}

func TestPayoutStatus_Mock_HTTPError(t *testing.T) {
	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/payouts/status")

		// Return HTTP 500 error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  "Internal server error",
		})
	})

	_, err := client.PayoutStatus(context.Background(), external.PayoutStatusRequest{
		Reference: "test-ref",
	})

	// Client should handle HTTP errors gracefully
	// The actual behavior depends on the client implementation
	// but typically this should result in an error
	require.Error(t, err)
}

func TestPayoutStatus_Mock_EmptyResponse(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "success",
		"data":   map[string]interface{}{},
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/payouts/status")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	resp, err := client.PayoutStatus(context.Background(), external.PayoutStatusRequest{
		Reference: "test-ref",
	})

	// Should succeed but return empty/default values
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Empty(t, resp.ID)
	require.Empty(t, resp.Status)
}

func TestPayoutStatus_Mock_FailedStatus(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"id":      "payout-failed-456",
			"amount":  50.00,
			"fee":     1.50,
			"type":    "interac",
			"issueID": "issue-failed-789",
			"status":  "FAILED",
		},
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/payouts/status")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	resp, err := client.PayoutStatus(context.Background(), external.PayoutStatusRequest{
		Reference: "failed-payout-ref",
		ChiWallet: "test-wallet",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "payout-failed-456", resp.ID)
	require.Equal(t, "FAILED", resp.Status)
	require.Equal(t, 50.00, resp.Amount)
}

func TestPayoutStatus_Mock_PendingStatus(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"id":      "payout-pending-789",
			"amount":  25.00,
			"fee":     0.75,
			"type":    "interac",
			"issueID": "issue-pending-012",
			"status":  "PENDING",
		},
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/payouts/status")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	resp, err := client.PayoutStatus(context.Background(), external.PayoutStatusRequest{
		Reference: "pending-payout-ref",
		ChiWallet: "test-wallet",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "payout-pending-789", resp.ID)
	require.Equal(t, "PENDING", resp.Status)
	require.Equal(t, 25.00, resp.Amount)
}

func TestPayoutStatus_Mock_CancelledStatus(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"id":      "payout-cancelled-321",
			"amount":  75.00,
			"fee":     2.00,
			"type":    "interac",
			"issueID": "issue-cancelled-654",
			"status":  "CANCELLED",
		},
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/payouts/status")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	resp, err := client.PayoutStatus(context.Background(), external.PayoutStatusRequest{
		Reference: "cancelled-payout-ref",
		ChiWallet: "test-wallet",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "payout-cancelled-321", resp.ID)
	require.Equal(t, "CANCELLED", resp.Status)
	require.Equal(t, 75.00, resp.Amount)
}

func TestVerifyPayment_Mock_Success(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"id":                  "payment-123",
			"valueInUSD":          10,
			"amount":              "13.00",
			"issueID":             "test-issue-123",
			"fee":                 1,
			"type":                "payment",
			"subAccount":          "test-wallet",
			"issuer":              "test-issuer",
			"t_id":                123,
			"chiRef":              "chi-ref-123",
			"currency":            "CAD",
			"interacFee":          0.5,
			"turnOffNotification": false,
			"issueDate":           "2024-01-01T00:00:00Z",
			"payerEmail":          "test@example.com",
			"status":              "paid",
			"redirect_url":        "https://example.com/callback",
			"meta":                nil,
			"integration":         map[string]interface{}{},
			"redeemData":          map[string]interface{}{},
		},
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/payment/verify")

		var req external.VerifyPaymentReq
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		require.Equal(t, "test-issue-123", req.IssueID)
		require.Equal(t, "test-wallet", req.ChiWallet)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	payment, err := client.VerifyPayment(context.Background(), external.VerifyPaymentReq{
		IssueID:   "test-issue-123",
		ChiWallet: "test-wallet",
	})

	require.NoError(t, err)
	require.NotNil(t, payment)
	require.Equal(t, "payment-123", payment.ID)
	require.Equal(t, "test-issue-123", payment.IssueID)
	require.Equal(t, "13.00", payment.Amount)
	require.Equal(t, "paid", payment.Status)
	require.Equal(t, "test@example.com", payment.PayerEmail)
}

func TestVerifyPayment_Mock_NotFound(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "error",
		"error":  "Payment not found",
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	_, err := client.VerifyPayment(context.Background(), external.VerifyPaymentReq{
		IssueID: "non-existent",
	})

	require.Error(t, err)
}

func TestVerifyPayment_Mock_PendingStatus(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"id":                  "payment-pending-999",
			"valueInUSD":          15,
			"amount":              "20.00",
			"issueID":             "test-issue-pending-888",
			"fee":                 1,
			"type":                "payment",
			"subAccount":          "test-wallet",
			"issuer":              "test-issuer",
			"t_id":                456,
			"chiRef":              "chi-ref-pending-777",
			"currency":            "CAD",
			"interacFee":          0.5,
			"turnOffNotification": false,
			"issueDate":           "2024-01-02T00:00:00Z",
			"payerEmail":          "pending@example.com",
			"status":              "pending",
			"redirect_url":        "https://example.com/callback",
			"meta":                nil,
			"integration":         map[string]interface{}{},
			"redeemData":          map[string]interface{}{},
		},
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/payment/verify")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	payment, err := client.VerifyPayment(context.Background(), external.VerifyPaymentReq{
		IssueID:   "test-issue-pending-888",
		ChiWallet: "test-wallet",
	})

	require.NoError(t, err)
	require.NotNil(t, payment)
	require.Equal(t, "payment-pending-999", payment.ID)
	require.Equal(t, "pending", payment.Status)
	require.Equal(t, "20.00", payment.Amount)
}

func TestVerifyPayment_Mock_ExpiredStatus(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"id":                  "payment-expired-111",
			"valueInUSD":          10,
			"amount":              "13.00",
			"issueID":             "test-issue-expired-222",
			"fee":                 1,
			"type":                "payment",
			"subAccount":          "test-wallet",
			"issuer":              "test-issuer",
			"t_id":                789,
			"chiRef":              "chi-ref-expired-333",
			"currency":            "CAD",
			"interacFee":          0.5,
			"turnOffNotification": false,
			"issueDate":           "2024-01-01T00:00:00Z",
			"payerEmail":          "expired@example.com",
			"status":              "expired",
			"redirect_url":        "https://example.com/callback",
			"meta":                nil,
			"integration":         map[string]interface{}{},
			"redeemData":          map[string]interface{}{},
		},
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/payment/verify")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	payment, err := client.VerifyPayment(context.Background(), external.VerifyPaymentReq{
		IssueID:   "test-issue-expired-222",
		ChiWallet: "test-wallet",
	})

	require.NoError(t, err)
	require.NotNil(t, payment)
	require.Equal(t, "payment-expired-111", payment.ID)
	require.Equal(t, "expired", payment.Status)
	require.Equal(t, "13.00", payment.Amount)
}

func TestDeposit_Mock_Success(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"paymentLink": "https://pay.chimoney.io/test-link",
			"issueID":     "test-issue-id-12345",
		},
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/payment/initiate")

		var req external.DepositReq
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		require.Equal(t, "10.00", req.Amount)
		require.Equal(t, "CAD", req.Currency)
		require.Equal(t, "test@example.com", req.Email)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	resp, err := client.Deposit(context.Background(), external.DepositReq{
		Amount:               "10.00",
		Currency:             "CAD",
		Email:                "test@example.com",
		ChimoneyWallet:       "test-wallet",
		TurnOffNotifications: true,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "https://pay.chimoney.io/test-link", resp.PaymentLink)
	require.Equal(t, "test-issue-id-12345", resp.IssueID)
}

func TestConvertToUSD_Mock_Success(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"originCurrency":         "CAD",
			"amountInOriginCurrency": "100",
			"amountInUSD":            75.50,
			"validUntil":             "2024-01-01T23:59:59Z",
			"expiresAt":              "2024-01-01T23:59:59Z",
			"expiresAtTimestamp":     1704153599,
		},
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Contains(t, r.URL.Path, "/info/convert/local-amount-to-usd")

		// Verify query parameters
		query := r.URL.Query()
		require.Equal(t, "100", query.Get("amountInOriginCurrency"))
		require.Equal(t, "CAD", query.Get("originCurrency"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	amt, err := client.ConvertToUSD(context.Background(), external.ConvertToUSDRequest{
		Currency: "CAD",
		Amount:   100,
	})

	require.NoError(t, err)
	require.NotNil(t, amt)
	require.Equal(t, currency.USD, amt.Currency)
	require.Equal(t, uint64(7550), amt.Value) // 75.50 USD in cents
}

func TestConvertToUSD_Mock_InvalidCurrency(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "error",
		"error":  "Invalid currency code",
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Contains(t, r.URL.Path, "/info/convert/local-amount-to-usd")

		query := r.URL.Query()
		require.Equal(t, "INVALID", query.Get("originCurrency"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	_, err := client.ConvertToUSD(context.Background(), external.ConvertToUSDRequest{
		Currency: "INVALID",
		Amount:   100,
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid currency code")
}

func TestConvertToUSD_Mock_UnsupportedCurrency(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "error",
		"error":  "Currency conversion not supported",
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Contains(t, r.URL.Path, "/info/convert/local-amount-to-usd")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	_, err := client.ConvertToUSD(context.Background(), external.ConvertToUSDRequest{
		Currency: "XYZ",
		Amount:   100,
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "Currency conversion not supported")
}

func TestConvertToUSD_Mock_InvalidAmount(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "error",
		"error":  "Amount must be greater than zero",
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Contains(t, r.URL.Path, "/info/convert/local-amount-to-usd")

		query := r.URL.Query()
		require.Equal(t, "0", query.Get("amountInOriginCurrency"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	_, err := client.ConvertToUSD(context.Background(), external.ConvertToUSDRequest{
		Currency: "CAD",
		Amount:   0,
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "Amount must be greater than zero")
}

func TestConvertToUSD_Mock_RateUnavailable(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "error",
		"error":  "Exchange rate temporarily unavailable",
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Contains(t, r.URL.Path, "/info/convert/local-amount-to-usd")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	_, err := client.ConvertToUSD(context.Background(), external.ConvertToUSDRequest{
		Currency: "EUR",
		Amount:   50,
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "Exchange rate temporarily unavailable")
}

func TestConvertToUSD_Mock_HTTPError(t *testing.T) {
	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Contains(t, r.URL.Path, "/info/convert/local-amount-to-usd")

		// Return HTTP 502 Bad Gateway
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  "Gateway error",
		})
	})

	_, err := client.ConvertToUSD(context.Background(), external.ConvertToUSDRequest{
		Currency: "GBP",
		Amount:   100,
	})

	require.Error(t, err)
}

func TestConvertToUSD_Mock_InvalidJSON(t *testing.T) {
	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Contains(t, r.URL.Path, "/info/convert/local-amount-to-usd")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json data"))
	})

	_, err := client.ConvertToUSD(context.Background(), external.ConvertToUSDRequest{
		Currency: "CAD",
		Amount:   100,
	})

	require.Error(t, err)
}

func TestConvertToUSD_Mock_MissingData(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "success",
		"data":   map[string]interface{}{},
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Contains(t, r.URL.Path, "/info/convert/local-amount-to-usd")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	amt, err := client.ConvertToUSD(context.Background(), external.ConvertToUSDRequest{
		Currency: "CAD",
		Amount:   100,
	})

	// Should succeed but return zero/default amount
	require.NoError(t, err)
	require.NotNil(t, amt)
	require.Equal(t, currency.USD, amt.Currency)
	require.Equal(t, uint64(0), amt.Value)
}

func TestCreateWallet_Mock_Success(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"id": "wallet-abc-123",
		},
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/multicurrency-wallets/create")

		var req external.CreateWalletReq
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		require.Equal(t, "Test Wallet", req.Name)
		require.Equal(t, "test@example.com", req.Email)
		require.Equal(t, "John", req.FirstName)
		require.Equal(t, "Doe", req.LastName)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	walletID, err := client.CreateWallet(context.Background(), external.CreateWalletReq{
		Name:      "Test Wallet",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
	})

	require.NoError(t, err)
	require.Equal(t, "wallet-abc-123", walletID)
}

func TestGetWallet_Mock_Success(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"id":         "wallet-123",
			"name":       "Test Wallet",
			"parent":     "parent-123",
			"uid":        "uid-123",
			"subAccount": true,
			"verification": map[string]interface{}{
				"status": "verified",
			},
		},
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Contains(t, r.URL.Path, "/multicurrency-wallets/get")

		query := r.URL.Query()
		require.Equal(t, "wallet-123", query.Get("id"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	wallet, err := client.GetWallet(context.Background(), "wallet-123")

	require.NoError(t, err)
	require.NotNil(t, wallet)
	require.Equal(t, "wallet-123", wallet.ID)
	require.Equal(t, "Test Wallet", wallet.Name)
	require.Equal(t, "parent-123", wallet.Parent)
	require.True(t, wallet.SubAccount)
	require.Equal(t, "verified", wallet.Verification.Status)
}

func TestTransfer_Mock_Success(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "success",
		"data":   map[string]interface{}{},
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/multicurrency-wallets/transfer")
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	err := client.Transfer(context.Background(), external.TransferReq{
		SenderSubAccount:   "sender-wallet",
		ReceiverSubAccount: "receiver-wallet",
		Amount:             currency.FromFloat64(10.50, currency.USD),
	})

	require.NoError(t, err)
}

func TestTransfer_Mock_Error(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "error",
		"error":  "Insufficient balance",
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/multicurrency-wallets/transfer")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	err := client.Transfer(context.Background(), external.TransferReq{
		SenderSubAccount:   "sender-wallet",
		ReceiverSubAccount: "receiver-wallet",
		Amount:             currency.FromFloat64(10.50, currency.USD),
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "Insufficient balance")
}

func TestTransfer_Mock_InvalidReceiver(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "error",
		"error":  "Invalid receiver account",
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/multicurrency-wallets/transfer")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	err := client.Transfer(context.Background(), external.TransferReq{
		SenderSubAccount:   "sender-wallet",
		ReceiverSubAccount: "invalid-receiver",
		Amount:             currency.FromFloat64(10.50, currency.USD),
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid receiver account")
}

func TestTransfer_Mock_InvalidAmount(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "error",
		"error":  "Amount must be greater than zero",
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/multicurrency-wallets/transfer")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	err := client.Transfer(context.Background(), external.TransferReq{
		SenderSubAccount:   "sender-wallet",
		ReceiverSubAccount: "receiver-wallet",
		Amount:             currency.FromFloat64(0, currency.USD),
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "Amount must be greater than zero")
}

func TestTransfer_Mock_AccountNotFound(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "error",
		"error":  "Sender account not found",
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/multicurrency-wallets/transfer")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	err := client.Transfer(context.Background(), external.TransferReq{
		SenderSubAccount:   "non-existent-sender",
		ReceiverSubAccount: "receiver-wallet",
		Amount:             currency.FromFloat64(10.50, currency.USD),
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "Sender account not found")
}

func TestTransfer_Mock_HTTPError(t *testing.T) {
	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/multicurrency-wallets/transfer")

		// Return HTTP 503 Service Unavailable
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  "Service temporarily unavailable",
		})
	})

	err := client.Transfer(context.Background(), external.TransferReq{
		SenderSubAccount:   "sender-wallet",
		ReceiverSubAccount: "receiver-wallet",
		Amount:             currency.FromFloat64(10.50, currency.USD),
	})

	require.Error(t, err)
}

func TestTransfer_Mock_InvalidJSON(t *testing.T) {
	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/multicurrency-wallets/transfer")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json response"))
	})

	err := client.Transfer(context.Background(), external.TransferReq{
		SenderSubAccount:   "sender-wallet",
		ReceiverSubAccount: "receiver-wallet",
		Amount:             currency.FromFloat64(10.50, currency.USD),
	})

	require.Error(t, err)
}

func TestWithdraw_Mock_Success(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":     "withdrawal-123",
					"type":   "interac",
					"chiref": "chi-ref-123",
				},
			},
		},
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/payouts/interac")

		var req external.WithdrawalReq
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		require.NotEmpty(t, req.Interacs)
		require.Equal(t, "John Doe", req.Interacs[0].Name)
		require.Equal(t, "john@example.com", req.Interacs[0].Email)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	resp, err := client.Withdraw(context.Background(), external.WithdrawalReq{
		Interacs: []external.Interacs{
			{
				Name:      "John Doe",
				Email:     "john@example.com",
				Amount:    50.00,
				Narration: "Test withdrawal",
			},
		},
		SubAccount:          "test-wallet",
		TurnOffNotification: true,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Data, 1)
	require.Equal(t, "withdrawal-123", resp.Data[0].ID)
	require.Equal(t, "interac", resp.Data[0].Type)
	require.Equal(t, "chi-ref-123", resp.Data[0].ChiRef)
}

func TestVerifyPayment_Mock_WithRedeemData(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"id":                  "payment-with-redeem-data",
			"valueInUSD":          10,
			"amount":              "13.00",
			"issueID":             "test-issue-redeem-123",
			"fee":                 1,
			"type":                "payment",
			"subAccount":          "test-wallet",
			"issuer":              "test-issuer",
			"t_id":                999,
			"chiRef":              "chi-ref-redeem-888",
			"currency":            "CAD",
			"interacFee":          0.5,
			"turnOffNotification": false,
			"issueDate":           "2024-01-01T00:00:00Z",
			"payerEmail":          "redeem@example.com",
			"status":              "paid",
			"redirect_url":        "https://example.com/callback",
			"meta":                nil,
			"integration":         map[string]interface{}{},
			"redeemData": map[string]interface{}{
				"valueInUSD": 10,
				"walletID":   "wallet-redeem-123",
				"amount":     "15.00",
				"currency":   "CAD",
				"interacFee": 0.75,
				"payerEmail": "payer@example.com",
				"subAccount": "sub-account-123",
			},
		},
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/payment/verify")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	payment, err := client.VerifyPayment(context.Background(), external.VerifyPaymentReq{
		IssueID:   "test-issue-redeem-123",
		ChiWallet: "test-wallet",
	})

	require.NoError(t, err)
	require.NotNil(t, payment)
	require.Equal(t, "payment-with-redeem-data", payment.ID)
	require.Equal(t, "13.00", payment.Amount)
	require.Equal(t, "paid", payment.Status)

	// Verify RedeemData fields, especially amount as string
	require.Equal(t, 10, payment.RedeemData.ValueInUSD)
	require.Equal(t, "wallet-redeem-123", payment.RedeemData.WalletID)
	require.Equal(t, "15.00", payment.RedeemData.Amount)
	require.Equal(t, "CAD", payment.RedeemData.Currency)
	require.Equal(t, 0.75, payment.RedeemData.InteracFee)
	require.Equal(t, "payer@example.com", payment.RedeemData.PayerEmail)
	require.Equal(t, "sub-account-123", payment.RedeemData.SubAccount)
}

func TestGetEstimatedFee_Mock_Success(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"amount":    100.00,
			"currency":  "CAD",
			"rail":      "interac",
			"direction": "payout",
			"totalFee":  1.50,
			"netAmount": 98.50,
			"note":      "Fee includes processing and Interac charges",
		},
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/info/fee-estimate")

		// Verify headers
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.Equal(t, "application/json", r.Header.Get("Accept"))
		require.Equal(t, "test-api-key", r.Header.Get("X-API-KEY"))

		// Verify request body
		var req external.EstimateFeeReq
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		require.Equal(t, "100.00", req.Amount)
		require.Equal(t, "CAD", req.Currency)
		require.Equal(t, "interac", req.Rail)
		require.Equal(t, "payout", req.Direction)

		// Send mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	resp, err := client.GetEstimatedFee(context.Background(), external.EstimateFeeReq{
		Amount:    "100.00",
		Currency:  "CAD",
		Rail:      "interac",
		Direction: "payout",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 100.00, resp.Amount)
	require.Equal(t, "CAD", resp.Currency)
	require.Equal(t, "interac", resp.Rail)
	require.Equal(t, "payout", resp.Direction)
	require.Equal(t, 1.50, resp.TotalFee)
	require.Equal(t, 98.50, resp.NetAmount)
	require.Equal(t, "Fee includes processing and Interac charges", resp.Note)
}

func TestGetEstimatedFee_Mock_MinimalResponse(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"totalFee": 2.00,
			"currency": "CAD",
		},
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/info/fee-estimate")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	resp, err := client.GetEstimatedFee(context.Background(), external.EstimateFeeReq{
		Amount:   "50.00",
		Currency: "CAD",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 2.00, resp.TotalFee)
	require.Equal(t, "CAD", resp.Currency)
}

func TestGetEstimatedFee_Mock_ErrorStatus(t *testing.T) {
	mockResponse := map[string]interface{}{
		"status": "error",
		"error":  "Invalid currency",
	}

	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/info/fee-estimate")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	})

	_, err := client.GetEstimatedFee(context.Background(), external.EstimateFeeReq{
		Amount:   "100.00",
		Currency: "INVALID",
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid currency")
}

func TestGetEstimatedFee_Mock_HTTPError(t *testing.T) {
	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/info/fee-estimate")

		// Return HTTP 400 Bad Request
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  "Bad request",
		})
	})

	_, err := client.GetEstimatedFee(context.Background(), external.EstimateFeeReq{
		Amount:   "invalid",
		Currency: "CAD",
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "request failed on estimating fee")
	require.Contains(t, err.Error(), "400")
}

func TestGetEstimatedFee_Mock_InvalidJSON(t *testing.T) {
	_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not valid json"))
	})

	_, err := client.GetEstimatedFee(context.Background(), external.EstimateFeeReq{
		Amount:   "100.00",
		Currency: "CAD",
	})

	require.Error(t, err)
}

func TestGetEstimatedFee_Mock_DifferentRails(t *testing.T) {
	tests := []struct {
		name      string
		rail      string
		direction string
		totalFee  float64
	}{
		{
			name:      "Interac payout",
			rail:      "interac",
			direction: "payout",
			totalFee:  1.50,
		},
		{
			name:      "Bank transfer",
			rail:      "bank",
			direction: "payout",
			totalFee:  2.00,
		},
		{
			name:      "Deposit",
			rail:      "interac",
			direction: "deposit",
			totalFee:  0.50,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockResponse := map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"amount":    100.00,
					"currency":  "CAD",
					"rail":      tc.rail,
					"direction": tc.direction,
					"totalFee":  tc.totalFee,
					"netAmount": 100.00 - tc.totalFee,
				},
			}

			_, client := mockChimoneyServer(t, func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)

				var req external.EstimateFeeReq
				json.NewDecoder(r.Body).Decode(&req)
				require.Equal(t, tc.rail, req.Rail)
				require.Equal(t, tc.direction, req.Direction)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(mockResponse)
			})

			resp, err := client.GetEstimatedFee(context.Background(), external.EstimateFeeReq{
				Amount:    "100.00",
				Currency:  "CAD",
				Rail:      tc.rail,
				Direction: tc.direction,
			})

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Equal(t, tc.totalFee, resp.TotalFee)
			require.Equal(t, tc.rail, resp.Rail)
			require.Equal(t, tc.direction, resp.Direction)
		})
	}
}
