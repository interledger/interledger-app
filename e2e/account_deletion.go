package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/playwright-community/playwright-go"
)

func (sc *E2EContext) iEnableDeleteAccountFeatureForMyWallet() error {
	walletID, err := sc.resolveCurrentWalletID()
	if err != nil {
		return err
	}
	return sc.enableDeleteAccountForWallet(walletID)
}

func (sc *E2EContext) aPendingAccountDeletionRequestExistsForMeWithStatus(status string) error {
	kratosID, err := sc.resolveCurrentKratosID()
	if err != nil {
		return err
	}
	return sc.seedAccountDeletionRequest(kratosID, status)
}

func (sc *E2EContext) theDeleteAccountSettingsLinkShouldBeVisible() error {
	link := sc.page.GetByRole(*playwright.AriaRoleLink, playwright.PageGetByRoleOptions{
		Name: "Delete account",
	})
	if err := link.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		_ = sc.iTakeAScreenshot("delete-account-link-missing")
		return fmt.Errorf("Delete account settings link not visible: %w", err)
	}
	debugPrintln("✓ Delete account settings link is visible")
	return nil
}

func (sc *E2EContext) theDeleteAccountSettingsLinkShouldNotBeVisible() error {
	link := sc.page.GetByRole(*playwright.AriaRoleLink, playwright.PageGetByRoleOptions{
		Name: "Delete account",
	})
	count, err := link.Count()
	if err != nil {
		return fmt.Errorf("failed to inspect Delete account link: %w", err)
	}
	if count > 0 {
		_ = sc.iTakeAScreenshot("delete-account-link-unexpected")
		return fmt.Errorf("Delete account settings link should not be visible (found %d)", count)
	}
	debugPrintln("✓ Delete account settings link is not visible")
	return nil
}

// AriaRoleButton disambiguates from the settings-index link of the same label.
func (sc *E2EContext) iClickTheDestructiveDeleteAccountButton() error {
	btn := sc.page.GetByRole(*playwright.AriaRoleButton, playwright.PageGetByRoleOptions{
		Name: "Delete account",
	})
	if err := btn.First().Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		_ = sc.iTakeAScreenshot("delete-account-button-click-failed")
		return fmt.Errorf("failed to click Delete account button: %w", err)
	}
	debugPrintln("✓ Clicked Delete account button")
	return nil
}

func (sc *E2EContext) iCompleteTheTOTPStepUpChallenge() error {
	popupInput := sc.page.Locator("input[name='totp_code']")
	if err := popupInput.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	}); err != nil {
		_ = sc.iTakeAScreenshot("totp-popup-missing")
		return fmt.Errorf("TOTP step-up popup did not appear: %w", err)
	}

	if err := sc.iTypeInMyGeneratedTotpForMyself(); err != nil {
		return fmt.Errorf("failed to fill TOTP code in step-up popup: %w", err)
	}

	verifyBtn := sc.page.Locator("button:text-is('Verify')")
	if err := verifyBtn.First().Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("failed to click Verify in TOTP popup: %w", err)
	}

	// Popup-hidden is the definitive AAL2-success signal (verify fetcher's onSuccess closes it).
	if err := popupInput.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateHidden,
		Timeout: playwright.Float(10000),
	}); err != nil {
		_ = sc.iTakeAScreenshot("totp-popup-stuck")
		return fmt.Errorf("TOTP step-up popup did not close after verify: %w", err)
	}
	return nil
}

func (sc *E2EContext) thePendingAccountDeletionIndicatorShouldBeVisible() error {
	return sc.waitForAccountDeletionIndicator("Account deletion pending", "pending-deletion-indicator-missing")
}

func (sc *E2EContext) theInProgressAccountDeletionIndicatorShouldBeVisible() error {
	return sc.waitForAccountDeletionIndicator("Account deletion in progress", "in-progress-deletion-indicator-missing")
}

func (sc *E2EContext) waitForAccountDeletionIndicator(text, screenshot string) error {
	indicator := sc.page.GetByText(text, playwright.PageGetByTextOptions{
		Exact: playwright.Bool(false),
	})
	if err := indicator.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		_ = sc.iTakeAScreenshot(screenshot)
		return fmt.Errorf("%q indicator not visible: %w", text, err)
	}
	debugPrintf("✓ %q indicator is visible\n", text)
	return nil
}

func (sc *E2EContext) theTOTPStepUpPopupShouldNotAppear() error {
	popupInput := sc.page.Locator("input[name='totp_code']")
	// Settle window — without it we'd race the fetcher response that would
	// open the popup if TOTP were configured.
	time.Sleep(2 * time.Second)
	count, err := popupInput.Count()
	if err != nil {
		return fmt.Errorf("failed to inspect TOTP popup: %w", err)
	}
	if count > 0 {
		visible, _ := popupInput.First().IsVisible()
		if visible {
			_ = sc.iTakeAScreenshot("totp-popup-unexpectedly-visible")
			return fmt.Errorf("TOTP step-up popup should not have appeared")
		}
	}
	debugPrintln("✓ TOTP step-up popup did not appear")
	return nil
}

func (sc *E2EContext) myTOTPEnrollmentIsRemoved() error {
	kratosID, err := sc.resolveCurrentKratosID()
	if err != nil {
		return err
	}

	kratosAdminURL := os.Getenv("KRATOS_ADMIN_URL")
	if kratosAdminURL == "" {
		kratosAdminURL = "http://localhost:4434"
	}

	url := fmt.Sprintf("%s/admin/identities/%s/credentials/totp", kratosAdminURL, kratosID)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("build TOTP delete request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("delete TOTP credential: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d removing TOTP for %s: %s", resp.StatusCode, kratosID, string(body))
	}

	debugPrintf("✓ Removed TOTP credential for %s\n", kratosID)
	return nil
}

func (sc *E2EContext) noAccountDeletionRequestShouldExistForMe() error {
	kratosID, err := sc.resolveCurrentKratosID()
	if err != nil {
		return err
	}

	got, err := sc.getAccountDeletionStatus(kratosID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			debugPrintf("✓ no account-deletion request exists for %s\n", kratosID)
			return nil
		}
		return fmt.Errorf("failed to query account-deletion status: %w", err)
	}
	return fmt.Errorf("expected no account-deletion request for user %s but found one with status %q", kratosID, got)
}

func (sc *E2EContext) anAccountDeletionRequestShouldExistForMeWithStatus(expected string) error {
	kratosID, err := sc.resolveCurrentKratosID()
	if err != nil {
		return err
	}

	got, err := sc.getAccountDeletionStatus(kratosID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("expected account-deletion request for user %s but found none", kratosID)
		}
		return fmt.Errorf("failed to query account-deletion status: %w", err)
	}
	if got != expected {
		return fmt.Errorf("expected account-deletion status %q, got %q", expected, got)
	}
	debugPrintf("✓ account-deletion request exists for me with status %q\n", got)
	return nil
}

func (sc *E2EContext) resolveCurrentKratosID() (string, error) {
	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return "", fmt.Errorf("cannot resolve current user email: %w", err)
	}
	kratosID := sc.getKratosUserIDByEmail(email)
	if kratosID == "" {
		return "", fmt.Errorf("cannot resolve kratos id for email %q", email)
	}
	return kratosID, nil
}

func (sc *E2EContext) resolveCurrentWalletID() (string, error) {
	kratosID, err := sc.resolveCurrentKratosID()
	if err != nil {
		return "", err
	}
	walletID, err := sc.getWalletIDForUser(kratosID)
	if err != nil {
		return "", fmt.Errorf("cannot resolve wallet id: %w", err)
	}
	return walletID, nil
}
