package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
)

func TestWalletMiddleware(t *testing.T) {
	t.Parallel()

	alice := &user.User{ID: "alice", Country: country.ZA}
	wallet := wallets.Wallet{ID: "wallet-1"}

	tests := []struct {
		name            string
		userInCtx       *user.User
		walletList      []wallets.Wallet
		listErr         error
		wantStatus      int
		wantWalletInCtx bool
	}{
		{
			name:            "user with wallet gets wallet attached to context",
			userInCtx:       alice,
			walletList:      []wallets.Wallet{wallet},
			wantStatus:      http.StatusOK,
			wantWalletInCtx: true,
		},
		{
			name:            "no user in context passes through without wallet",
			wantStatus:      http.StatusOK,
			wantWalletInCtx: false,
		},
		{
			name:            "user with no wallets creates one and attaches it",
			userInCtx:       alice,
			walletList:      []wallets.Wallet{},
			wantStatus:      http.StatusOK,
			wantWalletInCtx: true,
		},
		{
			name:            "wallet list error passes through without wallet",
			userInCtx:       alice,
			listErr:         errors.New("db is down"),
			wantStatus:      http.StatusOK,
			wantWalletInCtx: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var gotWallet *wallets.Wallet

			wc := &stubWalletClient{listErr: tt.listErr}

			wc.list = tt.walletList
			if len(tt.walletList) == 0 && tt.listErr == nil {
				wc.list = []wallets.Wallet{wallet}
			}

			r := chi.NewRouter()
			r.Use(MakeWalletMiddleware(&stubUserClient{}, wc))
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
