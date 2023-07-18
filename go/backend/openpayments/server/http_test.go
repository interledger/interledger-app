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

	"gitlab.com/fynbos/backend/features"
	features_mock "gitlab.com/fynbos/backend/features/client/mock"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/authorisation"
	mock_auth "gitlab.com/fynbos/backend/authorisation/client/mock"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/identities"
	identities_mock "gitlab.com/fynbos/backend/identities/client/mock"
	"gitlab.com/fynbos/backend/keys"
	mock_keys "gitlab.com/fynbos/backend/keys/client/mock"
	"gitlab.com/fynbos/backend/limits"
	limits_mock "gitlab.com/fynbos/backend/limits/client/mock"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linked_account_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
	"gitlab.com/fynbos/backend/providers/gmt"
	transactions_mock "gitlab.com/fynbos/backend/transactions/client/mock"
	"gitlab.com/fynbos/backend/user"
	users_mock "gitlab.com/fynbos/backend/user/client/mock"
	"gitlab.com/fynbos/env"
	"go.temporal.io/sdk/mocks"
)

func TestGetHandler(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	db := db.MigrateTestDB(t, ctx)
	idc := identities_mock.NewMockClient(ctrl)
	userClient := users_mock.NewMock()
	b := NewTestBackends(t, func(tb *testBackends) {
		tb.db = db
		tb.ids = idc
		tb.us = userClient
	})

	cases := []struct {
		name       string
		getPath    string
		pointer    *openpayments.PaymentPointer
		statusCode int
		identities []identities.Identity
	}{
		{
			name:       "not_found",
			getPath:    "https://fynbos.local.me/not_real",
			pointer:    nil,
			statusCode: http.StatusNotFound,
			identities: []identities.Identity{},
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
			identities: []identities.Identity{},
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
				wallet, err := userClient.CreateNewWallet(ctx, user.CreateWalletArgs{
					UserID: userID,
					Name:   tc.pointer.Alias,
				})
				require.NoError(t, err)

				tc.pointer.WalletID = wallet.ID

				err = ops.CreatePaymentPointer(ctx, b, *tc.pointer)
				require.NoError(t, err)
			}

			idc.EXPECT().ListPublic(gomock.Any(), gomock.Any()).Return(tc.identities, nil).AnyTimes()

			rr := httptest.NewRecorder()
			handler := catchAllHandler(b)
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.statusCode, rr.Code)

			if tc.pointer == nil {
				return
			}

			var resp JsonResponse
			err = json.NewDecoder(rr.Body).Decode(&resp)
			require.NoError(t, err)

			assert.Equal(t, tc.pointer.Alias, resp.PublicName)
			assert.Equal(t, req.URL.String(), resp.Id)
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
	auth.EXPECT().VerifyRequestSig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
	b := NewTestBackends(t, func(tb *testBackends) {
		tb.db = db
		tb.auth = auth
		tb.tr = tc
	})
	userClient := users_mock.NewMock()

	cases := []struct {
		name          string
		args          IncomingPaymentArgs
		statusCode    int
		contentDigest string
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
			statusCode:    http.StatusCreated,
			contentDigest: "sha-256=:OK8OHUEemX/nX+oImQMmwO81nLoOuSU1xiuqDFg2HE4=:",
		},
		{
			name: "success no amount",
			args: IncomingPaymentArgs{
				ToPP:   "https://fynbos.me/moneyplease2",
				FromPP: "https://fynbos.me/sendmoney2",
				Type:   "incoming_payment",
			},
			statusCode:    http.StatusCreated,
			contentDigest: "sha-256=:hgrR3MPMrqqKeknX+liGoxhDGIMtJ8onQerDh6wmirA=:",
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
		recvWallet, err := userClient.CreateNewWallet(ctx, user.CreateWalletArgs{
			UserID: recvUserID,
			Name:   "test",
		})
		require.NoError(t, err)
		sendWallet, err := userClient.CreateNewWallet(ctx, user.CreateWalletArgs{
			UserID: sendUserID,
			Name:   "test",
		})
		require.NoError(t, err)

		body, err := json.Marshal(tc.args)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/incoming", env.OpenPaymentsURL()), bytes.NewReader(body))
		req.Header.Set("authorization", "GNAP "+testToken)
		req.Header.Set("Content-Digest", tc.contentDigest)
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
		req, err = http.NewRequest(http.MethodGet, tc.args.ToPP+"/incoming/{payment_id}", nil)
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

	la_mock := linked_account_mock.NewMockClient(ctrl)
	ft_mock := features_mock.NewMockClient(ctrl)
	ft_mock.EXPECT().Features(gomock.Any(), gomock.Any()).Return(&features.WalletFeatures{
		SendEnabled:    true,
		ReceiveEnabled: true,
	}, nil).AnyTimes()
	tmp_mock := &mocks.Client{}
	lmt_mock := limits_mock.NewMockClient(ctrl)
	lmt_mock.EXPECT().Exceeds(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	lmt_mock.EXPECT().ExceedsKYCLimits(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, limits.LimitType(""), nil).AnyTimes()
	b := NewTestBackends(t, func(tb *testBackends) {
		tb.db = db
		tb.auth = auth
		tb.tr = tc
		tb.la = la_mock
		tb.lmt = lmt_mock
		tb.temp = tmp_mock
		tb.fc = ft_mock
	})
	auth.EXPECT().VerifyRequestSig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

	userClient := users_mock.NewMock()

	cases := []struct {
		name          string
		args          OutgoingPaymentArgs
		statusCode    int
		contentDigest string
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
			contentDigest: "sha-256=:hv/f0acdygqxt31nP2TtUXVycjowcR0oer8Sv4xw7g4=:",
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
		sendWallet, err := userClient.CreateNewWallet(ctx, user.CreateWalletArgs{
			UserID: sendUserID,
			Name:   "test",
		})
		require.NoError(t, err)
		recvWallet, err := userClient.CreateNewWallet(ctx, user.CreateWalletArgs{
			UserID: recvUserID,
			Name:   "test",
		})
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
			ID:         uuid.NewString(),
			Provider:   gmt.ProviderName,
			Type:       gmt.TypeBankAccount,
			CanSend:    true,
			CanReceive: true,
			State:      linkedaccounts.Verified,
		}, {
			ID:         uuid.NewString(),
			Provider:   gmt.ProviderName,
			Type:       gmt.TypeBankAccount,
			CanSend:    true,
			CanReceive: true,
			State:      linkedaccounts.Verified,
		}, {
			ID:         uuid.NewString(),
			Provider:   gmt.ProviderName,
			Type:       gmt.TypeBankAccount,
			CanSend:    true,
			CanReceive: true,
			State:      linkedaccounts.Verified,
		},
		}, nil).AnyTimes()

		ipAddress := "198.0.0.8"
		tmp_mock.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, ipAddress, mock.Anything).Return(nil, nil)

		body, err := json.Marshal(tc.args)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/outgoing", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Authorization", "GNAP "+testToken)
		req.Header.Set("X-Forwarded-For", ipAddress)
		req.Header.Set("Content-Digest", tc.contentDigest)

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
	keyClient := mock_keys.NewMockClient(ctrl)
	keyClient.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).AnyTimes()
	b := NewTestBackends(t, func(tb *testBackends) {
		tb.db = db.MigrateTestDB(t, ctx)
		tb.keys = keyClient
	})
	userClient := users_mock.NewMock()

	userID := uuid.NewString()
	// Create Wallets
	wallet, err := userClient.CreateNewWallet(ctx, user.CreateWalletArgs{
		UserID: userID,
		Name:   "test",
	})
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

	keyClient.EXPECT().List(gomock.Any(), wallet.ID).Return([]keys.Key{
		{
			ID:        uuid.NewString(),
			Type:      keys.NonCustodial,
			Reference: "",
			PublicKey: "encoded key",
		},
	}, nil)

	rr := httptest.NewRecorder()
	handler := catchAllHandler(b)

	req := httptest.NewRequest("GET", "https://fynbos.local.me/found_me/jwks.json", nil)
	handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	respBytes, err := io.ReadAll(rr.Body)
	require.NoError(t, err)

	var keySet struct {
		Keys []openpayments.Jwk
	}
	err = json.Unmarshal(respBytes, &keySet)
	require.NoError(t, err)

	require.Len(t, keySet.Keys, 1)
	assert.Equal(t, "OKP", keySet.Keys[0].Kty)
	assert.Equal(t, "encoded key", keySet.Keys[0].X)
}

