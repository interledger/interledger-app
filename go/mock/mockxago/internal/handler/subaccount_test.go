package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/interledger/interledger-app/go/mock/mockxago/internal/models"
)

func setupAuthHandler(t *testing.T) *Handler {
	t.Helper()
	return setupTestHandler(t)
}

func issueToken(t *testing.T, h *Handler) string {
	t.Helper()
	req := models.LoginRequest{
		PolicyID: "test-policy",
		Fields: []models.FieldData{
			{FieldName: "publicKey", FieldValue: "test-public-key"},
			{FieldName: "secret", FieldValue: "test-secret"},
		},
	}
	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/xago/v1/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Login(w, r)

	var resp models.LoginResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	return resp.TokenValue
}

func authorizedRequest(method, url, token string, body []byte) *http.Request {
	r := httptest.NewRequest(method, url, bytes.NewReader(body))
	r.Header.Set("Authorization", "Bearer "+token)
	r.Header.Set("Content-Type", "application/json")
	return r
}

func TestCreateSubAccount_Success(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	req := models.CreateSubAccountRequest{
		WalletID:                  "wallet_abc",
		FirstName:                 "John",
		LastName:                  "Doe",
		Email:                     "john@example.com",
		MobileNumber:              "+27123456789",
		IdentityType:              "individual",
		IDNumber:                  "9001011234567",
		PhysicalAddress:           "123 Main St, Cape Town, SA",
		ThirdPartyVerificationURL: "https://app.withpersona.com/dashboard/inquiries/inq_123",
	}

	body, _ := json.Marshal(req)
	r := authorizedRequest(http.MethodPost, "/xago/v1/company/accounts", token, body)
	w := httptest.NewRecorder()

	h.CreateSubAccount(w, r)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.CreateSubAccountResponse
	json.NewDecoder(w.Body).Decode(&resp)

	// Validate accountID is a proper UUID
	_, err := uuid.Parse(resp.AccountID)
	assert.NoError(t, err, "accountID should be a valid UUID")
	assert.True(t, strings.HasPrefix(resp.DepositAddress, "r"))
	assert.True(t, resp.DepositTag > 0)
	assert.NotNil(t, resp.BankDepositDetails["ZAR"])
	assert.NotNil(t, resp.BankDepositDetails["USD"])
	assert.Equal(t, 2, len(resp.Beneficiaries))
	assert.True(t, strings.Contains(resp.Beneficiaries[0].DepositReference, "wallet_abc"))
}

