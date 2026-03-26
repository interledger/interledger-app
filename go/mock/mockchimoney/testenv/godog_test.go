//go:build e2e
// +build e2e

package main

import (
	"flag"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var tags = flag.String("tags", "~@wip", "Godog tags expression")

func TestFeatures(t *testing.T) {
	opts := godog.Options{
		Output:   colors.Colored(os.Stdout),
		Format:   "pretty",
		Paths:    []string{"../features"},
		Tags:     *tags,
		TestingT: t,
	}

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}

	if suite.Run() != 0 {
		t.Fatalf("one or more feature scenarios failed")
	}
}
