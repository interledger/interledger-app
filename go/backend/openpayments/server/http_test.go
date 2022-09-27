package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
	test_utils "gitlab.com/fynbos/backend/utils"
)

func TestGetHandler(t *testing.T) {
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)

	b := NewTestBackends(t, db)

	cases := []struct {
		name       string
		getPath    string
		pointer    *openpayments.PaymentPointer
		statusCode int
	}{
		{
			name:       "not_found",
			getPath:    "https://fynbos.local.me/not_real",
			pointer:    nil,
			statusCode: http.StatusNotFound,
		},
		{
			name:    "found",
			getPath: "https://fynbos.local.me/found_me",
			pointer: &openpayments.PaymentPointer{
				Alias:      "Some Alias",
				Asset:      "USD",
				AssetScale: 2,
			},
			statusCode: http.StatusOK,
		},
	}

	for _, tc := range cases {

		req, err := http.NewRequest(http.MethodGet, tc.getPath, nil)
		require.NoError(t, err)

		fmt.Println(req.URL.String())

		// Setup the payment pointer
		if tc.pointer != nil {
			pp := *tc.pointer
			pp.URL = req.URL.String()
			err := ops.CreatePaymentPointer(ctx, b, *tc.pointer)
			require.NoError(t, err)
		}

		rr := httptest.NewRecorder()
		handler := catchAllHandler(b)
		handler.ServeHTTP(rr, req)

		require.Equal(t, tc.statusCode, rr.Code)

		if tc.pointer == nil {
			return
		}

		var pp openpayments.PaymentPointer

		err = json.NewDecoder(rr.Body).Decode(&pp)
		require.NoError(t, err)

		assert.Equal(t, tc.pointer.Alias, pp.Alias)
		assert.Equal(t, tc.pointer.Asset, pp.Asset)
		assert.Equal(t, tc.pointer.AssetScale, pp.AssetScale)
		assert.Equal(t, req.URL.String(), pp.URL)
	}
}
