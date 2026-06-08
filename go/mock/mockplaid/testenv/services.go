//go:build e2e
// +build e2e

package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"
)

func startServices() error {
	cmd := exec.Command("docker", "compose", "-f", "docker-compose.yml", "up", "-d", "--build", "--force-recreate")
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker compose up failed: %w\n%s", err, string(output))
	}
	return nil
}

func cleanup() {
	if os.Getenv("KEEP_CONTAINERS") != "" {
		return
	}
	cmd := exec.Command("docker", "compose", "-f", "docker-compose.yml", "down", "-v")
	_ = cmd.Run()
}

func waitForServices() error {
	for i := 0; i < maxWaitSeconds; i++ {
		resp, err := http.Get(mockPlaidURL + "/health")
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
