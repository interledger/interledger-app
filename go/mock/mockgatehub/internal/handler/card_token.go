package handler

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// cardDataTokenSecret is now sourced from config (see Config.CardDataTokenSecret).
// The const is kept removed deliberately: a hard-coded signing key turns this
// endpoint into a public encryption oracle.

// cardDataTokenTTL is how long a generated card-data token is valid.
const cardDataTokenTTL = 5 * time.Minute

// cardDataTokenMaxTTL is the maximum lifetime mockgatehub will accept on an
// incoming card-data token regardless of what `exp` claims. Since the data
// endpoint is unauthenticated at the HMAC layer, we reject tokens whose
// advertised validity exceeds this bound to prevent long-lived tokens from
// being reused as an encryption oracle.
const cardDataTokenMaxTTL = 10 * time.Minute

// cardDataPath is the path component of the card-data fetch endpoint.
const cardDataPath = "/cards/v1/token/card-data/data"

// CardDataClaims is the JWT payload mockgatehub embeds in card-data tokens.
type CardDataClaims struct {
	CardID    string `json:"cardId"`
	PublicKey string `json:"publicKey,omitempty"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

// generateCardDataJWT builds an HS256 JWT with the supplied claims, signed
// using the provided secret.
func generateCardDataJWT(secret string, claims CardDataClaims) (string, error) {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	encHeader := base64.RawURLEncoding.EncodeToString(headerBytes)
	encPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signingInput := encHeader + "." + encPayload

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return signingInput + "." + sig, nil
}

// parseCardDataJWT verifies the HS256 signature, requires a valid non-zero
// `exp` claim within an enforced maximum TTL window, and returns the claims.
func parseCardDataJWT(secret, token string) (*CardDataClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return nil, errors.New("invalid token signature")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid token payload: %w", err)
	}

	var claims CardDataClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("invalid token claims: %w", err)
	}

	// Require a valid `exp`: missing/zero values are treated as invalid to
	// avoid non-expiring tokens on an unauthenticated endpoint.
	if claims.ExpiresAt <= 0 {
		return nil, errors.New("token missing exp claim")
	}
	now := time.Now().Unix()
	if now > claims.ExpiresAt {
		return nil, errors.New("token expired")
	}
	// Enforce an upper bound on advertised validity so a malicious or
	// misconfigured issuer cannot mint extremely long-lived tokens.
	if claims.ExpiresAt-now > int64(cardDataTokenMaxTTL.Seconds()) {
		return nil, errors.New("token exceeds maximum allowed TTL")
	}

	return &claims, nil
}

// encryptWithBase64SPKI parses a base64-encoded SPKI RSA public key (matching
// the format produced by the browser's `crypto.subtle.exportKey('spki', ...)`
// followed by base64) and encrypts plaintext using RSA PKCS#1 v1.5 padding.
// It returns the base64-encoded ciphertext, matching the format protea's
// `decryptWithPrivateKey` expects.
func encryptWithBase64SPKI(publicKeyB64 string, plaintext []byte) (string, error) {
	derBytes, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return "", fmt.Errorf("public key is not valid base64: %w", err)
	}

	pubAny, err := x509.ParsePKIXPublicKey(derBytes)
	if err != nil {
		return "", fmt.Errorf("public key is not a valid SPKI key: %w", err)
	}

	rsaPub, ok := pubAny.(*rsa.PublicKey)
	if !ok {
		return "", errors.New("public key is not RSA")
	}

	cipher, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPub, plaintext)
	if err != nil {
		return "", fmt.Errorf("encryption failed: %w", err)
	}

	return base64.StdEncoding.EncodeToString(cipher), nil
}
