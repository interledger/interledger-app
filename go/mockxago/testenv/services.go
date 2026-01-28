package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"time"
)

func startServices() error {
	cmd := exec.Command("docker", "compose", "-f", "docker-compose.yml", "up", "-d")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func cleanup() {
	cmd := exec.Command("docker", "compose", "-f", "docker-compose.yml", "down", "-v")
	cmd.Stdout = nil
	cmd.Stderr = nil
	_ = cmd.Run()
}

func waitForServices() error {
	for i := 0; i < maxWaitSeconds; i++ {
		resp, err := http.Get(baseURL + "/health")
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
