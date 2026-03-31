package main

import (
	"fmt"
	"time"
)

// iCompletedThePTIKYCFlowFor runs the full post-signup KYC flow using the mockpti iframe.
// It mirrors iCompletedTheKYCFlowFor in gatehub_kyc.go but submits the PTI form.
func (sc *E2EContext) iCompletedThePTIKYCFlowFor(email string) error {
	debugPrintf("\n🔄 Executing complete PTI KYC flow for: %s\n", email)

	prefixedEmail, err := sc.getCurrentUserEmail()
	if err != nil {
		return fmt.Errorf("failed to get email for PTI KYC flow: %w", err)
	}

	password := sc.password
	if password == "" {
		return fmt.Errorf("no password set - signup must be completed before PTI KYC flow")
	}

	if err := sc.aSignupRecordShouldExistInTheDatabase(prefixedEmail); err != nil {
		return fmt.Errorf("signup verification failed: %w", err)
	}

	if err := sc.iTriggerUserVerificationFor(prefixedEmail); err != nil {
		return fmt.Errorf("trigger verification failed: %w", err)
	}

	if err := sc.iClearTheBrowserSession(); err != nil {
		return fmt.Errorf("clear session failed: %w", err)
	}

	if err := sc.iNavigateToTheLoginPage(); err != nil {
		return fmt.Errorf("navigate to login failed: %w", err)
	}

	if err := sc.iFillInLoginCredentials(prefixedEmail, password); err != nil {
		return fmt.Errorf("fill login credentials failed: %w", err)
	}

	if err := sc.iSubmitTheLogin(); err != nil {
		return fmt.Errorf("submit login failed: %w", err)
	}

	if err := sc.iShouldBeNavigatedToTheTOTPPage(); err != nil {
		return fmt.Errorf("TOTP page navigation failed: %w", err)
	}

	if err := sc.iTypeInMyGeneratedTotpForMyNewUser(); err != nil {
		return fmt.Errorf("TOTP generation failed: %w", err)
	}

	if err := sc.iSubmitTheTotpRegistration(); err != nil {
		return fmt.Errorf("TOTP registration failed: %w", err)
	}

	if err := sc.iShouldBeNavigatedToTheApplicationDashboard(); err != nil {
		return fmt.Errorf("dashboard navigation failed: %w", err)
	}

	if err := sc.iShouldBeOnTheWalletAddressCreationPage(); err != nil {
		return fmt.Errorf("wallet address page check failed: %w", err)
	}

	if err := sc.iFillInAndSubmitTheWalletAddressFormWithAUniqueAddress(); err != nil {
		return fmt.Errorf("fill wallet address form failed: %w", err)
	}

	if err := sc.iClickTheButtonOnTheWalletAddressForm("save"); err != nil {
		return fmt.Errorf("click save button failed: %w", err)
	}

	if err := sc.iShouldBeNavigatedBackToTheDashboardWithReservedWalletStatus(); err != nil {
		return fmt.Errorf("dashboard with reserved status failed: %w", err)
	}

	if err := sc.iNavigateToThePersonalDetailsPageToActivateWallet(); err != nil {
		return fmt.Errorf("navigate to personal details failed: %w", err)
	}

	if err := sc.iWaitForTheKYCIframeToLoad(); err != nil {
		return fmt.Errorf("PTI KYC iframe load failed: %w", err)
	}

	if err := sc.iFillAndSubmitTheMockptiKYCIframe(); err != nil {
		return fmt.Errorf("fill PTI KYC iframe failed: %w", err)
	}

	if err := sc.iWaitForTheKYCCompletion(); err != nil {
		return fmt.Errorf("KYC completion wait failed: %w", err)
	}

	if err := sc.iShouldBeNavigatedBackToTheDashboardWithApprovedKYCStatus(); err != nil {
		return fmt.Errorf("dashboard with approved KYC failed: %w", err)
	}

	if err := sc.iShouldSeeMyAccountBalanceWithKYCApproved(); err != nil {
		return fmt.Errorf("account balance verification failed: %w", err)
	}

	_ = sc.iTakeAScreenshot("pti-kyc-completed-dashboard")
	debugPrintf("✅ Complete PTI KYC flow finished successfully for %s\n", prefixedEmail)
	return nil
}

// iCompleteMinimalPTIKYCFlowWithDetails runs signup then PTI KYC for the given user details.
func (sc *E2EContext) iCompleteMinimalPTIKYCFlowWithDetails(firstName, lastName, email, country, phone, password string) error {
	debugPrintf("\n🧭 Running minimal PTI KYC flow for: %s\n", email)

	if err := sc.iCompleteSignupFlow(firstName, lastName, email, country, phone, password); err != nil {
		return fmt.Errorf("signup flow failed: %w", err)
	}

	if err := sc.iCompletedThePTIKYCFlowFor(email); err != nil {
		return fmt.Errorf("post-signup PTI KYC flow failed: %w", err)
	}

	return nil
}

