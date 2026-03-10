package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
)

type workflowArgs struct {
	APIBaseURL string `json:"apiBaseUrl"`
	TwoFAType  string `json:"twoFAType"`
}

var (
	tags                 = flag.String("tags", "", "Godog tags expression, e.g. @wip or @phone-debug")
	concurrency          = flag.Int("concurrency", 1, "Number of concurrent scenarios (default: 1)")
	reportPath           = flag.String("report", "debug/report.md", "Path for markdown report output")
	temporalWorfklowArgs = workflowArgs{APIBaseURL: "http://backend:8080/webhooks/gatehub", TwoFAType: "totp"}
)

// cleanupDebugScreenshots removes all screenshots from the debug directory
func cleanupDebugScreenshots() error {
	debugDir := "debug"
	if _, err := os.Stat(debugDir); os.IsNotExist(err) {
		return nil // Directory doesn't exist, nothing to clean
	}

	return os.RemoveAll(debugDir)
}

func TestFeatures(t *testing.T) {
	flag.Parse()

	// Clean up screenshots from previous test runs
	if err := cleanupDebugScreenshots(); err != nil {
		t.Logf("Warning: failed to cleanup debug screenshots: %v", err)
	}

	format := "pretty"
	if reportPath != nil && *reportPath != "" {
		reportDir := filepath.Dir(*reportPath)
		if err := os.MkdirAll(reportDir, 0755); err != nil {
			t.Fatalf("failed to create report directory: %v", err)
		}
		// Output to both stdout and report file
		format = fmt.Sprintf("pretty,pretty:%s", *reportPath)
	}

	// Setup before all tests
	prerequisite(t)

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:      format,
			Paths:       []string{"features"},
			TestingT:    t,
			Tags:        *tags,
			Concurrency: *concurrency,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func TestMain(m *testing.M) {
	// Setup before all tests
	status := m.Run()

	// Cleanup after all tests
	os.Exit(status)
}

func prerequisite(t *testing.T) error {
	// We need to update the GateHub organization config before running the tests.
	// This is done by starting the workflow that makes the update.
	// Notice: This was introduced with the SCA requirement, so it is only
	// relevant for GateHub users.
	temporalClient, err := client.Dial(client.Options{
		Namespace: "default",
		HostPort:  "localhost:7233",
	})
	if err != nil {
		t.Fatalf("failed to connect to temporal: %v", err)
	}
	defer temporalClient.Close()

	wo := client.StartWorkflowOptions{
		ID:                       uuid.NewString(),
		TaskQueue:                "backend",
		WorkflowExecutionTimeout: 1 * time.Minute,
	}

	run, err := temporalClient.ExecuteWorkflow(t.Context(), wo, "UpdateGateHubOrganizationConfig", temporalWorfklowArgs)
	if err != nil {
		t.Fatal("failed to execute workflow %w", err)
	}

	err = run.Get(t.Context(), nil)
	if err != nil {
		t.Fatal("failed to get workflow result %w", err)
	}

	return nil
}
