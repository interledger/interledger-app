# Quick Reference: E2E Test Step Library

This document provides a quick lookup for all available test steps organized by functionality.

## Current Step List (Feb 2026)

### Background and Setup
```gherkin
Given a random test identifier is generated
And the frontend is running at "https://interledger.test"
And Rafiki assets are seeded
Given the database is clean for email "signup-german-user@example.com"
```

### Signup Flow (full and minimal)
```gherkin
Given I complete the signup flow with first name "Anna" last name "Müller" email "anna@example.com" country "Germany" phone "+4917" and password "SecurePass2025!"
Given I complete the minimal KYC flow with first name "Bob" last name "Butler" email "bob@example.com" country "Germany" phone "+4917" and password "SecurePass2025!"
```

### Form Inputs
```gherkin
When I fill in "first name" with "Anna"
And I fill in "phone" with a random 10 digit number prefixed with "+49"
And I try to fill in "password" with "SecurePass2025!"
And I select "Germany" from the country dropdown
And I check the terms and conditions checkbox
And I try to submit without filling required fields
```

### Signup Assertions
```gherkin
Then the signup should be submitted
And a signup record should exist in the database for "signup-german-user@example.com"
And the signup should have first name "Anna"
And the signup should have last name "Müller"
And the signup should have country code "DE"
And I should be able to verify the full user status
And I should see validation errors or the form should validate on blur
```

### Login + TOTP
```gherkin
When I clear the browser session
And I navigate to the login page
And I fill in login credentials "user@example.com" with "SecurePass2025!"
And I submit the login
Then I should be navigated to the TOTP page
When I type in my generated totp for my new user
And I submit the totp registration
Then I should be navigated to the application dashboard
```

### Wallet Address + KYC
```gherkin
When I should be on the wallet address creation page
And I fill in and submit the wallet address form with a unique address
And I click the "save" button on the wallet-address-form
Then I should be navigated back to the dashboard with reserved wallet status

When I navigate to the personal details page to activate wallet
Then I should see the activate wallet button
When I click the "Continue" button
And I wait for the KYC iframe to load
And I fill and submit the mockgatehub KYC iframe
And I wait for the KYC completion
Then I should be navigated back to the dashboard with approved kyc status
And I should see my account balance with kyc approved
```

### Deposit
```gherkin
When I navigate to the deposit page
And I deposit "100" "EUR" via the deposit iframe
Then I should see my balance updated with "100" "EUR"
```

### Screenshots
```gherkin
And I take a screenshot "kyc-completed-dashboard"
```

### Debug Output
Enable or disable noisy debug output using the test flag:
```sh
go test -v -timeout 5m -debug=false
```

---

## Signup Flow

### Complete Full Signup
```gherkin
Given I complete the signup flow with first name "Anna" last name "Müller" email "anna@example.com" country "Germany" phone "+493012345678" and password "SecurePass2025!"
```
- Navigates to signup page
- Fills all form fields across multiple steps
- Handles password confirmation
- Accepts terms and conditions
- Submits form
- **Time:** 20-30 seconds
- **Use for:** Complete signup workflow tests

## Navigation

### Navigate to Signup Page
```gherkin
Given I navigate to the signup page
```
- Opens browser (if not already open)
- Navigates to base URL with /signup path
- **Time:** 2-3 seconds
- **Note:** Initializes Playwright browser on first use

### Navigate to Login Page
```gherkin
When I navigate to the login page
```
- Navigates to base URL with /login path
- **Time:** 2-3 seconds

### Navigate to Personal Details Page
```gherkin
When I navigate to the personal details or activation page
```
- Navigates to base URL with /personal-details path
- **Time:** 2-3 seconds
- **Use for:** Account activation and KYC flows

## Form Filling

### Fill Text Input (Required Field)
```gherkin
When I fill in "field name" with "value"
```
- Finds input by placeholder, name, label, or role
- Clears existing value
- Types new value
- Fails if field not found
- **Examples:**
  - `When I fill in "first name" with "Anna"`
  - `When I fill in "email" with "anna@example.com"`
  - `When I fill in "phone" with "+493012345678"`
  - `When I fill in "password" with "SecurePass2025!"`

