package configa

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// Resolve substitutes all {{ secret "name" "key" }} template expressions by fetching
// the referenced Kubernetes secrets, then unmarshals the result into T and validates it.
//
// If no template expressions were found during Parse, Resolve skips all Kubernetes
// API calls entirely, making it safe to call outside a Kubernetes environment.
//
// Returns an error if any secret cannot be fetched, the YAML cannot be unmarshalled
// into T, or struct validation fails.
func (c *Config[T]) Resolve(ctx context.Context) (T, error) {
	var zero T

	raw := c.raw

	if c.hasTemplates {
		if c.opts.client == nil {
			return zero, ErrNoSecretClient
		}
		resolved, err := resolveTemplates(ctx, raw, c.opts.client, c.opts.namespace)
		if err != nil {
			return zero, err
		}
		raw = resolved
	}

	var result T
	if err := yaml.Unmarshal(raw, &result); err != nil {
		return zero, fmt.Errorf("configa: unmarshal config: %w", err)
	}

	if err := validateStruct(result); err != nil {
		return zero, err
	}

	return result, nil
}

// resolveTemplates walks all string values in the YAML, executes any Go templates
// found, and returns the fully-substituted YAML bytes.
func resolveTemplates(ctx context.Context, raw []byte, client SecretClient, namespace string) ([]byte, error) {
	var m map[string]any
	if err := yaml.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("configa: parse yaml for template resolution: %w", err)
	}

	// cache deduplicates Kubernetes API calls within a single Resolve call.
	// Key: secret name → map of data keys to values.
	cache := make(map[string]map[string]string)

	secretFn := func(name, key string) (string, error) {
		if data, ok := cache[name]; ok {
			val, ok := data[key]
			if !ok {
				return "", fmt.Errorf("configa: key %q not found in secret %q", key, name)
			}
			return val, nil
		}
		data, err := client.GetSecret(ctx, namespace, name)
		if err != nil {
			return "", err
		}
		cache[name] = data
		val, ok := data[key]
		if !ok {
			return "", fmt.Errorf("configa: key %q not found in secret %q", key, name)
		}
		return val, nil
	}

	resolved, err := walkMap(m, secretFn)
	if err != nil {
		return nil, err
	}

	out, err := yaml.Marshal(resolved)
	if err != nil {
		return nil, fmt.Errorf("configa: marshal resolved config: %w", err)
	}

	return out, nil
}

func walkMap(m map[string]any, secretFn func(string, string) (string, error)) (map[string]any, error) {
	result := make(map[string]any, len(m))
	for k, v := range m {
		resolved, err := walkValue(v, secretFn)
		if err != nil {
			return nil, fmt.Errorf("key %q: %w", k, err)
		}
		result[k] = resolved
	}
	return result, nil
}

func walkValue(v any, secretFn func(string, string) (string, error)) (any, error) {
	switch val := v.(type) {
	case string:
		return executeTemplate(val, secretFn)
	case map[string]any:
		return walkMap(val, secretFn)
	case []any:
		result := make([]any, len(val))
		for i, item := range val {
			resolved, err := walkValue(item, secretFn)
			if err != nil {
				return nil, fmt.Errorf("[%d]: %w", i, err)
			}
			result[i] = resolved
		}
		return result, nil
	default:
		return v, nil
	}
}

// executeTemplate executes s as a Go template with the secret function registered.
// Returns s unchanged if no {{ }} markers are present (fast path).
func executeTemplate(s string, secretFn func(string, string) (string, error)) (string, error) {
	if !strings.Contains(s, "{{") {
		return s, nil
	}

	tmpl, err := template.New("").Funcs(template.FuncMap{
		"secret": secretFn,
	}).Parse(s)
	if err != nil {
		return "", fmt.Errorf("configa: parse template %q: %w", s, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return "", fmt.Errorf("configa: execute template %q: %w", s, err)
	}

	return buf.String(), nil
}