func TestCreateSubAccount_MissingFirstName(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	req := models.CreateSubAccountRequest{
		LastName:                  "Doe",
		Email:                     "john@example.com",
		MobileNumber:              "+27123456789",
		IdentityType:              "individual",
		IDNumber:                  "9001011234567",
		PhysicalAddress:           "123 Main St",
		ThirdPartyVerificationURL: "https://app.withpersona.com/dashboard/inquiries/inq_123",
	}
	body, _ := json.Marshal(req)
	r := authorizedRequest(http.MethodPost, "/xago/v1/company/accounts", token, body)
	w := httptest.NewRecorder()

	h.CreateSubAccount(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Contains(t, resp.Message, "firstName is required")
}

func TestCreateSubAccount_MissingLastName(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	req := models.CreateSubAccountRequest{
		FirstName:                 "John",
		Email:                     "john@example.com",
		MobileNumber:              "+27123456789",
		IdentityType:              "individual",
		IDNumber:                  "9001011234567",
		PhysicalAddress:           "123 Main St",
		ThirdPartyVerificationURL: "https://app.withpersona.com/dashboard/inquiries/inq_123",
	}
	body, _ := json.Marshal(req)
	r := authorizedRequest(http.MethodPost, "/xago/v1/company/accounts", token, body)
	w := httptest.NewRecorder()

	h.CreateSubAccount(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateSubAccount_MissingEmail(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	req := models.CreateSubAccountRequest{
		FirstName:                 "John",
		LastName:                  "Doe",
		MobileNumber:              "+27123456789",
		IdentityType:              "individual",
		IDNumber:                  "9001011234567",
		PhysicalAddress:           "123 Main St",
		ThirdPartyVerificationURL: "https://app.withpersona.com/dashboard/inquiries/inq_123",
	}
	body, _ := json.Marshal(req)
	r := authorizedRequest(http.MethodPost, "/xago/v1/company/accounts", token, body)
	w := httptest.NewRecorder()

	h.CreateSubAccount(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateSubAccount_Success(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	// create first
	createReq := models.CreateSubAccountRequest{
		WalletID:                  "wallet_update",
		FirstName:                 "Jane",
		LastName:                  "Doe",
		Email:                     "jane@example.com",
		MobileNumber:              "+27123456789",
		IdentityType:              "individual",
		IDNumber:                  "9001011234567",
		PhysicalAddress:           "123 Main St",
		ThirdPartyVerificationURL: "https://app.withpersona.com/dashboard/inquiries/inq_456",
	}
	body, _ := json.Marshal(createReq)
	r := authorizedRequest(http.MethodPost, "/xago/v1/company/accounts", token, body)
	w := httptest.NewRecorder()
	h.CreateSubAccount(w, r)

	var created models.CreateSubAccountResponse
	json.NewDecoder(w.Body).Decode(&created)

	// update
	updateReq := models.UpdateSubAccountRequest{
		ThirdPartyVerificationURL: "https://app.withpersona.com/dashboard/inquiries/inq_999",
		IDNumber:                  "9001019999999",
		PhysicalAddress:           "999 Updated St",
	}
	updateBody, _ := json.Marshal(updateReq)
	updateURL := "/xago/v1/company/accounts/" + created.AccountID
	ur := authorizedRequest(http.MethodPut, updateURL, token, updateBody)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", created.AccountID)
	ur = ur.WithContext(context.WithValue(ur.Context(), chi.RouteCtxKey, rctx))
	uw := httptest.NewRecorder()

	h.UpdateSubAccount(uw, ur)
	assert.Equal(t, http.StatusOK, uw.Code)

	// verify stored value changed
	sub, err := h.store.GetSubAccount(context.Background(), created.AccountID)
	assert.NoError(t, err)
	assert.Equal(t, "https://app.withpersona.com/dashboard/inquiries/inq_999", sub.ThirdPartyVerificationURL)
	assert.Equal(t, "999 Updated St", sub.PhysicalAddress)
}

func TestUpdateSubAccount_InvalidID(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	updateReq := models.UpdateSubAccountRequest{PhysicalAddress: "123"}
	body, _ := json.Marshal(updateReq)
	r := authorizedRequest(http.MethodPut, "/xago/v1/company/accounts/invalid_id", token, body)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "invalid_id")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.UpdateSubAccount(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Contains(t, strings.ToLower(resp.Message), "invalid account id format")
}

func TestGetSubAccountByWallet_Isolated(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	// Pre-fetch currencies to compare bank details later
	curReq := httptest.NewRequest(http.MethodGet, "/xago/v1/currencies", nil)
	curW := httptest.NewRecorder()
	h.ListCurrencies(curW, curReq)
	var currenciesNested []map[string]interface{}
	json.NewDecoder(curW.Body).Decode(&currenciesNested)

	// Convert nested format to flat for test compatibility
	currencies := make([]map[string]string, 0, len(currenciesNested))
	for _, curr := range currenciesNested {
		if providers, ok := curr["bankingProviders"].([]interface{}); ok && len(providers) > 0 {
			if provider, ok := providers[0].(map[string]interface{}); ok {
				if fields, ok := provider["depositFields"].(map[string]interface{}); ok {
					currencies = append(currencies, map[string]string{
						"currencyId":    curr["currencyCode"].(string),
						"bankName":      fields["bankName"].(string),
						"accountNumber": fields["accountNumber"].(string),
						"branchCode":    fields["branchCode"].(string),
						"swiftBIC":      fields["swiftBIC"].(string),
					})
				}
			}
		}
	}

	create := func(wallet string) models.CreateSubAccountResponse {
		req := models.CreateSubAccountRequest{
			WalletID:                  wallet,
			FirstName:                 "User",
			LastName:                  wallet,
			Email:                     wallet + "@example.com",
			MobileNumber:              "+27111111111",
			IdentityType:              "individual",
			IDNumber:                  "9001011234567",
			PhysicalAddress:           "123 Main St",
			ThirdPartyVerificationURL: "https://app.withpersona.com/dashboard/inquiries/inq_111",
		}
		body, _ := json.Marshal(req)
		r := authorizedRequest(http.MethodPost, "/xago/v1/company/accounts", token, body)
		w := httptest.NewRecorder()
		h.CreateSubAccount(w, r)

		var resp models.CreateSubAccountResponse
		json.NewDecoder(w.Body).Decode(&resp)
		return resp
	}

	first := create("wallet_abc")
	_ = create("wallet_xyz")

	getReq := httptest.NewRequest(http.MethodGet, "/xago/v1/company/accounts?walletId=wallet_abc", nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	h.GetSubAccountByWallet(w, getReq)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.CreateSubAccountResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, first.AccountID, resp.AccountID)
	assert.True(t, strings.Contains(resp.Beneficiaries[0].DepositReference, "wallet_abc"))

	// Ensure bankDepositDetails match currencies endpoint
	for _, cur := range currencies {
		bankList, ok := resp.BankDepositDetails[cur["currencyId"]]
		assert.True(t, ok)
		assert.GreaterOrEqual(t, len(bankList), 1)
		assert.Equal(t, cur["bankName"], bankList[0].BankName)
		assert.Equal(t, cur["accountNumber"], bankList[0].AccountNumber)
		assert.Equal(t, cur["branchCode"], bankList[0].BranchCode)
		assert.Equal(t, cur["swiftBIC"], bankList[0].SwiftBIC)
	}
}
