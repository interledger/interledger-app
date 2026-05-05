package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
	enums "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

type workflowArgs struct {
	APIBaseURL string `json:"apiBaseUrl"`
	TwoFAType  string `json:"twoFAType"`
}

var (
	tags                 = flag.String("tags", "", "Godog tags expression, e.g. @wip or @phone-debug")
	concurrency          = flag.Int("concurrency", 1, "Number of concurrent scenarios (default: 1)")
	reportPath           = flag.String("report", "debug/report.json", "Path for cucumber report output")
	temporalWorkflowArgs = workflowArgs{APIBaseURL: "http://backend:8080/webhooks/gatehub", TwoFAType: "totp"}
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
		// Output to both stdout and cucumber JSON report file
		format = fmt.Sprintf("pretty,cucumber:%s", *reportPath)
	}

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
	status := m.Run()
	os.Exit(status)
}

func prerequisite(t *testing.T) {
	// Wait for the backend HTTP server to be healthy before doing anything else.
	if err := waitForHealthEndpoint("http://localhost:8080/healthz", 2*time.Minute); err != nil {
		t.Fatalf("backend not ready: %v", err)
	}

	// The GateHub organization config workflow is only needed for GateHub/EU tests.
	// Skip it when running only Xago or other non-GateHub scenarios.
	if tags != nil && *tags != "" && !needsGateHubPrerequisite(*tags) {
		t.Logf("Skipping GateHub organization config prerequisite (tags: %s)", *tags)
		return
	}

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

	// Wait for the backend Temporal worker to register before starting the workflow.
	// In CI, the backend may still be booting when the test starts.
	waitForWorkers(t, temporalClient, "backend", 1*time.Minute)

	wo := client.StartWorkflowOptions{
		ID:                       uuid.NewString(),
		TaskQueue:                "backend",
		WorkflowExecutionTimeout: 2 * time.Minute,
	}

	run, err := temporalClient.ExecuteWorkflow(t.Context(), wo, "UpdateGateHubOrganizationConfig", temporalWorkflowArgs)
	if err != nil {
		t.Fatalf("failed to execute workflow %v", err)
	}

	err = run.Get(t.Context(), nil)
	if err != nil {
		t.Fatalf("failed to get workflow result %v", err)
	}
}

// needsGateHubPrerequisite returns true if the tag expression includes scenarios
// that require the GateHub organization config workflow.
func needsGateHubPrerequisite(tagExpr string) bool {
	// If tags explicitly target only xago, skip GateHub prerequisite
	lower := strings.ToLower(tagExpr)
	if strings.Contains(lower, "@xago") && !strings.Contains(lower, "@germany") && !strings.Contains(lower, "@gatehub") {
		return false
	}
	// Default: run the prerequisite (safe for mixed or unfiltered runs)
	return true
}

// waitForWorkers polls Temporal until at least one workflow worker is
// registered on the given task queue, or the timeout expires.
func waitForWorkers(t *testing.T, c client.Client, taskQueue string, timeout time.Duration) {
	t.Helper()
	t.Logf("Waiting for workers on task queue %q (timeout: %v)...", taskQueue, timeout)
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := c.DescribeTaskQueue(t.Context(), taskQueue, enums.TASK_QUEUE_TYPE_WORKFLOW)
		if err == nil && len(resp.Pollers) > 0 {
			t.Logf("Found %d worker(s) on task queue %q", len(resp.Pollers), taskQueue)
			return
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("timed out waiting for workers on task queue %q after %v", taskQueue, timeout)
}
