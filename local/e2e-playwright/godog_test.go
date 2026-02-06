package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"
)

var (
	tags = flag.String("tags", "", "Godog tags expression, e.g. @wip or @phone-debug")
)

// cleanupDebugScreenshots removes all PNG files from the debug directory
func cleanupDebugScreenshots() error {
	debugDir := "debug"
	if _, err := os.Stat(debugDir); os.IsNotExist(err) {
		return nil // Directory doesn't exist, nothing to clean
	}

	files, err := filepath.Glob(filepath.Join(debugDir, "*.png"))
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return err
		}
	}
	return nil
}

func TestFeatures(t *testing.T) {
	flag.Parse()

	// Clean up screenshots from previous test runs
	if err := cleanupDebugScreenshots(); err != nil {
		t.Logf("Warning: failed to cleanup debug screenshots: %v", err)
	}

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
			Tags:     *tags,
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
