//go:build e2e
// +build e2e

package main

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"
)

type cryptoState struct {
	signingPrivateKey ed25519.PrivateKey
	signingPublicKey  ed25519.PublicKey
}

var webhookCryptoState cryptoState

const testWebhookSigningPrivateKeyPEM = `-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEIO49I4fkirtnEKgxcZsToTO9y5FS6sRddjmOik17QhVS
-----END PRIVATE KEY-----`

func webhookCryptoComposeEnv() ([]string, error) {
	block, _ := pem.Decode([]byte(testWebhookSigningPrivateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("decode private key PEM")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	priv, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not ed25519")
	}
	pub := priv.Public().(ed25519.PublicKey)

	webhookCryptoState = cryptoState{
		signingPrivateKey: priv,
		signingPublicKey:  pub,
	}

	return nil, nil
}

func parseWebhookPayload(raw []byte, sigHeader string) (map[string]interface{}, bool, error) {
	payload := map[string]interface{}{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, false, fmt.Errorf("unmarshal webhook payload: %w", err)
	}

	if sigHeader == "" || webhookCryptoState.signingPublicKey == nil {
		return payload, false, nil
	}

	signed := false
	for _, part := range strings.Split(sigHeader, ",") {
		part = strings.TrimSpace(part)
		if !strings.HasPrefix(part, "v1=") {
			continue
		}
		sigBytes, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(part, "v1="))
		if err != nil {
			continue
		}
		if ed25519.Verify(webhookCryptoState.signingPublicKey, raw, sigBytes) {
			signed = true
			break
		}
	}

	return payload, signed, nil
}
