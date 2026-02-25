package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// iLinkedASABankAccount is a composite step that links a South African bank account
// by navigating to the connect form, filling it, and submitting it.
// This reuses the same UI flow as the bank account linking test.
func (sc *E2EContext) iLinkedASABankAccount(bankName, accountNumber string) error {
	debugPrintf("\n🏦 Linking SA bank account: %s / %s\n", bankName, accountNumber)

	// Navigate to /connect/bank/za directly (more reliable than clicking the CTA)
	url := sc.baseURL + "/connect/bank/za"
	_, err := sc.page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(15000),
	})
	if err != nil {
		return fmt.Errorf("failed to navigate to bank connect page: %w", err)
	}

	sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
	time.Sleep(500 * time.Millisecond)

	_ = sc.iTakeAScreenshot("link-bank-form")

	// Fill account number
	accountInput := sc.page.Locator("input#accountNumber")
	if err := accountInput.Fill(accountNumber); err != nil {
		return fmt.Errorf("failed to fill account number: %w", err)
	}
	debugPrintf("   ✓ Filled account number: %s\n", accountNumber)

	// Select bank from dropdown
	if err := sc.selectBankOption(bankName); err != nil {
		return fmt.Errorf("failed to select bank: %w", err)
	}

	_ = sc.iTakeAScreenshot("link-bank-filled")

	// Submit the form
	submitBtn := sc.page.Locator("button[type='submit'][form='connect-bank-za'], button:has-text('Continue')").First()
	if err := submitBtn.Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("failed to click Continue: %w", err)
	}

	// Wait for redirect to /accounts
	sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(15000),
	})
	time.Sleep(1 * time.Second)

	_ = sc.iTakeAScreenshot("link-bank-result")

	debugPrintf("   📍 Current URL after submit: %s\n", sc.page.URL())

	// Verify we ended up on /accounts or the account was created
	if !strings.Contains(sc.page.URL(), "/accounts") {
		return fmt.Errorf("expected redirect to /accounts after linking, but at: %s", sc.page.URL())
	}

	debugPrintf("✓ SA bank account linked: %s / %s\n", bankName, accountNumber)
	return nil
}

// iDepositedIntoMyXagoBackedWallet is a composite step that deposits funds via
// the MockXago test API (sub-account retrieval + test transaction + deposit webhook).
func (sc *E2EContext) iDepositedIntoMyXagoBackedWallet(amountStr, currency string) error {
	debugPrintf("\n💰 Depositing %s %s into Xago-backed wallet...\n", amountStr, currency)

	// Step 1: Get Xago sub-account details (polls until created by Temporal)
	if err := sc.iGetTheXagoSubAccountDetailsForTheCurrentUser(); err != nil {
		return fmt.Errorf("failed to get Xago sub-account details: %w", err)
	}

	// Step 2: Create a test transaction in MockXago
	debugPrintln("   📝 Creating test transaction in MockXago...")
	if err := sc.iCreateATestTransactionInMockXagoFor(amountStr, currency); err != nil {
		return fmt.Errorf("failed to create test transaction: %w", err)
	}

	// Step 3: Perform the test deposit (triggers webhook to backend)
	debugPrintln("   💳 Performing test deposit...")
	if err := sc.iPerformATestDepositOfInMockXago(amountStr, currency); err != nil {
		return fmt.Errorf("failed to perform test deposit: %w", err)
	}

	// Step 4: Wait for webhook processing
	debugPrintln("   ⏳ Waiting for deposit webhook processing...")
	time.Sleep(10 * time.Second)

	debugPrintf("✓ Deposited %s %s into Xago-backed wallet\n", amountStr, currency)
	return nil
}

// iSetTheWithdrawAmountTo fills the withdrawal amount input field.
func (sc *E2EContext) iSetTheWithdrawAmountTo(amount string) error {
	debugPrintf("\n💲 Setting withdraw amount to: %s\n", amount)

	_ = sc.iTakeAScreenshot("before-set-withdraw-amount")

	// The withdraw amount input has id="withdrawAmount" and name="withdrawAmount"
	input := sc.page.Locator("input#withdrawAmount, input[name='withdrawAmount'], [data-testid='pay-amount-input']").First()

	err := input.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		return fmt.Errorf("withdraw amount input not visible: %w", err)
	}

	// Clear and fill
	if err := input.Fill(amount); err != nil {
		return fmt.Errorf("failed to fill withdraw amount: %w", err)
	}

	debugPrintf("✓ Set withdraw amount to: %s\n", amount)
	return nil
}

// iSelectFirstAvailableLinkedAccountToWithdrawTo opens the bank dropdown
// and selects the first available linked account option.
func (sc *E2EContext) iSelectFirstAvailableLinkedAccountToWithdrawTo() error {
	debugPrintln("\n🏦 Selecting first available linked account for withdrawal...")

	// The bank select button has id="bank" (Headless UI Listbox)
	bankButton := sc.page.Locator("#bank, button[id='bank']").First()

	err := bankButton.Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		return fmt.Errorf("failed to open bank dropdown: %w", err)
	}

	time.Sleep(300 * time.Millisecond)

	_ = sc.iTakeAScreenshot("withdraw-bank-dropdown")

	// Select the first option in the listbox
	option := sc.page.Locator("[role='option']").First()
	if count, _ := option.Count(); count == 0 {
		// Fallback: try li elements within listbox
		option = sc.page.Locator("[role='listbox'] li").First()
	}

	err = option.Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		return fmt.Errorf("failed to select first linked account: %w", err)
	}

	time.Sleep(300 * time.Millisecond)

	debugPrintln("✓ Selected first available linked account")
	return nil
}

// iSetTheWithdrawNoteTo fills the withdrawal note input field.
func (sc *E2EContext) iSetTheWithdrawNoteTo(note string) error {
	debugPrintf("\n📝 Setting withdraw note to: %s\n", note)

	// The note input has id="note" and name="note"
	input := sc.page.Locator("input#note, input[name='note']").First()

	err := input.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		return fmt.Errorf("withdraw note input not visible: %w", err)
	}

	if err := input.Fill(note); err != nil {
		return fmt.Errorf("failed to fill withdraw note: %w", err)
	}

	_ = sc.iTakeAScreenshot("withdraw-form-filled")

	debugPrintf("✓ Set withdraw note to: %s\n", note)
	return nil
}
