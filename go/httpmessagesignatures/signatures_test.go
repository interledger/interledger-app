package httpmessagesignatures_test

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dunglas/httpsfv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/authorisation"
	"gitlab.com/fynbos/httpmessagesignatures"
)

type testSigner struct{}

func (s testSigner) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error) {
	return digest, nil
}

func (s testSigner) Public() crypto.PublicKey {
	return nil
}

type testEd25519Signer struct {
	privateKey ed25519.PrivateKey
}

func (s testEd25519Signer) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error) {
	return ed25519.Sign(s.privateKey, digest), nil
}

func (s testEd25519Signer) Public() crypto.PublicKey {
	return nil
}

type testVerifier struct {
	valid bool
}

func (s testVerifier) Verify(crypto.PublicKey, []byte, []byte) bool {
	return s.valid
}

type testEd25519Verifier struct {
}

func (s testEd25519Verifier) Verify(publicKey crypto.PublicKey, digest []byte, signature []byte) bool {
	return ed25519.Verify(publicKey.(ed25519.PublicKey), digest, signature)
}

func TestSignRequest(t *testing.T) {
	t.Run("builds signature base correctly", func(st *testing.T) {
		components := []string{"@method", "@target_uri", "@authority", "@scheme", "@request-target", "content-digest", "@path", "@query", "@query-params"}
		params := httpmessagesignatures.SignatureParams{
			Created: 1618884475,
			KeyID:   "test123",
		}
		expectedBase := []string{
			`"@method": POST`,
			`"@target_uri": https://www.example.com/movies/titanic/123?version=1`,
			`"@authority": www.example.com`,
			`"@scheme": https`,
			`"@request-target": /movies/titanic/123?version=1`,
			`"content-digest": digest`,
			`"@path": /movies/titanic/123`,
			`"@query": ?version=1`,
			`"@query-params"; name="version": 1`,
			`"@signature-params": ("@method" "@target_uri" "@authority" "@scheme" "@request-target" "content-digest" "@path" "@query" "@query-params");created=1618884475;keyid="test123"`,
		}

		r := httptest.NewRequest("POST", "https://www.example.com/movies/titanic/123?version=1", nil)
		r.Header.Add("Content-Digest", "digest")
		require.NoError(st, httpmessagesignatures.SignRequest(context.Background(), r, testSigner{}, components, params))

		signatureDictionary, err := httpsfv.UnmarshalDictionary([]string{r.Header.Get("Signature")})
		require.NoError(st, err)
		signatureMember, exists := signatureDictionary.Get("sig-1")
		require.True(st, exists)
		signatureItem, ok := signatureMember.(httpsfv.Item)
		require.True(st, ok)

		assert.Equal(st, strings.Join(expectedBase, "\n"), string(signatureItem.Value.([]byte)))
		assert.Equal(st, `sig-1=("@method" "@target_uri" "@authority" "@scheme" "@request-target" "content-digest" "@path" "@query" "@query-params");created=1618884475;keyid="test123"`, r.Header.Get("Signature-Input"))
	})

	t.Run("works with ed25519", func(st *testing.T) {
		// https://www.ietf.org/archive/id/draft-ietf-httpbis-message-signatures-15.html#appendix-B.2.6
		privateKeyPEM := `-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEIJ+DYvh6SEqVTm50DFtMDoQikTmiCqirVv9mWG9qfSnF
-----END PRIVATE KEY-----`

		b, _ := pem.Decode([]byte(privateKeyPEM))
		key, err := x509.ParsePKCS8PrivateKey(b.Bytes)
		require.NoError(t, err)

		privateKey, ok := key.(ed25519.PrivateKey)
		require.True(t, ok)

		components := []string{"date", "@method", "@path", "@authority", "content-type", "content-length"}
		params := httpmessagesignatures.SignatureParams{
			Created: 1618884473,
			KeyID:   "test-key-ed25519",
		}

		r := httptest.NewRequest("POST", "https://example.com/foo", nil)
		r.Header.Set("date", "Tue, 20 Apr 2021 02:07:55 GMT")
		r.Header.Set("content-type", "application/json")
		r.Header.Set("content-length", "18")
		require.NoError(st, httpmessagesignatures.SignRequest(context.Background(), r, testEd25519Signer{privateKey}, components, params))

		assert.Equal(st, `sig-1=("date" "@method" "@path" "@authority" "content-type" "content-length");created=1618884473;keyid="test-key-ed25519"`, r.Header.Get("Signature-Input"))
		assert.Equal(st, "sig-1=:wqcAqbmYJ2ji2glfAMaRy4gruYYnx2nEFN2HN6jrnDnQCK1u02Gb04v9EDgwUPiu4A0w6vuQv5lIp5WPpBKRCw==:", r.Header.Get("Signature"))
	})
}

