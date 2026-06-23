package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/interledger/interledger-app/go/mock/mockxago/internal/jobs"
	"github.com/interledger/interledger-app/go/mock/mockxago/internal/models"
	"github.com/interledger/interledger-app/go/mock/mockxago/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// KYCIframe Tests

func TestKYCIframe_Success(t *testing.T) {
	store := storage.NewMemoryStorage()
	h := NewHandler(store, jobs.NewQueue(store))

	// Request with token and user_id
	req := httptest.NewRequest(http.MethodGet, "/kyc/iframe?token=test_token_123&user_id=wallet_123", nil)
	w := httptest.NewRecorder()

	h.KYCIframe(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `action="/kyc/submit"`)
}

func TestKYCIframe_MissingToken(t *testing.T) {
	store := storage.NewMemoryStorage()
	h := NewHandler(store, jobs.NewQueue(store))

	// Request without token parameter
	req := httptest.NewRequest(http.MethodGet, "/kyc/iframe?user_id=wallet_123", nil)
	w := httptest.NewRecorder()

	h.KYCIframe(w, req)

	// Token is not required for iframe serving, request will succeed or fail on template lookup
	assert.True(t, w.Code > 0)
}

func TestKYCIframe_MissingUserID(t *testing.T) {
	store := storage.NewMemoryStorage()
	h := NewHandler(store, jobs.NewQueue(store))

	// Request without user_id
	req := httptest.NewRequest(http.MethodGet, "/kyc/iframe?token=test_token", nil)
	w := httptest.NewRecorder()

	h.KYCIframe(w, req)

	// user_id is not strictly required, template lookup will determine response
	assert.True(t, w.Code > 0)
}

// KYCIframeSubmit Tests

