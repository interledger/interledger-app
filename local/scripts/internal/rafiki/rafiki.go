package rafiki

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

// Config holds Rafiki API configuration
type Config struct {
	GraphQLEndpoint       string
	AdminAPISecret        string
	AdminSignatureVersion string
}

// Asset represents a currency asset
type Asset struct {
	Code  string
	Scale int
}

// DefaultAssets is the list of all supported assets
var DefaultAssets = []Asset{
	{Code: "USD", Scale: 2},
	{Code: "EUR", Scale: 2},
	{Code: "GBP", Scale: 2},
	{Code: "ZAR", Scale: 2},
	{Code: "MXN", Scale: 2},
	{Code: "SGD", Scale: 2},
	{Code: "CAD", Scale: 2},
	{Code: "EGG", Scale: 2},
	{Code: "PEB", Scale: 2},
	{Code: "PKR", Scale: 2},
}

// GraphQL structures
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors,omitempty"`
}

type GraphQLError struct {
	Message string `json:"message"`
}

type AssetsResponse struct {
	Assets struct {
		Edges []struct {
			Node struct {
				ID    string `json:"id"`
				Code  string `json:"code"`
				Scale int    `json:"scale"`
			} `json:"node"`
		} `json:"edges"`
	} `json:"assets"`
}

type CreateAssetResponse struct {
	CreateAsset struct {
		Asset struct {
			ID    string `json:"id"`
			Code  string `json:"code"`
			Scale int    `json:"scale"`
		} `json:"asset"`
	} `json:"createAsset"`
}

type AssetByCodeResponse struct {
	AssetByCodeAndScale struct {
		ID    string `json:"id"`
		Code  string `json:"code"`
		Scale int    `json:"scale"`
	} `json:"assetByCodeAndScale"`
}

type DepositLiquidityResponse struct {
	DepositAssetLiquidity struct {
		Success bool `json:"success"`
	} `json:"depositAssetLiquidity"`
}

// GraphQL queries and mutations
const (
	listAssetsQuery = `
		query Assets($first: Int) {
			assets(first: $first) {
				edges {
					node {
						id
						code
						scale
					}
				}
			}
		}
	`

	createAssetMutation = `
		mutation CreateAsset($input: CreateAssetInput!) {
			createAsset(input: $input) {
				asset {
					id
					code
					scale
				}
			}
		}
	`

	getAssetByCodeAndScaleQuery = `
		query AssetByCodeAndScale($code: String!, $scale: UInt8!) {
			assetByCodeAndScale(code: $code, scale: $scale) {
				id
				code
				scale
			}
		}
	`

	depositAssetLiquidityMutation = `
		mutation DepositAssetLiquidity($input: DepositAssetLiquidityInput!) {
			depositAssetLiquidity(input: $input) {
				success
			}
		}
	`
)

// LoadConfig loads Rafiki configuration from environment
func LoadConfig() Config {
	// Try to load .env from parent directory
	envPath := filepath.Join("..", ".env")
	_ = godotenv.Load(envPath)

	getEnv := func(key, fallback string) string {
		if value := os.Getenv(key); value != "" {
			return value
		}
		return fallback
	}

	return Config{
		GraphQLEndpoint:       getEnv("GRAPHQL_ENDPOINT", "http://localhost:3001/graphql"),
		AdminAPISecret:        getEnv("ADMIN_API_SECRET", "your_signature_secret"),
		AdminSignatureVersion: getEnv("ADMIN_SIGNATURE_VERSION", "1"),
	}
}

// canonicalize ensures consistent JSON representation for HMAC signing
func canonicalize(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		result := make(map[string]interface{})
		for _, k := range keys {
			result[k] = canonicalize(val[k])
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, elem := range val {
			result[i] = canonicalize(elem)
		}
		return result
	default:
		return val
	}
}

