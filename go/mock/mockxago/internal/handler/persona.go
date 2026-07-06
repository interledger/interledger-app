package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/interledger/interledger-app/go/mock/mockxago/internal/logger"
	"github.com/interledger/interledger-app/go/mock/mockxago/web"
)

type personaCreateInquiryRequest struct {
	Data struct {
		Attributes struct {
			ReferenceID string `json:"reference-id,omitempty"`
			CountryCode string `json:"country-code,omitempty"`
		} `json:"attributes"`
	} `json:"data"`
}

type personaInquiryResponse struct {
	Data personaInquiryData `json:"data"`
}

type personaInquiryData struct {
	Type       string                   `json:"type"`
	ID         string                   `json:"id"`
	Attributes personaInquiryAttributes `json:"attributes"`
	Meta       personaInquiryMeta       `json:"meta"`
}

type personaInquiryAttributes struct {
	Status      string `json:"status"`
	ReferenceID string `json:"reference-id"`
	CreatedAt   string `json:"created-at"`
	UpdatedAt   string `json:"updated-at"`
	CompletedAt string `json:"completed-at,omitempty"`
}

type personaInquiryMeta struct {
	SessionToken string `json:"session-token"`
}

// PersonaCreateInquiry handles POST /inquiries - Create or retrieve inquiry
// Compatible with Persona SDK API
func (h *Handler) PersonaCreateInquiry(w http.ResponseWriter, r *http.Request) {
	var req personaCreateInquiryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorf("Failed to decode inquiry request: %v", err)
		h.sendError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	refID := strings.TrimSpace(req.Data.Attributes.ReferenceID)
	inquiryID := refID
	if inquiryID == "" {
		inquiryID = fmt.Sprintf("inq_%d", time.Now().Unix())
	}

	now := time.Now().UTC().Format(time.RFC3339)
	logger.Infof("Created Persona inquiry: %s (reference-id=%s)", inquiryID, refID)

	response := personaInquiryResponse{
		Data: personaInquiryData{
			Type: "inquiry",
			ID:   inquiryID,
			Attributes: personaInquiryAttributes{
				Status:      "created",
				ReferenceID: refID,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			Meta: personaInquiryMeta{
				SessionToken: "mock-session-token",
			},
		},
	}

	h.sendJSON(w, http.StatusCreated, response)
}

// PersonaGetInquiry handles GET /inquiries/{id} - Get inquiry status
// Compatible with Persona SDK API
func (h *Handler) PersonaGetInquiry(w http.ResponseWriter, r *http.Request) {
	inquiryID := chi.URLParam(r, "inquiryId")
	if inquiryID == "" {
		h.sendError(w, http.StatusBadRequest, "missing_inquiry_id", "Inquiry ID is required")
		return
	}

	logger.Infof("Getting Persona inquiry: %s", inquiryID)

	now := time.Now().UTC().Format(time.RFC3339)

	// Return inquiry as approved with included session data (IP address)
	// The backend's dirty-sync path needs inquiry-session data for IP address
	response := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "inquiry",
			"id":   inquiryID,
			"attributes": map[string]interface{}{
				"status":       "approved",
				"reference-id": inquiryID,
				"created-at":   now,
				"updated-at":   now,
				"completed-at": now,
			},
			"meta": map[string]interface{}{
				"session-token": "mock-session-token",
			},
		},
		"included": []map[string]interface{}{
			{
				"type": "inquiry-session",
				"attributes": map[string]interface{}{
					"ip-address": "127.0.0.1",
				},
			},
		},
	}

	h.sendJSON(w, http.StatusOK, response)
}

