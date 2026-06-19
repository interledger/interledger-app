# configa

`configa` is a small Go library for loading YAML config into typed structs, supporting:

- layered config files (base + overlays)
- optional Kubernetes secret template resolution
- struct validation via `validate` tags

It is designed for apps that want simple file-based config in all environments, with Kubernetes Secret lookup only when needed.

## Features

- Parse one or more YAML files into a typed config model
- Deep-merge overlays (later files win)
- Resolve templates like `{{ secret "my-secret" "password" }}`
- Skip Kubernetes calls when no templates are present
- Validate config structs using `github.com/go-playground/validator/v10`

## Quick Start (No Secrets)

### 1. Define a typed config struct

```go
package main

type Config struct {
    Port   string `yaml:"port" validate:"required"`
    DBURL  string `yaml:"db_url" validate:"required,url"`
    LogLvl string `yaml:"log_level"`
}
```

### 2. Create YAML config

```yaml
# config.yaml
port: "8080"
db_url: "postgres://localhost:5432/app"
log_level: "info"
```

### 3. Parse and resolve

```go
package main

import (
    "context"
    "log"

    "github.com/interledger/interledger-app/go/configa"
)

func loadConfig() Config {
    parsed, err := configa.Parse[Config]([]string{"config.yaml"})
    if err != nil {
        log.Fatalf("parse config: %v", err)
    }

    cfg, err := parsed.Resolve(context.Background())
    if err != nil {
        log.Fatalf("resolve config: %v", err)
    }

    return cfg
}
```

## Overlay Files

You can pass multiple files to `Parse`. Files are merged in order.

```go
parsed, err := configa.Parse[Config]([]string{
    "config/base.yaml",
    "config/dev.yaml",
})
```

Merge behavior:

- Maps: deep-merged recursively
- Scalars and arrays: overlay value replaces base value
- Last file wins for conflicts

A common pattern is to load from an env var:

```go
files := strings.Split(os.Getenv("CONFIG"), ",")
parsed, err := configa.Parse[Config](files)
```

## Kubernetes Secret Templates

If YAML contains templates, register a secret client with `WithSecretClient`.

Template format:

```yaml
api_key: '{{ secret "wallet-secrets" "apiKey" }}'
database_url: 'postgres://user:{{ secret "wallet-secrets" "dbPassword" }}@db:5432/app'
```

### In-cluster usage

```go
secretClient := configa.NewInClusterSecretClient()

parsed, err := configa.Parse[Config](
    []string{"config.yaml"},
    configa.WithSecretClient(secretClient),
    // Optional: override namespace used for secret lookups.
    configa.WithNamespace("wallet"),
)
if err != nil {
    return err
}

cfg, err := parsed.Resolve(context.Background())
if err != nil {
    return err
}
```

Notes:

- If no `{{ ... }}` templates survive the merge, `Resolve` does not call Kubernetes.
- If templates are present but no secret client is set, `Resolve` returns `configa.ErrNoSecretClient`.

## Local Testing with a Stub SecretClient

`configa` exposes the `SecretClient` interface, so tests can use a fake implementation.

```go
package myapp_test

import (
    "context"

    "github.com/interledger/interledger-app/go/configa"
)

type fakeSecrets struct{}

func (f *fakeSecrets) GetSecret(ctx context.Context, namespace, name string) (map[string]string, error) {
    return map[string]string{
        "dbPassword": "test-pass",
        "apiKey":     "test-key",
    }, nil
}

func example() error {
    parsed, err := configa.Parse[Config](
        []string{"testdata/config.yaml"},
        configa.WithSecretClient(&fakeSecrets{}),
    )
    if err != nil {
        return err
    }

    _, err = parsed.Resolve(context.Background())
    return err
}
```

## Validation

`Resolve` unmarshals into your typed struct and then validates using `validate` tags.

```go
type Config struct {
    Port string `yaml:"port" validate:"required"`
}
```

If validation fails, `Resolve` returns an error.

## Error Behavior

Typical failure points:

- `Parse`: file read failures, invalid YAML, empty file list
- `Resolve`: missing secret client for templated config, secret fetch errors, missing secret keys, template parse/execute errors, unmarshal errors, validation errors

Kubernetes-related sentinel errors include:

- `configa.ErrNoSecretClient`
- `configa.ErrSecretNotFound`
- `configa.ErrSecretForbidden`
- `configa.ErrSecretFetchFailed`

## End-to-End Pattern

A practical app startup flow:

1. Read `CONFIG` as a comma-separated list of YAML files.
2. Build an in-cluster client with `configa.NewInClusterSecretClient()`.
3. Parse with `configa.Parse[T](files, configa.WithSecretClient(client))`.
4. Call `Resolve(context.Background())` once at startup.
5. Use the typed result everywhere else.
