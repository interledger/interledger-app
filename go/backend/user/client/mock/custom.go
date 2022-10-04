package mock

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/machinebox/graphql"
	"gitlab.com/fynbos/backend/user"
	"google.golang.org/grpc/metadata"
)

var _ user.Client = mockClient{}

type mockClient struct {
	wallets    map[string]*user.Wallet
	walletUser map[string]string
}

func (mc mockClient) ListUsers(ctx context.Context, walletID string) ([]user.User, error) {
	wallet := mc.wallets[walletID]
	if wallet == nil {
		return nil, user.ErrNoWalletFound
	}
	walletUser := mc.walletUser[wallet.ID]

	return []user.User{
		{
			ID:          walletUser,
			Email:       "info@fynbos.com",
			PhoneNumber: "+27836321959",
		},
	}, nil
}

func (mc mockClient) GetWallet(ctx context.Context, userID, id string) (*user.Wallet, error) {
	wallet := mc.wallets[id]
	if wallet == nil {
		return nil, user.ErrNoWalletFound
	}
	walletUser := mc.walletUser[wallet.ID]

	if walletUser != userID {
		return nil, user.ErrNoWalletFound
	}
	return wallet, nil
}

func (mc mockClient) UserForCookie(ctx context.Context, cookie string) (*user.User, error) {
	usr := user.User{}
	unescapedCookie, err := url.QueryUnescape(cookie)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(unescapedCookie), &usr)
	if err != nil {
		return nil, err
	}

	return &usr, nil
}

func (mc mockClient) UserForContext(ctx context.Context) (*user.User, error) {
	raw, ok := ctx.Value(userCtxKey).(*user.User)
	if !ok {
		return nil, user.ErrNoUserFound
	}
	return raw, nil
}

func (mc mockClient) WalletForContext(ctx context.Context) (*user.Wallet, error) {
	w, ok := ctx.Value(walletCtxKey).(*user.Wallet)
	if !ok || w == nil {
		return nil, user.ErrNoWalletFound
	}
	return w, nil
}

func (mc mockClient) CreateNewWallet(_ context.Context, userID, walletName string) (*user.Wallet, error) {
	wallet := &user.Wallet{
		ID:   uuid.NewString(),
		Name: walletName,
	}

	mc.wallets[wallet.ID] = wallet
	mc.walletUser[wallet.ID] = userID

	return wallet, nil
}

func (mc mockClient) ListWallets(_ context.Context, userID string) ([]user.Wallet, error) {
	var wallets []user.Wallet
	for key, element := range mc.walletUser {
		if element == userID {
			wallets = append(wallets, *mc.wallets[key])
		}
	}
	return wallets, nil
}

var _testCookieName = "ory_kratos_session"

const (
	userCtxKey   = user.UserCtxKey("user")
	walletCtxKey = user.WalletCtxKey("wallet")
)

func ActingAsContext(t *testing.T, ctx context.Context, usr *user.User) context.Context {
	if usr != nil {
		b, err := json.Marshal(usr)
		if err != nil {
			t.Fatal(err)
		}

		cookie := http.Cookie{
			Name:  _testCookieName,
			Value: url.QueryEscape(string(b)),
		}
		return metadata.AppendToOutgoingContext(ctx, "cookies", cookie.String())
	}

	return ctx
}

// Test helper function to set the cookie in the graphql request.
func ActingAs(req *graphql.Request, usr *user.User) error {
	if usr != nil {
		b, err := json.Marshal(usr)
		if err != nil {
			return err
		}

		cookie := http.Cookie{
			Name:  _testCookieName,
			Value: url.QueryEscape(string(b)),
		}
		req.Header.Add("Cookie", cookie.String())

		return nil
	}

	return nil
}

func NewMock() user.Client {
	return &mockClient{
		wallets:    make(map[string]*user.Wallet),
		walletUser: make(map[string]string),
	}
}
