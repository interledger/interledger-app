//go:build e2e
// +build e2e

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type webhookCapture struct {
	mu      sync.Mutex
	events  []capturedWebhook
	started bool
}

type capturedWebhook struct {
	payload   map[string]interface{}
	encrypted bool
}

var capture webhookCapture

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	payload, encrypted, err := parseWebhookPayload(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_webhook_payload"}`))
		return
	}

	capture.mu.Lock()
	capture.events = append(capture.events, capturedWebhook{payload: payload, encrypted: encrypted})
	capture.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func ensureWebhookServer() {
	capture.mu.Lock()
	alreadyStarted := capture.started
	capture.mu.Unlock()
	if alreadyStarted {
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/webhooks/pti", webhookHandler)
	server := &http.Server{Addr: ":24100", Handler: mux}

	capture.mu.Lock()
	capture.started = true
	capture.mu.Unlock()

	go func() {
		_ = server.ListenAndServe()
	}()

	// Give the listener a moment to bind before tests continue.
	time.Sleep(100 * time.Millisecond)
}

func resetWebhookCapture() {
	capture.mu.Lock()
	defer capture.mu.Unlock()
	capture.events = nil
}

func latestWebhook() capturedWebhook {
	capture.mu.Lock()
	defer capture.mu.Unlock()
	if len(capture.events) == 0 {
		return capturedWebhook{}
	}
	return capture.events[len(capture.events)-1]
}

func waitForWebhook(timeout time.Duration, predicate func(capturedWebhook) bool) (capturedWebhook, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		capture.mu.Lock()
		events := make([]capturedWebhook, len(capture.events))
		copy(events, capture.events)
		capture.mu.Unlock()

		for _, evt := range events {
			if predicate(evt) {
				return evt, nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return capturedWebhook{}, fmt.Errorf("timed out waiting for webhook")
}

func (tc *TestContext) webhookDeliveryIsConfiguredToBackend(path string) error {
	if path != "/webhooks/pti" {
		return fmt.Errorf("unsupported webhook path %q", path)
	}
	ensureWebhookServer()
	resetWebhookCapture()
	return nil
}

func (tc *TestContext) anExistingPTIUserAssessmentInState(state string) error {
	if err := tc.anExistingPTIUser(); err != nil {
		return err
	}
	if !strings.EqualFold(state, "ACCEPTED") {
		return fmt.Errorf("only ACCEPTED state is supported in phase 4 tests")
	}
	if _, err := tc.ptiRequest("POST", "/users/assessments", buildAssessmentPayload(tc.lastUserID), true); err != nil {
		return err
	}
	if err := tc.responseStatusShouldBe(200); err != nil {
		return err
	}
	return tc.responseShouldIncludeAssessmentRequestID()
}

func (tc *TestContext) mockptiProcessesTheAssessmentCompletion() error {
	evt, err := waitForWebhook(5*time.Second, func(evt capturedWebhook) bool {
		return asString(evt.payload["resourceType"]) == "USER_ASSESSMENT"
	})
	if err != nil {
		return err
	}
	tc.lastWebhook = evt.payload
	tc.lastWebhookEncrypted = evt.encrypted
	return nil
}

func (tc *TestContext) anExistingPTIDepositTransactionInState(state string) error {
	if state != "PENDING" {
		return fmt.Errorf("expected initial state PENDING, got %s", state)
	}
	return tc.anExistingPTIUserWithUSDWalletAndBankAccount()
}

func (tc *TestContext) anExistingPTIWithdrawalTransactionInState(state string) error {
	if state != "PENDING" {
		return fmt.Errorf("expected initial state PENDING, got %s", state)
	}
	return tc.anExistingPTIUserWithUSDWalletAndBankAccount()
}

func (tc *TestContext) mockptiTransitionsTheTransactionTo(status string) error {
	var headers map[string]string
	if status == "REFUSED" {
		headers = map[string]string{"x-pti-scenario-id": "REFUSE"}
	}

	path := "/transactions/deposits"
	payload := map[string]interface{}{
		"initiator": map[string]interface{}{"id": tc.lastUserID, "type": "PERSON"},
		"sourceMethod": map[string]interface{}{
			"currency":           "USD",
			"paymentMethodType":  "ACH",
			"paymentInformation": map[string]interface{}{"type": "BANK_ACCOUNT", "id": tc.lastPaymentInformationID},
		},
		"destinationMethod": map[string]interface{}{
			"paymentMethodType":  "WALLET",
			"paymentInformation": map[string]interface{}{"type": "WALLET", "id": tc.lastWalletID},
		},
		"amount": 100.0,
		"type":   "DEPOSIT",
	}

	if status == "REFUSED" {
		path = "/transactions/withdrawals"
		payload = map[string]interface{}{
			"initiator": map[string]interface{}{"id": tc.lastUserID, "type": "PERSON"},
			"sourceMethod": map[string]interface{}{
				"paymentMethodType":  "WALLET",
				"paymentInformation": map[string]interface{}{"type": "WALLET", "id": tc.lastWalletID},
			},
			"destinationMethod": map[string]interface{}{
				"currency":           "USD",
				"paymentMethodType":  "ACH",
				"paymentInformation": map[string]interface{}{"type": "BANK_ACCOUNT", "id": tc.lastPaymentInformationID},
			},
			"amount": 50.0,
			"type":   "WITHDRAWAL",
		}
	}

	if _, err := tc.ptiRequest("POST", path, payload, true, headers); err != nil {
		return err
	}
	if err := tc.responseStatusShouldBe(200); err != nil {
		return err
	}
	if err := tc.responseShouldIncludeTransactionRequestID(); err != nil {
		return err
	}

	evt, err := waitForWebhook(5*time.Second, func(evt capturedWebhook) bool {
		if asString(evt.payload["resourceType"]) != "TRANSACTION_STATUS" {
			return false
		}
		if asString(evt.payload["requestId"]) != tc.lastTransactionRequestID {
			return false
		}
		return asString(evt.payload["status"]) == status
	})
	if err != nil {
		return err
	}
	tc.lastWebhook = evt.payload
	tc.lastWebhookEncrypted = evt.encrypted
	return nil
}

func (tc *TestContext) theWebhookPayloadShouldBeSignedAndEncrypted() error {
	if tc.lastWebhook == nil {
		return fmt.Errorf("no webhook captured")
	}
	if !tc.lastWebhookEncrypted {
		return fmt.Errorf("expected encrypted webhook payload")
	}
	return nil
}

func (tc *TestContext) aWebhookShouldBeDeliveredWithResourceType(resourceType string) error {
	if tc.lastWebhook == nil {
		return fmt.Errorf("no webhook captured")
	}
	if asString(tc.lastWebhook["resourceType"]) != resourceType {
		return fmt.Errorf("expected resourceType %q, got %q", resourceType, asString(tc.lastWebhook["resourceType"]))
	}
	return nil
}

func (tc *TestContext) theWebhookPayloadShouldIncludeUserIDAndRequestID() error {
	if tc.lastWebhook == nil {
		return fmt.Errorf("no webhook captured")
	}
	if asString(tc.lastWebhook["userId"]) == "" || asString(tc.lastWebhook["requestId"]) == "" {
		b, _ := json.Marshal(tc.lastWebhook)
		return fmt.Errorf("expected both userId and requestId in payload: %s", string(b))
	}
	return nil
}

func (tc *TestContext) theWebhookPayloadShouldIncludeTransactionType(txType string) error {
	if tc.lastWebhook == nil {
		return fmt.Errorf("no webhook captured")
	}
	if asString(tc.lastWebhook["transactionType"]) != txType {
		return fmt.Errorf("expected transactionType %q, got %q", txType, asString(tc.lastWebhook["transactionType"]))
	}
	return nil
}

func (tc *TestContext) theWebhookPayloadShouldIncludeStatus(status string) error {
	if tc.lastWebhook == nil {
		return fmt.Errorf("no webhook captured")
	}
	if asString(tc.lastWebhook["status"]) != status {
		b, _ := json.Marshal(tc.lastWebhook)
		return fmt.Errorf("expected status %q, got %q payload=%s", status, asString(tc.lastWebhook["status"]), string(b))
	}
	return nil
}

func asString(v interface{}) string {
	if v == nil {
		return ""
	}
	s, ok := v.(string)
	if ok {
		return s
	}
	// JSON numbers may be float64, etc. Use compact conversion where needed.
	buf := bytes.Buffer{}
	_ = json.NewEncoder(&buf).Encode(v)
	return strings.TrimSpace(buf.String())
}
