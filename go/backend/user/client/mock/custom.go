package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/interledger/interledger-app/go/backend/db"

	"github.com/interledger/interledger-app/go/backend/user"
	userops "github.com/interledger/interledger-app/go/backend/user/ops"
	"github.com/machinebox/graphql"
	"google.golang.org/grpc/metadata"
)

var _ user.Client = &MockClient{}

// DefaultTestTotpURL returns a canonical otpauth URL for the given userID.
// The secret is deterministic so callers that also need to generate codes
// against it (e.g. sca_handler_test) can do so.
func DefaultTestTotpURL(userID string) string {
	return fmt.Sprintf("otpauth://totp/test:%s?algorithm=SHA1&digits=6&issuer=test&period=30&secret=EGO3DEBFSF6Q3RKNRENIQ7XT7JO76MFA", userID)
}

type MockClient struct {
	WalletUser  map[string]string
	UserTotpURL map[string]string
}

func (mc *MockClient) MapUserWallet(_ context.Context, userID, walletID string) {
	mc.WalletUser[walletID] = userID
}

func (mc *MockClient) MapUserTotpURL(_ context.Context, userID, totpURL string) {
	mc.UserTotpURL[userID] = totpURL
}

func (mc *MockClient) EnrollTotp(userID string) {
	mc.UserTotpURL[userID] = DefaultTestTotpURL(userID)
}

func (mc *MockClient) GetUser(_ context.Context, userID string) (*user.User, error) {
	return &user.User{
		ID:          userID,
		Email:       "info@fynbos.com",
		PhoneNumber: "+27836321959",
	}, nil

}

func (mc *MockClient) ListAllUsers(ctx context.Context, pagination db.Pagination) ([]user.User, error) {
	var res []user.User
	for _, uid := range mc.WalletUser {
		res = append(res, user.User{
			ID:          uid,
			Email:       "info@fynbos.com",
			PhoneNumber: "+27836321959",
		})
	}

	return res, nil
}

func (mc *MockClient) ListUsers(ctx context.Context, walletID string) ([]user.User, error) {
	uid, ok := mc.WalletUser[walletID]
	if !ok {
		return nil, user.ErrNoUserFound
	}

	return []user.User{
		{
			ID:          uid,
			Email:       "info@fynbos.com",
			PhoneNumber: "+27836321959",
		},
	}, nil
}

func (mc MockClient) UserForCookie(ctx context.Context, cookie string) (*user.User, error) {
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

// UserForToken TODO: Just to make the mock client happy for now?
func (mc MockClient) UserForToken(ctx context.Context, token string) (*user.User, error) {
	usr := user.User{}
	return &usr, nil
}

func (mc MockClient) CheckUserTotpEnabled(_ context.Context, userID string) (bool, error) {
	_, ok := mc.UserTotpURL[userID]
	return ok, nil
}

func (mc MockClient) Delete2FATotpEnrollment(ctx context.Context, token string) error {
	return nil
}

func (mc MockClient) GetTotpURL(ctx context.Context, userID string) (string, error) {
	totpURL, ok := mc.UserTotpURL[userID]
	if !ok {
		return "", user.ErrTotpNotConfigured
	}
	return totpURL, nil
}

func (mc MockClient) ValidateTotpCode(ctx context.Context, userID, code string, now time.Time) error {
	totpURL, err := mc.GetTotpURL(ctx, userID)
	if err != nil {
		return err
	}
	return userops.ValidateTotpAgainstURL(totpURL, code, now)
}

func (mc MockClient) UserForContext(ctx context.Context) (*user.User, error) {
	raw, ok := ctx.Value(user.CtxKey).(*user.User)
	if !ok {
		return nil, user.ErrNoUserFound
	}
	return raw, nil
}

var _testCookieName = "ory_kratos_session"

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

func (mc *MockClient) GetUserIDForWallet(_ context.Context, walletID string) (string, error) {
	uid, ok := mc.WalletUser[walletID]
	if !ok {
		return "", user.ErrNoUserFound
	}
	return uid, nil
}

func (mc *MockClient) SetPhoneVerified(_ context.Context, _ string) error {
	return nil
}

func (mc *MockClient) UpdateUserPhone(_ context.Context, _ string, _ string) error {
	return nil
}

func (mc *MockClient) FindWalletIDByEmail(_ context.Context, _ string) (string, error) {
	return "", nil
}

func (mc *MockClient) Cleanup() {
	mc.WalletUser = map[string]string{}
	mc.UserTotpURL = map[string]string{}
}

func NewMock() *MockClient {
	return &MockClient{
		WalletUser:  make(map[string]string),
		UserTotpURL: make(map[string]string),
	}
}