### Fill Text Input (Optional Field)
```gherkin
And I try to fill in "field name" with "value"
```
- Same as above but doesn't fail if field is missing
- Useful for optional or conditional fields
- **Examples:**
  - `And I try to fill in "password" with "SecurePass2025!"`
  - `And I try to fill in "password confirmation" with "SecurePass2025!"`

### Select from Country Dropdown
```gherkin
And I select "Germany" from the country dropdown
```
- Finds and opens country dropdown
- Selects option by country name
- Waits for page to settle after selection
- **Time:** 1-2 seconds
- **Examples:**
  - `And I select "Germany" from the country dropdown`
  - `And I select "Singapore" from the country dropdown`

### Check Terms and Conditions Checkbox
```gherkin
And I check the terms and conditions checkbox
```
- Finds T&C checkbox
- Checks it if unchecked
- **Time:** 1 second

## Button Clicks and Page Interactions

### Click Button
```gherkin
When I click the "button text" button
```
- Smart matching by text or data-testid
- Waits for page load after click
- **Time:** 1-3 seconds depending on page load
- **Examples:**
  - `When I click the "Continue" button`
  - `When I click the "Sign Up" button`
  - `When I click the "Save" button`
  - `When I click the "Confirm" button`

### Confirm Button is Available
```gherkin
Then I can confirm the "Save" button is available to click
```
- Verifies button exists and is visible/enabled
- Uses multiple selector patterns for robustness
- Returns success if button found anywhere on page
- Logs available buttons for debugging if not found
- **Time:** 1-2 seconds
- **Use for:** Validating form state before interaction

### Take Screenshot
```gherkin
And I take a screenshot "screenshot-name"
```
- Captures full page screenshot
- Saves to `debug/<feature>__<scenario>/<test-id>-NN-screenshot-name.png`
- Useful for documentation and debugging
- **Time:** < 0.5 seconds
- **Examples:**
  - `And I take a screenshot "step1-profile-filled"`
  - `And I take a screenshot "kyc-wallet-name-page"`
  - `And I take a screenshot "kyc-dashboard-after-wallet-creation"`

### Wait for Page Load
```gherkin
And I wait for the page to load
```
- Waits for network idle state
- Adds 2 second rendering buffer
- **Time:** 2-5 seconds
- **Use for:** Ensuring page is fully rendered after navigation

## User Verification and Authentication

### Trigger User Email Verification
```gherkin
When I trigger user verification for "anna@example.com"
```
- Marks user's email as verified in the system
- Allows user to proceed past verification step
- **Time:** < 1 second
- **Use for:** Simulating email verification in testing
- **Note:** User must already exist from signup

### Fill Login Credentials
```gherkin
And I fill in login credentials "anna@example.com" with "SecurePass2025!"
```
- Fills email and password fields on login page
- Does NOT submit the form
- Handles multiple input selector patterns
- **Time:** 2-3 seconds
- **Use with:** `And I submit the login`

### Submit Login
```gherkin
And I submit the login
```
- Clicks login submit button
- Waits for page transition
- **Time:** 3-5 seconds
- **Precondition:** Credentials must be filled first

### Navigate to TOTP Page
```gherkin
Then I should be navigated to the TOTP page
```
- Verifies current URL contains "/totp" or "two-factor"
- Confirms user is on TOTP registration page
- **Time:** < 1 second

### Generate and Enter TOTP Code
```gherkin
When I type in my generated totp for my new user
```
- Extracts TOTP secret from page
- Generates valid 6-digit TOTP code
- Enters code into TOTP input field
- **Time:** 2-3 seconds
- **Note:** Secret must be visible on page HTML

### Submit TOTP Registration
```gherkin
And I submit the totp registration
```
- Clicks TOTP submit/verify button
- Waits for page transition
- **Time:** 5 seconds

### Navigate to Application Dashboard
```gherkin
Then I should be navigated to the application dashboard
```
- Verifies navigation away from login/TOTP flow
- Checks for dashboard-related URL patterns
- **Time:** < 1 second
- **Expected URLs:** /dashboard, /app, /home, /wallet, /accounts, /personal-details, /activation

## Form and Page Validation

### Verify Signup Form Displayed
```gherkin
Then I should see the signup form
```
- Checks current URL contains "/signup"
- **Time:** < 1 second

