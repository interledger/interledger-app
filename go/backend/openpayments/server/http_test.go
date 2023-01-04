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

	"gitlab.com/fynbos/backend/currency"

	transactions_mock "gitlab.com/fynbos/backend/transactions/client/mock"

	"github.com/stretchr/testify/mock"

	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/providers/machnet"

	"gitlab.com/fynbos/backend/linkedaccounts"

	"go.temporal.io/sdk/mocks"

	"github.com/golang/mock/gomock"
	linked_account_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"

	users_client "gitlab.com/fynbos/backend/user/client"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
)

func TestGetHandler(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)

	b := NewTestBackends(t, db, nil, nil, nil, nil)
	userClient := users_client.New(b, "fakeURL", "fakeAdminURL")

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
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)

	ctrl := gomock.NewController(t)
	tc := transactions_mock.NewMockClient(ctrl)
	txID := uuid.NewString()
	tc.EXPECT().CreateTransactionTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(txID, nil).AnyTimes()

	b := NewTestBackends(t, db, nil, nil, nil, tc)
	userClient := users_client.New(b, "fakeURL", "fakeAdminURL")

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
				SendAmount: currency.Amount{
					Value:    100,
					Currency: currency.USD,
					Scale:    2,
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
			Asset:      tc.args.SendAmount.Currency,
			AssetScale: tc.args.SendAmount.Scale,
		})
		require.NoError(t, err)
		err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
			URL:        tc.args.ReceivePaymentPointer,
			Alias:      "alias",
			WalletID:   recvWallet.ID,
			Asset:      tc.args.SendAmount.Currency,
			AssetScale: tc.args.SendAmount.Scale,
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

		assert.Equal(t, tc.args.SendPaymentPointer, q.PaymentPointer)
		assert.Equal(t, tc.args.SendAmount.Value, q.SendAmount.Value)
		assert.Equal(t, tc.args.SendAmount.Currency, q.SendAmount.Currency)
		assert.Equal(t, tc.args.SendAmount.Scale, q.SendAmount.Scale)

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

		assert.Equal(t, tc.args.SendPaymentPointer, lq.PaymentPointer)
		assert.Equal(t, tc.args.SendAmount.Value, lq.SendAmount.Value)
		assert.Equal(t, tc.args.SendAmount.Currency, lq.SendAmount.Currency)
		assert.Equal(t, tc.args.SendAmount.Scale, lq.SendAmount.Scale)
	}
}

func TestHTTPCreateIncomingPaymentGet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	ctrl := gomock.NewController(t)
	tc := transactions_mock.NewMockClient(ctrl)
	txID := uuid.NewString()
	tc.EXPECT().CreateTransactionTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(txID, nil).AnyTimes()

	b := NewTestBackends(t, db, nil, nil, nil, tc)
	userClient := users_client.New(b, "fakeURL", "fakeAdminURL")

	cases := []struct {
		name       string
		args       openpayments.CreateIncomingPaymentArgs
		statusCode int
	}{
		{
			name: "success",
			args: openpayments.CreateIncomingPaymentArgs{
				PaymentPointer:     "https://fynbos.me/moneyplease",
				FromPaymentPointer: "https://fynbos.me/sendmoney",
				IncomingAmount: &currency.Amount{
					Value:    100,
					Currency: currency.USD,
					Scale:    2,
				},
				ExternalRef: "External",
				ExpiresAt:   time.Now().Add(time.Hour),
			},
			statusCode: http.StatusCreated,
		},
		{
			name: "success no amount",
			args: openpayments.CreateIncomingPaymentArgs{
				PaymentPointer:     "https://fynbos.me/moneyplease2",
				FromPaymentPointer: "https://fynbos.me/sendmoney2",
				ExternalRef:        "External",
				ExpiresAt:          time.Now().Add(time.Hour),
			},
			statusCode: http.StatusCreated,
		},
	}

	for _, tc := range cases {
		sendUserID := uuid.NewString()
		recvUserID := uuid.NewString()
		// Create Signups
		_, err := db.ExecContext(ctx, "INSERT INTO signups (id, user_id) VALUES ($1, $2),($3, $4)", uuid.NewString(), recvUserID, uuid.NewString(), sendUserID)
		require.NoError(t, err)
		// Create Wallets
		recvWallet, err := userClient.CreateNewWallet(ctx, recvUserID, "test")
		require.NoError(t, err)
		sendWallet, err := userClient.CreateNewWallet(ctx, sendUserID, "test")
		require.NoError(t, err)

		body, err := json.Marshal(tc.args)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/incoming-payments", tc.args.PaymentPointer), bytes.NewReader(body))
		require.NoError(t, err)

		// Setup the payment pointers
		asset := currency.USD
		assetScale := 2
		if tc.args.IncomingAmount != nil {
			asset = tc.args.IncomingAmount.Currency
			assetScale = tc.args.IncomingAmount.Scale
		}
		err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
			URL:        tc.args.PaymentPointer,
			Alias:      "alias",
			WalletID:   recvWallet.ID,
			Asset:      asset,
			AssetScale: assetScale,
		})
		require.NoError(t, err)
		err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
			URL:        tc.args.FromPaymentPointer,
			Alias:      "alias",
			WalletID:   sendWallet.ID,
			Asset:      asset,
			AssetScale: assetScale,
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
		if tc.args.IncomingAmount != nil {
			assert.Equal(t, tc.args.IncomingAmount.Value, ip.IncomingAmount.Value)
			assert.Equal(t, tc.args.IncomingAmount.Currency, ip.IncomingAmount.Currency)
			assert.Equal(t, tc.args.IncomingAmount.Scale, ip.IncomingAmount.Scale)
		} else {
			assert.Nil(t, ip.IncomingAmount)
		}

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
		if tc.args.IncomingAmount != nil {
			assert.Equal(t, tc.args.IncomingAmount.Value, lip.IncomingAmount.Value)
			assert.Equal(t, tc.args.IncomingAmount.Currency, lip.IncomingAmount.Currency)
			assert.Equal(t, tc.args.IncomingAmount.Scale, lip.IncomingAmount.Scale)
		} else {
			assert.Nil(t, lip.IncomingAmount)
		}
		assert.Equal(t, tc.args.ExternalRef, lip.ExternalRef)
	}
}

