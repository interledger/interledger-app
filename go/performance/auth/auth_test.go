package auth

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func responseWithBody(body []byte) *http.Response {
	return &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
}

// A rejected login: Kratos replies 400 with the whole flow, and the one useful
// sentence is buried in ui.messages. Trimmed from a real local response.
const rejectedLoginFlow = `{
  "id": "a5ca43e5-5f0a-4d65-87e7-d501a2513d8a",
  "type": "api",
  "request_url": "http://localhost:4433/self-service/login/api",
  "ui": {
    "action": "http://interledger.test/self-service/login?flow=a5ca43e5",
    "method": "POST",
    "nodes": [
      {"type":"input","group":"default","attributes":{"name":"csrf_token","type":"hidden"},"messages":[]},
      {"type":"input","group":"default","attributes":{"name":"identifier","type":"text"},"messages":[]},
      {"type":"input","group":"password","attributes":{"name":"password","type":"password"},"messages":[]}
    ],
    "messages": [
      {"id":4000006,"text":"The provided credentials are invalid, check for spelling mistakes in your password or username, email address, or phone number.","type":"error"}
    ]
  },
  "state": "choose_method"
}`

func TestFlowMessagesExtractsTheReason(t *testing.T) {
	reasons := flowMessages([]byte(rejectedLoginFlow))
	require.Len(t, reasons, 1)
	assert.Contains(t, reasons[0], "The provided credentials are invalid")
}

func TestFlowMessagesIncludesFieldErrors(t *testing.T) {
	// Provisioning hits these: a duplicate email or a rejected phone number comes
	// back as a per-node message, and the field name is the useful part.
	body := []byte(`{
      "ui": {
        "nodes": [
          {"attributes":{"name":"traits.email"},"messages":[{"text":"An account with the same identifier exists already.","type":"error"}]},
          {"attributes":{"name":"traits.phone"},"messages":[{"text":"must be a valid phone number","type":"error"}]}
        ],
        "messages": []
      }
    }`)

	reasons := flowMessages(body)
	require.Len(t, reasons, 2)
	assert.Equal(t, "traits.email: An account with the same identifier exists already.", reasons[0])
	assert.Equal(t, "traits.phone: must be a valid phone number", reasons[1])
}

func TestFlowMessagesIncludesTopLevelError(t *testing.T) {
	body := []byte(`{"error":{"message":"flow expired","reason":"the login flow has expired"}}`)

	reasons := flowMessages(body)
	require.Len(t, reasons, 2)
	assert.Equal(t, "flow expired", reasons[0])
	assert.Equal(t, "the login flow has expired", reasons[1])
}

func TestFlowMessagesSkipsNonErrors(t *testing.T) {
	// Flows carry info-level labels for every field; they are not failures.
	body := []byte(`{"ui":{"messages":[{"text":"Sign in with password","type":"info"}],"nodes":[]}}`)
	assert.Empty(t, flowMessages(body))
}

func TestFlowMessagesOnNonFlowBody(t *testing.T) {
	assert.Nil(t, flowMessages([]byte("not json at all")))
	assert.Empty(t, flowMessages([]byte(`{"unrelated":true}`)))
}

func TestKratosErrorAddsTheReason(t *testing.T) {
	underlying := errors.New("400 Bad Request")

	// No response and no SDK error body: the original error must survive untouched.
	assert.Equal(t, underlying, kratosError(nil, underlying))
}

func TestKratosErrorTruncatesUnrecognisedBodies(t *testing.T) {
	long := make([]byte, 2000)
	for i := range long {
		long[i] = 'x'
	}

	err := kratosError(responseWithBody(long), errors.New("500 Internal Server Error"))
	require.Error(t, err)
	assert.Less(t, len(err.Error()), 700, "an unparseable body is bounded, not dumped whole")
	assert.Contains(t, err.Error(), "500 Internal Server Error")
}

func TestKratosErrorPrefersFlowMessagesOverRawBody(t *testing.T) {
	err := kratosError(responseWithBody([]byte(rejectedLoginFlow)), errors.New("400 Bad Request"))
	require.Error(t, err)

	assert.Contains(t, err.Error(), "The provided credentials are invalid")
	assert.NotContains(t, err.Error(), "csrf_token", "the form definition is noise")
	assert.NotContains(t, err.Error(), "request_url")
}
