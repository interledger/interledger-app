//go:build e2e
// +build e2e

package main

import "github.com/cucumber/godog"

func InitializeScenario(ctx *godog.ScenarioContext) {
	tc := newTestContext()

	ctx.Before(tc.beforeScenario)
	ctx.After(tc.afterScenario)
	tc.registerScenarioSteps(ctx)
}
