package configa

import (
	"context"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// k8sHandler is a minimal mock of the Kubernetes secrets REST API.
type k8sHandler struct {
	// secrets maps "namespace/name" to the secret data (plain text values).
	secrets map[string]map[string]string
	// statusCode, when non-zero, is returned for every request regardless of path.
	statusCode int
}

func (h *k8sHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.statusCode != 0 {
		w.WriteHeader(h.statusCode)
		return
	}

	// Expected path: /api/v1/namespaces/{ns}/secrets/{name}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 6 || parts[0] != "api" || parts[2] != "namespaces" || parts[4] != "secrets" {
		http.NotFound(w, r)
		return
	}
	ns, name := parts[3], parts[5]
	key := ns+"/"+name

	data, ok := h.secrets[key]
	if !ok {
		http.NotFound(w, r)
		return
	}

	// Kubernetes encodes secret data as base64 in JSON.
	// encoding/json encodes []byte as base64, matching the real API.
	dataBytes := make(map[string][]byte, len(data))
	for k, v := range data {
		dataBytes[k] = []byte(v)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Data map[string][]byte `json:"data"`
	}{Data: dataBytes})
}

// setupInClusterClient starts a TLS test server and returns an InClusterSecretClient
// configured to use it. t.Cleanup handles server shutdown and env-var restoration.
func setupInClusterClient(t *testing.T, handler http.Handler) *InClusterSecretClient {
	t.Helper()

	ts := httptest.NewTLSServer(handler)
	t.Cleanup(ts.Close)

	dir := t.TempDir()

	// Token file
	tokenFile := filepath.Join(dir, "token")
	require.NoError(t, os.WriteFile(tokenFile, []byte("test-bearer-token"), 0600))

	// Namespace file
	nsFile := filepath.Join(dir, "namespace")
	require.NoError(t, os.WriteFile(nsFile, []byte("test-ns"), 0600))

	// CA cert — PEM-encode the test server's self-signed certificate.
	caFile := filepath.Join(dir, "ca.crt")
	rawCert := ts.TLS.Certificates[0].Certificate[0]
	caBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: rawCert})
	require.NoError(t, os.WriteFile(caFile, caBytes, 0600))

	// Point KUBERNETES_SERVICE_{HOST,PORT} at the test server.
	u, err := url.Parse(ts.URL)
	require.NoError(t, err)
	host, port, err := net.SplitHostPort(u.Host)
	require.NoError(t, err)
	t.Setenv("KUBERNETES_SERVICE_HOST", host)
	t.Setenv("KUBERNETES_SERVICE_PORT", port)

	return &InClusterSecretClient{
		tokenPath:     tokenFile,
		caPath:        caFile,
		namespacePath: nsFile,
	}
}

func TestNewInClusterSecretClient_NeverFails(t *testing.T) {
	// Construction must succeed even when no k8s files exist.
	c := NewInClusterSecretClient()
	assert.NotNil(t, c)
}

func TestInClusterSecretClient_MissingToken(t *testing.T) {
	// Credential errors surface on GetSecret, not on construction.
	c := &InClusterSecretClient{
		tokenPath:     "/nonexistent/token",
		caPath:        "/nonexistent/ca.crt",
		namespacePath: "/nonexistent/namespace",
	}
	_, err := c.GetSecret(context.Background(), "ns", "name")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service account token")
}

func TestInClusterSecretClient_MissingCACert(t *testing.T) {
	dir := t.TempDir()
	tokenFile := filepath.Join(dir, "token")
	require.NoError(t, os.WriteFile(tokenFile, []byte("tok"), 0600))

	c := &InClusterSecretClient{
		tokenPath:     tokenFile,
		caPath:        "/nonexistent/ca.crt",
		namespacePath: "/nonexistent/namespace",
	}
	_, err := c.GetSecret(context.Background(), "ns", "name")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "CA cert")
}

