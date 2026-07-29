package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const baseScenario = `
target:
  grpc_address: "localhost:8443"
  kratos_url: "http://localhost:4433"
run:
  amount: 1
senders:
  - email: a@example.com
    password: secret
receivers:
  - wallet_address: https://ilp.link/b
`

func write(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func TestLoadAppliesDefaults(t *testing.T) {
	cfg, err := Load(context.Background(), []string{write(t, "s.yaml", baseScenario)})
	require.NoError(t, err)

	assert.Equal(t, StopDrain, cfg.Run.Stop, "drain is the documented default shape")
	assert.Equal(t, PairingRoundRobin, cfg.Run.Pairing)
	assert.Equal(t, DefaultMaxInFlight, cfg.Run.MaxInFlight)
	assert.Equal(t, DefaultConnections, cfg.Target.Connections)
	assert.Equal(t, DefaultWorkers, cfg.Settlement.Workers)
	assert.Equal(t, DefaultPollInterval, cfg.Settlement.PollInterval)
	assert.Equal(t, DefaultJobLabel, cfg.Metrics.JobLabel)
	assert.Equal(t, "a@example.com", cfg.Senders[0].Label, "label falls back to email")
	assert.Equal(t, "https://ilp.link/b", cfg.Receivers[0].Label)
}

func TestLoadMergesOverlay(t *testing.T) {
	// The core workflow: run shape in a committed file, credentials in a local
	// overlay that replaces the empty sender and receiver lists.
	base := write(t, "base.yaml", `
target:
  grpc_address: "localhost:8443"
  kratos_url: "http://localhost:4433"
run:
  stop: count
  count_per_sender: 5
  arrival_rate: 25
senders: []
receivers: []
`)
	overlay := write(t, "overlay.yaml", `
senders:
  - email: one@example.com
    password: pw
  - email: two@example.com
    password: pw
receivers:
  - wallet_address: https://ilp.link/x
`)

	cfg, err := Load(context.Background(), []string{base, overlay})
	require.NoError(t, err)

	assert.Equal(t, StopCount, cfg.Run.Stop)
	assert.Equal(t, 5, cfg.Run.CountPerSender)
	assert.Len(t, cfg.Senders, 2)
	assert.Len(t, cfg.Receivers, 1)
	assert.Equal(t, 25, cfg.Run.Burst, "burst defaults to one second of arrivals")
}

func TestLoadExpandsWalletsIntoSendersAndReceivers(t *testing.T) {
	cfg, err := Load(context.Background(), []string{write(t, "s.yaml", `
target:
  grpc_address: "localhost:8443"
  kratos_url: "http://localhost:4433"
wallets:
  - label: za-001
    email: za-001@perf.interledger.test
    password: secret
    wallet_address: https://local.ilp.link/za-001
`)})
	require.NoError(t, err)

	require.Len(t, cfg.Senders, 1)
	assert.Equal(t, "za-001", cfg.Senders[0].Label)
	assert.Equal(t, "za-001@perf.interledger.test", cfg.Senders[0].Email)
	assert.Equal(t, "secret", cfg.Senders[0].Password)

	require.Len(t, cfg.Receivers, 1)
	assert.Equal(t, "za-001", cfg.Receivers[0].Label)
	assert.Equal(t, "https://local.ilp.link/za-001", cfg.Receivers[0].WalletAddress)
}

func TestLoadParsesDurations(t *testing.T) {
	cfg, err := Load(context.Background(), []string{write(t, "s.yaml", baseScenario+`
settlement:
  track: true
  timeout: 90s
  poll_interval: 250ms
`)})
	require.NoError(t, err)

	assert.Equal(t, 90*time.Second, cfg.Settlement.Timeout)
	assert.Equal(t, 250*time.Millisecond, cfg.Settlement.PollInterval)
}

func TestLoadRejectsInvalidScenarios(t *testing.T) {
	tests := map[string]struct {
		scenario string
		wantErr  string
	}{
		"missing grpc address": {
			scenario: `
target:
  kratos_url: "http://localhost:4433"
senders:
  - email: a@example.com
    password: pw
receivers:
  - wallet_address: https://ilp.link/b
`,
			wantErr: "GRPCAddress",
		},
		"no senders": {
			scenario: `
target:
  grpc_address: "localhost:8443"
  kratos_url: "http://localhost:4433"
senders: []
receivers:
  - wallet_address: https://ilp.link/b
`,
			wantErr: "at least one sender is required",
		},
		"sender without credentials": {
			scenario: `
target:
  grpc_address: "localhost:8443"
  kratos_url: "http://localhost:4433"
senders:
  - label: nameless
receivers:
  - wallet_address: https://ilp.link/b
`,
			wantErr: "either session_token or both email and password",
		},
		"index pairing with mismatched lists": {
			scenario: `
target:
  grpc_address: "localhost:8443"
  kratos_url: "http://localhost:4433"
run:
  pairing: index
senders:
  - email: a@example.com
    password: pw
  - email: b@example.com
    password: pw
receivers:
  - wallet_address: https://ilp.link/b
`,
			wantErr: "requires equal list lengths",
		},
		"count mode without a count": {
			scenario: `
target:
  grpc_address: "localhost:8443"
  kratos_url: "http://localhost:4433"
run:
  stop: count
senders:
  - email: a@example.com
    password: pw
receivers:
  - wallet_address: https://ilp.link/b
`,
			wantErr: "count_per_sender must be at least 1",
		},
		"duration mode without a duration": {
			scenario: `
target:
  grpc_address: "localhost:8443"
  kratos_url: "http://localhost:4433"
run:
  stop: duration
senders:
  - email: a@example.com
    password: pw
receivers:
  - wallet_address: https://ilp.link/b
`,
			wantErr: "duration must be positive",
		},
		"unknown stop mode": {
			scenario: `
target:
  grpc_address: "localhost:8443"
  kratos_url: "http://localhost:4433"
run:
  stop: forever
senders:
  - email: a@example.com
    password: pw
receivers:
  - wallet_address: https://ilp.link/b
`,
			wantErr: "Stop",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := Load(context.Background(), []string{write(t, "s.yaml", tt.scenario)})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestLoadAcceptsSessionTokenInsteadOfPassword(t *testing.T) {
	cfg, err := Load(context.Background(), []string{write(t, "s.yaml", `
target:
  grpc_address: "localhost:8443"
  kratos_url: "http://localhost:4433"
senders:
  - label: pre-authed
    session_token: abc123
receivers:
  - wallet_address: https://ilp.link/b
`)})
	require.NoError(t, err)
	assert.Equal(t, "abc123", cfg.Senders[0].SessionToken)
}

func TestLoadRequiresAFile(t *testing.T) {
	_, err := Load(context.Background(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one scenario file")
}
