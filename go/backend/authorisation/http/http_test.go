package http_test

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/authorisation"
	fynbos_http "gitlab.com/fynbos/backend/authorisation/http"
	"gitlab.com/fynbos/backend/authorisation/ops"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/httpmessagesignatures"
	"gotest.tools/assert"
)

type keyHandler struct {
	Keys []authorisation.Jwk `json:"keys"`
	T    *testing.T
}

func (h keyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, ".well-known/keys") {
		err := json.NewEncoder(w).Encode(h)
		require.NoError(h.T, err)
		return
	}

	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
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

func TestGrantRequest(t *testing.T) {
	ctx := context.Background()
	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx))
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	keyID := "test"
	opserver := httptest.NewServer(keyHandler{
		Keys: []authorisation.Jwk{
			{
				Kty: "OKP",
				Kid: keyID,
				Alg: "edDSA",
				Crv: "ed25519",
				Use: "sign",
				X:   base64.StdEncoding.EncodeToString(pub),
			},
		},
		T: t,
	})
	t.Cleanup(func() {
		opserver.Close()
	})

	_, err = ops.CreateClient(ctx, b, fmt.Sprintf("%s/.well-known/keys", opserver.URL))
	require.NoError(t, err)

	gr := authorisation.GrantRequest{
		Client: authorisation.ClientReq{
			Display: authorisation.Display{
				Name: "test",
				URI:  fmt.Sprintf("%s/.well-known/keys", opserver.URL),
			},
			Key: authorisation.Key{
				Proof: "httpsig",
				Jwk: authorisation.Jwk{
					Kty: "OKP",
					Kid: keyID,
					Alg: "edDSA",
					Crv: "ed25519",
					X:   base64.StdEncoding.EncodeToString(pub),
				},
			},
		},
		AccessToken: []authorisation.AccessTokenReq{{
			Access: []authorisation.Access{{
				Type:      "incoming-payments",
				Actions:   []string{"write", "read"},
				Locations: []string{"https://fynbos.me/bobby"},
				Datatypes: []string{"incoming-payments"},
			}},
			Label: "TestAccess1",
		}},
	}
	body, err := json.Marshal(gr)
	require.NoError(t, err)

	digest, err := httpmessagesignatures.CreateContentDigest(ctx, body, []string{"sha-256"})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "http://fynbos.test/grant", bytes.NewBuffer(body))
	req.Header.Set("Content-Digest", digest)
	err = httpmessagesignatures.SignRequest(
		ctx,
		req,
		testEd25519Signer{priv},
		[]string{"@method", "content-digest"},
		httpmessagesignatures.SignatureParams{
			Created: 1618884475,
			KeyID:   keyID,
		},
	)
	require.NoError(t, err)

	r := httptest.NewRecorder()
	handler := fynbos_http.AuthorisationHTTPHandler(b)

	handler.ServeHTTP(r, req)
	assert.Equal(t, http.StatusOK, r.Result().StatusCode)
}
