package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cucumber/godog"
	_ "github.com/lib/pq"
	"github.com/playwright-community/playwright-go"
)

// E2EContext holds the test context for signup scenarios
type E2EContext struct {
	pw          *playwright.Playwright
	browser     playwright.Browser
	context     playwright.BrowserContext
	page        playwright.Page
	db          *sql.DB
	baseURL     string
	email       string
	password    string
	firstName   string
	lastName    string
	country     string
	countryCode string
	dateOfBirth string

	// User details from Gherkin tables
	userDetails map[string]*UserDetails
	currentUser string // Currently impersonated user

	// Payment flow state
	receiverWalletAddress string // Wallet address/identifier for payment receiver

	// Test state
	currentStep     int
	signupID        string
	passwordFilled  bool   // Track if password was successfully filled
	testIdentifier  string // Random prefix for test emails to ensure uniqueness
	screenshotCount int    // Track number of screenshots taken in this scenario
	screenshotDir   string // Per-scenario screenshot directory
	totpSecret      string // TOTP secret for the current user
}

// InitializeScenario sets up the scenario context
func InitializeScenario(ctx *godog.ScenarioContext) {
	var sc *E2EContext

	// Background steps
	ctx.Before(func(goCtx context.Context, scenario *godog.Scenario) (context.Context, error) {
		// Create a fresh SignupContext for each scenario
		// This ensures complete isolation between scenarios
		sc = &E2EContext{}
		sc.screenshotDir = buildScenarioScreenshotDir(scenario)
		debugPrintf("\n🔧 Initialized new context for scenario: %s\n", scenario.Name)
		return goCtx, nil
	})

	ctx.After(func(goCtx context.Context, scenario *godog.Scenario, err error) (context.Context, error) {
		if sc == nil {
			return goCtx, nil
		}
		// Cleanup in reverse order
		if sc.page != nil {
			_ = sc.page.Close()
		}
		if sc.context != nil {
			_ = sc.context.Close()
		}
		if sc.browser != nil {
			_ = sc.browser.Close()
		}
		if sc.pw != nil {
			_ = sc.pw.Stop()
		}
		if sc.db != nil {
			_ = sc.db.Close()
		}
		debugPrintf("✓ Cleaned up context for scenario: %s\n", scenario.Name)
		return goCtx, nil
	})

	// Background steps - wrapped to ensure sc is initialized
	ctx.Step(`^a random test identifier is generated$`, func() error { return sc.aRandomTestIdentifierIsGenerated() })
	ctx.Step(`^the frontend is running at "([^"]*)"$`, func(url string) error { return sc.theFrontendIsRunningAt(url) })
	ctx.Step(`^mockgatehub is running at "([^"]*)"$`, func(url string) error { return sc.theMockgatehubIsRunningAt(url) })
	ctx.Step(`^mockxago is running at "([^"]*)"$`, func(url string) error { return sc.theMockxagoIsRunningAt(url) })
	ctx.Step(`^Rafiki assets are seeded$`, func() error { return sc.rafikiAssetsExist() })

	// User details and impersonation steps
	ctx.Step(`^the details of '([^']*)' are$`, func(userName string, table *godog.Table) error { return sc.theDetailsOfAre(userName, table) })
	ctx.Step(`^I impersonate '([^']*)'$`, func(userName string) error { return sc.iImpersonate(userName) })
	ctx.Step(`^that my "([^"]*)" is "([^"]*)"$`, func(fieldKey, value string) error { return sc.thatMyFieldIs(fieldKey, value) })
	ctx.Step(`^I completed the signup workflow$`, func() error { return sc.iCompletedTheSignupWorkflow() })
	ctx.Step(`^I completed the account verification workflow$`, func() error { return sc.iCompletedTheAccountVerificationWorkflow() })
	ctx.Step(`^I finished the TOTP registration workflow$`, func() error { return sc.iFinishedTheTOTPRegistrationWorkflow() })
	ctx.Step(`^I finished the wallet address creation workflow$`, func() error { return sc.iFinishedTheWalletAddressCreationWorkflow() })
	ctx.Step(`^I fill in "([^"]*)" with my "([^"]*)"$`, func(fieldName, fieldKey string) error { return sc.iFillInWithMy(fieldName, fieldKey) })
	ctx.Step(`^I fill in "([^"]*)" with "([^"]*)" prefixed with the random identifier$`, func(fieldName, fieldKey string) error { return sc.iFillInWithPrefixed(fieldName, fieldKey) })
	ctx.Step(`^I try to fill in "([^"]*)" with my "([^"]*)"$`, func(fieldName, fieldKey string) error { return sc.iTryToFillInWithMy(fieldName, fieldKey) })

	// Navigation steps
	ctx.Step(`^I navigate to the signup page$`, func() error { return sc.iNavigateToTheSignupPage() })
	ctx.Step(`^I click the "([^"]*)" button$`, func(buttonText string) error { return sc.iClickTheButton(buttonText) })
	ctx.Step(`^I should see the signup form$`, func() error { return sc.iShouldSeeTheSignupForm() })
	ctx.Step(`^I should be on step (\d+)$`, func(step int) error { return sc.iShouldBeOnStep(step) })
	ctx.Step("^I complete the minimal KYC flow `([^`]*)`$", func(userName string) error {
		return sc.iCompleteMinimalKYCFlow(userName)
	})

	// Form filling steps
	ctx.Step(`^I try to fill in "([^"]*)" with "([^"]*)"$`, func(fieldName, value string) error { return sc.iTryToFillInWith(fieldName, value) })
	ctx.Step(`^I select "([^"]*)" from the country dropdown$`, func(country string) error { return sc.iSelectFromTheCountryDropdown(country) })
	ctx.Step(`^I check the terms and conditions checkbox$`, func() error { return sc.iCheckTheTermsAndConditionsCheckbox() })
	ctx.Step(`^I take a screenshot "([^"]*)"$`, func(name string) error { return sc.iTakeAScreenshot(name) })
	ctx.Step(`^I try to submit without filling required fields$`, func() error { return sc.iTryToSubmitWithoutFillingRequiredFields() })

	// Assertion steps
	ctx.Step(`^the signup should be submitted$`, func() error { return sc.theSignupShouldBeSubmitted() })
	ctx.Step(`^I should see validation errors or the form should validate on blur$`, func() error { return sc.iShouldSeeValidationErrors() })
	ctx.Step(`^a signup record should exist in the database for myself$`, func() error { return sc.aSignupRecordShouldExistForMyself() })
	ctx.Step(`^a signup record should exist for myself$`, func() error { return sc.aSignupRecordShouldExistForMyself() })
	ctx.Step(`^the signup should have first name matching my "([^"]*)"$`, func(fieldKey string) error { return sc.theSignupShouldHaveFirstNameMatching(fieldKey) })
	ctx.Step(`^the signup should have last name matching my "([^"]*)"$`, func(fieldKey string) error { return sc.theSignupShouldHaveLastNameMatching(fieldKey) })
	ctx.Step(`^the signup should have country code matching my "([^"]*)"$`, func(fieldKey string) error { return sc.theSignupShouldHaveCountryCodeMatching(fieldKey) })
	ctx.Step(`^I should be able to verify the full user status$`, func() error { return sc.iShouldBeAbleToVerifyTheFullUserStatus() })

	// User verification steps
	ctx.Step(`^I trigger user verification for myself$`, func() error { return sc.iTriggerUserVerificationForMyself() })

	// Phone steps
	ctx.Step(`^I fill in "([^"]*)" with a unique valid phone number$`, func(fieldName string) error { return sc.iFillInWithUniquePhoneNumber(fieldName) })

	// Login steps
	ctx.Step(`^I clear the browser session$`, func() error { return sc.iClearTheBrowserSession() })
	ctx.Step(`^I navigate to (https?://.+)$`, func(url string) error { return sc.iNavigateToURL(url) })
	ctx.Step(`^I navigate to the login page$`, func() error { return sc.iNavigateToTheLoginPage() })
	ctx.Step(`^I fill in my login credentials$`, func() error { return sc.iFillInMyLoginCredentials() })
	ctx.Step(`^I fill in the login form with my details$`, func() error { return sc.iFillInTheLoginFormWithMyDetails() })
	ctx.Step(`^I submit the login$`, func() error { return sc.iSubmitTheLogin() })
	ctx.Step(`^I should be navigated to the TOTP page$`, func() error { return sc.iShouldBeNavigatedToTheTOTPPage() })
	ctx.Step(`^I log in as myself$`, func() error { return sc.iLogInAsMyself() })

	// TOTP registration steps
	ctx.Step(`^I type in my generated totp for myself$`, func() error { return sc.iTypeInMyGeneratedTotpForMyself() })
	ctx.Step(`^I type in my generated totp$`, func() error { return sc.iTypeInMyGeneratedTotpForMyself() })
	ctx.Step(`^I submit the totp registration$`, func() error { return sc.iSubmitTheTotpRegistration() })
	ctx.Step(`^I should be navigated to the application dashboard$`, func() error { return sc.iShouldBeNavigatedToTheApplicationDashboard() })

	// KYC and account activation steps
	ctx.Step(`^I navigate to the personal details page to activate wallet$`, func() error { return sc.iNavigateToThePersonalDetailsPageToActivateWallet() })
	ctx.Step(`^I should see the activate wallet button$`, func() error { return sc.iShouldSeeTheActivateWalletButton() })
	ctx.Step(`^I should be shown the "([^"]*)" prompt form$`, func(promptText string) error { return sc.iShouldBeShownTheActivateWalletPromptForm(promptText) })
	ctx.Step(`^I wait for the KYC iframe to load$`, func() error { return sc.iWaitForTheKYCIframeToLoad() })
	ctx.Step(`^I fill and submit the mockgatehub KYC iframe$`, func() error { return sc.iFillAndSubmitTheMockgatehubiframe() })
	ctx.Step(`^I fill and submit the mockxago KYC iframe$`, func() error { return sc.iFillAndSubmitTheMockxagoiframe() })
	ctx.Step(`^I wait for the KYC completion$`, func() error { return sc.iWaitForTheKYCCompletion() })

	// Wallet address creation steps
	ctx.Step(`^I should be redirected to the wallet address creation page$`, func() error { return sc.iShouldBeRedirectedToTheWalletAddressCreationPage() })
	ctx.Step(`^I fill in and submit the wallet address form with a unique address$`, func() error { return sc.iFillInAndSubmitTheWalletAddressFormWithAUniqueAddress() })
	ctx.Step(`^I click the "([^"]+)" button on the wallet-address-form$`, func(buttonText string) error { return sc.iClickTheButtonOnTheWalletAddressForm(buttonText) })
	ctx.Step(`^I should be navigated back to the dashboard with reserved wallet status$`, func() error { return sc.iShouldBeNavigatedBackToTheDashboardWithReservedWalletStatus() })
	ctx.Step(`^I should be navigated back to the dashboard with approved kyc status$`, func() error { return sc.iShouldBeNavigatedBackToTheDashboardWithApprovedKYCStatus() })
	ctx.Step(`^I should see my account balance with kyc approved$`, func() error { return sc.iShouldSeeMyAccountBalanceWithKYCApproved() })

	// Deposit steps
	ctx.Step(`^I navigate to the deposit page$`, func() error { return sc.iNavigateToTheDepositPage() })
	ctx.Step(`^I deposit "([^"]*)" "([^"]*)" via the deposit iframe$`, func(amount, currency string) error {
		return sc.iDepositViATheDepositIframe(amount, currency)
	})
	ctx.Step(`^I deposit "([^"]*)" "([^"]*)" via the deposit iframe as "([^"]*)"$`, func(amount, currency, userName string) error {
		return sc.iDepositViATheDepositIframeAsUser(amount, currency, userName)
	})
	ctx.Step(`^I should see my balance updated with "([^"]*)" "([^"]*)"$`, func(amount, currency string) error {
		return sc.iShouldSeeMyBalanceUpdatedWithAmount(amount, currency)
	})
	ctx.Step(`^that Gatehub charges my user a ([0-9.]+)% deposit fee$`, func(feePercent string) error {
		return sc.thatGatehubChargesDepositFee(feePercent)
	})

	// Xago deposit steps (test-mode via MockXago)
	ctx.Step(`^I get the Xago sub account details for the current user$`, func() error {
		return sc.iGetTheXagoSubAccountDetailsForTheCurrentUser()
	})
	ctx.Step(`^I create a test transaction in MockXago for "([^"]*)" "([^"]*)"$`, func(amount, currency string) error {
		return sc.iCreateATestTransactionInMockXagoFor(amount, currency)
	})
	ctx.Step(`^I perform a test deposit of "([^"]*)" "([^"]*)" in MockXago$`, func(amount, currency string) error {
		return sc.iPerformATestDepositOfInMockXago(amount, currency)
	})
	ctx.Step(`^I initiate a deposit for my Xago linked account$`, func() error {
		return sc.iInitiateADepositForMyXagoLinkedAccount()
	})
	ctx.Step(`^my Xago specific deposit instructions should be displayed to me$`, func() error {
		return sc.myXagoSpecificDepositInstructionsShouldBeDisplayedToMe()
	})
	ctx.Step(`^I simulate a "([^"]*)" "([^"]*)" EFT payment to Xago$`, func(amount, currency string) error {
		return sc.iSimulateAnEFTPaymentToXago(amount, currency)
	})
	ctx.Step(`^I wait "([^"]*)" seconds for the webhook to be processed$`, func(seconds string) error {
		return sc.iWaitSecondsForTheWebhookToBeProcessed(seconds)
	})
	ctx.Step(`^I navigate to the home page$`, func() error {
		return sc.iNavigateToTheHomePage()
	})
	ctx.Step(`^I should see my ZAR balance updated with "([^"]*)"$`, func(amount string) error {
		return sc.iShouldSeeMyZARBalanceUpdatedWith(amount)
	})
	ctx.Step(`^the test deposit should have been accepted by MockXago$`, func() error {
		return sc.iTheTestDepositShouldHaveBeenAcceptedByMockXago()
	})

	// Withdrawal steps
	ctx.Step(`^I navigate to the withdrawal page$`, func() error { return sc.iNavigateToTheWithdrawalPage() })
	ctx.Step(`^I withdraw "([^"]*)" "([^"]*)" via the withdrawal iframe$`, func(amount, currency string) error {
		return sc.iWithdrawViATheWithdrawalIframe(amount, currency)
	})
	ctx.Step(`^that Gatehub charges my user a ([0-9.]+)% withdrawal fee$`, func(feePercent string) error {
		return sc.thatGatehubChargesWithdrawalFee(feePercent)
	})

	// P2P Payment steps
	ctx.Step(`^I navigate to the dashboard$`, func() error { return sc.iNavigateToTheDashboard() })
	ctx.Step(`^I should see my account balance with "([^"]*)" "([^"]*)"$`, func(amount, currency string) error {
		return sc.iShouldSeeMyAccountBalanceWith(amount, currency)
	})
	ctx.Step(`^I navigate to the send payment page$`, func() error { return sc.iNavigateToTheSendPaymentPage() })
	ctx.Step(`^I fill in the receiver email with the "([^"]*)" email$`, func(userName string) error {
		return sc.iFillInTheReceiverEmailWith(userName)
	})
	ctx.Step(`^I fill in the payment amount "([^"]*)"$`, func(amount string) error {
		return sc.iFillInThePaymentAmount(amount)
	})
	ctx.Step(`^I select the payment currency "([^"]*)"$`, func(currency string) error {
		return sc.iSelectThePaymentCurrency(currency)
	})
	ctx.Step(`^I submit the payment$`, func() error { return sc.iSubmitThePayment() })
	ctx.Step(`^I should see a payment confirmation$`, func() error { return sc.iShouldSeeAPaymentConfirmation() })
	ctx.Step(`^I wait for the payment to complete$`, func() error { return sc.iWaitForThePaymentToComplete() })
	ctx.Step(`^the payment form should be accessible$`, func() error { return sc.thePaymentFormShouldBeAccessible() })
	ctx.Step(`^I should see the payments page$`, func() error { return sc.iShouldSeeThePaymentsPage() })
	ctx.Step(`^I navigate to the deposit page$`, func() error { return sc.iNavigateToTheDepositPage() })
	ctx.Step(`^I navigate to the payments history page$`, func() error { return sc.iNavigateToThePaymentsHistoryPage() })
	ctx.Step(`^I get the receiver wallet address for "([^"]*)"$`, func(userName string) error {
		return sc.iGetTheReceiverWalletAddressFor(userName)
	})
	ctx.Step(`^I fill in the receiver wallet address$`, func() error { return sc.iFillInTheReceiverWalletAddress() })
	ctx.Step(`^I wait "([^"]*)" seconds for the (payment|deposit|page load) to (.+)$`, func(seconds, typ, action string) error {
		return sc.iWaitSeconds(seconds)
	})
	ctx.Step(`^I wait "([^"]*)" seconds for the page to load$`, func(seconds string) error {
		return sc.iWaitSeconds(seconds)
	})
	ctx.Step(`^I should see the payment in my transaction history$`, func() error {
		return sc.iShouldSeeThePaymentInMyTransactionHistory()
	})
}

