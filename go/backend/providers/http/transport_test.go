package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	httplog "github.com/interledger/interledger-app/go/backend/providers/http"

	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testBackends struct {
	db *sqlx.DB
}

func (tb *testBackends) DB() *sqlx.DB {
	return tb.db
}

func TestLog(t *testing.T) {
	b := &testBackends{db: db.MigrateTestDB(t, context.Background())}
	client := &http.Client{
		Transport: httplog.NewTransport(http.DefaultTransport, b, nil),
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
		err := json.NewEncoder(w).Encode(map[string]string{
			"respKey": "456",
		})
		require.NoError(t, err)
	}))
	t.Cleanup(func() {
		ts.Close()
	})

	ctx := context.Background()
	ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
		Provider: "testapi",
		Context:  "testkey",
		Method:   "POST",
	})

	payload, err := json.Marshal(map[string]string{
		"reqKey": "123",
	})
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, "POST", ts.URL+"/", bytes.NewBuffer(payload))
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var data map[string]string
	require.NoError(t, json.Unmarshal(body, &data))
	assert.Equal(t, data["respKey"], "456")

	var log httplog.LogRecord
	err = b.db.GetContext(
		ctx,
		&log,
		"SELECT id, provider, context, method, request_body, request_path, response_body, response_status, created_at FROM external_api_logs WHERE context=$1 AND provider=$2;",
		"testkey",
		"testapi",
	)
	require.NoError(t, err)

	assert.Equal(t, "http://"+ts.Listener.Addr().String()+"/", log.RequestPath)
	assert.Equal(t, "{\"reqKey\":\"123\"}", log.RequestBody)
	assert.Equal(t, "{\"respKey\":\"456\"}\n", log.ResponseBody)
	assert.Equal(t, "200 OK", log.ResponseStatus)
	assert.Equal(t, "POST", log.Method)
}