// iCompleteMinimalPTIKYCFlow runs the minimal PTI KYC flow using impersonated user details.
// Usage: And I complete the minimal PTI KYC flow `deposit-pti-user`
func (sc *E2EContext) iCompleteMinimalPTIKYCFlow(userName string) error {
	debugPrintf("\n🧭 Running minimal PTI KYC flow for user: %s\n", userName)

	if err := sc.iImpersonate(userName); err != nil {
		return fmt.Errorf("failed to impersonate user '%s': %w", userName, err)
	}

	details, exists := sc.userDetails[userName]
	if !exists {
		return fmt.Errorf("no details defined for user '%s'", userName)
	}

	emailSuffix, ok := details.Fields["emailSuffix"]
	if !ok {
		return fmt.Errorf("emailSuffix not defined for user '%s'", userName)
	}

	password, ok := details.Fields["password"]
	if !ok {
		return fmt.Errorf("password not defined for user '%s'", userName)
	}

	country, ok := details.Fields["country"]
	if !ok {
		return fmt.Errorf("country not defined for user '%s'", userName)
	}

	firstName, ok := details.Fields["firstName"]
	if !ok {
		return fmt.Errorf("firstName not defined for user '%s'", userName)
	}

	lastName, ok := details.Fields["lastName"]
	if !ok {
		return fmt.Errorf("lastName not defined for user '%s'", userName)
	}

	phone := "+1202" // US phone prefix (digits will be overlaid by generateDeterministicPhone)

	return sc.iCompleteMinimalPTIKYCFlowWithDetails(firstName, lastName, emailSuffix, country, phone, password)
}

// iFillAndSubmitTheMockptiKYCIframe fills and submits the mockpti KYC iframe.
// The mockpti form has a date-of-birth field and a "Complete" button (not "Submit").
// On completion it calls /forms/complete on the mockpti service and posts
// { name: 'UserAssessmentCompleted' } to the parent window.
func (sc *E2EContext) iFillAndSubmitTheMockptiKYCIframe() error {
	debugPrintln("\n📝 Filling and submitting mockpti KYC iframe...")

	time.Sleep(500 * time.Millisecond)

	iframeLocator := sc.page.Locator("iframe").First()
	iframeSrc, _ := iframeLocator.GetAttribute("src")
	debugPrintf("   📍 Iframe src: %s\n", iframeSrc)

	// Listen for the PTI completion postMessage
	_, err := sc.page.Evaluate(`() => {
		window.ptiKycCompleted = false;
		window.addEventListener('message', (e) => {
			console.log('Parent received PTI message:', JSON.stringify(e.data));
			if (e.data && e.data.name === 'UserAssessmentCompleted') {
				window.ptiKycCompleted = true;
				console.log('PTI KYC completed message received');
			}
		});
	}`)
	if err != nil {
		debugPrintf("   ⚠️  Failed to set up PTI message listener: %v\n", err)
	}

	iframeCount, _ := iframeLocator.Count()
	if iframeCount == 0 {
		return fmt.Errorf("no iframe found on page")
	}

	frameLocator := sc.page.FrameLocator("iframe").First()

	// Wait for iframe content to load
	time.Sleep(500 * time.Millisecond)

	// Fill the date-of-birth field (id="dob")
	dobField := frameLocator.Locator("#dob")
	dobCount, _ := dobField.Count()
	if dobCount > 0 {
		dob := sc.dateOfBirth
		if dob == "" {
			dob = "1990-01-01"
		}
		debugPrintf("   📍 Filling date of birth: %s\n", dob)
		if err := dobField.Fill(dob); err != nil {
			debugPrintf("   ⚠️  Failed to fill dob field: %v\n", err)
		}
	} else {
		debugPrintf("   ⚠️  Date of birth field (#dob) not found in PTI iframe\n")
	}

	if err := sc.iTakeAScreenshot("pti-kyc-form-filled"); err != nil {
		debugPrintf("   ⚠️  Failed to take screenshot: %v\n", err)
	}

	// Click the "Complete" button (id="complete")
	completeButton := frameLocator.Locator("#complete")
	buttonCount, _ := completeButton.Count()
	if buttonCount == 0 {
		completeButton = frameLocator.Locator("button:has-text('Complete')")
		buttonCount, _ = completeButton.Count()
	}
	if buttonCount == 0 {
		return fmt.Errorf("PTI KYC iframe 'Complete' button not found")
	}

	debugPrintf("   ✓ Clicking 'Complete' button in PTI iframe\n")
	if err := completeButton.Click(); err != nil {
		return fmt.Errorf("failed to click Complete button in PTI iframe: %w", err)
	}

	time.Sleep(500 * time.Millisecond)
	debugPrintf("   ✓ PTI KYC iframe form submitted\n")
	return nil
}