### Verify User on Specific Signup Step
```gherkin
Then I should be on step 2
```
- Sets internal step counter
- Used for tracking multi-step flows
- **Time:** 2 seconds (includes load buffer)

### Attempt Form Submission Without Required Fields
```gherkin
When I try to submit without filling required fields
```
- Finds and clicks submit button
- Doesn't fill any form fields
- Used to test validation messages
- **Time:** 2 seconds

### Check for Validation Errors
```gherkin
Then I should see validation errors or the form should validate on blur
```
- Searches page content for validation messages
- Returns success if any validation indicators found
- Accepts forms that validate on blur (without errors shown)
- **Time:** < 1 second

## Database Verification

### Verify Signup Record Exists
```gherkin
Then a signup record should exist in the database for "anna@example.com"
```
- Queries PostgreSQL database for signup record
- Falls back to waitlist_signups table if needed
- Stores signup ID for later use
- **Time:** 1-2 seconds

### Verify First Name
```gherkin
And the signup should have first name "Anna"
```
- Queries database for stored first name
- Case-insensitive comparison
- Handles edge cases (e.g., country text prepended)
- **Requires:** Previous record lookup

### Verify Last Name
```gherkin
And the signup should have last name "Müller"
```
- Queries database for stored last name
- Supports special characters and accents
- **Requires:** Previous record lookup

### Verify Country Code
```gherkin
And the signup should have country code "DE"
```
- Queries database for country code
- Uses ISO 3166-1 alpha-2 format
- **Requires:** Previous record lookup

### Verify Full User Status
```gherkin
And I should be able to verify the full user status
```
- Comprehensive database check
- Verifies signup exists in all required tables
- Prints full user state for debugging
- **Time:** 2-3 seconds
- **Use for:** Confirming complete user setup

## Complete Example Scenarios

### Example 1: Full Signup to Dashboard
```gherkin
Feature: User Signup and Dashboard Access

  Scenario: Successfully complete signup and reach dashboard
    Given the frontend is running at "https://interledger.test"
    And the database is clean for email "anna.mueller@example.com"
    When I complete the signup flow with first name "Anna" last name "Müller" email "anna.mueller@example.com" country "Germany" phone "+493012345678" and password "InterlEdger2025!TestPassword"
    Then a signup record should exist in the database for "anna.mueller@example.com"
    And the signup should have first name "Anna"
    And the signup should have last name "Müller"
    And the signup should have country code "DE"
    When I trigger user verification for "anna.mueller@example.com"
    And I navigate to the login page
    And I fill in login credentials "anna.mueller@example.com" with "InterlEdger2025!TestPassword"
    And I submit the login
    Then I should be navigated to the TOTP page
    When I type in my generated totp for my new user
    And I submit the totp registration
    Then I should be navigated to the application dashboard
```
- **Time:** ~2 minutes
- **Coverage:** Complete signup + verification + login + TOTP + dashboard

### Example 2: Wallet Creation After Login
```gherkin
Feature: KYC and Account Activation

  Scenario: Successfully activate account and create wallet
    Given the frontend is running at "https://interledger.test"
    And the database is clean for email "anna.mueller@example.com"
    When I complete the signup flow with first name "Anna" last name "Müller" email "anna.mueller@example.com" country "Germany" phone "+493012345678" and password "InterlEdger2025!TestPassword"
    And I trigger user verification for "anna.mueller@example.com"
    And I navigate to the login page
    And I fill in login credentials "anna.mueller@example.com" with "InterlEdger2025!TestPassword"
    And I submit the login
    And I type in my generated totp for my new user
    And I submit the totp registration
    Then I should be navigated to the application dashboard
    When I take a screenshot "wallet-name-page"
    Then I can confirm the "Save" button is available to click
    When I click the "Save" button
    And I wait for the page to load
    And I take a screenshot "dashboard-after-wallet-creation"
    Then I should be navigated to the application dashboard
```
- **Time:** ~2 minutes
- **Coverage:** Complete flow with wallet creation

