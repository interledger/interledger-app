# configa — Implementation Plan

## Goal

A thin, focused Go configuration library for the interledger-app backend that:
- Loads YAML config files into typed Go structs
- Resolves Kubernetes secrets via Go template syntax (`{{ secret "name" "key" }}`)
- Validates the resulting struct using struct tags
- Fails fast at startup if any referenced secret cannot be fetched
- Adds zero new external dependencies

---

## Public API

```go
package configa

// Parse reads and parses a YAML configuration file.
// It detects whether any {{ secret }} template calls exist but does NOT
// contact Kubernetes yet. If no template calls are found, Resolve() will
// skip all k8s API calls entirely.
func Parse[T any](filename string, opts ...Option) (*Config[T], error)

// Config holds the parsed-but-not-yet-resolved configuration.
type Config[T any] struct { /* unexported */ }

// Resolve substitutes all {{ secret "name" "key" }} expressions by fetching
// the referenced Kubernetes secrets. If no template expressions were found,
// the Kubernetes client is never invoked.
// After substitution, Resolve unmarshals the result into T and runs
// go-playground/validator on the struct.
// Returns an error if: any secret cannot be fetched, the YAML cannot be
// unmarshalled into T, or validation fails.
func (c *Config[T]) Resolve(ctx context.Context) (T, error)

// Option is a functional option for Parse.
type Option func(*config[any])

// WithSecretClient sets the Kubernetes secret client. Required only when the
// config file contains {{ secret }} expressions.
func WithSecretClient(client SecretClient) Option

// WithNamespace overrides the Kubernetes namespace used when fetching secrets.
// Defaults to the value in /var/run/secrets/kubernetes.io/serviceaccount/namespace.
func WithNamespace(ns string) Option
```

### SecretClient interface

```go
// SecretClient fetches a single key from a Kubernetes secret.
type SecretClient interface {
    GetSecret(ctx context.Context, namespace, name, key string) (string, error)
}

// NewInClusterSecretClient returns a SecretClient that uses the Kubernetes
// in-cluster service account credentials. Credentials are loaded lazily on
// the first GetSecret call, so this is safe to construct in non-Kubernetes
// environments as long as no {{ secret }} templates are present in the config.
func NewInClusterSecretClient() SecretClient
```

---

## Template Syntax

Standard Go `text/template` syntax with a single registered function:

```yaml
# config.yaml
database_url: "postgres://user:{{ secret "db-credentials" "password" }}@host/db"
api_key: "{{ secret "api-credentials" "key" }}"
plain_value: "no-template-here"
```

Rules:
- `{{ secret "secret-name" "key" }}` — fetches `data["key"]` from secret `secret-name`
- Secret data values are base64-decoded automatically (the k8s API returns decoded values)
- Whole-value or embedded-in-string usage both work (`"prefix-{{ secret ... }}-suffix"`)
- Only the `secret` function is registered; any other template call is an error

---

## Internal Resolution Pipeline

```
Parse(filename)
  │
  ├─ Read file bytes
  ├─ Quick scan: does the file contain "{{"?
  │     no → hasTemplates = false
  │    yes → hasTemplates = true
  │
  ├─ Unmarshal YAML bytes → map[string]any (yaml.v3)
  └─ Return Config{raw, hasTemplates, opts}

Config.Resolve(ctx)
  │
  ├─ If !hasTemplates → skip to unmarshal step
  │
  ├─ Collect unique secret names referenced across all string values
  ├─ Fetch each unique secret once (single GET per secret name)
  │     → error if any fetch fails
  │
  ├─ Walk map[string]any, execute each string value as a Go template
  │     → the "secret" func reads from the already-fetched in-memory map
  │     → error if template execution fails for any value
  │
  ├─ Marshal resolved map → YAML bytes (yaml.v3)
  ├─ Unmarshal YAML bytes → T (yaml.v3, using yaml struct tags on T)
  │
  └─ Validate T (go-playground/validator/v10, using validate struct tags on T)
        → return (T, nil) or (zero, error)
```

The two-pass approach (collect references first, then execute templates) ensures each
Kubernetes secret is fetched exactly once even if referenced multiple times.

---

## Kubernetes REST Client

No `k8s.io/client-go` dependency. The in-cluster client is ~80 lines of stdlib HTTP.

**Lazy credential loading** — credentials are read on the first `GetSecret` call,
not at construction. `sync.Once` is used so they are read exactly once and cached
for subsequent calls within the same `Resolve()` invocation. This means:
- `NewInClusterSecretClient()` never fails, even outside Kubernetes
- Errors surface only when `Resolve()` actually needs to contact the API
- If the config has no `{{ secret }}` templates, credentials are never read

**In-cluster file paths read lazily on first GetSecret call:**
- Token: `/var/run/secrets/kubernetes.io/serviceaccount/token`
- CA cert: `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`
- Namespace: `/var/run/secrets/kubernetes.io/serviceaccount/namespace`

**Request:**
```
GET https://{KUBERNETES_SERVICE_HOST}:{KUBERNETES_SERVICE_PORT}/api/v1/namespaces/{namespace}/secrets/{name}
Authorization: Bearer {token}
```

