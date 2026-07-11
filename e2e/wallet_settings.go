package main

import (
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"
)

// Steps for the "Wallet Settings" botanist tab, backed by the entityconf
// library (go/backend/entityconf, go/backend/walletconf) rather than the
// wallet_features table that the existing "feature toggle" steps in
// botanist.go exercise. See botanist.go's featureToggleSwitch/
// theFeatureToggleShouldBe/iToggleTheFeatureOn/
// theFeatureShouldBeEnabledInTheDatabase for the analogous, older steps this
// mirrors.

func (sc *E2EContext) iNavigateToMyWalletSettingsInAdminPortal() error {
	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return fmt.Errorf("cannot resolve current user email: %w", err)
	}

	kratosID := sc.getKratosUserIDByEmail(email)
	if kratosID == "" {
		return fmt.Errorf("cannot resolve kratos ID for email %q", email)
	}

	details, err := sc.getWalletDetailsForUser(kratosID)
	if err != nil {
		return fmt.Errorf("cannot get wallet details for user %q: %w", email, err)
	}

	settingsURL := sc.botanistBaseURL + "/wallet/" + details.ID + "/wallet-settings"
	debugPrintf("\n🌐 Navigating to botanist wallet settings page: %s\n", settingsURL)

	_, err = sc.page.Goto(settingsURL, playwright.PageGotoOptions{
		Timeout:   playwright.Float(30000),
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		return fmt.Errorf("failed to navigate to wallet settings page: %w", err)
	}

	if err = sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(15000),
	}); err != nil {
		return fmt.Errorf("wallet settings page did not reach network idle: %w", err)
	}

	// Wait for React to finish client-side hydration before interacting.
	hydrationLocator := sc.page.Locator("html[data-hydrated='true']")
	if err = hydrationLocator.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(15000),
	}); err != nil {
		return fmt.Errorf("React hydration did not complete on wallet settings page: %w", err)
	}

	debugPrintf("✓ On botanist wallet settings page\n")
	return nil
}

// Exact-text anchor avoids prefix collisions between similarly-named confs.
func (sc *E2EContext) walletSettingToggleSwitch(displayName string) playwright.Locator {
	selector := fmt.Sprintf(
		`div:has(> dt:text-is("%s")) button[role="switch"]`,
		displayName,
	)
	return sc.page.Locator(selector).First()
}

func (sc *E2EContext) theWalletSettingToggleShouldBe(displayName, expectedState string) error {
	want := "false"
	if expectedState == "on" {
		want = "true"
	}

	toggle := sc.walletSettingToggleSwitch(displayName)
	if err := toggle.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("wallet setting toggle %q not visible: %w", displayName, err)
	}

	got, err := toggle.GetAttribute("aria-checked")
	if err != nil {
		return fmt.Errorf("failed to read aria-checked for %q: %w", displayName, err)
	}
	if got != want {
		return fmt.Errorf("expected wallet setting %q to be %s (aria-checked=%s), got aria-checked=%s",
			displayName, expectedState, want, got)
	}
	debugPrintf("✓ Wallet setting toggle %q is %s\n", displayName, expectedState)
	return nil
}

// Poll aria-checked: the Switch only flips once the useFetcher round-trip lands.
func (sc *E2EContext) iToggleTheWalletSettingOn(displayName string) error {
	toggle := sc.walletSettingToggleSwitch(displayName)

	got, err := toggle.GetAttribute("aria-checked")
	if err != nil {
		return fmt.Errorf("failed to read aria-checked for %q: %w", displayName, err)
	}
	if got == "true" {
		debugPrintf("✓ Wallet setting %q already on; nothing to toggle\n", displayName)
		return nil
	}

	if err := toggle.Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("failed to click toggle for %q: %w", displayName, err)
	}

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		got, err := toggle.GetAttribute("aria-checked")
		if err == nil && got == "true" {
			debugPrintf("✓ Wallet setting %q toggled on\n", displayName)
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return fmt.Errorf("wallet setting %q did not settle to aria-checked=true within 10s", displayName)
}

func (sc *E2EContext) theWalletSettingShouldBeEnabledInTheDatabase(displayName string) error {
	walletID, err := sc.resolveCurrentWalletID()
	if err != nil {
		return err
	}

	confKey, ok := walletSettingDisplayNameToConfKey[displayName]
	if !ok {
		return fmt.Errorf("no entityconf key mapping for wallet setting %q", displayName)
	}

	val, err := sc.getWalletConfBool(walletID, confKey)
	if err != nil {
		return fmt.Errorf("failed to read %s for wallet %s: %w", confKey, walletID, err)
	}
	if !val {
		return fmt.Errorf("expected %s=true for wallet %s, got false", confKey, walletID)
	}
	debugPrintf("✓ DB confirms %s=true for wallet %s\n", confKey, walletID)
	return nil
}

// Botanist "Wallet Settings" display names → go/backend/walletconf.Confs
// keys (see go/backend/walletconf/walletconf.go).
var walletSettingDisplayNameToConfKey = map[string]string{
	"Send": "wallet.send_enabled",
}
