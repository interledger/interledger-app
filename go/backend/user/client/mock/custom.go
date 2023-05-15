package mock

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"gitlab.com/fynbos/backend/db"

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

func (mc mockClient) GetUser(_ context.Context, userID string) (*user.User, error) {
	return &user.User{
		ID:          userID,
		Email:       "info@fynbos.com",
		PhoneNumber: "+27836321959",
	}, nil

}

func (mc mockClient) ListAllUsers(ctx context.Context, pagination db.Pagination) ([]user.User, error) {
	var res []user.User
	for _, w := range mc.wallets {
		res = append(res, user.User{
			ID:          mc.walletUser[w.ID],
			Email:       "info@fynbos.com",
			PhoneNumber: "+27836321959",
		})
	}

	return res, nil
}

func (mc mockClient) ListAllWallets(ctx context.Context, pagination db.Pagination) ([]user.Wallet, error) {
	var res []user.Wallet
	for _, w := range mc.wallets {
		res = append(res, *w)
	}

	return res, nil
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

func (mc mockClient) GetWallet(ctx context.Context, id string) (*user.Wallet, error) {
	wallet := mc.wallets[id]
	if wallet == nil {
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

func (mc mockClient) CreateNewWallet(_ context.Context, args user.CreateWalletArgs) (*user.Wallet, error) {
	wallet := &user.Wallet{
		ID:   uuid.NewString(),
		Name: args.Name,
	}

	mc.wallets[wallet.ID] = wallet
	mc.walletUser[wallet.ID] = args.UserID

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

func (mc mockClient) SetWalletName(_ context.Context, id, name string) error {
	_, ok := mc.wallets[id]
	if ok {
		mc.wallets[id].Name = name
		return nil
	}

	return user.ErrNoWalletFound
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