**Response parsing:** only `data` field is extracted; keys are already base64-decoded
by `encoding/json` since the response is `map[string][]byte`.

**Error handling:**
- HTTP 404 → `ErrSecretNotFound`
- HTTP 403/401 → `ErrSecretForbidden`
- Other non-2xx → `ErrSecretFetchFailed` with status code
- Network errors → wrapped error

---

## File Structure

```
go/configa/
├── configa.go          Public API: Config[T], Parse(), Option, errors
├── resolve.go          Resolve() implementation, template walking, template execution
├── k8s.go              SecretClient interface + InClusterSecretClient
├── configa_test.go     Tests for Parse + Resolve (mock SecretClient)
└── k8s_test.go         Tests for InClusterSecretClient (httptest.Server)
```

No subdirectories. The package is small enough to keep flat.

---

## Test Plan (target: ≥85% coverage)

### configa_test.go — uses a mock/stub SecretClient

| Test | What it covers |
|---|---|
| `TestParse_ValidYAML` | Happy path: file parsed, struct populated |
| `TestParse_FileNotFound` | Error on missing file |
| `TestParse_InvalidYAML` | Error on malformed YAML |
| `TestResolve_NoTemplates` | k8s client never called when no `{{` in file |
| `TestResolve_SingleSecret` | Template resolved, value substituted |
| `TestResolve_MultipleSecretsDeduped` | Same secret referenced twice → fetched once |
| `TestResolve_EmbeddedTemplate` | Template mid-string: `"prefix-{{ secret ... }}-suffix"` |
| `TestResolve_SecretFetchError` | Client returns error → Resolve returns error |
| `TestResolve_ValidationFailure` | Required field missing after resolution → validation error |
| `TestResolve_UnknownTemplateFunction` | `{{ unknown }}` → error |
| `TestResolve_NoClientWhenTemplatesPresent` | Templates found but no client set → error |
| `TestWithNamespace` | Namespace option is passed through to client |

Mock SecretClient is a simple struct with a map[string]map[string]string, no generated mocks needed.

### k8s_test.go — uses httptest.NewTLSServer

| Test | What it covers |
|---|---|
| `TestInClusterSecretClient_Success` | 200 with valid secret JSON → decoded value returned |
| `TestInClusterSecretClient_NotFound` | 404 → ErrSecretNotFound |
| `TestInClusterSecretClient_Forbidden` | 403 → ErrSecretForbidden |
| `TestInClusterSecretClient_ServerError` | 500 → ErrSecretFetchFailed |
| `TestInClusterSecretClient_MissingToken` | Token file absent → error on first GetSecret call, not construction |
| `TestNewInClusterSecretClient_NeverFails` | Constructor always succeeds, even with no k8s files present |

The test server serves a path like `/api/v1/namespaces/test-ns/secrets/my-secret` and returns
a JSON payload mimicking the Kubernetes API `v1.Secret` response shape (only `data` field needed).

---

## Struct Tag Conventions for Callers

```go
type AppConfig struct {
    DatabaseURL string `yaml:"database_url" validate:"required,url"`
    APIKey      string `yaml:"api_key"      validate:"required"`
    Port        int    `yaml:"port"         validate:"min=1,max=65535"`
    Optional    string `yaml:"optional"`     // no validate tag = not required
}
```

- `yaml:` tags control field mapping from YAML keys (standard yaml.v3)
- `validate:` tags control validation (standard go-playground/validator/v10)
- No custom `configa:` tag is needed

---

## Usage Example

```go
// At startup, in main.go or equivalent:
cfg, err := configa.Parse[AppConfig]("config.yaml",
    configa.WithSecretClient(k8sClient),
)
if err != nil {
    log.Fatal("failed to parse config", zap.Error(err))
}

app, err := cfg.Resolve(ctx)
if err != nil {
    log.Fatal("failed to resolve config", zap.Error(err))
}

// app is *AppConfig, fully populated and validated
```

Local development with no k8s secrets in the YAML:

```yaml
# config.yaml (local)
database_url: "postgres://user:localpassword@localhost/db"
api_key: "dev-key-abc123"
port: 8080
```

```go
// No WithSecretClient needed — Resolve() detects no templates and skips k8s
cfg, _ := configa.Parse[AppConfig]("config.yaml")
app, err := cfg.Resolve(ctx)
```

---

## Dependencies

| Package | Already in go.mod | Usage |
|---|---|---|
| `gopkg.in/yaml.v3` | Yes (indirect) | YAML parsing + marshaling |
| `github.com/go-playground/validator/v10` | Yes (direct) | Struct validation |
| `encoding/json` | stdlib | Kubernetes API response parsing |
| `net/http` | stdlib | Kubernetes REST calls |
| `text/template` | stdlib | Template execution |

**No new dependencies required.**

---

## Out of Scope (deliberately excluded)

- kubeconfig / OIDC auth (in-cluster REST only for now)
- Secret hot-reloading (startup-time resolution only)
- Watching secrets for changes
- Env var overrides or fallback chains
- Multiple config file merging
- Any config source other than YAML files
