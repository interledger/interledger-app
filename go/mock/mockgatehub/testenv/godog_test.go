//go:build e2e

package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var opts = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "progress",
	Paths:  []string{"../features"},
	Tags:   "~@skip && ~@stubbed",
}

func TestFeatures(t *testing.T) {
	// Start services
	if err := startServices(); err != nil {
		t.Fatalf("Failed to start services: %v", err)
	}
	defer cleanup()

	// Wait for services to be ready
	if err := waitForServices(); err != nil {
		t.Fatalf("Services failed to start: %v", err)
	}

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}

	if status := suite.Run(); status != 0 {
		t.Fatalf("one or more scenarios failed")
	}
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	tc := &TestContext{}
	tc.Reset()

	// Health check steps
	ctx.Step(`^MockGatehub is started via docker-compose in the test environment$`, tc.mockgatehubStarted)
	ctx.Step(`^the service URL is (.+)$`, tc.serviceURLIs)
	ctx.Step(`^I send a GET request to (.+) with valid HMAC headers$`, tc.sendGetRequestWithHMAC)
	ctx.Step(`^the response status is (\d+)$`, tc.responseStatusIs)
	ctx.Step(`^the response body contains "([^"]*)":"([^"]*)"$`, tc.responseBodyContains)

	// Auth steps
	ctx.Step(`^a clean MockGatehub instance$`, tc.cleanMockGatehub)
	ctx.Step(`^HMAC headers using app id "([^"]*)" and secret "([^"]*)"$`, tc.hmacHeaders)
	ctx.Step(`^I POST (.+) with email "([^"]*)"$`, tc.postWithEmail)
	ctx.Step(`^the payload includes a generated user id$`, tc.payloadHasUserID)
	ctx.Step(`^the managed flag is true$`, tc.managedFlagTrue)
	ctx.Step(`^an existing managed user$`, tc.existingManagedUserGeneric)
	ctx.Step(`^I POST (/[^ ]+)$`, tc.postEndpoint)
	ctx.Step(`^the response contains a token for the KYC flow$`, tc.responseHasKYCToken)
	ctx.Step(`^GET (.+) shows kyc_state "([^"]*)" and risk_level "([^"]*)"$`, tc.getShowsKYCState)
	// Specific GET patterns must come before generic GET pattern below
	ctx.Step(`^I GET /core/v1/wallets/\{walletAddress\}/balances$`, tc.getWalletBalance)
	ctx.Step(`^I GET /wallets/\{walletAddress\}/balances$`, tc.getWalletBalance)
	ctx.Step(`^I GET (/[^ ]+)$`, tc.getEndpoint)
	ctx.Step(`^the response is HTML that mentions "([^"]*)" and "([^"]*)"$`, tc.responseIsHTMLMentioning)

	// KYC 2FA steps
	ctx.Step(`^I submit the KYC form for user \{userId\} without 2FA$`, tc.submitKYCFormWithout2FA)
	ctx.Step(`^I submit the KYC form for user \{userId\} with 2FA and code "([^"]*)"$`, tc.submitKYCFormWith2FA)

	// Signature auth steps
	ctx.Step(`^a clean MockGatehub instance with authentication enforced$`, tc.cleanMockGatehubInstanceWithAuthenticationEnforced)
	ctx.Step(`^valid credentials with app id "([^"]*)" and secret "([^"]*)"$`, tc.validCredentialsWithAppIdAndSecret)
	ctx.Step(`^the current timestamp in milliseconds$`, tc.currentTimestampInMilliseconds)
	ctx.Step(`^a fixed timestamp value "([^"]*)"$`, tc.fixedTimestampValue)
	ctx.Step(`^body '([^']*)'$`, tc.bodyIs)
	ctx.Step(`^the base string "([^"]*)"$`, tc.baseStringIs)
	ctx.Step(`^the base string "([^"]*)" with no body component$`, tc.baseStringWithNoBodyComponent)
	ctx.Step(`^the base string includes the query string in the URL component$`, tc.baseStringIncludesQueryString)
	ctx.Step(`^the base string uses "([^"]*)" as the URL$`, tc.baseStringUsesURL)
	ctx.Step(`^the base string "([^"]*)" using path only$`, tc.baseStringUsingPathOnly)
	ctx.Step(`^the base string "([^"]*)" using the old simple format$`, tc.baseStringUsingOldSimpleFormat)
	ctx.Step(`^the signature is computed from "([^"]*)"$`, tc.signatureComputedFrom)
	ctx.Step(`^the signature is incorrectly computed using seconds "([^"]*)" in the base string$`, tc.signatureComputedUsingSeconds)
	ctx.Step(`^the request includes header X-Forwarded-Proto "([^"]*)"$`, tc.requestIncludesHeaderXForwardedProto)
	ctx.Step(`^the request includes header X-Forwarded-Host "([^"]*)"$`, tc.requestIncludesHeaderXForwardedHost)
	ctx.Step(`^the response contains a user id$`, tc.responseContainsUserID)
	ctx.Step(`^the response status is not (\d+)$`, tc.responseStatusIsNot)
	ctx.Step(`^I POST (/[^ ]+) with body '([^']*)' signed using the full URL$`, tc.postSignedUsingFullURL)
	ctx.Step(`^I GET (/[^ ]+) signed using the full URL$`, tc.getSignedUsingFullURL)
	ctx.Step(`^I POST (/[^ ]+) with body '([^']*)' signed using the full URL with query$`, tc.postSignedUsingFullURLWithQuery)
	ctx.Step(`^I POST (/[^ ]+) with body '([^']*)' signed using the proxied URL$`, tc.postSignedUsingProxiedURL)
	ctx.Step(`^I POST (/[^ ]+) with body '([^']*)' signed using only the path$`, tc.postSignedUsingPathOnly)
	ctx.Step(`^I POST (/[^ ]+) with body '([^']*)' signed using the simple format$`, tc.postSignedUsingSimpleFormat)
	ctx.Step(`^I POST (/[^ ]+) with body '([^']*)' without the x-gatehub-signature header$`, tc.postWithoutSignatureHeader)
	ctx.Step(`^I POST (/[^ ]+) with body '([^']*)' without the x-gatehub-timestamp header$`, tc.postWithoutTimestampHeader)
	ctx.Step(`^I POST (/[^ ]+) with body '([^']*)' without the x-gatehub-app-id header$`, tc.postWithoutAppIDHeader)
	ctx.Step(`^I POST (/[^ ]+) with body '([^']*)' using app id "([^"]*)"$`, tc.postWithUnknownAppID)
	ctx.Step(`^I POST (/[^ ]+) with body '([^']*)' signed with secret "([^"]*)"$`, tc.postSignedWithSecret)
	ctx.Step(`^I POST (/[^ ]+) with that body and x-gatehub-timestamp set to "([^"]*)"$`, tc.postSignedWithTimestamp)
	ctx.Step(`^I GET /health without any HMAC headers$`, tc.getHealthWithoutHMAC)
	ctx.Step(`^I GET / without any HMAC headers$`, tc.getRootWithoutHMAC)
	ctx.Step(`^I POST /iframe/submit without any HMAC headers$`, tc.postIframeSubmitWithoutHMAC)

	// Wallet steps
	ctx.Step(`^an authenticated request using HMAC headers$`, tc.authenticatedRequest)
	ctx.Step(`^a managed user id produced by (.+)$`, tc.managedUserFromEndpoint)
	ctx.Step(`^a wallets array is returned with at least one wallet$`, tc.walletsArrayReturned)
	ctx.Step(`^the first wallet address starts with "([^"]*)"$`, tc.firstWalletStartsWith)
	ctx.Step(`^the ([A-Z]+) balance is "([^"]*)"$`, tc.currencyBalanceIs)

	// Rates steps
	ctx.Step(`^MockGatehub is running and requests include valid HMAC headers$`, tc.mockgatehubRunningWithHeaders)
	ctx.Step(`^the payload contains a counter currency$`, tc.payloadHasCounterCurrency)
	ctx.Step(`^at least one currency rate entry besides the counter field$`, tc.rateEntryExists)
	ctx.Step(`^the response includes a non-empty vaults array$`, tc.responseHasVaults)

	// Card steps
	ctx.Step(`^a managed user with KYC state "([^"]*)"$`, tc.existingManagedUserWithKYC)
	ctx.Step(`^I POST (.+) with managed user UUID header, nameOnCard "([^"]*)", account productCode "([^"]*)", currency "([^"]*)", and card productCode "([^"]*)"$`, tc.postCardCustomerFromString)
	ctx.Step(`^the response contains a customer with id, sourceId, type "([^"]*)", and kycStatus "([^"]*)"$`, tc.responseHasCustomer)
	ctx.Step(`^the customer has one account with currency "([^"]*)", status "([^"]*)", and type "([^"]*)"$`, tc.customerHasAccount)
	ctx.Step(`^the account has one card with status "([^"]*)" and nameOnCard "([^"]*)"$`, tc.accountHasCardHelper)
	ctx.Step(`^I GET (.+) with managed user UUID header$`, tc.getWithManagedUserHeader)
	ctx.Step(`^the response has a data array with at least one card$`, tc.responseHasCardData)
	ctx.Step(`^each card has id, accountId, customerId, nameOnCard, maskedPan, status, and productCode$`, tc.eachCardHasFields)
	ctx.Step(`^the response includes pagination with pageNumber, pageSize, and totalPages$`, tc.responsePaginated)
	ctx.Step(`^the response contains the card object with all required fields$`, tc.responseHasCardObject)
	ctx.Step(`^the card status is "([^"]*)"$`, tc.cardStatusIs)
	ctx.Step(`^I PUT (.+) with managed user UUID header and note "([^"]*)"$`, tc.putWithNote)
	ctx.Step(`^the card status is changed to "([^"]*)"$`, tc.cardStatusChangedTo)
	ctx.Step(`^the card status is changed back to "([^"]*)"$`, tc.cardStatusChangedBackTo)

	// Card lifecycle setup steps
	ctx.Step(`^a managed customer with a card exists$`, tc.managedCustomerWithCard)
	ctx.Step(`^a managed customer with a locked card exists$`, tc.managedCustomerWithLockedCard)
	ctx.Step(`^a managed customer with a card and transaction exists$`, tc.managedCustomerWithCardAndTransaction)
	ctx.Step(`^a managed customer with a pending 3DS challenge exists$`, tc.managedCustomerWithPending3DS)

	// Card token and transaction steps
	ctx.Step(`^I POST (.+) with cardId and managed user UUID header$`, tc.postCardTokenWithManagedUser)
	ctx.Step(`^the response contains a links array with at least one entry$`, tc.responseContainsLinksArray)
	ctx.Step(`^I PUT (.+) with managed user UUID header and updated dailyOverall limit (\d+)$`, tc.putCardLimitsWithUpdate)
	ctx.Step(`^I POST (.+) with cardId, amount "([^"]*)", currency "([^"]*)", and managed user UUID header$`, tc.postCardTransaction)
	ctx.Step(`^the response contains a card transaction with transactionId and ghResponseCode "([^"]*)"$`, tc.responseContainsCardTransactionWithGH)
	ctx.Step(`^the response contains the card transaction with transactionId and transactionAmount$`, tc.responseContainsCardTransactionDetails)
	ctx.Step(`^the response contains pending 3DS confirmations$`, tc.responseContainsPendingConfirmations)

	// Additional card steps
	ctx.Step(`^I DELETE (.+) with managed user UUID header$`, tc.deleteWithManagedUserHeader)
	ctx.Step(`^I POST (.+) with managed user UUID header$`, tc.postWithManagedUserHeader)
	ctx.Step(`^the response status is (\d+) with status "([^"]*)"$`, tc.responseStatusWithStatus)
	ctx.Step(`^the response contains a token starting with "([^"]*)"$`, tc.responseContainsTokenStarting)
	ctx.Step(`^the response contains pending (\d+)DS confirmations$`, tc.responseContainsPending3DSConfirmations)
	ctx.Step(`^each confirmation includes transactionId, merchantName, purchaseAmount, purchaseCurrency, and timeout$`, tc.eachConfirmationHasRequiredFields)
	ctx.Step(`^I POST (.+) with managed user UUID header, confirmed (true|false), authMethod "([^"]*)"$`, tc.postWith3DSConfirmation)
	ctx.Step(`^the response indicates success with status "([^"]*)"$`, tc.responseIndicatesSuccess)
	ctx.Step(`^the response indicates declined with status "([^"]*)"$`, tc.responseIndicatesDeclined)
	ctx.Step(`^the response contains a data array with card product objects$`, tc.responseHasCardProducts)
	ctx.Step(`^the response contains orderId, cardId, status "([^"]*)", and type "([^"]*)"$`, tc.responseContainsPlasticCardOrder)
	ctx.Step(`^the response is an array of card limit objects$`, tc.responseIsArrayOfCardLimits)
	ctx.Step(`^each limit has type, limit amount, currency "([^"]*)", and isDisabled flag$`, tc.eachLimitHasRequiredFields)
	ctx.Step(`^the limits include types: dailyOverall, perTransaction, monthlyOverall, dailyAtm, dailyEcomm$`, tc.limitsIncludeRequiredTypes)
	ctx.Step(`^the response returns the updated limits array$`, tc.responseReturnsUpdatedLimits)
	ctx.Step(`^the dailyOverall limit is changed to (\d+)$`, tc.dailyOverallLimitChanged)
	ctx.Step(`^the card no longer appears in list cards query \(filtered out\)$`, tc.cardNotInListCardsQuery)

	// Transaction steps
	ctx.Step(`^a managed user with at least one wallet address$`, tc.managedUserWithWalletAddress)
	ctx.Step(`^authenticated requests signed with HMAC headers$`, tc.authenticatedRequestsWithHMAC)
	ctx.Step(`^an iframe token obtained with scope \["([^"]*)"\] and mapped to the managed user$`, tc.iframeTokenWithScope)
	ctx.Step(`^I POST (.+) with amount "([^"]*)" and currency "([^"]*)" and Authorization header "Bearer (.+)"$`, tc.postTransactionWithAuth)
	ctx.Step(`^the user balance for EUR increases by (\d+)\.(\d+)$`, tc.userBalanceIncreasesBy)
	ctx.Step(`^I POST (.+) with user_id, amount (\d+)\.(\d+), currency "([^"]*)", type (\d+), and deposit_type "([^"]*)"$`, tc.postHostedTransfer)
	ctx.Step(`^the response includes an id or uuid$`, tc.responseHasIDOrUUID)
	ctx.Step(`^the amount echoes (\d+)\.(\d+)$`, tc.amountEchoes)
	ctx.Step(`^the status is completed \(integer (\d+) or string "([^"]*)"\)$`, tc.statusIsCompleted)
	ctx.Step(`^deposit_type is "([^"]*)"$`, tc.depositTypeIs)
	ctx.Step(`^a core\.deposit\.completed webhook is emitted for the user$`, tc.coreDepositWebhookEmitted)
	ctx.Step(`^I POST (.+) with type (\d+), deposit_type "([^"]*)", receiving_address set to a wallet, amount (\d+)\.(\d+), currency "([^"]*)", and a valid vault_uuid$`, tc.postExternalDeposit)
	ctx.Step(`^fields amount, total_amount, and fee are string formatted with two decimals$`, tc.fieldsAreStringFormatted)
	ctx.Step(`^status is integer (\d+)$`, tc.statusIsInteger)
	ctx.Step(`^the transaction can be retrieved via GET \/core\/v(\d+)\/transactions\/{id} with the same format$`, tc.transactionCanBeRetrievedFormatted)

	// Fee configuration steps
	ctx.Step(`^deposit fee is configured to ([\d.]+)%$`, tc.depositFeeConfigured)
	ctx.Step(`^withdrawal fee is configured to ([\d.]+)%$`, tc.withdrawalFeeConfigured)
	ctx.Step(`^I GET /admin/fees without authentication$`, tc.getAdminFeesWithoutAuth)
	ctx.Step(`^I PUT /admin/fees without authentication and body (.+)$`, tc.putAdminFeesWithoutAuth)
	ctx.Step(`^I PUT /admin/fees without any HMAC headers and body (.+)$`, tc.putAdminFeesWithoutAnyHMAC)
	ctx.Step(`^the deposit fee percentage is ([\d.]+)$`, tc.depositFeePercentageIs)
	ctx.Step(`^the withdrawal fee percentage is ([\d.]+)$`, tc.withdrawalFeePercentageIs)
	ctx.Step(`^the transaction fee is "([^"]*)"$`, tc.transactionFeeIs)
	ctx.Step(`^the transaction total_amount is "([^"]*)"$`, tc.transactionTotalAmountIs)
	ctx.Step(`^the transaction amount is "([^"]*)"$`, tc.transactionAmountIs)
	ctx.Step(`^I POST /core/v1/transactions with type (\d+), deposit_type "([^"]*)", amount ([\d.]+), currency "([^"]*)", and a valid vault_uuid$`, tc.postDepositWithFeeFields)
	ctx.Step(`^I GET /core/v1/transactions/\{txId\}$`, tc.getTransactionByID)
	ctx.Step(`^I create a withdrawal transaction for ([\d.]+ [A-Z]+)$`, func(amountCurrency string) error {
		amount, currency, err := parseAmountCurrency(amountCurrency)
		if err != nil {
			return err
		}
		return tc.createWithdrawalTransaction(amount, currency)
	})
	ctx.Step(`^I POST /core/v1/transactions with user_id, amount ([\d.]+), currency "([^"]*)", type (\d+), and deposit_type "([^"]*)"$`, tc.postHostedTransferWithFeeFields)

	// User-specific fee configuration steps
	ctx.Step(`^I GET /admin/users/\{userId\}/fees without authentication$`, func() error {
		return tc.getUserFeesWithoutAuth(tc.userID)
	})
	ctx.Step(`^I PUT /admin/users/\{userId\}/fees without authentication and body (.+)$`, func(bodyJSON string) error {
		return tc.setUserFeesWithoutAuth(tc.userID, bodyJSON)
	})
	ctx.Step(`^I DELETE /admin/users/\{userId\}/fees without authentication$`, func() error {
		return tc.clearUserFeesWithoutAuth(tc.userID)
	})
	ctx.Step(`^the (deposit|withdrawal) fee source is "([^"]*)"$`, tc.userFeeSourceIs)
	ctx.Step(`^user-specific deposit fee is set to ([\d.]+)%$`, func(percent float64) error {
		bodyJSON := fmt.Sprintf(`{"deposit_fee_percentage": %.1f}`, percent)
		return tc.setUserFeesWithoutAuth(tc.userID, bodyJSON)
	})

	// Organization configuration steps
	InitializeOrganizationScenario(ctx, tc)
}
