package external

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"
)

// adminSigningRoundTripper signs GraphQL Admin API requests for Rafiki v2
//
// Headers:
//
//	signature: t=<millis>, v<version>=<hmac_sha256_hex(timestamp + "." + canonical_json(request))>
//	tenant-id: <tenant_uuid>
//
// Env vars:
//
//	OPERATOR_TENANT_ID
//	ADMIN_API_SECRET
//	SIGNATURE_VERSION
//
// If required env vars are missing, it falls back to the base transport without modifying the request.
type adminSigningRoundTripper struct {
	base             http.RoundTripper
	operatorTenantID string
	adminAPISecret   string
	signatureVersion string
}

func (rt *adminSigningRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var bodyBytes []byte
	if req.Body != nil {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		bodyBytes = b
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	var requestData map[string]interface{}
	if len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, &requestData); err == nil {
			canonicalJSON, err := canonicalizeAndStringify(requestData)
			if err != nil {
				return nil, err
			}

			timestamp := time.Now().UnixMilli()
			payload := fmt.Sprintf("%d.%s", timestamp, canonicalJSON)

			h := hmac.New(sha256.New, []byte(rt.adminAPISecret))
			_, _ = h.Write([]byte(payload))
			digest := hex.EncodeToString(h.Sum(nil))

			req.Header.Set("signature", fmt.Sprintf("t=%d, v%s=%s", timestamp, rt.signatureVersion, digest))
			req.Header.Set("tenant-id", rt.operatorTenantID)
		}
	}

	if req.Body != nil && len(bodyBytes) > 0 {
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	return rt.base.RoundTrip(req)
}

func canonicalize(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		result := make(map[string]interface{}, len(val))
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
