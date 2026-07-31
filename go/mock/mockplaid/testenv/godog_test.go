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

var tags = flag.String("tags", "", "Godog tags expression")

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

	// Health
	ctx.Step(`^mockplaid is running$`, tc.mockplaidIsRunning)
	ctx.Step(`^I GET "([^"]*)"$`, tc.iGET)
	ctx.Step(`^the response status is (\d+)$`, tc.responseStatusIs)
	ctx.Step(`^the response field "([^"]*)" equals "([^"]*)"$`, tc.responseFieldEquals)
	ctx.Step(`^the response field "([^"]*)" is present$`, tc.responseFieldPresent)

	// Link flow
	ctx.Step(`^I create a link token for user "([^"]*)"$`, tc.createLinkToken)
	ctx.Step(`^a link token for user "([^"]*)"$`, tc.createLinkToken)
	ctx.Step(`^I select institution "([^"]*)" account "([^"]*)"$`, tc.selectAccount)
	ctx.Step(`^I select institution "([^"]*)" account "([^"]*)" again$`, tc.selectAccount)
	ctx.Step(`^both selected account ids are equal$`, tc.accountIDsEqual)
	ctx.Step(`^the selected account ids differ$`, tc.accountIDsDiffer)
	ctx.Step(`^I exchange the public token$`, tc.exchangePublicToken)
	ctx.Step(`^I resolve the item institution$`, tc.resolveInstitution)
	ctx.Step(`^the institution name is "([^"]*)"$`, tc.institutionNameIs)
}
