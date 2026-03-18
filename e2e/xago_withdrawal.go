package main

import (
	"fmt"
	"strings"
	"time"
)

// iLinkedASABankAccount is a composite step that links a South African bank account.
func (sc *E2EContext) iLinkedASABankAccount(bankName, accountNumber string) error {
	debugPrintf("\n🏦 Linking SA bank account: %s / %s\n", bankName, accountNumber)

	url := sc.baseURL + "/connect/bank/za"
	_, err := sc.page.Goto(url)
	if err != nil {
		return fmt.Errorf("failed to navigate to bank connect page: %w", err)
	}

	time.Sleep(500 * time.Millisecond)
	_ = sc.iTakeAScreenshot("link-bank-form")

	accountInput := sc.page.Locator("input#accountNumber")
	if err := accountInput.Fill(accountNumber); err != nil {
		return fmt.Errorf("failed to fill account number: %w", err)
	}

	if err := sc.selectBankOption(bankName); err != nil {
		return fmt.Errorf("failed to select bank: %w", err)
	}

	_ = sc.iTakeAScreenshot("link-bank-filled")

	submitBtn := sc.page.Locator("button[type='submit'][form='connect-bank-za'], button:has-text('Continue')").First()
	if err := submitBtn.Click(); err != nil {
		return fmt.Errorf("failed to click Continue: %w", err)
	}

	time.Sleep(1 * time.Second)
	_ = sc.iTakeAScreenshot("link-bank-result")

	if !strings.Contains(sc.page.URL(), "/accounts") {
		return fmt.Errorf("expected redirect to /accounts after linking, but at: %s", sc.page.URL())
	}

	debugPrintf("✓ SA bank account linked: %s / %s\n", bankName, accountNumber)
	return nil
}

// iDepositedIntoMyXagoBackedWallet is a composite step that deposits funds via MockXago.
func (sc *E2EContext) iDepositedIntoMyXagoBackedWallet(amountStr, currency string) error {
	debugPrintf("\n💰 Depositing %s %s into Xago-backed wallet...\n", amountStr, currency)

	if err := sc.iGetTheXagoSubAccountDetailsForTheCurrentUser(); err != nil {
		return fmt.Errorf("failed to get Xago sub-account details: %w", err)
	}

	if err := sc.iCreateATestTransactionInMockXagoFor(amountStr, currency); err != nil {
		return fmt.Errorf("failed to create test transaction: %w", err)
	}

	if err := sc.iPerformATestDepositOfInMockXago(amountStr, currency); err != nil {
		return fmt.Errorf("failed to perform test deposit: %w", err)
	}

	debugPrintln("   ⏳ Waiting for deposit webhook processing...")
	time.Sleep(10 * time.Second)

	debugPrintf("✓ Deposited %s %s into Xago-backed wallet\n", amountStr, currency)
	return nil
}

// iSetTheWithdrawAmountTo fills the withdrawal amount field.
func (sc *E2EContext) iSetTheWithdrawAmountTo(amount string) error {
	debugPrintf("\n💵 Setting withdraw amount to: %s\n", amount)

	amountInput := sc.page.Locator("input[name='amount'], input#amount, input[placeholder*='amount' i]")
	if count, _ := amountInput.Count(); count > 0 {
		if err := amountInput.First().Fill(amount); err == nil {
			debugPrintf("✓ Set withdraw amount: %s\n", amount)
			return nil
		}
	}

	return fmt.Errorf("could not find withdraw amount input")
}

// iSelectFirstAvailableLinkedAccountToWithdrawTo selects the first available linked account.
func (sc *E2EContext) iSelectFirstAvailableLinkedAccountToWithdrawTo() error {
	debugPrintln("\n🏦 Selecting first available linked account for withdrawal...")

	accountOption := sc.page.Locator("[data-testid='withdraw-account-option'], input[type='radio']").First()
	if count, _ := accountOption.Count(); count > 0 {
		if err := accountOption.Click(); err == nil {
			debugPrintln("✓ Selected first linked account")
			return nil
		}
	}

	return fmt.Errorf("could not find linked account options for withdrawal")
}

// iSetTheWithdrawNoteTo fills the withdrawal reference/note field.
func (sc *E2EContext) iSetTheWithdrawNoteTo(note string) error {
	debugPrintf("\n📝 Setting withdraw note to: %s\n", note)

	noteInput := sc.page.Locator("input[name='reference'], input[name='note'], input[placeholder*='note' i], input[placeholder*='reference' i]")
	if count, _ := noteInput.Count(); count > 0 {
		if err := noteInput.First().Fill(note); err == nil {
			debugPrintf("✓ Set withdraw note: %s\n", note)
			return nil
		}
	}

	return fmt.Errorf("could not find withdraw note/reference input")
}