### Example 3: Simple Signup Without Verification
```gherkin
Feature: Signup Form Validation

  Scenario: Test signup form displays and validates
    Given the frontend is running at "https://interledger.test"
    And the database is clean for email "new.user@example.com"
    When I navigate to the signup page
    And I click the "Sign Up" button
    And I click the "Let's get started" button
    Then I should see the signup form
    When I fill in "first name" with "John"
    And I fill in "email" with "john@example.com"
    And I select "Germany" from the country dropdown
    And I click the "Continue" button
    Then I should be on step 2
    When I fill in "phone" with "+493012345678"
    And I try to fill in "password" with "SecurePass2025!"
    And I click the "Continue" button
    Then I take a screenshot "signup-step-3"
```
- **Time:** ~30 seconds
- **Coverage:** Signup form flow and validation

## Tips & Best Practices

### Setup Requirements
- ✅ **Always start with:** `Given the frontend is running at "https://interledger.test"`
  - Establishes base URL for all navigation
  - Sets up browser context
  
- ✅ **Clean database before tests:** `And the database is clean for email "user@example.com"`
  - Prevents conflicts with previous test data
  - Removes old Kratos identities
  - Essential for test repeatability

### Debugging Failed Steps
1. **Check screenshots:** Look in `/tmp/` for `*.png` files
   - Screenshots are auto-saved on failures
   - Helpful for visual debugging
   
2. **Check database state:** Use verification steps to query database
   - `Then a signup record should exist...`
   - `And the signup should have first name...`
   
3. **Add debugging screenshots:** Insert before failing step
   - `And I take a screenshot "before-failing-step"`
   
4. **Add waits:** Some pages need rendering time
   - `And I wait for the page to load`
   - `And I wait for 2 seconds`
   - Check for timeout errors in output

### Writing Robust Tests

**For Optional Fields:**
- Use `And I try to fill in` instead of `And I fill in`
- Won't fail if field doesn't exist

**For Dynamic Page Content:**
- Use flexible button selectors: `I can confirm the "button" button is available`
- Checks multiple selector patterns

**For Multi-Step Processes:**
- Take screenshots between major steps
- Helps with debugging and documentation
- Example: `And I take a screenshot "step-N-description"`

**For Form Validation:**
- Test both happy path (all fields filled) and validation path (missing fields)
- Use `And I try to submit without filling required fields`

### Performance Tips
- Screenshots take time; use sparingly in CI
- Database queries are fast (< 1 second)
- Waits for network idle add 2-3 seconds per page load
- Total test should complete in 2-3 minutes

### Common Mistakes to Avoid
- ❌ Forgetting `Given the frontend is running at` → Tests will fail with no base URL
- ❌ Skipping `And the database is clean for` → Tests conflict with previous runs
- ❌ Not waiting for TOTP page after login → Test rushes past authentication
- ❌ Wrong email format in database assertions → SQL queries fail silently
- ❌ Submitting form before all fields filled → Form validation errors
- ❌ Taking too many screenshots → Tests run slowly
- ❌ Not checking step output for button naming → Click fails on wrong button text

### Troubleshooting Common Issues

**Issue: "Button not found"**
- Check the actual button text in page
- Use `I can confirm the "X" button is available` to debug
- Check screenshot for actual button text

**Issue: "Page navigation failed"**
- Ensure base URL is set: `Given the frontend is running at`
- Check if page exists and is accessible
- Add screenshot before navigation

**Issue: "Database query returned no results"**
- Verify email matches exactly (case-sensitive for queries)
- Check database was cleaned: `the database is clean for email`
- Run manually: `SELECT * FROM signups WHERE email = 'user@example.com'`

**Issue: "TOTP code invalid"**
- TOTP secret extraction is time-sensitive
- Code expires every 30 seconds
- Check page source includes secret in expected format
- Add screenshot of TOTP page if failing

**Issue: "Test hangs or times out"**
- Add screenshots to find where it's stuck
- May be waiting for slow page load
- Use shorter timeouts in CI (test should < 3 minutes)
- Check browser console for JavaScript errors

### Running Tests Locally
```bash
cd /home/stephan/interledger/interledger-app/e2e
go test -v -timeout 300s
```

Check for:
- ✅ All 65+ steps passing
- ✅ Screenshots created in `/tmp/`
- ✅ No timeout errors
- ✅ Database connections successful

