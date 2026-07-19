package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/interledger/interledger-app/go/mock/mockxago/internal/models"
)

func TestAddBeneficiary_InvalidAccountID(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	req := models.AddBeneficiaryRequest{
		Name:          "John Doe",
		Scope:         "domestic",
		CurrencyCode:  "ZAR",
		AccountNumber: "123456789",
	}

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/xago/v1/accounts/invalid123/beneficiaries", bytes.NewReader(body))
	r.Header.Set("Authorization", "Bearer "+token)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "invalid123")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.AddBeneficiary(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAddBeneficiary_MissingName(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	// Create an account first
	subreq := models.CreateSubAccountRequest{
		WalletID:                  "wallet_noname",
		FirstName:                 "Test",
		LastName:                  "User",
		Email:                     "test@example.com",
		MobileNumber:              "+27123456789",
		IdentityType:              "individual",
		IDNumber:                  "9001011234567",
		PhysicalAddress:           "123 Test St",
		ThirdPartyVerificationURL: "https://example.com",
	}
	subbody, _ := json.Marshal(subreq)
	subr := authorizedRequest(http.MethodPost, "/xago/v1/company/accounts", token, subbody)
	subw := httptest.NewRecorder()
	h.CreateSubAccount(subw, subr)

	var subresp models.CreateSubAccountResponse
	json.NewDecoder(subw.Body).Decode(&subresp)

	// Now try to add beneficiary without name
	req := models.AddBeneficiaryRequest{
		Scope:         "domestic",
		CurrencyCode:  "ZAR",
		AccountNumber: "123456789",
	}

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/xago/v1/accounts/"+subresp.AccountID+"/beneficiaries", bytes.NewReader(body))
	r.Header.Set("Authorization", "Bearer "+token)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", subresp.AccountID)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.AddBeneficiary(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddBeneficiary_MissingAccountNumber(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	// Create an account first
	subreq := models.CreateSubAccountRequest{
		WalletID:                  "wallet_noaccount",
		FirstName:                 "Test",
		LastName:                  "User",
		Email:                     "test2@example.com",
		MobileNumber:              "+27123456789",
		IdentityType:              "individual",
		IDNumber:                  "9001011234567",
		PhysicalAddress:           "123 Test St",
		ThirdPartyVerificationURL: "https://example.com",
	}
	subbody, _ := json.Marshal(subreq)
	subr := authorizedRequest(http.MethodPost, "/xago/v1/company/accounts", token, subbody)
	subw := httptest.NewRecorder()
	h.CreateSubAccount(subw, subr)

	var subresp models.CreateSubAccountResponse
	json.NewDecoder(subw.Body).Decode(&subresp)

	// Now try to add beneficiary without account number
	req := models.AddBeneficiaryRequest{
		Name:         "John Doe",
		Scope:        "domestic",
		CurrencyCode: "ZAR",
	}

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/xago/v1/accounts/"+subresp.AccountID+"/beneficiaries", bytes.NewReader(body))
	r.Header.Set("Authorization", "Bearer "+token)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", subresp.AccountID)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.AddBeneficiary(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListBeneficiaries_InvalidAccountID(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	r := httptest.NewRequest(http.MethodGet, "/xago/v1/accounts/invalid456/beneficiaries", nil)
	r.Header.Set("Authorization", "Bearer "+token)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "invalid456")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.ListBeneficiaries(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAddBeneficiaryGlobal_MissingToken(t *testing.T) {
	h := setupAuthHandler(t)

	req := models.AddBeneficiaryRequest{
		Name:          "Global Test",
		Scope:         "international",
		CurrencyCode:  "USD",
		AccountNumber: "999888777",
	}

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/xago/v1/beneficiaries", bytes.NewReader(body))

	w := httptest.NewRecorder()
	h.AuthMiddleware(http.HandlerFunc(h.AddBeneficiaryGlobal)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestListBeneficiariesGlobal_MissingToken(t *testing.T) {
	h := setupAuthHandler(t)

	r := httptest.NewRequest(http.MethodGet, "/xago/v1/beneficiaries", nil)

	w := httptest.NewRecorder()
	h.AuthMiddleware(http.HandlerFunc(h.ListBeneficiariesGlobal)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAddBeneficiaryGlobal_Success(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	// Create a sub-account first
	subreq := models.CreateSubAccountRequest{
		WalletID:                  "wallet_global_ben",
		FirstName:                 "Global",
		LastName:                  "User",
		Email:                     "global@example.com",
		MobileNumber:              "+27123456789",
		IdentityType:              "individual",
		IDNumber:                  "9001011234567",
		PhysicalAddress:           "456 Global St",
		ThirdPartyVerificationURL: "https://example.com",
	}
	subbody, _ := json.Marshal(subreq)
	subr := authorizedRequest(http.MethodPost, "/xago/v1/company/accounts", token, subbody)
	subw := httptest.NewRecorder()
	h.CreateSubAccount(subw, subr)

	var subresp models.CreateSubAccountResponse
	json.NewDecoder(subw.Body).Decode(&subresp)

	// Associate token with account
	h.store.SaveTokenAccount(context.Background(), token, subresp.AccountID)

	// Add beneficiary via global endpoint
	req := models.AddBeneficiaryRequest{
		Name:          "Global Beneficiary",
		Scope:         "international",
		CurrencyCode:  "USD",
		AccountNumber: "999888777",
		BankName:      "Global Bank",
		Reference:     "Payment",
	}

	body, _ := json.Marshal(req)
	r := authorizedRequest(http.MethodPost, "/xago/v1/beneficiaries", token, body)

	w := httptest.NewRecorder()
	h.AuthMiddleware(http.HandlerFunc(h.AddBeneficiaryGlobal)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.BeneficiaryItem
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "Global Beneficiary", resp.Name)
	assert.Equal(t, "999888777", resp.AccountNumber)
	assert.Equal(t, "USD", resp.CurrencyCode)
	assert.Equal(t, "pending", resp.Status)
}

func TestAddBeneficiaryGlobal_InvalidToken(t *testing.T) {
	h := setupAuthHandler(t)

	req := models.AddBeneficiaryRequest{
		Name:          "Test",
		AccountNumber: "123456",
		CurrencyCode:  "ZAR",
	}

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/xago/v1/beneficiaries", bytes.NewReader(body))
	r.Header.Set("Authorization", "Bearer invalid-token-xyz")

	w := httptest.NewRecorder()
	h.AuthMiddleware(http.HandlerFunc(h.AddBeneficiaryGlobal)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAddBeneficiaryGlobal_NoAccountAssociated(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	// Don't associate token with any account

	req := models.AddBeneficiaryRequest{
		Name:          "Orphan Beneficiary",
		AccountNumber: "111222333",
		CurrencyCode:  "EUR",
	}

	body, _ := json.Marshal(req)
	r := authorizedRequest(http.MethodPost, "/xago/v1/beneficiaries", token, body)

	w := httptest.NewRecorder()
	h.AuthMiddleware(http.HandlerFunc(h.AddBeneficiaryGlobal)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddBeneficiaryGlobal_MissingName(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	// Create sub-account and associate
	subreq := models.CreateSubAccountRequest{
		WalletID:                  "wallet_global_noname",
		FirstName:                 "Test",
		LastName:                  "User",
		Email:                     "noname@example.com",
		MobileNumber:              "+27123456789",
		IdentityType:              "individual",
		IDNumber:                  "9001011234567",
		PhysicalAddress:           "123 Test St",
		ThirdPartyVerificationURL: "https://example.com",
	}
	subbody, _ := json.Marshal(subreq)
	subr := authorizedRequest(http.MethodPost, "/xago/v1/company/accounts", token, subbody)
	subw := httptest.NewRecorder()
	h.CreateSubAccount(subw, subr)

	var subresp models.CreateSubAccountResponse
	json.NewDecoder(subw.Body).Decode(&subresp)
	h.store.SaveTokenAccount(context.Background(), token, subresp.AccountID)

	// Try to add beneficiary without name
	req := models.AddBeneficiaryRequest{
		AccountNumber: "444555666",
		CurrencyCode:  "GBP",
	}

	body, _ := json.Marshal(req)
	r := authorizedRequest(http.MethodPost, "/xago/v1/beneficiaries", token, body)

	w := httptest.NewRecorder()
	h.AuthMiddleware(http.HandlerFunc(h.AddBeneficiaryGlobal)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListBeneficiariesGlobal_Success(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	// Create sub-account
	subreq := models.CreateSubAccountRequest{
		WalletID:                  "wallet_list_global",
		FirstName:                 "List",
		LastName:                  "Test",
		Email:                     "listglobal@example.com",
		MobileNumber:              "+27123456789",
		IdentityType:              "individual",
		IDNumber:                  "9001011234567",
		PhysicalAddress:           "789 List St",
		ThirdPartyVerificationURL: "https://example.com",
	}
	subbody, _ := json.Marshal(subreq)
	subr := authorizedRequest(http.MethodPost, "/xago/v1/company/accounts", token, subbody)
	subw := httptest.NewRecorder()
	h.CreateSubAccount(subw, subr)

	var subresp models.CreateSubAccountResponse
	json.NewDecoder(subw.Body).Decode(&subresp)
	h.store.SaveTokenAccount(context.Background(), token, subresp.AccountID)

	// Add some beneficiaries
	for i := 1; i <= 3; i++ {
		ben := &models.Beneficiary{
			ID:            fmt.Sprintf("ben-global-%d", i),
			AccountID:     subresp.AccountID,
			WalletID:      "wallet_list_global",
			Name:          fmt.Sprintf("Beneficiary %d", i),
			AccountNumber: fmt.Sprintf("ACC%d", i),
			Currency:      "ZAR",
			Status:        "approved",
		}
		h.store.SaveBeneficiary(context.Background(), ben)
	}

	// List via global endpoint
	r := authorizedRequest(http.MethodGet, "/xago/v1/beneficiaries", token, nil)

	w := httptest.NewRecorder()
	h.AuthMiddleware(http.HandlerFunc(h.ListBeneficiariesGlobal)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.ListBeneficiariesResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Len(t, resp.Data, 3)
	assert.Equal(t, 3, resp.Pagination.Total)
}

func TestListBeneficiariesGlobal_WithPagination(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	// Create sub-account
	subreq := models.CreateSubAccountRequest{
		WalletID:                  "wallet_paginate_global",
		FirstName:                 "Page",
		LastName:                  "Test",
		Email:                     "pageglobal@example.com",
		MobileNumber:              "+27123456789",
		IdentityType:              "individual",
		IDNumber:                  "9001011234567",
		PhysicalAddress:           "111 Page St",
		ThirdPartyVerificationURL: "https://example.com",
	}
	subbody, _ := json.Marshal(subreq)
	subr := authorizedRequest(http.MethodPost, "/xago/v1/company/accounts", token, subbody)
	subw := httptest.NewRecorder()
	h.CreateSubAccount(subw, subr)

	var subresp models.CreateSubAccountResponse
	json.NewDecoder(subw.Body).Decode(&subresp)
	h.store.SaveTokenAccount(context.Background(), token, subresp.AccountID)

	// Add 15 beneficiaries
	for i := 1; i <= 15; i++ {
		ben := &models.Beneficiary{
			ID:            fmt.Sprintf("ben-page-%d", i),
			AccountID:     subresp.AccountID,
			WalletID:      "wallet_paginate_global",
			Name:          fmt.Sprintf("Page Ben %d", i),
			AccountNumber: fmt.Sprintf("PAGE%d", i),
			Currency:      "USD",
			Status:        "approved",
		}
		h.store.SaveBeneficiary(context.Background(), ben)
	}

	// Request page 2 with limit 5
	r := authorizedRequest(http.MethodGet, "/xago/v1/beneficiaries?limit=5&page=2", token, nil)

	w := httptest.NewRecorder()
	h.AuthMiddleware(http.HandlerFunc(h.ListBeneficiariesGlobal)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.ListBeneficiariesResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Len(t, resp.Data, 5)
	assert.Equal(t, 5, resp.Pagination.Limit)
	assert.Equal(t, 2, resp.Pagination.Page)
	assert.Equal(t, 15, resp.Pagination.Total)
	assert.Equal(t, 3, resp.Pagination.NumberOfPages) // ceil(15/5)
}

func TestListBeneficiariesGlobal_InvalidToken(t *testing.T) {
	h := setupAuthHandler(t)

	r := httptest.NewRequest(http.MethodGet, "/xago/v1/beneficiaries", nil)
	r.Header.Set("Authorization", "Bearer bad-token")

	w := httptest.NewRecorder()
	h.AuthMiddleware(http.HandlerFunc(h.ListBeneficiariesGlobal)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestListBeneficiariesGlobal_NoAccountAssociated(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	// Don't associate token with account

	r := authorizedRequest(http.MethodGet, "/xago/v1/beneficiaries", token, nil)

	w := httptest.NewRecorder()
	h.AuthMiddleware(http.HandlerFunc(h.ListBeneficiariesGlobal)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
