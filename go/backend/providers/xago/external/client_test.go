package external

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyTraefikRoutingHostRewritesLocalMockXagoRequests(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "http://mockxago:8080/v1/company/accounts/testdeposit", strings.NewReader(`{}`))
	require.NoError(t, err)

	applyTraefikRoutingHost(req)

	assert.Equal(t, "https", req.URL.Scheme)
	assert.Equal(t, "traefik:443", req.URL.Host)
	assert.Equal(t, "mockxago.interledger.test", req.Host)
}