func TestInClusterSecretClient_MissingEnvVars(t *testing.T) {
	dir := t.TempDir()

	tokenFile := filepath.Join(dir, "token")
	require.NoError(t, os.WriteFile(tokenFile, []byte("tok"), 0600))

	caFile := filepath.Join(dir, "ca.crt")
	// Write a minimal self-signed PEM so the CA parse step passes.
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(ts.Close)
	rawCert := ts.TLS.Certificates[0].Certificate[0]
	caBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: rawCert})
	require.NoError(t, os.WriteFile(caFile, caBytes, 0600))

	nsFile := filepath.Join(dir, "namespace")
	require.NoError(t, os.WriteFile(nsFile, []byte("test-ns"), 0600))

	// Ensure the env vars are unset.
	t.Setenv("KUBERNETES_SERVICE_HOST", "")
	t.Setenv("KUBERNETES_SERVICE_PORT", "")

	c := &InClusterSecretClient{
		tokenPath:     tokenFile,
		caPath:        caFile,
		namespacePath: nsFile,
	}
	_, err := c.GetSecret(context.Background(), "ns", "name")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "KUBERNETES_SERVICE_HOST")
}

func TestInClusterSecretClient_Success(t *testing.T) {
	handler := &k8sHandler{
		secrets: map[string]map[string]string{
			"test-ns/my-secret": {"password": "secret-value", "user": "admin"},
		},
	}
	c := setupInClusterClient(t, handler)

	data, err := c.GetSecret(context.Background(), "test-ns", "my-secret")
	require.NoError(t, err)

	assert.Equal(t, "secret-value", data["password"])
	assert.Equal(t, "admin", data["user"])
}

func TestInClusterSecretClient_DefaultNamespace(t *testing.T) {
	// When namespace is empty, the client uses the service account namespace file.
	handler := &k8sHandler{
		secrets: map[string]map[string]string{
			"test-ns/my-secret": {"key": "value"},
		},
	}
	c := setupInClusterClient(t, handler)

	data, err := c.GetSecret(context.Background(), "" /* empty → uses "test-ns" */, "my-secret")
	require.NoError(t, err)
	assert.Equal(t, "value", data["key"])
}

func TestInClusterSecretClient_NotFound(t *testing.T) {
	handler := &k8sHandler{secrets: map[string]map[string]string{}}
	c := setupInClusterClient(t, handler)

	_, err := c.GetSecret(context.Background(), "test-ns", "nonexistent")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSecretNotFound))
}

func TestInClusterSecretClient_Forbidden(t *testing.T) {
	handler := &k8sHandler{statusCode: http.StatusForbidden}
	c := setupInClusterClient(t, handler)

	_, err := c.GetSecret(context.Background(), "test-ns", "my-secret")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSecretForbidden))
}

func TestInClusterSecretClient_Unauthorized(t *testing.T) {
	handler := &k8sHandler{statusCode: http.StatusUnauthorized}
	c := setupInClusterClient(t, handler)

	_, err := c.GetSecret(context.Background(), "test-ns", "my-secret")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSecretForbidden))
}

func TestInClusterSecretClient_ServerError(t *testing.T) {
	handler := &k8sHandler{statusCode: http.StatusInternalServerError}
	c := setupInClusterClient(t, handler)

	_, err := c.GetSecret(context.Background(), "test-ns", "my-secret")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSecretFetchFailed))
}

func TestInClusterSecretClient_CredentialsLoadedOnce(t *testing.T) {
	// Verifies that sync.Once prevents re-reading credential files on repeated calls.
	handler := &k8sHandler{
		secrets: map[string]map[string]string{
			"test-ns/s1": {"k": "v1"},
			"test-ns/s2": {"k": "v2"},
		},
	}
	c := setupInClusterClient(t, handler)

	_, err := c.GetSecret(context.Background(), "test-ns", "s1")
	require.NoError(t, err)

	// Remove the token file after the first successful call.
	require.NoError(t, os.Remove(c.tokenPath))

	// Second call should still succeed because credentials are cached via sync.Once.
	_, err = c.GetSecret(context.Background(), "test-ns", "s2")
	require.NoError(t, err)
}
