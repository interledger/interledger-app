//go:build e2e

package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Private RSA state kept on the test context so the generated public key used
// to issue a card-data token matches the private key we try to decrypt with
// later in the scenario.
var cardDataPrivateKey *rsa.PrivateKey
var cardDataToken string
var cardDataLinkHref string

func (tc *TestContext) generateCardDataKeyPair() (string, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", err
	}
	cardDataPrivateKey = key

	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(der), nil
}

// postCardTokenWithPublicKey issues a POST /cards/v1/token/card-data including
// an RSA publicKey field so the scenario can later decrypt the card data.
func (tc *TestContext) postCardTokenWithPublicKey(path string) error {
	pubKeyB64, err := tc.generateCardDataKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate key pair: %w", err)
	}
	body := map[string]interface{}{
		"cardId":    tc.cardID,
		"publicKey": pubKeyB64,
	}
	_, err = tc.sendWithManagedUserHeader("POST", path, body)
	if err != nil {
		return err
	}

	var result struct {
		Token string `json:"token"`
		Links []struct {
			Href   string `json:"href"`
			Method string `json:"method"`
			Rel    string `json:"rel"`
		} `json:"links"`
	}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}
	cardDataToken = result.Token
	if len(result.Links) > 0 {
		cardDataLinkHref = result.Links[0].Href
	}
	return nil
}

// responseContainsJWT verifies the response's `token` is a JWT whose payload
// carries the expected cardId and a non-empty publicKey claim.
func (tc *TestContext) responseContainsJWT() error {
	var result struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	parts := strings.Split(result.Token, ".")
	if len(parts) != 3 {
		return fmt.Errorf("token is not a JWT: %s", result.Token)
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("cannot decode JWT payload: %w", err)
	}

	var claims struct {
		CardID    string `json:"cardId"`
		PublicKey string `json:"publicKey"`
	}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return fmt.Errorf("cannot parse JWT claims: %w", err)
	}

	if claims.CardID != tc.cardID {
		return fmt.Errorf("expected cardId %q in JWT, got %q", tc.cardID, claims.CardID)
	}
	if claims.PublicKey == "" {
		return fmt.Errorf("JWT is missing publicKey claim")
	}
	return nil
}

// firstLinkHrefIsAbsoluteURLTo asserts the first link is an absolute URL that
// ends with the expected path.
func (tc *TestContext) firstLinkHrefIsAbsoluteURLTo(expectedPath string) error {
	var result struct {
		Links []struct {
			Href string `json:"href"`
		} `json:"links"`
	}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}
	if len(result.Links) == 0 {
		return fmt.Errorf("no links in response")
	}
	href := result.Links[0].Href
	if !(strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://")) {
		return fmt.Errorf("expected absolute URL, got %q", href)
	}
	if !strings.HasSuffix(href, expectedPath) {
		return fmt.Errorf("expected href to end with %q, got %q", expectedPath, href)
	}
	return nil
}

// cardDataTokenIssuedForCard is a setup step that performs the POST
// /cards/v1/token/card-data call with a freshly generated RSA key pair so the
// subsequent browser-style GET can be decrypted.
func (tc *TestContext) cardDataTokenIssuedForCard() error {
	return tc.postCardTokenWithPublicKey("/cards/v1/token/card-data")
}

// browserGetCardDataWithBearer simulates the browser following the link
// returned by the card-data token endpoint: GET with Bearer auth and no HMAC
// headers at all.
func (tc *TestContext) browserGetCardDataWithBearer() error {
	if cardDataLinkHref == "" {
		return fmt.Errorf("no card-data link href has been captured")
	}
	return tc.browserGetWithToken(cardDataLinkHref, cardDataToken)
}

// browserGetCardDataWithInvalidToken tests the 401 path: invalid Bearer, no HMAC.
func (tc *TestContext) browserGetCardDataWithInvalidToken(path string) error {
	return tc.browserGetWithToken(tc.baseURL+path, "not-a-valid-token")
}

func (tc *TestContext) browserGetWithToken(fullURL, token string) error {
	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	// Deliberately NO HMAC headers — this simulates the browser calling the
	// endpoint directly, which is why it must be excluded from HMAC auth.
	resp, err := tc.doRequest(req)
	if err != nil {
		return err
	}
	tc.lastResponse = resp
	return nil
}

// responseContainsCypherField checks the JSON body contains a non-empty
// `cypher` field.
func (tc *TestContext) responseContainsCypherField() error {
	var result map[string]string
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}
	if result["cypher"] == "" {
		return fmt.Errorf("response missing cypher field: %s", string(tc.lastResponseBody))
	}
	return nil
}

// cypherDecryptsToSensitiveData decrypts the response's `cypher` field with
// the scenario's private key and verifies the JSON shape the wallet frontend
// expects (Pan / ExpiryDate / Cvc2).
func (tc *TestContext) cypherDecryptsToSensitiveData() error {
	if cardDataPrivateKey == nil {
		return fmt.Errorf("no card-data private key for this scenario")
	}

	var result map[string]string
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	cipher, err := base64.StdEncoding.DecodeString(result["cypher"])
	if err != nil {
		return fmt.Errorf("cypher is not base64: %w", err)
	}

	plain, err := rsa.DecryptPKCS1v15(rand.Reader, cardDataPrivateKey, cipher)
	if err != nil {
		return fmt.Errorf("RSA decryption failed: %w", err)
	}

	var sensitive struct {
		Pan        string `json:"Pan"`
		ExpiryDate string `json:"ExpiryDate"`
		Cvc2       string `json:"Cvc2"`
	}
	if err := json.Unmarshal(plain, &sensitive); err != nil {
		return fmt.Errorf("decrypted payload is not JSON: %w", err)
	}
	if sensitive.Pan == "" || sensitive.ExpiryDate == "" || sensitive.Cvc2 == "" {
		return fmt.Errorf("decrypted payload missing required fields: %s", string(plain))
	}
	return nil
}