// Background step implementations

func (sc *E2EContext) aRandomTestIdentifierIsGenerated() error {
	// Generate a random test identifier based on current timestamp
	// This ensures uniqueness across test runs
	sc.testIdentifier = fmt.Sprintf("%d", time.Now().UnixNano()%1000000)
	debugPrintf("✓ Generated random test identifier: %s\n", sc.testIdentifier)
	return nil
}

func (sc *E2EContext) theFrontendIsRunningAt(urlStr string) error {
	sc.baseURL = urlStr

	// Extract hostname from URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("failed to parse frontend URL: %w", err)
	}

	// First ensure DNS resolves
	if err := sc.ensureHostsResolve([]string{parsedURL.Hostname()}); err != nil {
		return err
	}

	// Then verify the frontend is actually serving HTML
	debugPrintf("🔍 Verifying frontend is serving content at %s...\n", urlStr)
	return sc.waitForHTMLToBeServed(urlStr, 60*time.Second)
}

func (sc *E2EContext) theMockgatehubIsRunningAt(urlStr string) error {
	// Extract hostname from URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("failed to parse mockgatehub URL: %w", err)
	}

	// First ensure DNS resolves
	if err := sc.ensureHostsResolve([]string{parsedURL.Hostname()}); err != nil {
		return err
	}

	// Then verify mockgatehub health endpoint is responding
	debugPrintf("🔍 Verifying mockgatehub health endpoint at %s...\n", urlStr)
	healthURL := strings.TrimSuffix(urlStr, "/") + "/health"
	return sc.waitForHealthEndpoint(healthURL, 30*time.Second)
}