// PersonaGetInquiryIframe handles GET /inquiries/{id}/iframe - Serves the KYC iframe
func (h *Handler) PersonaGetInquiryIframe(w http.ResponseWriter, r *http.Request) {
	inquiryID := chi.URLParam(r, "inquiryId")
	if inquiryID == "" {
		h.sendError(w, http.StatusBadRequest, "missing_inquiry_id", "Inquiry ID is required")
		return
	}

	logger.Infof("Serving Persona iframe for inquiry: %s", inquiryID)

	// Use inquiry ID as the wallet identifier for mock Persona flows.
	userID := inquiryID

	tmpl, err := template.ParseFS(web.Assets, "kyc-iframe.html")
	if err != nil {
		logger.Errorf("Failed to parse KYC iframe template: %v", err)
		h.sendError(w, http.StatusInternalServerError, "template_error", "Failed to parse template")
		return
	}

	data := map[string]string{
		"Token":      r.URL.Query().Get("token"),
		"UserID":     userID,
		"SubmitPath": fmt.Sprintf("/v1/inquiries/%s/submit", inquiryID),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Frame-Options", "ALLOWALL") // Allow framing from any origin
	if err := tmpl.Execute(w, data); err != nil {
		logger.Errorf("Failed to execute KYC iframe template: %v", err)
		h.sendError(w, http.StatusInternalServerError, "template_execution_error", "Failed to execute template")
		return
	}
}

// PersonaSDK handles GET /persona-sdk.js - Serves a Persona-compatible SDK mock
func (h *Handler) PersonaSDK(w http.ResponseWriter, r *http.Request) {
	data, err := web.Assets.ReadFile("persona-sdk.js")
	if err != nil {
		logger.Errorf("Failed to read embedded persona-sdk.js: %v", err)
		h.sendError(w, http.StatusInternalServerError, "script_not_found", "Persona SDK script not found")
		return
	}

	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Write(data)
}

// PersonaInquirySubmit handles POST /inquiries/{id}/submit - Form submission callback
func (h *Handler) PersonaInquirySubmit(w http.ResponseWriter, r *http.Request) {
	inquiryID := chi.URLParam(r, "inquiryId")
	if inquiryID == "" {
		h.sendError(w, http.StatusBadRequest, "missing_inquiry_id", "Inquiry ID is required")
		return
	}

	logger.Infof("Submitting Persona inquiry: %s", inquiryID)

	// Parse form submission
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		if err := r.ParseForm(); err != nil {
			logger.Errorf("Failed to parse form: %v", err)
			h.sendError(w, http.StatusBadRequest, "invalid_form", "Invalid form data")
			return
		}
	}

	// Extract form fields
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	dob := r.FormValue("dob")
	address := r.FormValue("address")

	logger.Infof("KYC submitted for inquiry %s: %s %s (DOB: %s)", inquiryID, firstName, lastName, dob)

	if firstName == "" || lastName == "" {
		h.sendError(w, http.StatusBadRequest, "missing_name", "First and last name are required")
		return
	}

	if dob == "" {
		h.sendError(w, http.StatusBadRequest, "missing_dob", "Date of birth is required")
		return
	}

	if _, err := time.Parse("2006-01-02", dob); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_dob", "Date of birth must be in YYYY-MM-DD format")
		return
	}

	if address == "" {
		address = "123 Main Street"
	}

	if err := h.saveKYCAndFireWebhooks(r.Context(), inquiryID, firstName, lastName, dob, address); err != nil {
		logger.Errorf("Failed to save KYC for inquiry %s: %v", inquiryID, err)
		h.sendError(w, http.StatusInternalServerError, "save_error", "Failed to save account")
		return
	}

	// Return success response
	h.sendJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"message": "Inquiry submitted successfully",
	})
}

// PersonaGetAccount handles GET /v1/accounts/{accountId} - Get account details
// Compatible with Persona SDK API. The backend calls this after receiving
// an account.tag-added webhook to fetch full account attributes.
func (h *Handler) PersonaGetAccount(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "accountId")
	if accountID == "" {
		h.sendError(w, http.StatusBadRequest, "missing_account_id", "Account ID is required")
		return
	}

	logger.Infof("Getting Persona account: %s", accountID)

	// Resolve wallet and profile data from stored sub-account.
	// For Xago-style records, accountID is a distinct sub-account ID.
	// For older mock/persona flows, accountID may equal walletID.
	var walletID, nameFirst, nameLast, dateOfBirth, physicalAddress string
	subAccount, err := h.store.GetSubAccount(r.Context(), accountID)
	if err != nil {
		subAccount, err = h.store.GetSubAccountByWalletID(r.Context(), accountID)
	}
	if err != nil || subAccount == nil {
		logger.Errorf("account not found: %s", accountID)
		h.sendError(w, http.StatusNotFound, "account_not_found", "Account not found")
		return
	}
	walletID = subAccount.WalletID
	nameFirst = subAccount.FirstName
	nameLast = subAccount.LastName
	dateOfBirth = subAccount.DateOfBirth
	physicalAddress = subAccount.PhysicalAddress
	if dateOfBirth == "" {
		logger.Errorf("date of birth missing for wallet %s", walletID)
		h.sendError(w, http.StatusBadRequest, "missing_dob", "Date of birth not found for account")
		return
	}
	if physicalAddress == "" {
		physicalAddress = "123 Main Street"
	}

	// Build identification numbers with a valid ZA ID
	// Format: YYMMDDSSSSCZD (13 digits with Luhn check)
	identificationNumbers := map[string]interface{}{
		"pp": []map[string]interface{}{
			{
				"issuing-country":       "ZA",
				"identification-class":  "dl",
				"identification-number": "8406270000087",
			},
		},
	}

	response := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "account",
			"id":   accountID,
			"attributes": map[string]interface{}{
				"reference-id":           walletID,
				"name-first":             nameFirst,
				"name-last":              nameLast,
				"country-code":           "ZA",
				"birthdate":              dateOfBirth,
				"address-street-1":       physicalAddress,
				"address-city":           "Johannesburg",
				"address-postal-code":    "2000",
				"tags":                   []string{"DIRTY", "STATUS:KYC-LEVEL:1"},
				"identification-numbers": identificationNumbers,
			},
		},
	}

	h.sendJSON(w, http.StatusOK, response)
}

// PersonaRemoveTag handles POST /v1/accounts/{accountId}/remove-tag
// Compatible with Persona SDK API. Called by the backend after dirty-sync
// to remove the DIRTY tag from the account.
func (h *Handler) PersonaRemoveTag(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "accountId")
	if accountID == "" {
		h.sendError(w, http.StatusBadRequest, "missing_account_id", "Account ID is required")
		return
	}

	logger.Infof("Removing tag from Persona account: %s", accountID)

	// Return the account without the DIRTY tag (just STATUS:KYC-LEVEL:1)
	response := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "account",
			"id":   accountID,
			"attributes": map[string]interface{}{
				"reference-id": accountID,
				"tags":         []string{"STATUS:KYC-LEVEL:1"},
			},
		},
	}

	h.sendJSON(w, http.StatusOK, response)
}
