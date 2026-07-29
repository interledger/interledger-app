package provision

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/interledger/interledger-app/go/performance/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func TestSplitPrefersFundedWalletsAsSenders(t *testing.T) {
	// An unfunded sender has nothing to drain, so funded wallets go first.
	wallets := []Wallet{
		{Label: "a", Balance: 0},
		{Label: "b", Balance: 500},
		{Label: "c", Balance: 0},
		{Label: "d", Balance: 900},
	}

	senders, receivers := Split(wallets, 2)

	require.Len(t, senders, 2)
	assert.Equal(t, "b", senders[0].Label)
	assert.Equal(t, "d", senders[1].Label)
	require.Len(t, receivers, 2)
}

func TestSplitDefaultsToHalf(t *testing.T) {
	wallets := []Wallet{{Label: "a"}, {Label: "b"}, {Label: "c"}, {Label: "d"}}

	senders, receivers := Split(wallets, 0)

	assert.Len(t, senders, 2)
	assert.Len(t, receivers, 2)
}

func TestSplitClampsOversizedSenderCount(t *testing.T) {
	wallets := []Wallet{{Label: "a"}, {Label: "b"}}

	senders, receivers := Split(wallets, 99)

	assert.Len(t, senders, 1, "an out-of-range count falls back to half")
	assert.Len(t, receivers, 1)
}

func TestWriteScenarioOverlayProducesALoadableFile(t *testing.T) {
	// The point of the overlay: it must layer cleanly over a committed scenario.
	wallets := []Wallet{
		{Label: "perf-001", Email: "perf-001@perf.interledger.test", Password: "pw", WalletAddress: "https://ilp.link/perf-001", Balance: 500},
		{Label: "perf-002", Email: "perf-002@perf.interledger.test", Password: "pw", WalletAddress: "https://ilp.link/perf-002"},
	}

	var buf bytes.Buffer
	require.NoError(t, WriteScenarioOverlay(&buf, wallets))

	overlay := buf.String()
	assert.Contains(t, overlay, "Do not commit this file")
	assert.Contains(t, overlay, "perf-001@perf.interledger.test")
	assert.Contains(t, overlay, "https://ilp.link/perf-002")
	assert.NotContains(t, overlay, "session_token", "tokens expire; credentials re-authenticate")

	base := writeTemp(t, "base.yaml", `
target:
  grpc_address: "localhost:8443"
  kratos_url: "http://localhost:4433"
senders: []
receivers: []
`)
	overlayPath := writeTemp(t, "overlay.yaml", overlay)

	cfg, err := config.Load(context.Background(), []string{base, overlayPath})
	require.NoError(t, err, "the generated overlay must load as a scenario")
	require.Len(t, cfg.Senders, 2)
	require.Len(t, cfg.Receivers, 2)
	assert.Equal(t, "perf-001@perf.interledger.test", cfg.Senders[0].Email)
	assert.Equal(t, "https://ilp.link/perf-002", cfg.Receivers[1].WalletAddress)
}

func TestApplyDefaults(t *testing.T) {
	opts := Options{}
	applyDefaults(&opts)

	assert.Equal(t, "perf", opts.Prefix)
	assert.Equal(t, int64(5000), opts.TargetMajor)
	assert.Equal(t, 100, opts.PerCountry)
	assert.NotEmpty(t, opts.Password)
}

func TestRunRejectsZeroCount(t *testing.T) {
	_, err := Run(context.Background(), Options{Countries: []string{"za"}, PerCountry: 0}, &bytes.Buffer{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "per-country must be at least 1")
}
