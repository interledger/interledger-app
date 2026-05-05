package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gitlab.com/fynbos/mock/mockgatehub/internal/auth"
	"gitlab.com/fynbos/mock/mockgatehub/internal/logger"
	"gitlab.com/fynbos/mock/mockgatehub/internal/models"
	"gitlab.com/fynbos/mock/mockgatehub/internal/storage"
	"gitlab.com/fynbos/mock/mockgatehub/internal/utils"

	"go.uber.org/zap"
)

// Manager handles webhook delivery
type Manager struct {
	webhookURL            string
	webhookSecret         string
	httpClient            *http.Client
	queue                 *Queue // Redis-backed job queue
	store                 storage.Storage
	defaultOrganizationID string
}

// WebhookPayload represents the webhook request body (matches wallet-backend IWebhookData)
type WebhookPayload struct {
	UUID        string                 `json:"uuid"`        // Webhook UUID (required by controller check)
	Timestamp   string                 `json:"timestamp"`   // Milliseconds since epoch as string (e.g., "1768920404045")
	EventType   string                 `json:"event_type"`  // e.g., "core.deposit.completed"
	UserUUID    string                 `json:"user_uuid"`   // GateHub user UUID
	Environment string                 `json:"environment"` // "sandbox" or "production"
	Data        map[string]interface{} `json:"data"`        // Event-specific data (IDepositWebhookData, etc.)
}

// NewManager creates a new webhook manager
func NewManager(webhookURL, webhookSecret string, queue *Queue, store storage.Storage, defaultOrgID string) *Manager {
	logger.Info("initializing webhook manager",
		zap.String("url", webhookURL),
		zap.Int("secret_length", len(webhookSecret)),
		zap.String("default_org_id", defaultOrgID),
	)

	return &Manager{
		webhookURL:    webhookURL,
		webhookSecret: webhookSecret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		queue:                 queue,
		store:                 store,
		defaultOrganizationID: defaultOrgID,
	}
}

// SendAsync enqueues a webhook job for asynchronous delivery.
// offsetDelaySeconds adds extra seconds on top of the queue's minimum delay
// before the job becomes eligible for delivery.
func (m *Manager) SendAsync(eventType, userID string, data any, offsetDelaySeconds float64) {
	// Internal events (like 2FA callback tests) are always enqueued regardless of URL config
	isInternal := len(eventType) > 9 && eventType[:9] == "internal."
	if !isInternal {
		// For regular webhooks, check that a callback URL is configured
		callbackURL := m.resolveCallbackURL()
		if callbackURL == "" {
			logger.Info("skipping webhook send - no url configured", zap.String("event", eventType), zap.String("user", userID))
			return
		}
	}

	logger.Info("enqueueing webhook", zap.String("event", eventType), zap.String("user", userID), zap.Float64("offset_delay_seconds", offsetDelaySeconds))

	if m.queue == nil {
		logger.Warn("webhook queue not available, skipping enqueue", zap.String("event", eventType))
		return
	}

	ctx := context.Background()
	jobID, err := m.queue.Enqueue(ctx, eventType, userID, data, offsetDelaySeconds)
	if err != nil {
		logger.Error("failed to enqueue webhook", zap.Error(err))
		return
	}

	logger.Info("webhook enqueued successfully", zap.String("job_id", jobID))
}

// HasURL reports whether a webhook URL is configured (either via org config or global).
func (m *Manager) HasURL() bool {
	return m != nil && m.resolveCallbackURL() != ""
}

// ResolveCallbackURL determines where to send callbacks.
// Organization config takes priority, then falls back to global WEBHOOK_URL.
// Exported so handler code can reuse the same resolution logic.
// ResolveCallbackURL returns the callback URL, preferring org config over global WEBHOOK_URL.
func (m *Manager) ResolveCallbackURL() string {
	return m.resolveCallbackURL()
}

