package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

const (
	kycStatusUnknown           = 0
	kycStatusPending           = 1
	kycStatusDocumentsRequired = 2
	kycWebhookPollInterval     = 500 * time.Millisecond
	kycWebhookPollTimeout      = 30 * time.Second
)

// triggerGatehubKYCWebhook POSTs to MockGatehub's /ui/actions/kyc endpoint to fire a signed KYC webhook.
func (sc *E2EContext) triggerGatehubKYCWebhook(gatehubUserID, eventType, gateway string) error {
	baseURL := "https://mockgatehub.interledger.test"
	if sc.mockgatehubBaseURL != "" {
		baseURL = sc.mockgatehubBaseURL
	}

	status, err := gatehubKYCEventToUIStatus(eventType)
	if err != nil {
		return err
	}
	if gateway == "" {
		gateway = "paywiser-eu-sandbox"
	}

	formData := url.Values{
		"userID":  {gatehubUserID},
		"gateway": {gateway},
		"status":  {status},
	}

	reqURL := baseURL + "/ui/actions/kyc"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, reqURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("triggerGatehubKYCWebhook: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := mockgatehubHTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("triggerGatehubKYCWebhook: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther && resp.StatusCode != http.StatusFound {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("triggerGatehubKYCWebhook: unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	debugPrintf("   ✓ Triggered GateHub KYC webhook %s for user %s\n", eventType, gatehubUserID)
	return nil
}

func gatehubKYCEventToUIStatus(eventType string) (string, error) {
	switch eventType {
	case "id.verification.action_required":
		return "action_required", nil
	case "id.verification.resubmission":
		return "resubmission", nil
	case "id.document_notice.expired":
		return "document_expired", nil
	case "id.document_notice.warning":
		return "document_warning", nil
	case "id.verification.accepted":
		return "accepted", nil
	case "id.verification.rejected":
		return "rejected", nil
	default:
		return "", fmt.Errorf("unsupported GateHub KYC webhook event: %s", eventType)
	}
}

func (sc *E2EContext) iTriggerGateHubKYCWebhookForMyself(eventType string) error {
	debugPrintf("\n📨 Triggering GateHub KYC webhook %s...\n", eventType)

	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return fmt.Errorf("trigger KYC webhook: %w", err)
	}

	gatehubUserID, err := sc.getGatehubUserIDByEmail(email)
	if err != nil {
		return fmt.Errorf("trigger KYC webhook: %w", err)
	}

	// Only fire the webhook here; the resulting KYC status is asserted by the
	// explicit "my KYC status should be ..." step, which polls with a timeout.
	return sc.triggerGatehubKYCWebhook(gatehubUserID, eventType, "")
}

func (sc *E2EContext) waitForKYCStatusByEmail(email string, expectedStatus int, timeout time.Duration) error {
	walletID, err := sc.getWalletIDByEmail(email)
	if err != nil {
		return fmt.Errorf("waitForKYCStatusByEmail: %w", err)
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		status, err := sc.getKYCStatusByWalletID(walletID)
		if err != nil {
			return fmt.Errorf("waitForKYCStatusByEmail: %w", err)
		}
		if status == expectedStatus {
			debugPrintf("   ✓ KYC status reached %d for wallet %s\n", expectedStatus, walletID)
			return nil
		}
		time.Sleep(kycWebhookPollInterval)
	}

	return fmt.Errorf("KYC status did not reach %d within %s", expectedStatus, timeout)
}

func (sc *E2EContext) myKYCStatusShouldBe(statusName string) error {
	expectedStatus, err := parseKYCStatusName(statusName)
	if err != nil {
		return err
	}

	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return fmt.Errorf("myKYCStatusShouldBe: %w", err)
	}

	return sc.waitForKYCStatusByEmail(email, expectedStatus, kycWebhookPollTimeout)
}

func parseKYCStatusName(statusName string) (int, error) {
	switch strings.ToLower(strings.TrimSpace(statusName)) {
	case "unknown":
		return kycStatusUnknown, nil
	case "pending":
		return kycStatusPending, nil
	case "documents required", "documents_required":
		return kycStatusDocumentsRequired, nil
	default:
		return -1, fmt.Errorf("unsupported KYC status name: %s", statusName)
	}
}

func (sc *E2EContext) iShouldSeeTheReactivateWalletPromptOnTheDashboard() error {
	debugPrintln("\n👁️  Checking for reactivate wallet prompt on dashboard...")

	if sc.browser == nil {
		if err := sc.initializeBrowser(); err != nil {
			return fmt.Errorf("initialize browser: %w", err)
		}
	}

	if _, err := sc.page.Goto(sc.baseURL+"/", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		Timeout:   playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("navigate to dashboard: %w", err)
	}

	deadline := time.Now().Add(kycWebhookPollTimeout)
	for time.Now().Before(deadline) {
		reactivateLink := sc.page.Locator("a:has-text('Reactivate wallet')")
		count, _ := reactivateLink.Count()
		if count > 0 {
			debugPrintln("   ✓ Found Reactivate wallet link")
			return nil
		}

		reservedChip := sc.page.Locator("text=Reserved")
		chipCount, _ := reservedChip.Count()
		if chipCount > 0 {
			content, _ := sc.page.Content()
			if strings.Contains(content, "resubmit") {
				debugPrintln("   ✓ Found documents-required banner with Reserved status")
				return nil
			}
		}

		_, _ = sc.page.Reload()
		time.Sleep(kycWebhookPollInterval)
	}

	return fmt.Errorf("reactivate wallet prompt not found on dashboard")
}
