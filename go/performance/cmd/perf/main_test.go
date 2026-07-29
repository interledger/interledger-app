package main

import (
	"testing"
	"time"

	"github.com/interledger/interledger-app/go/performance/config"

	"github.com/stretchr/testify/assert"
)

func TestApplyOverridesReplacesScenarioValues(t *testing.T) {
	cfg := &config.Config{}
	cfg.Target.GRPCAddress = "localhost:8443"
	cfg.Run.ArrivalRate = 10
	cfg.Run.Duration = time.Minute
	cfg.Run.CountPerSender = 5
	cfg.Run.Stop = config.StopDrain
	cfg.Run.Pairing = config.PairingIndex
	cfg.Settlement.Track = true

	applyOverrides(cfg, overrides{
		target:       "127.0.0.1:9443",
		rate:         50,
		duration:     5 * time.Minute,
		count:        20,
		stopMode:     "count",
		pairing:      "fan_in",
		metricsAddr:  ":9999",
		noSettlement: true,
	})

	assert.Equal(t, "127.0.0.1:9443", cfg.Target.GRPCAddress)
	assert.Equal(t, 50.0, cfg.Run.ArrivalRate)
	assert.Equal(t, 5*time.Minute, cfg.Run.Duration)
	assert.Equal(t, 20, cfg.Run.CountPerSender)
	assert.Equal(t, config.StopCount, cfg.Run.Stop)
	assert.Equal(t, config.PairingFanIn, cfg.Run.Pairing)
	assert.Equal(t, ":9999", cfg.Metrics.Listen)
	assert.False(t, cfg.Settlement.Track)
}

func TestApplyOverridesKeepsScenarioValuesWhenFlagsAreUnset(t *testing.T) {
	// Zero-valued flags must not silently clobber the scenario file.
	cfg := &config.Config{}
	cfg.Target.GRPCAddress = "localhost:8443"
	cfg.Run.ArrivalRate = 10
	cfg.Run.Duration = time.Minute
	cfg.Run.CountPerSender = 5
	cfg.Run.Stop = config.StopDrain
	cfg.Run.Pairing = config.PairingIndex
	cfg.Metrics.Listen = ":9464"
	cfg.Settlement.Track = true

	applyOverrides(cfg, overrides{})

	assert.Equal(t, "localhost:8443", cfg.Target.GRPCAddress)
	assert.Equal(t, 10.0, cfg.Run.ArrivalRate)
	assert.Equal(t, time.Minute, cfg.Run.Duration)
	assert.Equal(t, 5, cfg.Run.CountPerSender)
	assert.Equal(t, config.StopDrain, cfg.Run.Stop)
	assert.Equal(t, config.PairingIndex, cfg.Run.Pairing)
	assert.Equal(t, ":9464", cfg.Metrics.Listen)
	assert.True(t, cfg.Settlement.Track)
}

func TestConfigFlagCollectsRepeatedValues(t *testing.T) {
	// Layering is the mechanism that keeps credentials out of committed scenarios.
	var files configFlag
	assert.NoError(t, files.Set("base.yaml"))
	assert.NoError(t, files.Set("overlay.yaml"))

	assert.Equal(t, configFlag{"base.yaml", "overlay.yaml"}, files)
	assert.Equal(t, "base.yaml,overlay.yaml", files.String())
}