// ResolveOrgBaseURL returns the organization's apiBaseUrl strictly from org config.
// Returns empty string if no org config or no apiBaseUrl is set (no fallback to WEBHOOK_URL).
func (m *Manager) ResolveOrgBaseURL() string {
	if m.store != nil && m.defaultOrganizationID != "" {
		org, err := m.store.GetOrganization(m.defaultOrganizationID)
		if err == nil && org.APIBaseURL != "" {
			logger.Debug("resolved org base URL from organization config",
				zap.String("org_id", m.defaultOrganizationID),
				zap.String("api_base_url", org.APIBaseURL),
			)
			return strings.TrimRight(org.APIBaseURL, "/")
		}
	}
	return ""
}

// resolveCallbackURL determines where to send callbacks.
// Organization config takes priority, then falls back to global WEBHOOK_URL.
func (m *Manager) resolveCallbackURL() string {
	if m.store != nil && m.defaultOrganizationID != "" {
		org, err := m.store.GetOrganization(m.defaultOrganizationID)
		if err == nil && org.APIBaseURL != "" {
			logger.Debug("resolved callback URL from organization",
				zap.String("org_id", m.defaultOrganizationID),
				zap.String("api_base_url", org.APIBaseURL),
			)
			return org.APIBaseURL
		}
	}
	return m.webhookURL
}

// Send performs the actual HTTP webhook request (called by worker)
func (m *Manager) send(eventType, userID string, data any) error {
	// Special handling for 2FA callback test
	if eventType == "internal.2fa_callback_test" {
		return m.execute2FACallbackTest(data)
	}

	// Use the global WEBHOOK_URL for regular webhook delivery.
	// The org apiBaseUrl is a base URL (no path) meant for 2FA callbacks;
	// WEBHOOK_URL is the full URL including the webhook handler path.
	callbackURL := m.webhookURL
	if callbackURL == "" {
		logger.Warn("no webhook URL configured")
		return fmt.Errorf("no webhook URL available")
	}

	normalized := normalizeVerificationPayload(eventType, data)

	// Build payload - testnet wallet-backend expects timestamp as milliseconds string
	now := time.Now()
	payload := WebhookPayload{
		UUID:        utils.GenerateUUID(),               // Generate unique webhook UUID
		Timestamp:   fmt.Sprintf("%d", now.UnixMilli()), // Milliseconds since epoch as string
		EventType:   eventType,                          // e.g., "core.deposit.completed"
		UserUUID:    userID,                             // GateHub user UUID
		Environment: "sandbox",                          // Always sandbox for mockgatehub
		Data:        normalized,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", callbackURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")

	// Generate signature - GateHub expects SHA256(body) signed with the secret
	// The signature should use the entire JSON payload as the message
	signature := auth.GenerateGateHubWebhookSignature(string(body), m.webhookSecret)
	req.Header.Set("X-GH-Webhook-Signature", signature)

	logger.Info("sending webhook request",
		zap.String("url", callbackURL),
		zap.String("event_type", eventType),
		zap.String("user_id", userID),
		zap.String("signature", signature),
	)

	// Send request
	start := time.Now()
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(start)
	logger.Info("webhook response received",
		zap.Duration("duration", duration),
		zap.Int("status_code", resp.StatusCode),
	)

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d %s", resp.StatusCode, resp.Status)
	}

	return nil
}

func normalizeVerificationPayload(eventType string, data any) map[string]interface{} {
	converted := coerceToMap(data)

	switch eventType {
	case "id.verification.accepted", "id.verification.rejected", "id.verification.action_required":
		if _, ok := converted["gateway"]; !ok {
			converted["gateway"] = "paywiser"
		}
		if _, ok := converted["verified"]; !ok {
			short := "action_required"
			status := 0
			switch eventType {
			case "id.verification.accepted":
				short = "accepted"
				status = 1
			case "id.verification.rejected":
				short = "rejected"
				status = 2
			}
			converted["verified"] = map[string]interface{}{
				"short":  short,
				"status": status,
			}
		}
	}

	return converted
}

