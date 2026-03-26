//go:build e2e
// +build e2e

package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
)

type cryptoState struct {
	signingPrivateJWK    string
	encryptionPrivateJWK string
	signingPublicKey     *rsa.PublicKey
	encryptionPrivateKey *rsa.PrivateKey
}

var webhookCryptoState cryptoState

func webhookCryptoComposeEnv() ([]string, error) {
	signingPrivate, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generate signing key: %w", err)
	}
	encryptionPrivate, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generate encryption key: %w", err)
	}

	signingJWK, err := marshalPrivateJWK(signingPrivate)
	if err != nil {
		return nil, err
	}
	encryptionJWK, err := marshalPrivateJWK(encryptionPrivate)
	if err != nil {
		return nil, err
	}

	webhookCryptoState = cryptoState{
		signingPrivateJWK:    signingJWK,
		encryptionPrivateJWK: encryptionJWK,
		signingPublicKey:     &signingPrivate.PublicKey,
		encryptionPrivateKey: encryptionPrivate,
	}

	return []string{
		"MOCKPTI_WEBHOOK_SIGNING_JWK=" + signingJWK,
		"MOCKPTI_WEBHOOK_ENCRYPTION_JWK=" + encryptionJWK,
	}, nil
}

func parseWebhookPayload(raw []byte) (map[string]interface{}, bool, error) {
	// First try plain JSON payload (webhook crypto disabled/fallback mode)
	plain := map[string]interface{}{}
	if err := json.Unmarshal(raw, &plain); err == nil {
		if _, hasCipher := plain["ciphertext"]; !hasCipher {
			return plain, false, nil
		}
	}

	decrypted, err := jwe.Decrypt(raw, jwe.WithKey(jwa.RSA_OAEP_256, webhookCryptoState.encryptionPrivateKey))
	if err != nil {
		return nil, false, fmt.Errorf("decrypt webhook payload: %w", err)
	}

	verified, err := jws.Verify(decrypted, jws.WithKey(jwa.RS512, webhookCryptoState.signingPublicKey))
	if err != nil {
		return nil, false, fmt.Errorf("verify webhook payload: %w", err)
	}

	payload := map[string]interface{}{}
	if err := json.Unmarshal(verified, &payload); err != nil {
		return nil, false, fmt.Errorf("unmarshal verified payload: %w", err)
	}

	return payload, true, nil
}

func marshalPrivateJWK(key *rsa.PrivateKey) (string, error) {
	jwkKey, err := jwk.FromRaw(key)
	if err != nil {
		return "", fmt.Errorf("create jwk from private key: %w", err)
	}

	b, err := json.Marshal(jwkKey)
	if err != nil {
		return "", fmt.Errorf("marshal jwk: %w", err)
	}

	return string(b), nil
}
