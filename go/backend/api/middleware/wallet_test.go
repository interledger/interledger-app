package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/interledger/interledger-app/go/backend/country"
	"github.com/interledger/interledger-app/go/backend/user"
	"github.com/interledger/interledger-app/go/backend/wallets"
	wc_mock "github.com/interledger/interledger-app/go/backend/wallets/client/mock"
	"github.com/stretchr/testify/assert"
)

func TestWalletMiddleware(t *testing.T) {
	t.Parallel()

	alice := &user.User{ID: "alice", Country: country.ZA}
	wallet := wallets.Wallet{ID: "wallet-1"}

	tests := []struct {
		name            string
		userInCtx       *user.User
		setupWC         func(*gomock.Controller) *wc_mock.MockClient
		wantStatus      int
		wantWalletInCtx bool
	}{
		{
			name:      "user with wallet gets wallet attached to context",
			userInCtx: alice,
			setupWC: func(ctrl *gomock.Controller) *wc_mock.MockClient {
				wc := wc_mock.NewMockClient(ctrl)
				wc.EXPECT().List(gomock.Any(), alice.ID).Return([]wallets.Wallet{wallet}, nil)
				return wc
			},
			wantStatus:      http.StatusOK,
			wantWalletInCtx: true,
		},
		{
			name: "no user in context passes through without wallet",
			setupWC: func(ctrl *gomock.Controller) *wc_mock.MockClient {
				return wc_mock.NewMockClient(ctrl)
			},
			wantStatus:      http.StatusOK,
			wantWalletInCtx: false,
		},
		{
			name:      "user with no wallets creates one and attaches it",
			userInCtx: alice,
			setupWC: func(ctrl *gomock.Controller) *wc_mock.MockClient {
				wc := wc_mock.NewMockClient(ctrl)
				wc.EXPECT().List(gomock.Any(), alice.ID).Return([]wallets.Wallet{}, nil)
				wc.EXPECT().Create(gomock.Any(), wallets.CreateArgs{UserID: alice.ID, Country: alice.Country}).Return(&wallet, nil)
				wc.EXPECT().List(gomock.Any(), alice.ID).Return([]wallets.Wallet{wallet}, nil)
				return wc
			},
			wantStatus:      http.StatusOK,
			wantWalletInCtx: true,
		},
		{
			name:      "wallet list error passes through without wallet",
			userInCtx: alice,
			setupWC: func(ctrl *gomock.Controller) *wc_mock.MockClient {
				wc := wc_mock.NewMockClient(ctrl)
				wc.EXPECT().List(gomock.Any(), alice.ID).Return(nil, errors.New("db is down"))
				return wc
			},
			wantStatus:      http.StatusOK,
			wantWalletInCtx: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			var gotWallet *wallets.Wallet

			r := chi.NewRouter()
			r.Use(MakeWalletMiddleware(&stubUserClient{}, tt.setupWC(ctrl)))
			r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
				gotWallet, _ = walletForContext(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.userInCtx != nil {
				ctx := context.WithValue(req.Context(), user.CtxKey, tt.userInCtx)
				req = req.WithContext(ctx)
			}
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			if tt.wantWalletInCtx {
				assert.NotNil(t, gotWallet)
			} else {
				assert.Nil(t, gotWallet)
			}
		})
	}
}

func walletForContext(ctx context.Context) (*wallets.Wallet, error) {
	w, ok := ctx.Value(wallets.CtxKey).(*wallets.Wallet)
	if !ok || w == nil {
		return nil, wallets.ErrNoWalletFound
	}
	return w, nil
}
