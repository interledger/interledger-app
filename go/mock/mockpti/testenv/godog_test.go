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
	tags = flag.String("tags", "", "Godog tags expression")
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
		if err := tc.resetBackend(); err != nil {
			return ctx, fmt.Errorf("failed to reset backend: %w", err)
		}
		tc.Reset()
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		return ctx, nil
	})

	// Health steps
	ctx.Step(`^mockpti is running$`, tc.mockptiIsRunning)
	ctx.Step(`^I send a GET request to "([^"]*)"$`, tc.sendGETRequest)
	ctx.Step(`^the response status should be (\d+)$`, tc.responseStatusShouldBe)
	ctx.Step(`^the response body should contain "([^"]*)" as "([^"]*)"$`, tc.responseBodyShouldContainAs)

	// Auth/header steps
	ctx.Step(`^valid PTI headers are present$`, tc.validPTIHeadersArePresent)

	// Token steps
	ctx.Step(`^I POST "([^"]*)" with url and method payload$`, tc.postWithURLAndMethodPayload)
	ctx.Step(`^the response should include "([^"]*)"$`, tc.responseBodyShouldIncludeField)

	// User/KYC steps
	ctx.Step(`^I POST "([^"]*)" with a valid PTI user payload$`, tc.postWithValidPTIUserPayload)
	ctx.Step(`^the response should include a PTI user id$`, tc.responseShouldIncludePTIUserID)
	ctx.Step(`^an existing PTI user in mockpti$`, tc.anExistingPTIUser)
	ctx.Step(`^I GET "([^"]*)"$`, tc.sendGETDynamic)
	ctx.Step(`^the response should include "([^"]*)" equal to "([^"]*)"$`, tc.responseBodyShouldIncludeEqual)
	ctx.Step(`^I POST "([^"]*)" with scenario id "([^"]*)"$`, tc.postAssessmentWithScenarioID)
	ctx.Step(`^the response should include an assessment request id$`, tc.responseShouldIncludeAssessmentRequestID)
	ctx.Step(`^the response should include "assessment"$`, tc.responseShouldIncludeAssessment)

	// Wallet steps
	ctx.Step(`^I POST "([^"]*)" with USD wallet payload$`, tc.postWithUSDWalletPayload)
	ctx.Step(`^the response should include a wallet id$`, tc.responseShouldIncludeWalletID)
	ctx.Step(`^an existing PTI user with at least one wallet in mockpti$`, tc.anExistingPTIUserWithWallet)
	ctx.Step(`^the response should include at least one wallet$`, tc.responseShouldIncludeAtLeastOneWallet)

	// Payment information steps
	ctx.Step(`^I POST "([^"]*)" with bank account payload$`, tc.postWithBankAccountPayload)
	ctx.Step(`^the response should include a payment information id$`, tc.responseShouldIncludePaymentInformationID)

	// Transaction steps
	ctx.Step(`^an existing PTI user with a USD wallet and bank account$`, tc.anExistingPTIUserWithUSDWalletAndBankAccount)
	ctx.Step(`^two PTI users each with a USD wallet$`, tc.twoPTIUsersEachWithUSDWallet)
	ctx.Step(`^I POST "([^"]*)" with a valid deposit payload$`, tc.postWithValidDepositPayload)
	ctx.Step(`^I POST "([^"]*)" with a valid withdrawal payload$`, tc.postWithValidWithdrawalPayload)
	ctx.Step(`^I POST "([^"]*)" with a valid transfer payload$`, tc.postWithValidTransferPayload)
	ctx.Step(`^the response should include a transaction request id$`, tc.responseShouldIncludeTransactionRequestID)
	ctx.Step(`^an existing PTI transaction request id$`, tc.anExistingPTITransactionRequestID)
	ctx.Step(`^I POST "([^"]*)" with feedback payload$`, tc.postWithFeedbackPayload)
	ctx.Step(`^the response should include an update id$`, tc.responseShouldIncludeUpdateID)
}
