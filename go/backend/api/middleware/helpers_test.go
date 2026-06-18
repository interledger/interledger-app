package middleware

import (
	"context"
	"time"

	"github.com/interledger/interledger-app/go/backend/user"
)

type stubUserClient struct {
	tokenUser  *user.User
	tokenErr   error
	cookieUser *user.User
	cookieErr  error
}

func (s *stubUserClient) UserForToken(_ context.Context, _ string) (*user.User, error) {
	return s.tokenUser, s.tokenErr
}
func (s *stubUserClient) UserForCookie(_ context.Context, _ string) (*user.User, error) {
	return s.cookieUser, s.cookieErr
}
func (s *stubUserClient) UserForContext(ctx context.Context) (*user.User, error) {
	u, ok := ctx.Value(user.CtxKey).(*user.User)
	if !ok || u == nil {
		return nil, user.ErrNoUserFound
	}
	return u, nil
}
func (s *stubUserClient) GetUser(_ context.Context, _ string) (*user.User, error) {
	panic("unexpected")
}
func (s *stubUserClient) ListUsers(_ context.Context, _ string) ([]user.User, error) {
	panic("unexpected")
}
func (s *stubUserClient) CheckUserTotpEnabled(_ context.Context, _ string) (bool, error) {
	panic("unexpected")
}
func (s *stubUserClient) Delete2FATotpEnrollment(_ context.Context, _ string) error {
	panic("unexpected")
}
func (s *stubUserClient) GetTotpURL(_ context.Context, _ string) (string, error) { panic("unexpected") }
func (s *stubUserClient) ValidateTotpCode(_ context.Context, _, _ string, _ time.Time) error {
	panic("unexpected")
}
func (s *stubUserClient) GetUserIDForWallet(_ context.Context, _ string) (string, error) {
	panic("unexpected")
}
func (s *stubUserClient) SetPhoneVerified(_ context.Context, _ string) error {
	panic("unexpected")
}
func (s *stubUserClient) UpdateUserPhone(_ context.Context, _, _ string) error {
	panic("unexpected")
}
func (s *stubUserClient) FindWalletIDByEmail(_ context.Context, _ string) (string, error) {
	panic("unexpected")
}
