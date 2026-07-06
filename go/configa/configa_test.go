package configa_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/interledger/interledger-app/go/configa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockClient is a simple stub SecretClient for unit tests.
type mockClient struct {
	secrets   map[string]map[string]string // secret name → {key → value}
	callCount map[string]int
	err       error
}

func newMockClient(secrets map[string]map[string]string) *mockClient {
	return &mockClient{
		secrets:   secrets,
		callCount: make(map[string]int),
	}
}

func (m *mockClient) GetSecret(_ context.Context, _, name string) (map[string]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.callCount[name]++
	if data, ok := m.secrets[name]; ok {
		return data, nil
	}
	return nil, configa.ErrSecretNotFound
}

// captureNSClient records the namespace argument passed to GetSecret.
type captureNSClient struct {
	secrets map[string]map[string]string
	lastNS  string
}

func (c *captureNSClient) GetSecret(_ context.Context, namespace, name string) (map[string]string, error) {
	c.lastNS = namespace
	if data, ok := c.secrets[name]; ok {
		return data, nil
	}
	return nil, configa.ErrSecretNotFound
}

// testConfig is the struct used across all tests.
type testConfig struct {
	DatabaseURL string `yaml:"database_url" validate:"required"`
	APIKey      string `yaml:"api_key"`
	Port        int    `yaml:"port"`
}

func writeYAML(t *testing.T, content string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(f, []byte(content), 0600))
	return f
}

func writeYAMLNamed(t *testing.T, name, content string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(f, []byte(content), 0600))
	return f
}

// --- Parse tests ---

