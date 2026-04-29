package e2e_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cucumber/godog"
)

// TTTContext holds per-scenario state.
type TTTContext struct {
	dbPath  string
	lastOut string
	lastErr string
	lastRC  int
}

// InitializeScenario wires up godog step definitions for each scenario.
func InitializeScenario(ctx *godog.ScenarioContext) {
	var sc *TTTContext

	ctx.Before(func(goCtx context.Context, scenario *godog.Scenario) (context.Context, error) {
		f, err := os.CreateTemp("", "ttt-e2e-*.db")
		if err != nil {
			return goCtx, err
		}
		_ = f.Close()
		sc = &TTTContext{dbPath: f.Name()}
		return goCtx, nil
	})

	ctx.After(func(goCtx context.Context, scenario *godog.Scenario, err error) (context.Context, error) {
		if sc != nil && sc.dbPath != "" {
			if odsErr := writeScenarioODS(sc.dbPath, scenario); odsErr != nil {
				_ = os.Remove(sc.dbPath)
				return goCtx, odsErr
			}
			_ = os.Remove(sc.dbPath)
		}
		return goCtx, nil
	})

	ctx.Step(`^I run "([^"]*)"$`, func(cmd string) error { return sc.iRun(cmd) })
	ctx.Step(`^the exit code is (\d+)$`, func(code int) error { return sc.theExitCodeIs(code) })
	ctx.Step(`^the output contains "([^"]*)"$`, func(s string) error { return sc.theOutputContains(s) })
	ctx.Step(`^the error output contains "([^"]*)"$`, func(s string) error { return sc.theErrorOutputContains(s) })
}

func (sc *TTTContext) iRun(cmdline string) error {
	args := strings.Fields(cmdline)
	if len(args) == 0 {
		return fmt.Errorf("empty command")
	}
	if args[0] != "ttt" {
		return fmt.Errorf("expected command to start with ttt, got %q", args[0])
	}

	cmd := exec.Command(tttBinary, args[1:]...)
	cmd.Env = append(os.Environ(), "TTT_DB="+sc.dbPath)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	sc.lastOut = stdout.String()
	sc.lastErr = stderr.String()
	sc.lastRC = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			sc.lastRC = exitErr.ExitCode()
		} else {
			sc.lastRC = -1
			sc.printDebug(cmdline)
			return err
		}
	}

	sc.printDebug(cmdline)
	return nil
}

func (sc *TTTContext) printDebug(cmdline string) {
	if debugFlag == nil || !*debugFlag {
		return
	}

	fmt.Fprintf(os.Stderr, "\n>> $ %s\n", cmdline)
	if sc.lastRC != 0 {
		fmt.Fprintf(os.Stderr, ">> exit code: %d\n", sc.lastRC)
	}
	if sc.lastOut != "" {
		fmt.Fprint(os.Stderr, sc.lastOut)
		if !strings.HasSuffix(sc.lastOut, "\n") {
			fmt.Fprintln(os.Stderr)
		}
	}
	if sc.lastErr != "" {
		fmt.Fprint(os.Stderr, "\033[31m")
		fmt.Fprint(os.Stderr, sc.lastErr)
		if !strings.HasSuffix(sc.lastErr, "\n") {
			fmt.Fprintln(os.Stderr)
		}
		fmt.Fprint(os.Stderr, "\033[0m")
	}
	if sc.lastOut == "" && sc.lastErr == "" {
		fmt.Fprintln(os.Stderr)
	}
}

func (sc *TTTContext) theExitCodeIs(code int) error {
	if sc.lastRC != code {
		return fmt.Errorf("expected exit code %d, got %d\nstdout:\n%s\nstderr:\n%s", code, sc.lastRC, sc.lastOut, sc.lastErr)
	}
	return nil
}

func (sc *TTTContext) theOutputContains(s string) error {
	if !strings.Contains(sc.lastOut, s) {
		return fmt.Errorf("expected stdout to contain %q\nstdout:\n%s\nstderr:\n%s", s, sc.lastOut, sc.lastErr)
	}
	return nil
}

func (sc *TTTContext) theErrorOutputContains(s string) error {
	if !strings.Contains(sc.lastErr, s) {
		return fmt.Errorf("expected stderr to contain %q\nstdout:\n%s\nstderr:\n%s", s, sc.lastOut, sc.lastErr)
	}
	return nil
}
