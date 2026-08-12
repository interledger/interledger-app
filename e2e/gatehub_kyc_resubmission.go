package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mxschmitt/playwright-go"
)

const (
	kycStatusUnknown           = 0
	kycStatusPending           = 1
	kycStatusDocumentsRequired = 2
	kycWebhookPollInterval     = 500 * time.Millisecond
	kycWebhookPollTimeout      = 30 * time.Second

	// Must match consts.DefaultKYCGateway in mockgatehub (separate module).
	defaultKYCGateway = "DINARO d.o.o."
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
		gateway = defaultKYCGateway
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

	// Only fire the webhook; the resulting status is asserted by the separate
	// "my KYC status should be ..." step, which polls with a timeout.
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

func (sc *E2EContext) setMockGatehubUserKYCState(gatehubUserID, kycState string) error {
	baseURL := "https://mockgatehub.interledger.test"
	if sc.mockgatehubBaseURL != "" {
		baseURL = sc.mockgatehubBaseURL
	}

	body := fmt.Sprintf(`{"kyc_state":%q}`, kycState)
	reqURL := fmt.Sprintf("%s/admin/users/%s/kyc-state", baseURL, url.PathEscape(gatehubUserID))
	req, err := http.NewRequest(http.MethodPut, reqURL, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("setMockGatehubUserKYCState: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := mockgatehubHTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("setMockGatehubUserKYCState: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("setMockGatehubUserKYCState: unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	debugPrintf("   ✓ Set MockGatehub user %s kyc_state to %s\n", gatehubUserID, kycState)
	return nil
}

func (sc *E2EContext) iSetMyMockGatehubUserKYCStateTo(state string) error {
	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return err
	}

	gatehubUserID, err := sc.getGatehubUserIDByEmail(email)
	if err != nil {
		return err
	}

	return sc.setMockGatehubUserKYCState(gatehubUserID, state)
}

func (sc *E2EContext) iShouldSeeTheGatehubKYCWidgetUnavailableMessage() error {
	debugPrintln("\n👁️  Checking for GateHub KYC widget unavailable message...")

	deadline := time.Now().Add(kycWebhookPollTimeout)
	for time.Now().Before(deadline) {
		unavailable := sc.page.Locator("text=Document resubmission is not available right now")
		count, _ := unavailable.Count()
		if count > 0 {
			debugPrintln("   ✓ Found widget unavailable message")
			return nil
		}

		iframeCount, _ := sc.page.Locator("iframe").Count()
		if iframeCount > 0 {
			return fmt.Errorf("expected no GateHub KYC iframe when widget is unavailable")
		}

		time.Sleep(kycWebhookPollInterval)
	}

	return fmt.Errorf("GateHub KYC widget unavailable message not found")
}

func (sc *E2EContext) iShouldNotSeeTheGatehubKYCWidgetUnavailableMessage() error {
	debugPrintln("\n👁️  Verifying GateHub KYC widget unavailable message is absent...")

	// Wait for the page to render (the "Continue" button) before asserting absence,
	// so the check can't pass trivially against a not-yet-rendered page.
	activateButton := sc.page.Locator("button:has-text('Continue'), button:has-text('Activate')")
	if err := activateButton.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("activation page did not render: %w", err)
	}

	unavailable := sc.page.Locator("text=Document resubmission is not available right now")
	count, _ := unavailable.Count()
	if count > 0 {
		return fmt.Errorf("did not expect GateHub KYC widget unavailable message")
	}

	debugPrintln("   ✓ Widget unavailable message not shown")
	return nil
}
