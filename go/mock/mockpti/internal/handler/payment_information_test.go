package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/interledger/interledger-app/go/mock/mockpti/internal/models"
)

func TestCreatePaymentInformation_Success(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)
	userID := createTestUser(t, h)

	body := models.CreatePaymentInformationRequest{
		Type:              "BANK_ACCOUNT",
		BankAccountNumber: "123456789",
		BankRoutingNumber: "021000021",
		AccountBankName:   "Test Bank",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users/"+userID+"/payment-information", bytes.NewReader(payload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp models.PaymentInformation
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if resp.ID == "" {
		t.Error("expected non-empty payment information id")
	}
	if resp.Type != "BANK_ACCOUNT" {
		t.Errorf("expected type BANK_ACCOUNT, got %s", resp.Type)
	}
	if resp.BankAccountNumber != "123456789" {
		t.Errorf("expected bank account number 123456789, got %s", resp.BankAccountNumber)
	}
	if resp.BankRoutingNumber != "021000021" {
		t.Errorf("expected routing number 021000021, got %s", resp.BankRoutingNumber)
	}
}

func TestCreatePaymentInformation_UserNotFound(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)

	body := models.CreatePaymentInformationRequest{
		Type:              "BANK_ACCOUNT",
		BankAccountNumber: "123456789",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users/nonexistent/payment-information", bytes.NewReader(payload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestCreatePaymentInformation_InvalidBody(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)
	userID := createTestUser(t, h)

	req := httptest.NewRequest(http.MethodPost, "/users/"+userID+"/payment-information", bytes.NewReader([]byte("not json")))
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestCreatePaymentInformation_MissingClientID(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)

	body := models.CreatePaymentInformationRequest{Type: "BANK_ACCOUNT"}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users/user-1/payment-information", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestGetPaymentInformation_Success(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)
	userID := createTestUser(t, h)

	// Create payment information
	body := models.CreatePaymentInformationRequest{
		Type:              "BANK_ACCOUNT",
		BankAccountNumber: "999888777",
		BankAccountType:   "CHECKING",
		BankRoutingNumber: "021000021",
		AccountBankName:   "Test Bank",
	}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/users/"+userID+"/payment-information", bytes.NewReader(payload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("setup: expected 200, got %d", rr.Code)
	}

	var created models.PaymentInformation
	_ = json.NewDecoder(rr.Body).Decode(&created)

	// Get payment information
	req = httptest.NewRequest(http.MethodGet, "/users/"+userID+"/payment-information/"+created.ID, nil)
	ptiHeaders(req)
	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp models.PaymentInformation
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if resp.ID != created.ID {
		t.Errorf("expected ID %s, got %s", created.ID, resp.ID)
	}
	if resp.Type != "BANK_ACCOUNT" {
		t.Errorf("expected type BANK_ACCOUNT, got %s", resp.Type)
	}
	if resp.BankAccountNumber != "999888777" {
		t.Errorf("expected bank account number 999888777, got %s", resp.BankAccountNumber)
	}
	if resp.BankAccountType != "CHECKING" {
		t.Errorf("expected bank account type CHECKING, got %s", resp.BankAccountType)
	}
}

func TestGetPaymentInformation_NotFound(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)
	userID := createTestUser(t, h)

	req := httptest.NewRequest(http.MethodGet, "/users/"+userID+"/payment-information/nonexistent", nil)
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}
