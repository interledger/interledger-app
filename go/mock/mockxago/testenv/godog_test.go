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
}
