package e2e_test

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"
)

var (
	debugFlag = flag.Bool("debug", false, "print each executed CLI command with stdout and stderr")
	tagsFlag  = flag.String("tags", "", "godog tag expression, for example @standard&&@xago-to-gatehub")
	tttBinary string
)

func TestMain(m *testing.M) {
	flag.Parse()

	bin := os.Getenv("TTT_BINARY")
	if bin == "" {
		tmp, err := os.MkdirTemp("", "ttt-e2e-*")
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to create temp dir:", err)
			os.Exit(1)
		}
		defer os.RemoveAll(tmp)

		bin = filepath.Join(tmp, "ttt")
		cmd := exec.Command("go", "build", "-o", bin, ".")
		cmd.Dir = ".."
		cmd.Env = append(os.Environ(), "GOWORK=off")
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "failed to build ttt binary:", err)
			os.Exit(1)
		}
	}

	tttBinary = bin
	os.Exit(m.Run())
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			Tags:     *tagsFlag,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
