package persona_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/kyc/persona"
)

func TestClient_ValidateWebhook(t *testing.T) {
	pc := persona.New()

	req, err := http.NewRequest(http.MethodPost, "some_url", bytes.NewReader([]byte("this is content")))
	require.NoError(t, err)
	req.Header.Set("Persona-Signature", "t=123,v1=sjpFkbUQTj9Ej8IwV3gqRWuVM7ZP7jeS5zLWuyVdAp0=")

	valid := pc.ValidateWebhook(req)
	require.True(t, valid)
}