func TestHTTPCreateOutgoingPaymentGet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	ctrl := gomock.NewController(t)
	tc := transactions_mock.NewMockClient(ctrl)
	txID := uuid.NewString()
	tc.EXPECT().CreateTransactionTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(txID, nil).AnyTimes()

	la_mock := linked_account_mock.NewMockClient(ctrl)
	tmp_mock := &mocks.Client{}

	b := NewTestBackends(t, db, la_mock, tmp_mock, nil, tc)
	userClient := users_client.New(b, "fakeURL", "fakeAdminURL")

	cases := []struct {
		name       string
		quoteArgs  openpayments.CreateQuoteArgs
		opArgs     openpayments.CreateOutgoingPaymentArgs
		statusCode int
	}{
		{
			name: "success",
			quoteArgs: openpayments.CreateQuoteArgs{
				SendPaymentPointer:    "https://fynbos.me/paysend",
				ReceivePaymentPointer: "https://fynbos.me/payrecv",
				ExpiresAt:             time.Now().Add(time.Hour),
				SendAmount: currency.Amount{
					Value:    100,
					Currency: currency.USD,
					Scale:    2,
				}},
			opArgs: openpayments.CreateOutgoingPaymentArgs{
				Description: "description",
				ExternalRef: "external reference",
			},
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

		body, err := json.Marshal(tc.quoteArgs)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/quotes", tc.quoteArgs.SendPaymentPointer), bytes.NewReader(body))
		require.NoError(t, err)

		// Setup the payment pointers
		err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
			URL:        tc.quoteArgs.SendPaymentPointer,
			Alias:      "alias",
			WalletID:   sendWallet.ID,
			Asset:      tc.quoteArgs.SendAmount.Currency,
			AssetScale: tc.quoteArgs.SendAmount.Scale,
		})
		require.NoError(t, err)
		err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
			URL:        tc.quoteArgs.ReceivePaymentPointer,
			Alias:      "alias",
			WalletID:   recvWallet.ID,
			Asset:      tc.quoteArgs.SendAmount.Currency,
			AssetScale: tc.quoteArgs.SendAmount.Scale,
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

		assert.Equal(t, tc.quoteArgs.SendPaymentPointer, q.PaymentPointer)
		assert.Equal(t, tc.quoteArgs.SendAmount.Value, q.SendAmount.Value)
		assert.Equal(t, tc.quoteArgs.SendAmount.Currency, q.SendAmount.Currency)
		assert.Equal(t, tc.quoteArgs.SendAmount.Scale, q.SendAmount.Scale)

		// Create outgoing payment

		la_mock.EXPECT().ListByWalletId(gomock.Any(), gomock.Any()).Return([]linkedaccounts.LinkedAccount{{
			ID:       uuid.NewString(),
			Provider: machnet.ProviderName,
			Type:     machnet.TypeSendCard,
		}, {
			ID:       uuid.NewString(),
			Provider: machnet.ProviderName,
			Type:     machnet.TypeReceiveBankAccount,
		}, {
			ID:       uuid.NewString(),
			Provider: machnet.ProviderName,
			Type:     machnet.TypeWallet,
		},
		}, nil).Times(2)

		ipAddress := "198.0.0.8"
		tmp_mock.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything, ipAddress).Return(nil, nil)

		tc.opArgs.QuoteID = q.ID

		body, err = json.Marshal(tc.opArgs)
		require.NoError(t, err)

		req, err = http.NewRequest(http.MethodPost, tc.quoteArgs.SendPaymentPointer+"/outgoing-payments", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("X-Forwarded-For", ipAddress)

		rr = httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusCreated, rr.Code)

		respBytes, err = io.ReadAll(rr.Body)
		require.NoError(t, err)

		var op openpayments.OutgoingPayment
		err = json.Unmarshal(respBytes, &op)
		require.NoError(t, err)

		assert.Equal(t, tc.opArgs.Description, op.Description)
	}
}