func (sc *E2EContext) theMockxagoIsRunningAt(urlStr string) error {
	// Extract hostname from URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("failed to parse mockxago URL: %w", err)
	}

	// First ensure DNS resolves
	if err := sc.ensureHostsResolve([]string{parsedURL.Hostname()}); err != nil {
		return err
	}

	// Then verify mockxago health endpoint is responding
	debugPrintf("🔍 Verifying mockxago health endpoint at %s...\n", urlStr)
	healthURL := strings.TrimSuffix(urlStr, "/") + "/health"
	return sc.waitForHealthEndpoint(healthURL, 30*time.Second)
}

func (sc *E2EContext) ensureHostsResolve(hosts []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	for _, host := range hosts {
		ips, err := net.DefaultResolver.LookupHost(ctx, host)
		if err != nil || len(ips) == 0 {
			return fmt.Errorf("required host %s does not resolve; run `make hosts` from local to update /etc/hosts", host)
		}
	}

	return nil
}

// waitForHTMLToBeServed polls a URL until it returns HTML content or times out
func (sc *E2EContext) waitForHTMLToBeServed(url string, timeout time.Duration) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	deadline := time.Now().Add(timeout)
	attempt := 0

	for time.Now().Before(deadline) {
		attempt++
		req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()

			// Read first few bytes to verify it's actually HTML
			buf := make([]byte, 512)
			n, _ := io.ReadFull(resp.Body, buf)
			content := string(buf[:n])

			// Check if response looks like HTML
			if strings.Contains(strings.ToLower(content), "<html") ||
				strings.Contains(strings.ToLower(content), "<!doctype") ||
				strings.Contains(content, "<head") {
				debugPrintf("✅ Service is serving HTML content (attempt %d)\n", attempt)
				return nil
			}

			debugPrintf("⚠️  Service returned 200 but content doesn't look like HTML (attempt %d)\n", attempt)
		} else if err != nil {
			debugPrintf("⏳ Waiting for service to be ready... (attempt %d: %v)\n", attempt, err)
		} else {
			debugPrintf("⏳ Service returned status %d (attempt %d), waiting...\n", resp.StatusCode, attempt)
			if resp.Body != nil {
				resp.Body.Close()
			}
		}

		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("service at %s did not serve HTML content within %v (tried %d times)", url, timeout, attempt)
}

