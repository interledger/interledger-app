package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	users_client "gitlab.com/fynbos/backend/user/client"

	"github.com/google/uuid"

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
	userClient := users_client.New(b, "fakeURL")

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
				WalletID:   uuid.NewString(),
				Alias:      "Some Alias",
				Asset:      "USD",
				AssetScale: 2,
				URL:        "https://fynbos.local.me/found_me",
			},
			statusCode: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			req, err := http.NewRequest(http.MethodGet, tc.getPath, nil)
			require.NoError(t, err)

			fmt.Println(req.URL.String())

			// Setup the payment pointer
			if tc.pointer != nil {
				userID := uuid.NewString()
				_, err = db.ExecContext(ctx, "INSERT INTO signups (id, user_id) VALUES ($1, $2)", uuid.NewString(), userID)
				require.NoError(t, err)
				// Create Wallets
				wallet, err := userClient.CreateNewWallet(ctx, userID, "test")
				require.NoError(t, err)

				tc.pointer.WalletID = wallet.ID

				err = ops.CreatePaymentPointer(ctx, b, *tc.pointer)
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
		})
	}
}

func TestHTTPCreateQuoteGet(t *testing.T) {
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)

	b := NewTestBackends(t, db)
	userClient := users_client.New(b, "fakeURL")

	cases := []struct {
		name       string
		args       openpayments.CreateQuoteArgs
		statusCode int
	}{
		{
			name: "success",
			args: openpayments.CreateQuoteArgs{
				SendPaymentPointer:    "https://fynbos.me/paysend",
				ReceivePaymentPointer: "https://fynbos.me/payrecv",
				ExpiresAt:             time.Now().Add(time.Hour),
				SendAmount: openpayments.Amount{
					Value:      100,
					Asset:      "USD",
					AssetScale: 2,
				}},
			statusCode: http.StatusCreated,
		},
	}

	for _, tc := range cases {

		sendUserID := uuid.NewString()
		recvUserID := uuid.NewString()
		// Create Signups
		_, err := db.ExecContext(ctx, "INSERT INTO signups (id, user_id) VALUES ($1, $2), ($3, $4)", uuid.NewString(), sendUserID, uuid.NewString(), recvUserID)
		require.NoError(t, err)
		// Create Wallets
		sendWallet, err := userClient.CreateNewWallet(ctx, sendUserID, "test")
		require.NoError(t, err)
		recvWallet, err := userClient.CreateNewWallet(ctx, recvUserID, "test")
		require.NoError(t, err)

		body, err := json.Marshal(tc.args)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/quotes", tc.args.SendPaymentPointer), bytes.NewReader(body))
		require.NoError(t, err)

		// Setup the payment pointers
		err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
			URL:        tc.args.SendPaymentPointer,
			Alias:      "alias",
			WalletID:   sendWallet.ID,
			Asset:      tc.args.SendAmount.Asset,
			AssetScale: tc.args.SendAmount.AssetScale,
		})
		require.NoError(t, err)
		err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
			URL:        tc.args.ReceivePaymentPointer,
			Alias:      "alias",
			WalletID:   recvWallet.ID,
			Asset:      tc.args.SendAmount.Asset,
			AssetScale: tc.args.SendAmount.AssetScale,
		})
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := catchAllHandler(b)
		handler.ServeHTTP(rr, req)

		require.Equal(t, tc.statusCode, rr.Code)

		respBytes, err := io.ReadAll(rr.Body)
		require.NoError(t, err)

		var q openpayments.Quote
		err = json.Unmarshal(respBytes, &q)
		require.NoError(t, err)

		assert.Equal(t, tc.args.ReceivePaymentPointer, q.PaymentPointer)
		assert.Equal(t, tc.args.SendAmount.Value, q.SendAmount.Value)
		assert.Equal(t, tc.args.SendAmount.Asset, q.SendAmount.Asset)
		assert.Equal(t, tc.args.SendAmount.AssetScale, q.SendAmount.AssetScale)

		// Do a get and get the same values
		req, err = http.NewRequest(http.MethodGet, q.ID, nil)
		require.NoError(t, err)

		rr = httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		respBytes, err = io.ReadAll(rr.Body)
		require.NoError(t, err)

		var lq openpayments.Quote
		err = json.Unmarshal(respBytes, &lq)
		require.NoError(t, err)

		assert.Equal(t, tc.args.ReceivePaymentPointer, lq.PaymentPointer)
		assert.Equal(t, tc.args.SendAmount.Value, lq.SendAmount.Value)
		assert.Equal(t, tc.args.SendAmount.Asset, lq.SendAmount.Asset)
		assert.Equal(t, tc.args.SendAmount.AssetScale, lq.SendAmount.AssetScale)
	}
}

