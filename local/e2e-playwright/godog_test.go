package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"
)

var (
	tags        = flag.String("tags", "", "Godog tags expression, e.g. @wip or @phone-debug")
	concurrency = flag.Int("concurrency", 1, "Number of concurrent scenarios (default: 1)")
	reportPath  = flag.String("report", "debug/report.md", "Path for markdown report output")
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
		format = fmt.Sprintf("pretty:%s", *reportPath)
	}

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
