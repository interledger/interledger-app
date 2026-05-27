package persona_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/kyc/persona"
)

func TestClient_ValidateWebhook(t *testing.T) {
	pc := persona.New(persona.Config{})

	req, err := http.NewRequest(http.MethodPost, "some_url", bytes.NewReader([]byte("this is content")))
	require.NoError(t, err)
	req.Header.Set("Persona-Signature", "t=123,v1=b23a4591b5104e3f448fc23057782a456b9533b64fee3792e732d6bb255d029d")

	valid := pc.ValidateWebhook(req, []byte("this is content"))
	require.True(t, valid)
}
