//go:build e2e
// +build e2e

package main

import (
	"crypto/ed25519"
	"crypto/rand"
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

func webhookCryptoComposeEnv() ([]string, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate signing key: %w", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("marshal private key: %w", err)
	}
	privPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}))

	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, fmt.Errorf("marshal public key: %w", err)
	}
	pubPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}))

	privB64 := base64.StdEncoding.EncodeToString([]byte(privPEM))
	pubOneLine := strings.ReplaceAll(strings.TrimSpace(pubPEM), "\n", `\n`)

	webhookCryptoState = cryptoState{
		signingPrivateKey: priv,
		signingPublicKey:  pub,
	}

	return []string{
		"MOCKPTI_WEBHOOK_SIGNING_KEY_B64=" + privB64,
		"PTI_PUBLIC_KEY_JWK=" + pubOneLine,
	}, nil
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
