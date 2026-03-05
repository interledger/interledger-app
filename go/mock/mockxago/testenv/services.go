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
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// dumpLogs saves all container logs to testenv/lastlogs.txt for post-mortem analysis.
func dumpLogs() {
	cmd := exec.Command("docker", "compose", "-f", "docker-compose.yml", "logs", "--no-color")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to dump container logs: %v\n", err)
		return
	}
	if err := os.WriteFile("lastlogs.txt", output, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to write lastlogs.txt: %v\n", err)
	}
}

func cleanup() {
	dumpLogs()

	if os.Getenv("KEEP_CONTAINERS") != "" {
		fmt.Println("KEEP_CONTAINERS is set — skipping container teardown")
		return
	}

	cmd := exec.Command("docker", "compose", "-f", "docker-compose.yml", "down", "-v")
	cmd.Stdout = nil
	cmd.Stderr = nil
	_ = cmd.Run()
}

func waitForServices() error {
	for i := 0; i < maxWaitSeconds; i++ {
		resp, err := http.Get(mockXagoURL + "/health")
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