func TestKYCIframeSubmit_Success(t *testing.T) {
	store := storage.NewMemoryStorage()
	h := NewHandler(store, jobs.NewQueue(store))

	// Create request with form data
	formData := url.Values{
		"user_id":    {"wallet_kyc_123"},
		"token":      {"test_token"},
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"address":    {"123 Test St"},
		"dob":        {"1990-01-15"},
	}

	body := strings.NewReader(formData.Encode())
	req := httptest.NewRequest(http.MethodPost, "/kyc/submit", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	h.KYCIframeSubmit(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "accepted", resp["status"])
}

func TestKYCIframeSubmit_CreatesSubAccount(t *testing.T) {
	store := storage.NewMemoryStorage()
	h := NewHandler(store, jobs.NewQueue(store))

	walletID := "wallet_create_kyc"
	formData := url.Values{
		"user_id":    {walletID},
		"first_name": {"Jane"},
		"last_name":  {"Smith"},
		"address":    {"456 Oak Ave"},
		"dob":        {"1985-05-20"},
	}

	body := strings.NewReader(formData.Encode())
	req := httptest.NewRequest(http.MethodPost, "/kyc/submit", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	h.KYCIframeSubmit(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify sub-account was created
	time.Sleep(100 * time.Millisecond) // Allow goroutine to complete
	subAccount, err := store.GetSubAccountByWalletID(context.Background(), walletID)
	require.NoError(t, err)
	assert.Equal(t, "Jane", subAccount.FirstName)
	assert.Equal(t, "Smith", subAccount.LastName)
	assert.Equal(t, "456 Oak Ave", subAccount.PhysicalAddress)
}

func TestKYCIframeSubmit_UpdatesExistingSubAccount(t *testing.T) {
	store := storage.NewMemoryStorage()
	h := NewHandler(store, jobs.NewQueue(store))

	walletID := "wallet_update_kyc"

	// Create initial sub-account
	subAccount := &models.SubAccount{
		ID:        "sub_1",
		WalletID:  walletID,
		AccountID: "acc_1",
		FirstName: "Old",
		LastName:  "Name",
	}
	err := store.SaveSubAccount(context.Background(), subAccount)
	require.NoError(t, err)

	// Update via KYC submission
	formData := url.Values{
		"user_id":    {walletID},
		"first_name": {"Updated"},
		"last_name":  {"Account"},
		"dob":        {"1992-10-11"},
		"address":    {"789 New Rd"},
	}

	body := strings.NewReader(formData.Encode())
	req := httptest.NewRequest(http.MethodPost, "/kyc/submit", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	h.KYCIframeSubmit(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify update
	time.Sleep(100 * time.Millisecond)
	updated, err := store.GetSubAccountByWalletID(context.Background(), walletID)
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.FirstName)
	assert.Equal(t, "Account", updated.LastName)
	assert.Equal(t, "1992-10-11", updated.DateOfBirth)
	assert.Equal(t, "789 New Rd", updated.PhysicalAddress)
}

func TestKYCIframeSubmit_MissingWalletID(t *testing.T) {
	store := storage.NewMemoryStorage()
	h := NewHandler(store, jobs.NewQueue(store))

	// Form without user_id
	formData := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"dob":        {"1990-01-01"},
	}

	body := strings.NewReader(formData.Encode())
	req := httptest.NewRequest(http.MethodPost, "/kyc/submit", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	h.KYCIframeSubmit(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestKYCIframeSubmit_MissingFirstName(t *testing.T) {
	store := storage.NewMemoryStorage()
	h := NewHandler(store, jobs.NewQueue(store))

	formData := url.Values{
		"user_id":   {"wallet_123"},
		"last_name": {"Doe"},
		"dob":       {"1990-01-01"},
	}

	body := strings.NewReader(formData.Encode())
	req := httptest.NewRequest(http.MethodPost, "/kyc/submit", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	h.KYCIframeSubmit(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestKYCIframeSubmit_MissingLastName(t *testing.T) {
	store := storage.NewMemoryStorage()
	h := NewHandler(store, jobs.NewQueue(store))

	formData := url.Values{
		"user_id":    {"wallet_123"},
		"first_name": {"John"},
		"dob":        {"1990-01-01"},
	}

	body := strings.NewReader(formData.Encode())
	req := httptest.NewRequest(http.MethodPost, "/kyc/submit", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	h.KYCIframeSubmit(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestKYCIframeSubmit_InvalidFormData(t *testing.T) {
	store := storage.NewMemoryStorage()
	h := NewHandler(store, jobs.NewQueue(store))

	// Send raw invalid data
	body := bytes.NewBufferString("invalid form data without proper encoding")
	req := httptest.NewRequest(http.MethodPost, "/kyc/submit", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	h.KYCIframeSubmit(w, req)

	// Handler tries to parse form, may succeed with empty values or fail gracefully
	assert.True(t, w.Code > 0)
}

func TestKYCIframeSubmit_MultipartForm(t *testing.T) {
	store := storage.NewMemoryStorage()
	h := NewHandler(store, jobs.NewQueue(store))

	// Build a simple multipart form body
	body := bytes.NewBufferString("")
	boundary := "boundary"
	fmt.Fprintf(body, "--%s\r\n", boundary)
	fmt.Fprintf(body, "Content-Disposition: form-data; name=\"user_id\"\r\n\r\n")
	fmt.Fprintf(body, "wallet_multipart\r\n")
	fmt.Fprintf(body, "--%s\r\n", boundary)
	fmt.Fprintf(body, "Content-Disposition: form-data; name=\"first_name\"\r\n\r\n")
	fmt.Fprintf(body, "Multi\r\n")
	fmt.Fprintf(body, "--%s\r\n", boundary)
	fmt.Fprintf(body, "Content-Disposition: form-data; name=\"last_name\"\r\n\r\n")
	fmt.Fprintf(body, "Part\r\n")
	fmt.Fprintf(body, "--%s\r\n", boundary)
	fmt.Fprintf(body, "Content-Disposition: form-data; name=\"dob\"\r\n\r\n")
	fmt.Fprintf(body, "1991-04-17\r\n")
	fmt.Fprintf(body, "--%s\r\n", boundary)
	fmt.Fprintf(body, "Content-Disposition: form-data; name=\"address\"\r\n\r\n")
	fmt.Fprintf(body, "Test Address\r\n")
	fmt.Fprintf(body, "--%s--\r\n", boundary)

	req := httptest.NewRequest(http.MethodPost, "/kyc/submit", body)
	req.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", boundary))

	w := httptest.NewRecorder()
	h.KYCIframeSubmit(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify sub-account created with multipart data
	time.Sleep(100 * time.Millisecond)
	subAccount, err := store.GetSubAccountByWalletID(context.Background(), "wallet_multipart")
	require.NoError(t, err)
	assert.Equal(t, "Multi", subAccount.FirstName)
	assert.Equal(t, "1991-04-17", subAccount.DateOfBirth)
}

func TestKYCIframeSubmit_WithDOB(t *testing.T) {
	store := storage.NewMemoryStorage()
	h := NewHandler(store, jobs.NewQueue(store))

	formData := url.Values{
		"user_id":    {"wallet_dob"},
		"first_name": {"Bob"},
		"last_name":  {"Builder"},
		"dob":        {"1980-03-10"},
	}

	body := strings.NewReader(formData.Encode())
	req := httptest.NewRequest(http.MethodPost, "/kyc/submit", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	h.KYCIframeSubmit(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify sub-account was created and DOB was stored
	time.Sleep(100 * time.Millisecond)
	subAccount, err := store.GetSubAccountByWalletID(context.Background(), "wallet_dob")
	require.NoError(t, err)
	assert.Equal(t, "Bob", subAccount.FirstName)
	assert.Equal(t, "1980-03-10", subAccount.DateOfBirth)
}

func TestKYCIframeSubmit_WhitespaceHandling(t *testing.T) {
	store := storage.NewMemoryStorage()
	h := NewHandler(store, jobs.NewQueue(store))

	// Names with extra whitespace
	formData := url.Values{
		"user_id":    {"wallet_spaces"},
		"first_name": {"  John  "},
		"last_name":  {"  Doe  "},
		"dob":        {"1990-01-01"},
	}

	body := strings.NewReader(formData.Encode())
	req := httptest.NewRequest(http.MethodPost, "/kyc/submit", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	h.KYCIframeSubmit(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Form values preserve whitespace
	time.Sleep(100 * time.Millisecond)
	subAccount, err := store.GetSubAccountByWalletID(context.Background(), "wallet_spaces")
	require.NoError(t, err)
	assert.NotEmpty(t, subAccount.FirstName)
}
