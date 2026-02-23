package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gitlab.com/fynbos/mockxago/internal/logger"
	"gitlab.com/fynbos/mockxago/internal/models"
)

// KYCIframe serves the KYC verification iframe
// GET /kyc/iframe?token={token}&user_id={wallet_id}
func (h *Handler) KYCIframe(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	userID := r.URL.Query().Get("user_id")
	walletID := r.URL.Query().Get("wallet_id")

	logger.Infof("Serving KYC iframe - token: %s, user_id: %s, wallet_id: %s", token, userID, walletID)

	// Try to find the KYC iframe template
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
		"Token":  token,
		"UserID": userID,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		logger.Errorf("Failed to execute KYC iframe template: %v", err)
		h.sendError(w, http.StatusInternalServerError, "template_execution_error", "Failed to execute template")
		return
	}
}

// KYCIframeSubmit handles KYC form submission
// POST /kyc/submit
func (h *Handler) KYCIframeSubmit(w http.ResponseWriter, r *http.Request) {
	logger.Infof("KYC iframe form submitted")

	// Parse form - handle both multipart and urlencoded
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		if err := r.ParseForm(); err != nil {
			logger.Errorf("Failed to parse form: %v", err)
			h.sendError(w, http.StatusBadRequest, "invalid_form", "Invalid form data")
			return
		}
	}

	// Extract form fields
	walletID := r.FormValue("user_id")
	_ = r.FormValue("token") // token may be used for webhook authentication in future
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	address := r.FormValue("address")
	_ = r.FormValue("city")    // city not currently stored in sub-account
	_ = r.FormValue("country") // country not currently stored in sub-account
	dob := r.FormValue("dob")

	logger.Infof("KYC submitted - wallet_id: %s, name: %s %s", walletID, firstName, lastName)

	if walletID == "" {
		logger.Errorf("wallet_id is required")
		h.sendError(w, http.StatusBadRequest, "missing_wallet_id", "Wallet ID is required")
		return
	}

	if firstName == "" || lastName == "" {
		logger.Errorf("name fields are required")
		h.sendError(w, http.StatusBadRequest, "missing_name", "First and last name are required")
		return
	}

	// Get or create sub-account for this wallet
	ctx := r.Context()
	subAccount, err := h.store.GetSubAccountByWalletID(ctx, walletID)
	if err != nil {
		logger.Infof("SubAccount not found for wallet %s, creating new one", walletID)
		// Create a new sub-account
		subAccount = &models.SubAccount{
			ID:        generateID(),
			WalletID:  walletID,
			AccountID: generateID(),
		}
	}

	// Update sub-account with KYC information
	subAccount.FirstName = firstName
	subAccount.LastName = lastName
	subAccount.PhysicalAddress = address

	// Parse date of birth
	if dob != "" {
		parts := strings.Split(dob, "-")
		if len(parts) == 3 {
			if year, err := strconv.Atoi(parts[0]); err == nil {
				logger.Infof("Parsed birth year: %d", year)
			}
		}
	}

	// Save or update sub-account
	if err := h.store.SaveSubAccount(ctx, subAccount); err != nil {
		logger.Errorf("Failed to save sub-account: %v", err)
		h.sendError(w, http.StatusInternalServerError, "save_error", "Failed to save account")
		return
	}

	logger.Infof("KYC submission successful for wallet %s", walletID)

	// Send webhook to backend notifying of KYC completion
	// This triggers the backend's Temporal workflow for Xago onboarding
	go h.sendPersonaInquiryApproved(walletID)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"accepted","message":"KYC verification submitted successfully"}`)
}

func (h *Handler) sendPersonaInquiryApproved(walletID string) {
	webhookURL := os.Getenv("PERSONA_WEBHOOK_URL")
	if webhookURL == "" {
		webhookURL = "http://backend:8080/webhooks/persona"
	}
	secret := os.Getenv("PERSONA_WEBHOOK_TOKEN")

	now := time.Now().UTC().Format(time.RFC3339)
	inquiry := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "inquiry",
			"id":   walletID,
			"attributes": map[string]interface{}{
				"status":       "approved",
				"reference-id": walletID,
				"created-at":   now,
				"updated-at":   now,
				"completed-at": now,
			},
		},
	}

	webhook := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "event",
			"id":   fmt.Sprintf("evt_%d", time.Now().UnixNano()),
			"attributes": map[string]interface{}{
				"name":       "inquiry.approved",
				"payload":    inquiry,
				"created-at": now,
			},
		},
	}

	body, err := json.Marshal(webhook)
	if err != nil {
		logger.Errorf("Failed to marshal persona webhook: %v", err)
		return
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewReader(body))
	if err != nil {
		logger.Errorf("Failed to create persona webhook request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Persona-Version", "2023-01-05")

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(append([]byte(timestamp+"."), body...))
	signature := hex.EncodeToString(mac.Sum(nil))
	req.Header.Set("Persona-Signature", fmt.Sprintf("t=%s,v1=%s", timestamp, signature))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("Persona webhook request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		logger.Errorf("Persona webhook failed with status %d: %s", resp.StatusCode, string(bodyBytes))
		return
	}

	logger.Infof("Persona inquiry approved webhook sent for wallet %s", walletID)
}

// sendWebhookWithSignature sends a POST request with HMAC signature
func sendWebhookWithSignature(webhookURL string, secret string, payload map[string]interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add HMAC signature if secret is provided
	if secret != "" {
		h := hmac.New(sha256.New, []byte(secret))
		h.Write(body)
		signature := hex.EncodeToString(h.Sum(nil))
		req.Header.Set("X-Signature", signature)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// generateID generates a random ID
func generateID() string {
	return uuid.NewString()
}