func TestVerifyReqeust(t *testing.T) {
	components := []string{"@method", "@target_uri", "@authority", "@scheme", "@request-target", "content-digest", "@path", "@query", "@query-params"}
	params := httpmessagesignatures.SignatureParams{
		Created: 1618884475,
		KeyID:   "test123",
	}

	t.Run("returns true when verification succeeds", func(st *testing.T) {
		r := httptest.NewRequest("POST", "https://www.example.com/movies/titanic/123?version=1", nil)
		r.Header.Add("Content-Digest", "digest")
		require.NoError(t, httpmessagesignatures.SignRequest(context.Background(), r, testSigner{}, components, params))

		assert.True(
			st,
			httpmessagesignatures.VerifySignature(
				context.Background(),
				r,
				[]byte{},
				testVerifier{true},
			),
		)
	})

	t.Run("returns false when verification fails", func(st *testing.T) {
		r := httptest.NewRequest("POST", "https://www.example.com/movies/titanic/123?version=1", nil)
		r.Header.Add("Content-Digest", "digest")
		require.NoError(t, httpmessagesignatures.SignRequest(context.Background(), r, testSigner{}, components, params))

		assert.False(
			st,
			httpmessagesignatures.VerifySignature(
				context.Background(),
				r,
				[]byte{},
				testVerifier{false},
			),
		)
	})

	t.Run("works with ed25519", func(st *testing.T) {
		// https://www.ietf.org/archive/id/draft-ietf-httpbis-message-signatures-15.html#appendix-B.2.6
		r := httptest.NewRequest("POST", "https://example.com/foo", nil)
		r.Header.Set("date", "Tue, 20 Apr 2021 02:07:55 GMT")
		r.Header.Set("content-type", "application/json")
		r.Header.Set("content-length", "18")
		r.Header.Set("Signature-Input", `sig-1=("date" "@method" "@path" "@authority" "content-type" "content-length");created=1618884473;keyid="test-key-ed25519"`)
		r.Header.Set("Signature", "sig-1=:wqcAqbmYJ2ji2glfAMaRy4gruYYnx2nEFN2HN6jrnDnQCK1u02Gb04v9EDgwUPiu4A0w6vuQv5lIp5WPpBKRCw==:")

		var jwk authorisation.Jwk
		err := json.Unmarshal([]byte(`{
  "kty": "OKP",
  "crv": "Ed25519",
  "kid": "test-key-ed25519",
  "d": "n4Ni-HpISpVObnQMW0wOhCKROaIKqKtW_2ZYb2p9KcU",
  "x": "JrQLj5P/89iXES9+vFgrIy29clF9CC/oPPsw3c5D0bs="
}`), &jwk)
		require.NoError(st, err)
		pubKeyBytes, err := base64.StdEncoding.DecodeString(jwk.X)
		require.NoError(st, err)

		assert.True(st, httpmessagesignatures.VerifySignature(context.Background(), r, ed25519.PublicKey(pubKeyBytes), testEd25519Verifier{}))
	})
}