func coerceToMap(data any) map[string]interface{} {
	if data == nil {
		return map[string]interface{}{}
	}

	if typed, ok := data.(map[string]interface{}); ok {
		return typed
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return map[string]interface{}{}
	}

	var converted map[string]interface{}
	if err := json.Unmarshal(bytes, &converted); err != nil {
		return map[string]interface{}{}
	}

	return converted
}

// execute2FACallbackTest handles the internal 2FA callback test job
func (m *Manager) execute2FACallbackTest(data any) error {
	dataMap := coerceToMap(data)
	orgID, _ := dataMap["organization_id"].(string)
	testUserID, _ := dataMap["test_user_id"].(string)

	if m.store == nil {
		return fmt.Errorf("store not available for 2FA callback test")
	}

	org, err := m.store.GetOrganization(orgID)
	if err != nil {
		logger.Error("organization not found for 2fa test", zap.String("org_id", orgID), zap.Error(err))
		return err
	}

	if org.APIBaseURL == "" {
		logger.Warn("organization has no apiBaseUrl, skipping 2FA callback test", zap.String("org_id", orgID))
		return nil
	}

	logger.Info("executing 2FA callback test",
		zap.String("org_id", orgID),
		zap.String("type2fa", org.TwoFAType),
		zap.String("api_base_url", org.APIBaseURL),
	)

	switch org.TwoFAType {
	case "sms":
		return m.test2FASMSWorkflow(org, testUserID)
	case "totp":
		return m.test2FATOTPWorkflow(org, testUserID)
	default:
		return fmt.Errorf("unknown 2fa type: %s", org.TwoFAType)
	}
}

func (m *Manager) test2FASMSWorkflow(org *models.Organization, userID string) error {
	// Step 1: Send INITIATE callback
	initiatePayload := map[string]interface{}{
		"action": "INITIATE",
	}

	resp, err := m.post2FACallback(org, userID, initiatePayload)
	if err != nil {
		logger.Error("2FA INITIATE callback failed",
			zap.String("org_id", org.ID),
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return err
	}
	resp.Body.Close()

	logger.Info("2FA INITIATE callback succeeded",
		zap.String("org_id", org.ID),
		zap.String("user_id", userID),
		zap.Int("status_code", resp.StatusCode),
	)

	// Step 2: Send VERIFY callback (simulating user entering code)
	verifyPayload := map[string]interface{}{
		"action": "VERIFY",
		"code":   "123456",
	}

	resp, err = m.post2FACallback(org, userID, verifyPayload)
	if err != nil {
		logger.Error("2FA VERIFY callback failed",
			zap.String("org_id", org.ID),
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return err
	}
	resp.Body.Close()

	logger.Info("2FA VERIFY callback succeeded",
		zap.String("org_id", org.ID),
		zap.String("user_id", userID),
		zap.Int("status_code", resp.StatusCode),
	)

	return nil
}

func (m *Manager) test2FATOTPWorkflow(org *models.Organization, userID string) error {
	// TOTP only has verification (no INITIATE)
	verifyPayload := map[string]interface{}{
		"action": "VERIFY",
		"code":   "123456",
	}

	resp, err := m.post2FACallback(org, userID, verifyPayload)
	if err != nil {
		logger.Error("2FA TOTP callback failed",
			zap.String("org_id", org.ID),
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return err
	}
	resp.Body.Close()

	logger.Info("2FA TOTP callback succeeded",
		zap.String("org_id", org.ID),
		zap.String("user_id", userID),
		zap.Int("status_code", resp.StatusCode),
	)

	return nil
}

func (m *Manager) post2FACallback(org *models.Organization, userID string, payload map[string]interface{}) (*http.Response, error) {
	normalizedBase := strings.TrimRight(org.APIBaseURL, "/")
	endpoint := fmt.Sprintf("%s/v1/users/managed/%s/2fa", normalizedBase, userID)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal 2FA payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create 2FA request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return m.httpClient.Do(req)
}