func TestGetIdentitiesHandler(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	db := db.MigrateTestDB(t, ctx)
	idc := identities_mock.NewMockClient(ctrl)
	userClient := users_mock.NewMock()
	b := NewTestBackends(t, func(tb *testBackends) {
		tb.db = db
		tb.ids = idc
		tb.us = userClient
	})

	cases := []struct {
		name        string
		getPath     string
		contentType string
		pointer     *openpayments.PaymentPointer
		statusCode  int
		identities  []identities.Identity
	}{
		{
			name:       "not_found_pp",
			getPath:    "https://fynbos.local.me/not_real/identities/MTIzNA==",
			pointer:    nil,
			statusCode: http.StatusNotFound,
			identities: []identities.Identity{},
		},
		{
			name:    "not_found_identity",
			getPath: "https://fynbos.local.me/found_me/identities/MTIzNA==",
			pointer: &openpayments.PaymentPointer{
				WalletID:   uuid.NewString(),
				Alias:      "Some Alias",
				Asset:      "USD",
				AssetScale: 2,
				URL:        "https://fynbos.local.me/found_me",
			},
			statusCode: http.StatusNotFound,
			identities: []identities.Identity{},
		},
		{
			name:    "identity_not_same_as_wallet",
			getPath: "https://fynbos.local.me/found_me1/identities/MTIzNA==",
			pointer: &openpayments.PaymentPointer{
				WalletID:   uuid.NewString(),
				Alias:      "Some Alias",
				Asset:      "USD",
				AssetScale: 2,
				URL:        "https://fynbos.local.me/found_me1",
			},
			statusCode: http.StatusNotFound,
			identities: []identities.Identity{
				{
					ID:            "1234",
					Platform:      identities.PlatformTwitter,
					State:         identities.StateVerified,
					SignatureHash: []byte("1234"),
				},
			},
		},
		{
			name:    "identity_not_public",
			getPath: "https://fynbos.local.me/found_me2/identities/MTIzNA==",
			pointer: &openpayments.PaymentPointer{
				WalletID:   "841a60bc-112d-4bd7-ac24-3e48d52b3fe8",
				Alias:      "Some Alias",
				Asset:      "USD",
				AssetScale: 2,
				URL:        "https://fynbos.local.me/found_me2",
			},
			statusCode: http.StatusNotFound,
			identities: []identities.Identity{
				{
					ID:            "dd192bbd-1b08-4e41-8c37-e909af79d10e",
					Platform:      identities.PlatformTwitter,
					State:         identities.StateVerified,
					Public:        false,
					SignatureHash: []byte("1234"),
				},
			},
		},
		{
			name:    "found",
			getPath: "https://fynbos.local.me/found_me3/identities/MTIzNA==",
			pointer: &openpayments.PaymentPointer{
				WalletID:   uuid.NewString(),
				Alias:      "Some Alias",
				Asset:      "USD",
				AssetScale: 2,
				URL:        "https://fynbos.local.me/found_me3",
			},
			statusCode: http.StatusOK,
			identities: []identities.Identity{
				{
					ID:            "dd192bbd-1b08-4e41-8c37-e909af79d10e",
					Platform:      identities.PlatformTwitter,
					State:         identities.StateVerified,
					Public:        true,
					SignatureHash: []byte("1234"),
				},
			},
		},
		{
			name:    "redirect_html",
			getPath: "https://fynbos.local.me/found_me4/identities/MTIzNA==",
			pointer: &openpayments.PaymentPointer{
				WalletID:   uuid.NewString(),
				Alias:      "Some Alias",
				Asset:      "USD",
				AssetScale: 2,
				URL:        "https://fynbos.local.me/found_me4",
			},
			contentType: "text/html; charset=utf-8",
			statusCode:  http.StatusSeeOther,
			identities: []identities.Identity{
				{
					ID:            "dd192bbd-1b08-4e41-8c37-e909af79d10e",
					Platform:      identities.PlatformTwitter,
					State:         identities.StateVerified,
					Public:        true,
					SignatureHash: []byte("1234"),
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			req, err := http.NewRequest(http.MethodGet, tc.getPath, nil)
			req.Header.Add("X-Forwarded-For", "8.8.8.8")
			if tc.contentType != "" {
				req.Header.Add("Accept", tc.contentType)
			}
			require.NoError(t, err)

			// Setup the payment pointer
			if tc.pointer != nil {
				userID := uuid.NewString()
				// Create Wallets
				wallet, err := userClient.CreateNewWallet(ctx, user.CreateWalletArgs{
					UserID: userID,
					Name:   tc.pointer.Alias,
				})
				require.NoError(t, err)

				tc.pointer.WalletID = wallet.ID

				// set wallet id on identity on empty ones
				for i := range tc.identities {
					if tc.identities[i].WalletID == "" {
						tc.identities[i].WalletID = wallet.ID
					}
				}

				err = ops.CreatePaymentPointer(ctx, b, *tc.pointer)
				require.NoError(t, err)
			}
			idc.EXPECT().GetBySignatureHash(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, sigHash []byte) (*identities.Identity, error) {
				// find the identity in the list
				for _, i := range tc.identities {
					if bytes.Equal(i.SignatureHash, sigHash) {
						return &i, nil
					}
				}
				return nil, identities.ErrNotFound
			}).AnyTimes()

			rr := httptest.NewRecorder()
			handler := OpenPaymentsHTTPHandler(b)
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.statusCode, rr.Code)

			if tc.statusCode != http.StatusOK {
				return
			}

			var resp JsonResponse
			err = json.NewDecoder(rr.Body).Decode(&resp)
			require.NoError(t, err)
		})
	}
}
