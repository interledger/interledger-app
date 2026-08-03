package provision

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	backenddb "github.com/interledger/interledger-app/go/backend/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func responseWithBody(body []byte) *http.Response {
	return &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
}

func TestFormatRegistrationErrorPreservesKratosReason(t *testing.T) {
	body := []byte(`{
		"ui": {
			"nodes": [
				{"attributes":{"name":"traits.email"},"messages":[{"text":"An account with the same identifier exists already.","type":"error"}]}
			],
			"messages": []
		}
	}`)

	err := formatRegistrationError(responseWithBody(body), errors.New("400 Bad Request"))
	require.Error(t, err)
	assert.ErrorContains(t, err, "400 Bad Request")
	assert.ErrorContains(t, err, "same identifier")
}

func TestLogProgressWritesReadableStatus(t *testing.T) {
	var buf bytes.Buffer
	logProgress(&buf, "perf-za-001", "registering identity")
	assert.Equal(t, "  perf-za-001: registering identity\n", buf.String())
}

func TestCheckHTTPReachableSucceedsForHealthyEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	require.NoError(t, checkHTTPReachable(srv.URL, time.Second))
}

func TestCheckHTTPReachableFailsForClosedPort(t *testing.T) {
	require.Error(t, checkHTTPReachable("http://127.0.0.1:1", time.Second))
}

func TestFundingErrorAddsLocalServiceHintForTimeouts(t *testing.T) {
	err := fundingError("Xago", context.DeadlineExceeded)
	require.Error(t, err)
	assert.ErrorContains(t, err, "Xago provider timed out")
	assert.ErrorContains(t, err, "local xago service")
}

func TestApproveKYCSeedsIndividualDetails(t *testing.T) {
	ctx := context.Background()
	dbc := backenddb.MigrateTestDB(t, ctx, "")
	defer dbc.Close()

	userID := uuid.NewString()
	walletID := uuid.NewString()
	_, err := dbc.ExecContext(ctx, `INSERT INTO wallets (id, name, country) VALUES ($1, $2, 'ZA')`, walletID, "perf-wallet")
	require.NoError(t, err)
	_, err = dbc.ExecContext(ctx, `INSERT INTO user_wallets (user_id, wallet_id) VALUES ($1, $2)`, userID, walletID)
	require.NoError(t, err)

	err = approveKYC(ctx, dbc, userID, countrySpec{code: "za", country: "ZA", currencyCode: "ZAR"}, "perf-za-001")
	require.NoError(t, err)

	var count int
	err = dbc.GetContext(ctx, &count, `SELECT COUNT(*) FROM individual_kyc_details WHERE wallet_id=$1`, walletID)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	var inquiryCount int
	err = dbc.GetContext(ctx, &inquiryCount, `SELECT COUNT(*) FROM kyc_persona_inquiries WHERE wallet_id=$1 AND state='approved'`, walletID)
	require.NoError(t, err)
	require.Equal(t, 1, inquiryCount)

	var personaAccountCount int
	err = dbc.GetContext(ctx, &personaAccountCount, `SELECT COUNT(*) FROM kyc_persona_accounts WHERE wallet_id=$1`, walletID)
	require.NoError(t, err)
	require.Equal(t, 1, personaAccountCount)

	var got struct {
		CountryCode string `db:"country_code"`
		FirstName   string `db:"first_name"`
		LastName    string `db:"last_name"`
		IPAddr      string `db:"ip_address"`
	}
	err = dbc.GetContext(ctx, &got, `SELECT country_code, first_name, last_name, ip_address FROM individual_kyc_details WHERE wallet_id=$1 ORDER BY revision DESC LIMIT 1`, walletID)
	require.NoError(t, err)
	assert.Equal(t, "ZA", got.CountryCode)
	assert.Equal(t, "Perf", got.FirstName)
	assert.Equal(t, "perf-za-001", got.LastName)
	assert.Equal(t, "127.0.0.1", got.IPAddr)
}
