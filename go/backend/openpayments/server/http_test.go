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

	"gitlab.com/fynbos/env"

	"github.com/go-chi/chi/v5"

	"gitlab.com/fynbos/backend/authorisation"
	mock_auth "gitlab.com/fynbos/backend/authorisation/client/mock"
	"gitlab.com/fynbos/backend/currency"

	transactions_mock "gitlab.com/fynbos/backend/transactions/client/mock"

	machnet_mock "gitlab.com/fynbos/backend/providers/machnet/client/mock"

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

	b := NewTestBackends(t, db, nil, nil, nil, nil, nil)
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
			req.Header.Add("X-Forwarded-For", "8.8.8.8")
			require.NoError(t, err)

			// Setup the payment pointer
			if tc.pointer != nil {
				userID := uuid.NewString()
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

func TestHTTPCreateIncomingPaymentGet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	ctrl := gomock.NewController(t)
	tc := transactions_mock.NewMockClient(ctrl)
	auth := mock_auth.NewMockInternalClient(ctrl)
	txID := uuid.NewString()
	tc.EXPECT().CreateTransactionTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(txID, nil).AnyTimes()

	b := NewTestBackends(t, db, nil, nil, nil, tc, auth)
	userClient := users_client.New(b, "fakeURL", "fakeAdminURL")

	cases := []struct {
		name       string
		args       IncomingPaymentArgs
		statusCode int
	}{
		{
			name: "success",
			args: IncomingPaymentArgs{
				FromPP: "https://fynbos.me/sendmoney",
				Type:   "incoming_payment",
				ToPP:   "https://fynbos.me/moneyplease",
				IncomingAmount: &struct {
					Amount   float64 `json:"amount,string"`
					Currency string  `json:"currency"`
				}{
					Amount:   10,
					Currency: "USD",
				},
			},
			statusCode: http.StatusCreated,
		},
		{
			name: "success no amount",
			args: IncomingPaymentArgs{
				ToPP:   "https://fynbos.me/moneyplease2",
				FromPP: "https://fynbos.me/sendmoney2",
				Type:   "incoming_payment",
			},
			statusCode: http.StatusCreated,
		},
	}

	for _, tc := range cases {
		testToken := uuid.NewString()
		auth.EXPECT().Introspect(gomock.Any(), testToken).Return(&authorisation.Grant{
			ID:     uuid.NewString(),
			Client: "http://fynbos.me/client",
			Tokens: []authorisation.AccessToken{
				{
					Value: testToken,
					Access: []authorisation.Access{
						{
							Type:       "incoming-payment",
							Actions:    []string{"read", "write"},
							Identifier: tc.args.ToPP,
						},
					},
				},
			},
		}, nil).AnyTimes()
		sendUserID := uuid.NewString()
		recvUserID := uuid.NewString()
		// Create Wallets
		recvWallet, err := userClient.CreateNewWallet(ctx, recvUserID, "test")
		require.NoError(t, err)
		sendWallet, err := userClient.CreateNewWallet(ctx, sendUserID, "test")
		require.NoError(t, err)

		body, err := json.Marshal(tc.args)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/incoming_payment", env.OpenPaymentsURL()), bytes.NewReader(body))
		req.Header.Set("authorization", "GNAP "+testToken)
		require.NoError(t, err)

		// Setup the payment pointers
		asset := currency.USD
		assetScale := 2
		err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
			URL:        tc.args.ToPP,
			Alias:      "alias",
			WalletID:   recvWallet.ID,
			Asset:      asset,
			AssetScale: assetScale,
		})
		require.NoError(t, err)
		err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
			URL:        tc.args.FromPP,
			Alias:      "alias",
			WalletID:   sendWallet.ID,
			Asset:      asset,
			AssetScale: assetScale,
		})
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := createIncomingPayment(b)
		handler.ServeHTTP(rr, req)

		require.Equal(t, tc.statusCode, rr.Code)

		respBytes, err := io.ReadAll(rr.Body)
		require.NoError(t, err)

		var ip openpayments.IncomingPayment
		err = json.Unmarshal(respBytes, &ip)
		require.NoError(t, err)

		assert.Equal(t, tc.args.ToPP, ip.PaymentPointer)
		assert.Equal(t, tc.args.FromPP, ip.FromPaymentPointer)
		if tc.args.IncomingAmount != nil {
			assert.Equal(t, tc.args.IncomingAmount.Amount, ip.IncomingAmount.Float64())
			assert.Equal(t, tc.args.IncomingAmount.Currency, ip.IncomingAmount.Currency.String())
		} else {
			assert.Nil(t, ip.IncomingAmount)
		}

		assert.Equal(t, tc.args.ExternalRef, ip.ExternalRef)

		// Do a get and get the same values
		req, err = http.NewRequest(http.MethodGet, tc.args.ToPP+"/incoming-payment/{payment_id}", nil)
		require.NoError(t, err)
		req.Header.Set("authorization", "GNAP "+testToken)
		rCtx := chi.NewRouteContext()
		rCtx.URLParams.Add("payment_id", ip.ID)
		req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rCtx))

		rr = httptest.NewRecorder()
		handler = getIncomingPayment(b)
		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		respBytes, err = io.ReadAll(rr.Body)
		require.NoError(t, err)

		var lip openpayments.IncomingPayment
		err = json.Unmarshal(respBytes, &lip)
		require.NoError(t, err)

		assert.Equal(t, tc.args.FromPP, lip.FromPaymentPointer)
		assert.Equal(t, tc.args.ToPP, lip.PaymentPointer)
		if tc.args.IncomingAmount != nil {
			assert.Equal(t, tc.args.IncomingAmount.Amount, lip.IncomingAmount.Float64())
			assert.Equal(t, tc.args.IncomingAmount.Currency, lip.IncomingAmount.Currency.String())
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
	auth := mock_auth.NewMockInternalClient(ctrl)
	tc := transactions_mock.NewMockClient(ctrl)
	txID := uuid.NewString()
	tc.EXPECT().CreateTransactionTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(txID, nil).AnyTimes()

	mc := machnet_mock.NewMockClient(ctrl)
	mc.EXPECT().GetUserByWalletID(gomock.Any(), gomock.Any()).Return(&machnet.User{KYCStatus: machnet.KYCStatusVerified}, nil).AnyTimes()
	mc.EXPECT().GetUserLimits(gomock.Any(), gomock.Any()).Return(&machnet.UserLimits{
		FundWallet: machnet.RemainingLimit{
			Annual:     currency.FromFloat64(1100, currency.USD),
			Daily:      currency.FromFloat64(1100, currency.USD),
			Monthly:    currency.FromFloat64(1100, currency.USD),
			WalletHold: currency.FromFloat64(1100, currency.USD),
		},
		Withdrawal: machnet.RemainingLimit{
			Annual:  currency.FromFloat64(1100, currency.USD),
			Daily:   currency.FromFloat64(1100, currency.USD),
			Monthly: currency.FromFloat64(1100, currency.USD),
		},
		Transfer: machnet.RemainingLimit{
			Annual:  currency.FromFloat64(1100, currency.USD),
			Daily:   currency.FromFloat64(1100, currency.USD),
			Monthly: currency.FromFloat64(1100, currency.USD),
		},
	}, nil)
	la_mock := linked_account_mock.NewMockClient(ctrl)
	tmp_mock := &mocks.Client{}

	b := NewTestBackends(t, db, la_mock, tmp_mock, mc, tc, auth)
	userClient := users_client.New(b, "fakeURL", "fakeAdminURL")

	cases := []struct {
		name       string
		args       OutgoingPaymentArgs
		statusCode int
	}{
		{
			name:       "success",
			statusCode: http.StatusCreated,
			args: OutgoingPaymentArgs{
				FromPP:      "https://fynbos.me/paysend",
				Type:        "outgoing_payment",
				ToPP:        "https://fynbos.me/payrecv",
				ExternalRef: "external_ref",
				SendAmount: struct {
					Amount   float64 `json:"amount,string"`
					Currency string  `json:"currency"`
				}{
					Amount:   10,
					Currency: "USD",
				},
			},
		},
	}

	for _, tc := range cases {
		testToken := uuid.NewString()
		auth.EXPECT().Introspect(gomock.Any(), testToken).Return(&authorisation.Grant{
			ID:     uuid.NewString(),
			Client: "http://fynbos.me/client",
			Tokens: []authorisation.AccessToken{
				{
					Value: testToken,
					Access: []authorisation.Access{
						{
							Type:       "outgoing-payment",
							Actions:    []string{"read", "write"},
							Identifier: tc.args.FromPP,
						},
					},
				},
			},
		}, nil).AnyTimes()
		sendUserID := uuid.NewString()
		recvUserID := uuid.NewString()
		// Create Wallets
		sendWallet, err := userClient.CreateNewWallet(ctx, sendUserID, "test")
		require.NoError(t, err)
		recvWallet, err := userClient.CreateNewWallet(ctx, recvUserID, "test")
		require.NoError(t, err)

		// Setup the payment pointers
		err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
			URL:        tc.args.FromPP,
			Alias:      "alias",
			WalletID:   sendWallet.ID,
			Asset:      currency.ParseCurrency(tc.args.SendAmount.Currency),
			AssetScale: 2,
		})
		require.NoError(t, err)
		err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
			URL:        tc.args.ToPP,
			Alias:      "alias",
			WalletID:   recvWallet.ID,
			Asset:      currency.ParseCurrency(tc.args.SendAmount.Currency),
			AssetScale: 2,
		})
		require.NoError(t, err)

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
		tmp_mock.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, ipAddress).Return(nil, nil)

		body, err := json.Marshal(tc.args)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/outgoing-payment", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Authorization", "GNAP "+testToken)
		req.Header.Set("X-Forwarded-For", ipAddress)

		rr := httptest.NewRecorder()
		handler := createOutgoingPayment(b)
		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusCreated, rr.Code)

		respBytes, err := io.ReadAll(rr.Body)
		require.NoError(t, err)

		var op openpayments.OutgoingPayment
		err = json.Unmarshal(respBytes, &op)
		require.NoError(t, err)

		assert.Equal(t, tc.args.ExternalRef, op.Description)
	}
}

