package ops_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/providers/gatehub/ops"
)

func TestVerifyWebhook(t *testing.T) {
	if os.Getenv("GATEHUB_WEBHOOK_SECRET") == "" {
		t.SkipNow()
	}
	key, err := hex.DecodeString(os.Getenv("GATEHUB_WEBHOOK_SECRET"))
	if err != nil {
		require.NoError(t, err)
	}
	payload := `{"uuid":"44534833-c733-4c84-aac8-4daf758c0723","timestamp":"2024-04-22T12:12:31.220Z","event_type":"id.verification.accepted","user_uuid":"19227839-caa1-458f-a5ec-a3f03aa3e0e5","environment":"sandbox","data":{"gateway":"GateHub Crypto","verified":{"status":1,"short":"accepted"}}}`

	r, err := http.NewRequest(http.MethodPost, "https://en8pxah8zi194.x.pipedream.net", bytes.NewBufferString(payload))
	require.NoError(t, err)
	r.Header.Set("x-gh-webhook-timestamp", "2024-04-22T12:12:31.220Z")
	r.Header.Set("x-gh-webhook-signature", "785fd686e035d642a04c81f9c65bef1fd1bf69f9280ecf73b6ee41c55910fcd7")

	body, err := ops.Verify(context.Background(), r, key)
	require.NoError(t, err)

	assert.Equal(t, payload, string(body))
}
