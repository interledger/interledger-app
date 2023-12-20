package external

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientSignature(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err)

	req, err := http.NewRequest("GET", "users/123", nil)
	require.NoError(t, err)

	err = Sign(context.Background(), req, nil, key)
	require.NoError(t, err)

	assert.NotEmpty(t, req.Header.Get(ptiSignatureHeader))

	require.NoError(t, Verify(context.Background(), req, &key.PublicKey))
}