func TestListKeys(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	db := db.MigrateTestDB(t, ctx)
	authClient := mock_auth.NewMockInternalClient(ctrl)
	b := NewTestBackends(t, db, nil, nil, nil, nil, authClient)
	userClient := users_client.New(b, "fakeURL", "fakeAdminURL")

	userID := uuid.NewString()
	// Create Wallets
	wallet, err := userClient.CreateNewWallet(ctx, userID, "test")
	require.NoError(t, err)
	asset := currency.USD
	assetScale := 2
	err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
		URL:        "https://fynbos.local.me/found_me",
		Alias:      "alias",
		WalletID:   wallet.ID,
		Asset:      asset,
		AssetScale: assetScale,
	})
	require.NoError(t, err)

	authClient.EXPECT().ListKeys(gomock.Any(), "https://fynbos.local.me/found_me").Return([]authorisation.Jwk{
		{
			Kty: "OKP",
			X:   "encoded key",
		},
	}, nil)

	rr := httptest.NewRecorder()
	handler := catchAllHandler(b)

	req := httptest.NewRequest("GET", "https://fynbos.local.me/found_me/.well-known/keys", nil)
	handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	respBytes, err := io.ReadAll(rr.Body)
	require.NoError(t, err)

	var keySet struct {
		Keys []authorisation.Jwk
	}
	err = json.Unmarshal(respBytes, &keySet)
	require.NoError(t, err)

	require.Len(t, keySet.Keys, 1)
	assert.Equal(t, "OKP", keySet.Keys[0].Kty)
	assert.Equal(t, "encoded key", keySet.Keys[0].X)
}
