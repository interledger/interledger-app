//go:build e2e
// +build e2e

package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"gitlab.com/fynbos/mock/mockpti/internal/logger"
	"go.uber.org/zap"
)

func startServices() error {
	env, err := webhookCryptoComposeEnv()
	if err != nil {
		return fmt.Errorf("failed to initialize webhook crypto env: %w", err)
	}

	cmd := exec.Command("docker", "compose", "-f", "docker-compose.yml", "up", "-d", "--build", "--force-recreate")
	cmd.Env = append(os.Environ(), env...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		redactedOutput := strings.ReplaceAll(string(output), webhookCryptoState.signingPrivatePEM, "[REDACTED_SIGNING_KEY]")
		return fmt.Errorf("docker compose up failed: %w\n%s", err, redactedOutput)
	}
	return nil
}

func dumpLogs() {
	cmd := exec.Command("docker", "compose", "-f", "docker-compose.yml", "logs", "--no-color")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Warn("failed to dump container logs", zap.Error(err))
		return
	}
	if err := os.WriteFile("lastlogs.txt", output, 0644); err != nil {
		logger.Warn("failed to write lastlogs.txt", zap.Error(err))
	}
}

func cleanup() {
	dumpLogs()

	if os.Getenv("KEEP_CONTAINERS") != "" {
		logger.Info("KEEP_CONTAINERS is set — skipping container teardown")
		return
	}

	cmd := exec.Command("docker", "compose", "-f", "docker-compose.yml", "down", "-v")
	cmd.Stdout = nil
	cmd.Stderr = nil
	_ = cmd.Run()
}

func waitForServices() error {
	for i := 0; i < maxWaitSeconds; i++ {
		resp, err := http.Get(mockPTIURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("health check timed out after %d seconds", maxWaitSeconds)
}
