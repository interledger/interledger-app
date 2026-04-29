package e2e_test

import (
	"fmt"
	"path/filepath"
	"strings"

	"ttt/engine/sqlite"
	"ttt/ods"

	"github.com/cucumber/godog"
)

func writeScenarioODS(dbPath string, scenario *godog.Scenario) error {
	store, err := sqlite.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open scenario db for ods: %w", err)
	}
	defer func() { _ = store.Close() }()

	path := scenarioODSPath(scenario)
	if err := ods.ExportStore(store, path); err != nil {
		return fmt.Errorf("write ods %s: %w", path, err)
	}
	return nil
}

func scenarioODSPath(scenario *godog.Scenario) string {
	group := "generic"
	if scenario != nil && scenario.Uri != "" {
		parts := strings.Split(filepath.ToSlash(scenario.Uri), "/")
		for i, part := range parts {
			if part == "features" && i+1 < len(parts)-1 {
				group = parts[i+1]
				break
			}
		}
	}

	name := "scenario"
	if scenario != nil && scenario.Name != "" {
		name = scenario.Name
	}
	return filepath.Join("..", "output", ods.Slug(group)+"__"+ods.Slug(name)+".ods")
}
