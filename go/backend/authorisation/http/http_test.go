package http_test

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"

	"gitlab.com/fynbos/backend/wallets"

	wallets_mock "gitlab.com/fynbos/backend/wallets/client/mock"

	"github.com/golang/mock/gomock"

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
	if strings.Contains(r.URL.Path, "jwks.json") {
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
	ctrl := gomock.NewController(t)
	wc := wallets_mock.NewMockClient(ctrl)
	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx), wc)
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	keyID := "test"
	opserver := httptest.NewServer(keyHandler{
		Keys: []authorisation.Jwk{
			{
				Kty: "OKP",
				Kid: keyID,
				Alg: "EdDSA",
				Crv: "Ed25519",
				Use: "sign",
				X:   base64.StdEncoding.EncodeToString(pub),
			},
		},
		T: t,
	})
	t.Cleanup(func() {
		opserver.Close()
	})

	clientPaymentPointer := opserver.URL
	waURL, err := url.Parse(clientPaymentPointer)
	require.NoError(t, err)
	wa := wallets.TestAddress(t, waURL)

	wc.EXPECT().GetFromAddress(gomock.Any(), clientPaymentPointer).Return(&wallets.Wallet{
		ID:        uuid.NewString(),
		Addresses: []wallets.Address{wa},
	}, nil).AnyTimes()
	_, err = ops.CreateClient(ctx, b, clientPaymentPointer)
	require.NoError(t, err)

	gr := map[string]interface{}{
		"client": clientPaymentPointer,
		"access_token": []map[string]interface{}{
			{
				"access": []map[string]interface{}{
					{
						"type":       "incoming-payment",
						"actions":    []string{"write", "read"},
						"identifier": clientPaymentPointer,
					},
				},
				"label": "test",
			},
		},
	}
	body, err := json.Marshal(gr)
	require.NoError(t, err)

	digest, err := httpmessagesignatures.CreateContentDigest(ctx, body, []string{"sha-256"})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "http://auth.interledger.test/grant", bytes.NewBuffer(body))
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
		[]string{"Content-Digest"},
	)
	require.NoError(t, err)

	r := httptest.NewRecorder()
	handler := fynbos_http.AuthorisationHTTPHandler(b)

	handler.ServeHTTP(r, req)
	assert.Equal(t, http.StatusOK, r.Result().StatusCode)
}
