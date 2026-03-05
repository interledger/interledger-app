//go:build e2e

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var (
	tags = flag.String("tags", "", "Godog tags expression, e.g. @wip or @phone-debug")
)

func TestFeatures(t *testing.T) {
	if err := startServices(); err != nil {
		t.Fatalf("Failed to start services: %v", err)
	}
	defer cleanup()

	if err := waitForServices(); err != nil {
		t.Fatalf("Services failed to start: %v", err)
	}

	opts := godog.Options{
		Output:   colors.Colored(os.Stdout),
		TestingT: t,
		Format:   "pretty",
		Paths:    []string{"../features"},
		Tags:     *tags,
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

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		// Ensure backend state is clean before each scenario
		if err := tc.resetBackend(); err != nil {
			return ctx, fmt.Errorf("failed to reset backend: %w", err)
		}
		// Also re-run client-side reset just in case
		tc.Reset()
		return ctx, nil
	})

	// Background steps
	ctx.Step(`^the Xago mock service is running$`, tc.xagoMockServiceRunning)
	ctx.Step(`^the environment variables are set:$`, tc.environmentVariablesAreSet)

	// Authentication-related steps
	ctx.Step(`^I have obtained a valid access token$`, tc.obtainValidAccessToken)
	ctx.Step(`^I have obtained an access token that is about to expire$`, tc.obtainExpiringAccessToken)
	ctx.Step(`^I have not obtained an access token$`, tc.clearAccessToken)
	ctx.Step(`^I do not have a valid access token$`, tc.clearAccessToken)
	ctx.Step(`^I have an invalid access token "([^"]*)"$`, tc.useInvalidAccessToken)

	ctx.Step(`^I request a login token with valid credentials$`, tc.requestLoginTokenWithValidCredentials)
	ctx.Step(`^I request a login token with invalid credentials$`, tc.requestLoginTokenWithInvalidCredentials)
	ctx.Step(`^I request a login token with missing fields$`, tc.requestLoginTokenWithMissingFields)
	ctx.Step(`^I request a new login token with valid credentials$`, tc.requestNewLoginTokenWithValidCredentials)
	ctx.Step(`^I attempt to use the expired token$`, tc.attemptToUseExpiredToken)

	ctx.Step(`^I receive a successful response with status code (\d+)$`, tc.responseStatusIs)
	ctx.Step(`^I receive an error response with status code (\d+)$`, tc.responseStatusIs)
	ctx.Step(`^the request succeeds with status code (\d+)$`, tc.responseStatusIs)
	ctx.Step(`^the response contains a valid access token$`, tc.responseContainsValidAccessToken)
	ctx.Step(`^the token expires in 55 minutes$`, tc.tokenExpiresIn55Minutes)
	ctx.Step(`^the error message is "([^"]*)"$`, tc.errorMessageIs)
	ctx.Step(`^the error message contains "([^"]*)"$`, tc.errorMessageContains)
	ctx.Step(`^the new token is different from the expired token$`, tc.newTokenIsDifferentFromExpired)
	ctx.Step(`^the new token is valid$`, tc.newTokenIsValid)

	ctx.Step(`^I use the token to call a protected route$`, tc.callProtectedRouteWithToken)
	ctx.Step(`^I use the same token to call a protected route$`, tc.callProtectedRouteWithToken)

	// Sub-account creation steps
	ctx.Step(`^I create a sub-account with the following details:$`, tc.createSubAccountWithDetails)
	ctx.Step(`^I create a sub-account with only required fields:$`, tc.createSubAccountWithDetails)
	ctx.Step(`^I attempt to create a sub-account without firstName:$`, tc.attemptCreateSubAccountWithoutFirstName)
	ctx.Step(`^I attempt to create a sub-account without lastName:$`, tc.attemptCreateSubAccountWithoutLastName)
	ctx.Step(`^I attempt to create a sub-account without email:$`, tc.attemptCreateSubAccountWithoutEmail)

	// Sub-account assertion steps
	ctx.Step(`^the sub-account is created with:$`, tc.subAccountIsCreatedWith)
	ctx.Step(`^the response includes bank deposit details for ZAR$`, tc.responseIncludesBankDetailsForZAR)
	ctx.Step(`^the response includes bank deposit details for USD$`, tc.responseIncludesBankDetailsForUSD)
	ctx.Step(`^the response includes beneficiaries with deposit references$`, tc.responseIncludesBeneficiariesWithDepositRefs)
	ctx.Step(`^a new sub-account is created$`, tc.newSubAccountIsCreated)
	ctx.Step(`^the sub-account has the provided email address$`, tc.subAccountHasProvidedEmail)
	ctx.Step(`^the beneficiaries in the response include:$`, tc.beneficiariesInclude)
	ctx.Step(`^each currency has a unique deposit reference$`, tc.depositReferencesAreUnique)

	// Sub-account update steps
	ctx.Step(`^I have created a sub-account for wallet "([^"]*)"$`, tc.createSubAccountForWallet)
	ctx.Step(`^I update the sub-account with new details:$`, tc.updateSubAccountWithDetails)
	ctx.Step(`^the sub-account is updated with the new verification URL$`, tc.subAccountUpdatedWithVerificationURL)
	ctx.Step(`^the response contains updated status confirmation$`, tc.responseContainsUpdatedStatus)
	ctx.Step(`^I attempt to update a sub-account with invalid ID "([^"]*)"$`, tc.attemptUpdateSubAccountInvalidID)

	// Sub-account wallet association steps
	ctx.Step(`^I have created two sub-accounts for different wallets$`, tc.createTwoSubAccountsDifferentWallets)
	ctx.Step(`^I retrieve sub-account information for "([^"]*)"$`, tc.retrieveSubAccountInfoForWallet)
	ctx.Step(`^I get the correct sub-account associated with "([^"]*)"$`, tc.correctSubAccountAssociated)
	ctx.Step(`^I do not get sub-accounts from other wallets$`, tc.subAccountIsolationConfirmed)

	// Beneficiary management steps
	ctx.Step(`^I add a beneficiary with the following details:$`, tc.addBeneficiaryWithDetails)
	ctx.Step(`^I add a beneficiary with only required fields:$`, tc.addBeneficiaryWithRequiredFieldsOnly)
	ctx.Step(`^I add a beneficiary with specific details:$`, tc.addBeneficiaryWithSpecificDetails)
	ctx.Step(`^I attempt to add a beneficiary without name:$`, tc.attemptAddBeneficiaryWithoutName)
	ctx.Step(`^I attempt to add a beneficiary without account number:$`, tc.attemptAddBeneficiaryWithoutAccountNumber)
	ctx.Step(`^I attempt to add a beneficiary without authentication$`, tc.attemptAddBeneficiaryWithoutAuthentication)
	ctx.Step(`^I attempt to list beneficiaries without authentication$`, tc.attemptListBeneficiariesWithoutAuthentication)

	ctx.Step(`^I list the beneficiaries$`, tc.listBeneficiaries)
	ctx.Step(`^I request to list beneficiaries with limit (\d+)$`, tc.requestListWithLimit)
	ctx.Step(`^I request to list beneficiaries with limit (\d+) and page (\d+)$`, tc.requestListWithLimitAndPage)
	ctx.Step(`^I list beneficiaries for wallet_aaa$`, tc.listBeneficiariesForWalletAAA)

	ctx.Step(`^I have added a beneficiary with status "([^"]*)"$`, tc.haveAddedBeneficiaryWithStatus)
	ctx.Step(`^I have added (\d+) beneficiaries to the sub-account$`, tc.haveAddedNBeneficiariesToSubAccount)
	ctx.Step(`^I have added (\d+) beneficiaries$`, tc.haveAddedNBeneficiaries)
	ctx.Step(`^I have created sub-accounts for two wallets$`, tc.haveCreatedSubAccountsForTwoWallets)
	ctx.Step(`^I have added beneficiaries to wallet_aaa$`, tc.haveAddedBeneficiariesToWalletAAA)

	ctx.Step(`^I wait (\d+) seconds for auto-approval$`, tc.waitNSecondsForAutoApproval)

	ctx.Step(`^the response includes a beneficiary with:$`, tc.responseIncludesBeneficiaryWith)
	ctx.Step(`^the beneficiary is created with status "([^"]*)"$`, tc.beneficiaryCreatedWithStatus)
	ctx.Step(`^the beneficiary is created$`, tc.beneficiaryIsCreated)
	ctx.Step(`^the beneficiary status has transitioned to "([^"]*)"$`, tc.beneficiaryStatusTransitionedTo)
	ctx.Step(`^the response includes (\d+) beneficiaries$`, tc.responseIncludesNBeneficiaries)
	ctx.Step(`^each beneficiary includes required fields$`, tc.eachBeneficiaryHasRequiredFields)
	ctx.Step(`^the pagination shows numberOfPages is (\d+)$`, tc.paginationShowsNumberOfPages)
	ctx.Step(`^the UUIDs are different$`, tc.uuidsAreDifferent)
	ctx.Step(`^each UUID is unique$`, tc.eachUUIDIsUnique)
	ctx.Step(`^I get the correct beneficiaries$`, tc.getCorrectBeneficiaries)
	ctx.Step(`^I do not get beneficiaries from wallet_bbb$`, tc.doNotGetBeneficiariesFromWalletBBB)
	ctx.Step(`^the beneficiary details match exactly:$`, tc.beneficiaryDetailsMatchExactly)
}
