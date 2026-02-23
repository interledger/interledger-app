package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"gitlab.com/fynbos/mockxago/internal/logger"
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
	response := personaInquiryResponse{
		Data: personaInquiryData{
			Type: "inquiry",
			ID:   inquiryID,
			Attributes: personaInquiryAttributes{
				Status:      "created",
				ReferenceID: inquiryID,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			Meta: personaInquiryMeta{
				SessionToken: "mock-session-token",
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

	// Load and serve the KYC iframe template (same pattern as KYCIframe in kyc.go)
	possiblePaths := []string{
		"web/kyc-iframe.html",
		"./web/kyc-iframe.html",
		"../../web/kyc-iframe.html",
		"../../../web/kyc-iframe.html",
	}

	var templatePath string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			templatePath = path
			break
		}
	}

	if templatePath == "" {
		logger.Errorf("Could not find KYC iframe template, tried: %v", possiblePaths)
		h.sendError(w, http.StatusInternalServerError, "template_not_found", "KYC template not found")
		return
	}

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		logger.Errorf("Failed to parse KYC iframe template: %v", err)
		h.sendError(w, http.StatusInternalServerError, "template_error", "Failed to parse template")
		return
	}

	data := map[string]string{
		"Token":  r.URL.Query().Get("token"),
		"UserID": userID,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Frame-Options", "ALLOWALL") // Allow framing from any origin
	if err := tmpl.Execute(w, data); err != nil {
		logger.Errorf("Failed to execute KYC iframe template: %v", err)
		h.sendError(w, http.StatusInternalServerError, "template_execution_error", "Failed to execute template")
		return
	}
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

	logger.Infof("KYC submitted for inquiry %s: %s %s (DOB: %s)", inquiryID, firstName, lastName, dob)

	// Update inquiry status to approved
	// In a real implementation, this would trigger background verification
	// For testing, we mark it as approved immediately

	// Send webhook notification to backend
	go h.sendKYCWebhook(inquiryID)

	// Return success response
	h.sendJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"message": "Inquiry submitted successfully",
	})
}
