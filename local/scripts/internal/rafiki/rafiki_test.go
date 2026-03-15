package rafiki

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	cfg := LoadConfig()

	assert.NotEmpty(t, cfg.GraphQLEndpoint)
	assert.NotEmpty(t, cfg.AdminAPISecret)
	assert.NotEmpty(t, cfg.AdminSignatureVersion)
}

func TestCanonicalize(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name: "simple map with sorted keys",
			input: map[string]interface{}{
				"z": "last",
				"a": "first",
				"m": "middle",
			},
			expected: map[string]interface{}{
				"a": "first",
				"m": "middle",
				"z": "last",
			},
		},
		{
			name: "nested map",
			input: map[string]interface{}{
				"outer": map[string]interface{}{
					"z": "inner_last",
					"a": "inner_first",
				},
			},
			expected: map[string]interface{}{
				"outer": map[string]interface{}{
					"a": "inner_first",
					"z": "inner_last",
				},
			},
		},
		{
			name:     "array unchanged",
			input:    []interface{}{"a", "b", "c"},
			expected: []interface{}{"a", "b", "c"},
		},
		{
			name:     "primitive unchanged",
			input:    "hello",
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := canonicalize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCanonicalizeAndStringify(t *testing.T) {
	input := map[string]interface{}{
		"z": "last",
		"a": "first",
	}

	result, err := canonicalizeAndStringify(input)
	require.NoError(t, err)

	// Keys should be sorted alphabetically
	expected := `{"a":"first","z":"last"}`
	assert.Equal(t, expected, result)
}

func TestSignRequest(t *testing.T) {
	cfg := Config{
		GraphQLEndpoint:       "http://localhost:3001/graphql",
		AdminAPISecret:        "test_secret",
		AdminSignatureVersion: "1",
	}

	req := GraphQLRequest{
		Query: "query { test }",
		Variables: map[string]interface{}{
			"id": "123",
		},
	}

	timestamp := int64(1704067200000) // Fixed timestamp for testing

	signature, err := signRequest(req, cfg, timestamp)
	require.NoError(t, err)

	assert.Contains(t, signature, "t=1704067200000")
	assert.Contains(t, signature, "v1=")
	assert.NotEmpty(t, signature)
}

func TestGraphQLRequest(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.NotEmpty(t, r.Header.Get("signature"))

		// Return success response
		response := GraphQLResponse{
			Data: json.RawMessage(`{"test": "success"}`),
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := Config{
		GraphQLEndpoint:       server.URL,
		AdminAPISecret:        "test_secret",
		AdminSignatureVersion: "1",
	}

	req := GraphQLRequest{
		Query: "query { test }",
	}

	data, err := graphqlRequest(req, cfg)
	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.Contains(t, string(data), "success")
}

func TestGraphQLRequest_WithErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := GraphQLResponse{
			Errors: []GraphQLError{
				{Message: "Something went wrong"},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := Config{
		GraphQLEndpoint:       server.URL,
		AdminAPISecret:        "test_secret",
		AdminSignatureVersion: "1",
	}

	req := GraphQLRequest{
		Query: "query { test }",
	}

	_, err := graphqlRequest(req, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Something went wrong")
}

func TestEnsureAssets(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if callCount == 1 {
			// First call: list assets (empty)
			response := GraphQLResponse{
				Data: json.RawMessage(`{"assets": {"edges": []}}`),
			}
			json.NewEncoder(w).Encode(response)
		} else {
			// Subsequent calls: create asset
			response := GraphQLResponse{
				Data: json.RawMessage(`{"createAsset": {"asset": {"id": "123", "code": "USD", "scale": 2}}}`),
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	cfg := Config{
		GraphQLEndpoint:       server.URL,
		AdminAPISecret:        "test_secret",
		AdminSignatureVersion: "1",
	}

	assets := []Asset{
		{Code: "USD", Scale: 2},
	}

	err := EnsureAssets(cfg, assets)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, callCount, 2) // List + Create
}

func TestEnsureLiquidity(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if callCount == 1 {
			// First call: get asset by code and scale
			response := GraphQLResponse{
				Data: json.RawMessage(`{"assetByCodeAndScale": {"id": "asset-123", "code": "USD", "scale": 2}}`),
			}
			json.NewEncoder(w).Encode(response)
		} else {
			// Second call: deposit liquidity
			response := GraphQLResponse{
				Data: json.RawMessage(`{"depositAssetLiquidity": {"success": true}}`),
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	cfg := Config{
		GraphQLEndpoint:       server.URL,
		AdminAPISecret:        "test_secret",
		AdminSignatureVersion: "1",
	}

	assets := []Asset{
		{Code: "USD", Scale: 2},
	}

	err := EnsureLiquidity(cfg, assets)
	require.NoError(t, err)
	assert.Equal(t, 2, callCount) // Get asset + Deposit
}

func TestDefaultAssets(t *testing.T) {
	assert.Len(t, DefaultAssets, 10)

	// Verify specific assets exist
	codes := make(map[string]bool)
	for _, asset := range DefaultAssets {
		codes[asset.Code] = true
		assert.Equal(t, 2, asset.Scale) // All should have scale 2
	}

	expectedCodes := []string{"USD", "EUR", "GBP", "ZAR", "MXN", "SGD", "CAD", "EGG", "PEB", "PKR"}
	for _, code := range expectedCodes {
		assert.True(t, codes[code], "Expected asset %s to exist", code)
	}
}
