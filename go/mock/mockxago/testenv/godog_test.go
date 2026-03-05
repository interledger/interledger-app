//go:build e2e
// +build e2e

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

	ctx.Step(`^the Xago mock service is running$`, tc.xagoMockServiceRunning)
	ctx.Step(`^the environment variables are set:$`, tc.environmentVariablesAreSet)
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

	ctx.Step(`^I use the token to create a sub-account$`, tc.createSubAccountWithToken)
	ctx.Step(`^I use the same token to list currencies$`, tc.listCurrenciesWithToken)

	ctx.Step(`^I create a sub-account with the following details:$`, tc.createSubAccountWithDetails)
	ctx.Step(`^I create a sub-account with only required fields:$`, tc.createSubAccountWithOnlyRequiredFields)
	ctx.Step(`^I attempt to create a sub-account without firstName:$`, tc.attemptCreateSubAccountWithoutFirstName)
	ctx.Step(`^I attempt to create a sub-account without lastName:$`, tc.attemptCreateSubAccountWithoutLastName)
	ctx.Step(`^I attempt to create a sub-account without email:$`, tc.attemptCreateSubAccountWithoutEmail)
	ctx.Step(`^I attempt to create a sub-account without a token$`, tc.attemptCreateSubAccountWithoutToken)
	ctx.Step(`^I attempt to create a sub-account with the invalid token$`, tc.attemptCreateSubAccountWithInvalidToken)

	ctx.Step(`^the sub-account is created with:$`, tc.subAccountIsCreatedWith)
	ctx.Step(`^the response includes bank deposit details for ZAR$`, tc.responseIncludesBankDetailsForZAR)
	ctx.Step(`^the response includes bank deposit details for USD$`, tc.responseIncludesBankDetailsForUSD)
	ctx.Step(`^the response includes beneficiaries with deposit references$`, tc.responseIncludesBeneficiariesWithDepositRefs)
	ctx.Step(`^a new sub-account is created$`, tc.newSubAccountIsCreated)
	ctx.Step(`^the sub-account has the provided email address$`, tc.subAccountHasProvidedEmail)
	ctx.Step(`^the beneficiaries in the response include:$`, tc.beneficiariesInclude)
	ctx.Step(`^each currency has a unique deposit reference$`, tc.depositReferencesAreUnique)

	ctx.Step(`^I have created a sub-account for wallet "([^"]*)"$`, tc.createSubAccountForWallet)
	ctx.Step(`^I update the sub-account with new details:$`, tc.updateSubAccountWithDetails)
	ctx.Step(`^the sub-account is updated with the new verification URL$`, tc.subAccountUpdatedWithVerificationURL)
	ctx.Step(`^the response contains updated status confirmation$`, tc.responseContainsUpdatedStatus)
	ctx.Step(`^I attempt to update a sub-account with invalid ID "([^"]*)"$`, tc.attemptUpdateSubAccountInvalidID)

	ctx.Step(`^I have created two sub-accounts for different wallets$`, tc.createTwoSubAccountsDifferentWallets)
	ctx.Step(`^I have created sub-accounts for two different wallets$`, tc.createTwoSubAccountsDifferentWallets)
	ctx.Step(`^I retrieve sub-account information for "([^"]*)"$`, tc.retrieveSubAccountInfoForWallet)
	ctx.Step(`^I get the correct sub-account associated with "([^"]*)"$`, tc.correctSubAccountAssociated)
	ctx.Step(`^I do not get sub-accounts from other wallets$`, tc.subAccountIsolationConfirmed)

	ctx.Step(`^I request the list of available currencies$`, tc.requestCurrencyList)
	ctx.Step(`^the response includes at least (\d+) currencies:$`, tc.responseIncludesAtLeastCurrencies)
	ctx.Step(`^the ZAR currency includes:$`, tc.zarCurrencyIncludes)
	ctx.Step(`^the USD currency includes:$`, tc.usdCurrencyIncludes)
	ctx.Step(`^I have retrieved the currency list$`, tc.storeCurrencyList)
	ctx.Step(`^I request the currency list again$`, tc.requestCurrencyListAgain)
	ctx.Step(`^the response is identical to the previous response$`, tc.responseIdenticalToPrevious)
	ctx.Step(`^all account numbers and bank codes remain the same$`, tc.accountNumbersAndBankCodesSame)
	ctx.Step(`^I request the list of available currencies without authentication$`, tc.requestCurrencyListWithoutAuth)
	ctx.Step(`^the response includes available currencies$`, tc.responseIncludesAvailableCurrencies)

	// Nested currency format steps (backend compatibility)
	ctx.Step(`^the response includes currencies in nested format$`, tc.responseIncludesCurrenciesInNestedFormat)
	ctx.Step(`^the ZAR currency has nested banking providers with:$`, tc.zarCurrencyHasNestedBankingProvidersWith)
	ctx.Step(`^the first ZAR banking provider includes:$`, tc.firstZarBankingProviderIncludes)
	ctx.Step(`^the first ZAR provider deposit fields include:$`, tc.firstZarProviderDepositFieldsInclude)
	ctx.Step(`^the USD currency has nested banking providers$`, tc.usdCurrencyHasNestedBankingProviders)
	ctx.Step(`^the response structure matches backend expectations$`, tc.responseStructureMatchesBackendExpectations)
	ctx.Step(`^each currency has required fields:$`, tc.eachCurrencyHasRequiredFields)
	ctx.Step(`^each banking provider has required fields:$`, tc.eachBankingProviderHasRequiredFields)

	ctx.Step(`^I have created a sub-account$`, tc.createSubAccount)
	ctx.Step(`^I retrieve the created sub-account details$`, tc.retrieveCreatedSubAccountDetails)
	ctx.Step(`^the bankDepositDetails in the sub-account match the currencies endpoint$`, tc.bankDetailsMatchCurrenciesEndpoint)
	ctx.Step(`^the ZAR bank details match exactly$`, tc.zarBankDetailsMatchExactly)
	ctx.Step(`^the USD bank details match exactly$`, tc.usdBankDetailsMatchExactly)

	ctx.Step(`^I have created a sub-account for wallet with ID "([^"]*)"$`, tc.createSubAccountForWalletID)
	ctx.Step(`^I retrieve the sub-account details$`, tc.retrieveSubAccountDetails)
	ctx.Step(`^the ZAR deposit reference contains "([^"]*)"$`, tc.zarDepositReferenceContains)
	ctx.Step(`^the USD deposit reference contains "([^"]*)"$`, tc.usdDepositReferenceContains)
	ctx.Step(`^I have created two sub-accounts$`, tc.createTwoSubAccounts)
	ctx.Step(`^I retrieve both sub-account details$`, tc.retrieveBothSubAccounts)
	ctx.Step(`^the deposit references are different$`, tc.depositReferencesAreDifferent)
	ctx.Step(`^deposit reference for wallet_aaa is unique$`, tc.depositReferenceWalletAAAUnique)
	ctx.Step(`^deposit reference for wallet_bbb is unique$`, tc.depositReferenceWalletBBBUnique)

	ctx.Step(`^I request the balance for the sub-account$`, tc.requestBalanceForSubAccount)
	ctx.Step(`^the balance response includes:$`, tc.balanceResponseIncludes)
	ctx.Step(`^the balance includes ZAR currency with:$`, tc.balanceIncludesZAR)
	ctx.Step(`^the balance includes USD currency with:$`, tc.balanceIncludesUSD)
	ctx.Step(`^I request the balance for an invalid account ID "([^"]*)"$`, tc.requestBalanceForInvalidAccount)
	ctx.Step(`^I request the balance for a sub-account without authentication$`, tc.requestBalanceWithoutAuth)
	ctx.Step(`^the sub-account has:$`, tc.setBalanceForSubAccount)
	ctx.Step(`^the sub-account has a balance of:$`, tc.setBalanceForSubAccount)
	ctx.Step(`^the sub-account starts with zero balance$`, tc.subAccountStartsWithZeroBalance)
	ctx.Step(`^a deposit of ([0-9.]+) (ZAR|USD) is received and processed$`, tc.depositReceivedAndProcessed)
	ctx.Step(`^a transfer of ([0-9.]+) (ZAR|USD) is initiated and completed$`, tc.transferInitiatedAndCompleted)
	ctx.Step(`^the balance response shows:$`, tc.balanceResponseShows)
	ctx.Step(`^the available (ZAR|USD) balance is reduced to ([0-9.]+)$`, tc.availableBalanceIs)
	ctx.Step(`^the total (ZAR|USD) balance is reduced to ([0-9.]+)$`, tc.totalBalanceIs)
	ctx.Step(`^the (ZAR|USD) balance is ([0-9.]+)$`, tc.availableBalanceIs)
	ctx.Step(`^the total (ZAR|USD) balance is ([0-9.]+)$`, tc.totalBalanceIs)
	ctx.Step(`^the balances are independent$`, tc.balancesAreIndependent)
	ctx.Step(`^I deposit ([0-9.]+) (ZAR|USD) to (wallet_[^ ]+)$`, tc.depositToWallet)
	ctx.Step(`^the balance for (wallet_[^ ]+) is ([0-9.]+)$`, tc.balanceForWalletIs)

	// Integration workflow steps
	ctx.Step(`^the wallet backend is configured to use Xago$`, tc.walletBackendConfigured)
	ctx.Step(`^I submit KYC for wallet "([^"]*)" with name "([^"]*)" "([^"]*)"$`, tc.submitKYCForWallet)
	ctx.Step(`^the KYC submission is accepted$`, tc.kycSubmissionIsAccepted)
	ctx.Step(`^the sub-account for "([^"]*)" is retrievable$`, tc.subAccountForWalletIsRetrievable)
	ctx.Step(`^I create the following test transactions:$`, tc.createFollowingTestTransactions)
	ctx.Step(`^the transaction history contains at least (\d+) records$`, tc.transactionHistoryContainsAtLeast)

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

	// Transaction management steps
	registerTransactionSteps(ctx, tc)
	registerDepositSteps(ctx, tc)
}