func TestHTTPCreateIncomingPaymentGet(t *testing.T) {
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)

	b := NewTestBackends(t, db)
	userClient := users_client.New(b, "fakeURL")

	cases := []struct {
		name       string
		args       openpayments.CreateIncomingPaymentArgs
		statusCode int
	}{
		{
			name: "success",
			args: openpayments.CreateIncomingPaymentArgs{
				PaymentPointer: "https://fynbos.me/moneyplease",
				IncomingAmount: openpayments.Amount{
					Value:      100,
					Asset:      "USD",
					AssetScale: 2,
				},
				ExternalRef: "External",
				ExpiresAt:   time.Now().Add(time.Hour),
			},
			statusCode: http.StatusCreated,
		},
	}

	for _, tc := range cases {

		recvUserID := uuid.NewString()
		// Create Signups
		_, err := db.ExecContext(ctx, "INSERT INTO signups (id, user_id) VALUES ($1, $2)", uuid.NewString(), recvUserID)
		require.NoError(t, err)
		// Create Wallets
		recvWallet, err := userClient.CreateNewWallet(ctx, recvUserID, "test")
		require.NoError(t, err)

		body, err := json.Marshal(tc.args)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/incoming-payments", tc.args.PaymentPointer), bytes.NewReader(body))
		require.NoError(t, err)

		// Setup the payment pointers
		err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
			URL:        tc.args.PaymentPointer,
			Alias:      "alias",
			WalletID:   recvWallet.ID,
			Asset:      tc.args.IncomingAmount.Asset,
			AssetScale: tc.args.IncomingAmount.AssetScale,
		})
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := catchAllHandler(b)
		handler.ServeHTTP(rr, req)

		require.Equal(t, tc.statusCode, rr.Code)

		respBytes, err := io.ReadAll(rr.Body)
		require.NoError(t, err)

		var ip openpayments.IncomingPayment
		err = json.Unmarshal(respBytes, &ip)
		require.NoError(t, err)

		assert.Equal(t, tc.args.PaymentPointer, ip.PaymentPointer)
		assert.Equal(t, tc.args.IncomingAmount.Value, ip.IncomingAmount.Value)
		assert.Equal(t, tc.args.IncomingAmount.Asset, ip.IncomingAmount.Asset)
		assert.Equal(t, tc.args.IncomingAmount.AssetScale, ip.IncomingAmount.AssetScale)
		assert.Equal(t, tc.args.ExternalRef, ip.ExternalRef)

		// Do a get and get the same values
		req, err = http.NewRequest(http.MethodGet, ip.ID, nil)
		require.NoError(t, err)

		rr = httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		respBytes, err = io.ReadAll(rr.Body)
		require.NoError(t, err)

		var lip openpayments.IncomingPayment
		err = json.Unmarshal(respBytes, &lip)
		require.NoError(t, err)

		assert.Equal(t, tc.args.PaymentPointer, lip.PaymentPointer)
		assert.Equal(t, tc.args.IncomingAmount.Value, lip.IncomingAmount.Value)
		assert.Equal(t, tc.args.IncomingAmount.Asset, lip.IncomingAmount.Asset)
		assert.Equal(t, tc.args.IncomingAmount.AssetScale, lip.IncomingAmount.AssetScale)
		assert.Equal(t, tc.args.ExternalRef, lip.ExternalRef)
	}
}
