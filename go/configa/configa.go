package configa

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ErrNoSecretClient is returned by Resolve when the config file contains
// {{ secret }} template expressions but no SecretClient was provided.
var ErrNoSecretClient = errors.New("configa: config contains {{ secret }} templates but no SecretClient was provided via WithSecretClient")

// SecretClient fetches Kubernetes secrets by name, returning all data keys.
type SecretClient interface {
	GetSecret(ctx context.Context, namespace, name string) (map[string]string, error)
}

// Option configures a Config.
type Option func(*options)

type options struct {
	client    SecretClient
	namespace string
}

// WithSecretClient sets the SecretClient used to resolve {{ secret }} template expressions
// during Resolve. Only required when the config file contains such expressions.
func WithSecretClient(client SecretClient) Option {
	return func(o *options) {
		o.client = client
	}
}

// WithNamespace sets the Kubernetes namespace used for secret lookups.
// When not set, InClusterSecretClient defaults to the service account namespace.
func WithNamespace(ns string) Option {
	return func(o *options) {
		o.namespace = ns
	}
}

// Config holds a parsed YAML configuration file, ready for secret resolution.
type Config[T any] struct {
	raw          []byte
	hasTemplates bool
	opts         options
}

// Parse reads one or more YAML configuration files and merges them in order.
// Later files act as overlays: their values take precedence over earlier ones.
// For nested maps the merge is deep; for scalars and arrays the overlay wins entirely.
//
// Template detection runs on the merged result, so Resolve will skip all
// Kubernetes API calls if no {{ }} expressions survive after merging.
func Parse[T any](filenames []string, optFns ...Option) (*Config[T], error) {
	if len(filenames) == 0 {
		return nil, fmt.Errorf("configa: at least one filename is required")
	}

	data, err := os.ReadFile(filenames[0])
	if err != nil {
		return nil, fmt.Errorf("configa: read config file %q: %w", filenames[0], err)
	}

	// Validate base YAML syntax before any merge work.
	var check map[string]any
	if err := yaml.Unmarshal(data, &check); err != nil {
		return nil, fmt.Errorf("configa: invalid yaml in %q: %w", filenames[0], err)
	}

	for _, f := range filenames[1:] {
		overlay, err := os.ReadFile(f)
		if err != nil {
			return nil, fmt.Errorf("configa: read config file %q: %w", f, err)
		}
		data, err = mergeYAML(data, overlay)
		if err != nil {
			return nil, fmt.Errorf("configa: merge overlay %q: %w", f, err)
		}
	}

	opts := options{}
	for _, fn := range optFns {
		fn(&opts)
	}

	return &Config[T]{
		raw:          data,
		hasTemplates: bytes.Contains(data, []byte("{{")),
		opts:         opts,
	}, nil
}
