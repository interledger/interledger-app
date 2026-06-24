package configa

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	ErrSecretNotFound    = errors.New("configa: kubernetes secret not found")
	ErrSecretForbidden   = errors.New("configa: forbidden: insufficient permissions to read kubernetes secret")
	ErrSecretFetchFailed = errors.New("configa: failed to fetch kubernetes secret")
)

const (
	defaultSATokenPath     = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	defaultSACAPath        = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	defaultSANamespacePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

// InClusterSecretClient fetches Kubernetes secrets using the pod's in-cluster
// service account credentials. Credentials are loaded lazily on the first
// GetSecret call so the client is safe to construct outside Kubernetes.
type InClusterSecretClient struct {
	once      sync.Once
	initErr   error
	client    *http.Client
	token     string
	apiBase   string
	defaultNS string

	// Overridable paths — set by NewInClusterSecretClient; tests may override
	// by constructing the struct directly (same package).
	tokenPath     string
	caPath        string
	namespacePath string
}

// NewInClusterSecretClient returns a SecretClient backed by the Kubernetes
// in-cluster service account. Construction never fails — any errors in reading
// service account files surface on the first GetSecret call.
func NewInClusterSecretClient() SecretClient {
	return &InClusterSecretClient{
		tokenPath:     defaultSATokenPath,
		caPath:        defaultSACAPath,
		namespacePath: defaultSANamespacePath,
	}
}

// k8sSecretResponse is the subset of the Kubernetes v1.Secret JSON we need.
// encoding/json automatically base64-decodes []byte fields from JSON strings.
type k8sSecretResponse struct {
	Data map[string][]byte `json:"data"`
}

func (c *InClusterSecretClient) loadCredentials() {
	c.once.Do(func() {
		tokenBytes, err := os.ReadFile(c.tokenPath)
		if err != nil {
			c.initErr = fmt.Errorf("configa: read service account token: %w", err)
			return
		}

		caBytes, err := os.ReadFile(c.caPath)
		if err != nil {
			c.initErr = fmt.Errorf("configa: read service account CA cert: %w", err)
			return
		}

		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caBytes) {
			c.initErr = fmt.Errorf("configa: failed to parse service account CA cert")
			return
		}

		nsBytes, err := os.ReadFile(c.namespacePath)
		if err != nil {
			c.initErr = fmt.Errorf("configa: read service account namespace: %w", err)
			return
		}

		host := os.Getenv("KUBERNETES_SERVICE_HOST")
		port := os.Getenv("KUBERNETES_SERVICE_PORT")
		if host == "" || port == "" {
			c.initErr = fmt.Errorf("configa: KUBERNETES_SERVICE_HOST or KUBERNETES_SERVICE_PORT not set")
			return
		}

		c.token = strings.TrimSpace(string(tokenBytes))
		c.defaultNS = strings.TrimSpace(string(nsBytes))
		c.apiBase = fmt.Sprintf("https://%s:%s", host, port)
		c.client = &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: pool,
				},
			},
		}
	})
}

// GetSecret fetches all data keys for the named secret.
// When namespace is empty the service account's own namespace is used.
func (c *InClusterSecretClient) GetSecret(ctx context.Context, namespace, name string) (map[string]string, error) {
	c.loadCredentials()
	if c.initErr != nil {
		return nil, c.initErr
	}

	ns := namespace
	if ns == "" {
		ns = c.defaultNS
	}

	url := fmt.Sprintf("%s/api/v1/namespaces/%s/secrets/%s", c.apiBase, ns, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("configa: build k8s request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("configa: k8s request failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// handled below
	case http.StatusNotFound:
		return nil, fmt.Errorf("%w: %s/%s", ErrSecretNotFound, ns, name)
	case http.StatusForbidden, http.StatusUnauthorized:
		return nil, fmt.Errorf("%w: %s/%s", ErrSecretForbidden, ns, name)
	default:
		return nil, fmt.Errorf("%w: %s/%s (HTTP %d)", ErrSecretFetchFailed, ns, name, resp.StatusCode)
	}

	var secret k8sSecretResponse
	if err := json.NewDecoder(resp.Body).Decode(&secret); err != nil {
		return nil, fmt.Errorf("configa: decode k8s secret response: %w", err)
	}

	result := make(map[string]string, len(secret.Data))
	for k, v := range secret.Data {
		result[k] = string(v)
	}

	return result, nil
}