func canonicalizeAndStringify(v interface{}) (string, error) {
	canonical := canonicalize(v)
	data, err := json.Marshal(canonical)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func signRequest(req GraphQLRequest, cfg Config, timestamp int64) (string, error) {
	requestData := map[string]interface{}{
		"query":         req.Query,
		"variables":     req.Variables,
		"operationName": req.OperationName,
	}
	if req.Variables == nil {
		requestData["variables"] = map[string]interface{}{}
	}

	canonical, err := canonicalizeAndStringify(requestData)
	if err != nil {
		return "", err
	}

	payload := fmt.Sprintf("%d.%s", timestamp, canonical)

	h := hmac.New(sha256.New, []byte(cfg.AdminAPISecret))
	h.Write([]byte(payload))
	digest := hex.EncodeToString(h.Sum(nil))

	return fmt.Sprintf("t=%d, v%s=%s", timestamp, cfg.AdminSignatureVersion, digest), nil
}

func graphqlRequest(req GraphQLRequest, cfg Config) (json.RawMessage, error) {
	timestamp := time.Now().UnixMilli()
	signature, err := signRequest(req, cfg, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", cfg.GraphQLEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("signature", signature)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var gqlResp GraphQLResponse
	if err := json.Unmarshal(bodyBytes, &gqlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		var messages []string
		for _, e := range gqlResp.Errors {
			messages = append(messages, e.Message)
		}
		return nil, fmt.Errorf("GraphQL errors: %s", strings.Join(messages, "; "))
	}

	return gqlResp.Data, nil
}

// EnsureAssets creates the specified assets in Rafiki
func EnsureAssets(cfg Config, assets []Asset) error {
	// First, list existing assets
	req := GraphQLRequest{
		Query: listAssetsQuery,
		Variables: map[string]interface{}{
			"first": 200,
		},
	}

	data, err := graphqlRequest(req, cfg)
	existingCodes := make(map[string]bool)

	if err != nil {
		fmt.Printf("Warning: Asset list failed, continuing to create assets... %v\n", err)
	} else {
		var assetsResp AssetsResponse
		if err := json.Unmarshal(data, &assetsResp); err == nil {
			for _, edge := range assetsResp.Assets.Edges {
				existingCodes[edge.Node.Code] = true
			}
		}
	}

	// Create missing assets
	for _, asset := range assets {
		if existingCodes[asset.Code] {
			fmt.Printf("  ✓ Asset %s already exists\n", asset.Code)
			continue
		}

		fmt.Printf("  Creating asset %s...\n", asset.Code)
		createReq := GraphQLRequest{
			Query: createAssetMutation,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"code":  asset.Code,
					"scale": asset.Scale,
				},
			},
		}

		_, err := graphqlRequest(createReq, cfg)
		if err != nil {
			errMsg := strings.ToLower(err.Error())
			if strings.Contains(errMsg, "already exists") || strings.Contains(errMsg, "duplicate") {
				fmt.Printf("  ✓ Asset %s already exists\n", asset.Code)
				continue
			}
			return fmt.Errorf("failed to create asset %s: %w", asset.Code, err)
		}

		fmt.Printf("  ✓ Asset %s created\n", asset.Code)
	}

	return nil
}

// WaitForReady waits for Rafiki GraphQL endpoint to be ready
func WaitForReady(cfg Config, timeoutSeconds int) error {
	start := time.Now()
	timeout := time.Duration(timeoutSeconds) * time.Second
	pollInterval := 2 * time.Second

	for {
		// Try a simple introspection query to check if the endpoint is ready
		req := GraphQLRequest{
			Query: `query { __typename }`,
		}

		_, err := graphqlRequest(req, cfg)
		if err == nil {
			return nil // Success!
		}

		// Check if we've exceeded the timeout
		if time.Since(start) >= timeout {
			return fmt.Errorf("timeout waiting for Rafiki after %v: %w", timeout, err)
		}

		// Wait before retrying
		time.Sleep(pollInterval)
	}
}

// EnsureLiquidity adds liquidity to the specified assets
func EnsureLiquidity(cfg Config, assets []Asset) error {
	for _, asset := range assets {
		// Get asset by code and scale
		getReq := GraphQLRequest{
			Query: getAssetByCodeAndScaleQuery,
			Variables: map[string]interface{}{
				"code":  asset.Code,
				"scale": asset.Scale,
			},
		}

		data, err := graphqlRequest(getReq, cfg)
		if err != nil {
			fmt.Printf("  ⚠ Lookup failed for %s: %v\n", asset.Code, err)
			continue
		}

		var assetResp AssetByCodeResponse
		if err := json.Unmarshal(data, &assetResp); err != nil {
			fmt.Printf("  ⚠ Failed to parse asset response for %s: %v\n", asset.Code, err)
			continue
		}

		if assetResp.AssetByCodeAndScale.ID == "" {
			fmt.Printf("  ⚠ Skipping liquidity for %s: asset id not found\n", asset.Code)
			continue
		}

		// Calculate amount: 100,000 * 10^scale
		baseAmount := int64(100000)
		scale := int64(asset.Scale)
		multiplier := int64(1)
		for i := int64(0); i < scale; i++ {
			multiplier *= 10
		}
		amount := baseAmount * multiplier

		fmt.Printf("  Depositing liquidity for %s: %d (scale %d)\n", asset.Code, amount, asset.Scale)

		depositReq := GraphQLRequest{
			Query: depositAssetLiquidityMutation,
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"id":             uuid.New().String(),
					"assetId":        assetResp.AssetByCodeAndScale.ID,
					"amount":         strconv.FormatInt(amount, 10),
					"idempotencyKey": uuid.New().String(),
				},
			},
		}

		data, err = graphqlRequest(depositReq, cfg)
		if err != nil {
			fmt.Printf("  ⚠ Liquidity deposit error for %s: %v\n", asset.Code, err)
			continue
		}

		var depositResp DepositLiquidityResponse
		if err := json.Unmarshal(data, &depositResp); err != nil {
			fmt.Printf("  ⚠ Failed to parse deposit response for %s: %v\n", asset.Code, err)
			continue
		}

		if !depositResp.DepositAssetLiquidity.Success {
			fmt.Printf("  ✗ Liquidity deposit failed for %s\n", asset.Code)
		} else {
			fmt.Printf("  ✓ Liquidity deposited for %s\n", asset.Code)
		}
	}

	return nil
}