// waitForHealthEndpoint polls a health endpoint until it returns 200 OK or times out
func (sc *E2EContext) waitForHealthEndpoint(url string, timeout time.Duration) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	deadline := time.Now().Add(timeout)
	attempt := 0

	for time.Now().Before(deadline) {
		attempt++
		req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create health check request: %w", err)
		}

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			debugPrintf("✅ Health endpoint is responding (attempt %d)\n", attempt)
			return nil
		}

		if err != nil {
			debugPrintf("⏳ Waiting for health endpoint... (attempt %d: %v)\n", attempt, err)
		} else {
			debugPrintf("⏳ Health endpoint returned status %d (attempt %d), waiting...\n", resp.StatusCode, attempt)
			if resp.Body != nil {
				resp.Body.Close()
			}
		}

		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("health endpoint at %s did not respond with 200 OK within %v (tried %d times)", url, timeout, attempt)
}

type kratosIdentity struct {
	ID          string                 `json:"id"`
	Traits      map[string]interface{} `json:"traits"`
	Credentials map[string]interface{} `json:"credentials"`
}

func (sc *E2EContext) getKratosUserIDByEmail(email string) string {
	kratosAdminURL := os.Getenv("KRATOS_ADMIN_URL")
	if kratosAdminURL == "" {
		kratosAdminURL = "http://localhost:4434"
	}

	client := &http.Client{Timeout: 30 * time.Second}
	listReq, err := http.NewRequestWithContext(context.Background(), "GET", kratosAdminURL+"/admin/identities", nil)
	if err != nil {
		debugPrintf("⚠️  getKratosUserIDByEmail: failed to build request: %v\n", err)
		return ""
	}

	debugPrintf("→ GET %s/admin/identities\n", kratosAdminURL)
	listResp, err := client.Do(listReq)
	if err != nil {
		debugPrintf("⚠️  getKratosUserIDByEmail: request error: %v\n", err)
		return ""
	}
	defer listResp.Body.Close()

	if listResp.StatusCode != http.StatusOK {
		// Read response body for diagnostics
		var bodyBytes []byte
		bodyBytes, _ = io.ReadAll(listResp.Body)
		debugPrintf("⚠️  getKratosUserIDByEmail: unexpected status %d: %s\n", listResp.StatusCode, string(bodyBytes))
		return ""
	}

	var identities []kratosIdentity
	if err := json.NewDecoder(listResp.Body).Decode(&identities); err != nil {
		debugPrintf("⚠️  getKratosUserIDByEmail: decode error: %v\n", err)
		return ""
	}

	for _, identity := range identities {
		traits := identity.Traits
		if v, ok := traits["email"]; ok {
			if s, ok2 := v.(string); ok2 && s == email {
				debugPrintf("   ✓ Found Kratos identity id=%s for email=%s\n", identity.ID, email)
				return identity.ID
			}
		}
	}

	debugPrintf("   - No Kratos identity found for %s\n", email)
	return ""
}

var nonAlphaNumeric = regexp.MustCompile(`[^a-z0-9]+`)

func normalizeName(value string) string {
	value = strings.ToLower(value)
	value = nonAlphaNumeric.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	return value
}

func featureNameFromScenario(scenario *godog.Scenario) string {
	if scenario == nil {
		return "feature"
	}
	if scenario.Uri == "" {
		return "feature"
	}
	fileName := filepath.Base(scenario.Uri)
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

func buildScenarioScreenshotDir(scenario *godog.Scenario) string {
	featureSlug := normalizeName(featureNameFromScenario(scenario))
	if featureSlug == "" {
		featureSlug = "feature"
	}
	scenarioName := ""
	if scenario != nil {
		scenarioName = scenario.Name
	}
	scenarioSlug := normalizeName(scenarioName)
	if scenarioSlug == "" {
		scenarioSlug = "scenario"
	}
	folderName := fmt.Sprintf("%s__%s", featureSlug, scenarioSlug)
	return filepath.Join("debug", folderName)
}