func TestParse_ValidYAML(t *testing.T) {
	f := writeYAML(t, `
database_url: postgres://localhost/db
port: 5432
`)
	cfg, err := configa.Parse[testConfig]([]string{f})
	require.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestParse_FileNotFound(t *testing.T) {
	_, err := configa.Parse[testConfig]([]string{"/nonexistent/config.yaml"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read config file")
}

func TestParse_InvalidYAML(t *testing.T) {
	f := writeYAML(t, "key: {unclosed")
	_, err := configa.Parse[testConfig]([]string{f})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid yaml")
}

func TestParse_EmptyFilenames(t *testing.T) {
	_, err := configa.Parse[testConfig]([]string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one filename")
}

// --- Resolve tests ---

func TestResolve_NoTemplates(t *testing.T) {
	f := writeYAML(t, `
database_url: postgres://localhost/db
port: 5432
`)
	client := newMockClient(nil)
	cfg, err := configa.Parse[testConfig]([]string{f}, configa.WithSecretClient(client))
	require.NoError(t, err)

	result, err := cfg.Resolve(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "postgres://localhost/db", result.DatabaseURL)
	assert.Equal(t, 5432, result.Port)
	assert.Empty(t, client.callCount, "k8s client must not be called when no templates are present")
}

func TestResolve_SingleSecret(t *testing.T) {
	f := writeYAML(t, `
database_url: postgres://localhost/db
api_key: '{{ secret "my-secret" "key" }}'
`)
	client := newMockClient(map[string]map[string]string{
		"my-secret": {"key": "super-secret-value"},
	})
	cfg, err := configa.Parse[testConfig]([]string{f}, configa.WithSecretClient(client))
	require.NoError(t, err)

	result, err := cfg.Resolve(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "super-secret-value", result.APIKey)
}

func TestResolve_MultipleSecretsDeduped(t *testing.T) {
	// Both fields reference "db-creds"; the client must be called exactly once.
	f := writeYAML(t, `
database_url: 'postgres://user:{{ secret "db-creds" "password" }}@localhost/db'
api_key: '{{ secret "db-creds" "api_key" }}'
port: 8080
`)
	client := newMockClient(map[string]map[string]string{
		"db-creds": {
			"password": "s3cr3t",
			"api_key":  "key-value",
		},
	})
	cfg, err := configa.Parse[testConfig]([]string{f}, configa.WithSecretClient(client))
	require.NoError(t, err)

	result, err := cfg.Resolve(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "postgres://user:s3cr3t@localhost/db", result.DatabaseURL)
	assert.Equal(t, "key-value", result.APIKey)
	assert.Equal(t, 1, client.callCount["db-creds"], "db-creds should be fetched exactly once")
}

func TestResolve_EmbeddedTemplate(t *testing.T) {
	f := writeYAML(t, `
database_url: 'postgres://user:{{ secret "db-creds" "password" }}@localhost/mydb'
port: 5432
`)
	client := newMockClient(map[string]map[string]string{
		"db-creds": {"password": "p@ssw0rd"},
	})
	cfg, err := configa.Parse[testConfig]([]string{f}, configa.WithSecretClient(client))
	require.NoError(t, err)

	result, err := cfg.Resolve(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "postgres://user:p@ssw0rd@localhost/mydb", result.DatabaseURL)
}

func TestResolve_NestedYAML(t *testing.T) {
	type nestedConfig struct {
		Database struct {
			URL      string `yaml:"url"      validate:"required"`
			Password string `yaml:"password"`
		} `yaml:"database"`
	}
	f := writeYAML(t, `
database:
  url: postgres://localhost/db
  password: '{{ secret "db-secret" "password" }}'
`)
	client := newMockClient(map[string]map[string]string{
		"db-secret": {"password": "nested-pass"},
	})
	cfg, err := configa.Parse[nestedConfig]([]string{f}, configa.WithSecretClient(client))
	require.NoError(t, err)

	result, err := cfg.Resolve(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "nested-pass", result.Database.Password)
}

func TestResolve_SecretFetchError(t *testing.T) {
	f := writeYAML(t, `
database_url: postgres://localhost/db
api_key: '{{ secret "missing-secret" "key" }}'
`)
	client := newMockClient(nil) // no secrets configured → returns ErrSecretNotFound
	cfg, err := configa.Parse[testConfig]([]string{f}, configa.WithSecretClient(client))
	require.NoError(t, err)

	_, err = cfg.Resolve(context.Background())
	require.Error(t, err)
}

func TestResolve_SecretKeyMissing(t *testing.T) {
	// Secret exists but the requested key is absent.
	f := writeYAML(t, `
database_url: postgres://localhost/db
api_key: '{{ secret "my-secret" "nonexistent-key" }}'
`)
	client := newMockClient(map[string]map[string]string{
		"my-secret": {"other-key": "value"},
	})
	cfg, err := configa.Parse[testConfig]([]string{f}, configa.WithSecretClient(client))
	require.NoError(t, err)

	_, err = cfg.Resolve(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent-key")
}

func TestResolve_ValidationFailure(t *testing.T) {
	// database_url has validate:"required" but is empty.
	f := writeYAML(t, `
database_url: ""
port: 5432
`)
	cfg, err := configa.Parse[testConfig]([]string{f})
	require.NoError(t, err)

	_, err = cfg.Resolve(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configa:")
}

func TestResolve_UnknownTemplateFunction(t *testing.T) {
	f := writeYAML(t, `
database_url: postgres://localhost/db
api_key: '{{ unknown "arg" }}'
`)
	client := newMockClient(nil)
	cfg, err := configa.Parse[testConfig]([]string{f}, configa.WithSecretClient(client))
	require.NoError(t, err)

	_, err = cfg.Resolve(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configa:")
}

func TestResolve_NoClientWhenTemplatesPresent(t *testing.T) {
	f := writeYAML(t, `
database_url: postgres://localhost/db
api_key: '{{ secret "my-secret" "key" }}'
`)
	cfg, err := configa.Parse[testConfig]([]string{f}) // no WithSecretClient
	require.NoError(t, err)

	_, err = cfg.Resolve(context.Background())
	require.Error(t, err)
	assert.True(t, errors.Is(err, configa.ErrNoSecretClient))
}

func TestWithNamespace(t *testing.T) {
	f := writeYAML(t, `
database_url: postgres://localhost/db
api_key: '{{ secret "my-secret" "key" }}'
`)
	client := &captureNSClient{
		secrets: map[string]map[string]string{
			"my-secret": {"key": "value"},
		},
	}
	cfg, err := configa.Parse[testConfig]([]string{f},
		configa.WithSecretClient(client),
		configa.WithNamespace("custom-namespace"),
	)
	require.NoError(t, err)

	_, err = cfg.Resolve(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "custom-namespace", client.lastNS)
}

func TestResolve_ArrayValues(t *testing.T) {
	type arrayConfig struct {
		Items []string `yaml:"items"`
	}
	f := writeYAML(t, `
items:
  - plain-value
  - '{{ secret "list-secret" "item" }}'
`)
	client := newMockClient(map[string]map[string]string{
		"list-secret": {"item": "resolved-item"},
	})
	cfg, err := configa.Parse[arrayConfig]([]string{f}, configa.WithSecretClient(client))
	require.NoError(t, err)

	result, err := cfg.Resolve(context.Background())
	require.NoError(t, err)

	require.Len(t, result.Items, 2)
	assert.Equal(t, "plain-value", result.Items[0])
	assert.Equal(t, "resolved-item", result.Items[1])
}

// --- Overlay tests ---

func TestParse_Overlay_ScalarOverride(t *testing.T) {
	// Overlay replaces a top-level scalar value from the base.
	base := writeYAMLNamed(t, "base.yaml", `
database_url: postgres://base/db
port: 5432
`)
	overlay := writeYAMLNamed(t, "overlay.yaml", `
port: 9999
`)
	cfg, err := configa.Parse[testConfig]([]string{base, overlay})
	require.NoError(t, err)

	result, err := cfg.Resolve(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "postgres://base/db", result.DatabaseURL, "base value should be preserved")
	assert.Equal(t, 9999, result.Port, "overlay value should win")
}

func TestParse_Overlay_DeepMerge(t *testing.T) {
	// Overlay changes one key inside a nested map; sibling keys are preserved.
	type nestedCfg struct {
		Database struct {
			URL  string `yaml:"url"  validate:"required"`
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		} `yaml:"database"`
	}
	base := writeYAMLNamed(t, "base.yaml", `
database:
  url: postgres://base/db
  host: base-host
  port: 5432
`)
	overlay := writeYAMLNamed(t, "overlay.yaml", `
database:
  host: overlay-host
`)
	cfg, err := configa.Parse[nestedCfg]([]string{base, overlay})
	require.NoError(t, err)

	result, err := cfg.Resolve(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "postgres://base/db", result.Database.URL, "unset key preserved from base")
	assert.Equal(t, "overlay-host", result.Database.Host, "overlay key wins")
	assert.Equal(t, 5432, result.Database.Port, "unset key preserved from base")
}

func TestParse_Overlay_ThirdFileWins(t *testing.T) {
	// When three files are given, the last one takes final precedence.
	base := writeYAMLNamed(t, "base.yaml", `
database_url: postgres://base/db
port: 1111
`)
	mid := writeYAMLNamed(t, "mid.yaml", `
port: 2222
`)
	top := writeYAMLNamed(t, "top.yaml", `
port: 3333
`)
	cfg, err := configa.Parse[testConfig]([]string{base, mid, top})
	require.NoError(t, err)

	result, err := cfg.Resolve(context.Background())
	require.NoError(t, err)

	assert.Equal(t, 3333, result.Port)
}

func TestParse_Overlay_TemplateInOverlay(t *testing.T) {
	// A {{ secret }} expression introduced by the overlay file is resolved.
	base := writeYAMLNamed(t, "base.yaml", `
database_url: postgres://base/db
port: 5432
`)
	overlay := writeYAMLNamed(t, "overlay.yaml", `
api_key: '{{ secret "my-secret" "key" }}'
`)
	client := newMockClient(map[string]map[string]string{
		"my-secret": {"key": "overlay-secret-value"},
	})
	cfg, err := configa.Parse[testConfig]([]string{base, overlay}, configa.WithSecretClient(client))
	require.NoError(t, err)

	result, err := cfg.Resolve(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "overlay-secret-value", result.APIKey)
}

func TestParse_Overlay_FileNotFound(t *testing.T) {
	base := writeYAML(t, `
database_url: postgres://base/db
port: 5432
`)
	_, err := configa.Parse[testConfig]([]string{base, "/nonexistent/overlay.yaml"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read config file")
}
