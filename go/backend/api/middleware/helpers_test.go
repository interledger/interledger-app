package middleware

import (
	"context"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
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
func (s *stubUserClient) GetUserIDForWallet(_ context.Context, _ string) (string, error) {
	panic("unexpected")
}

type stubWalletClient struct {
	list    []wallets.Wallet
	listErr error
}

func (s *stubWalletClient) List(_ context.Context, _ string) ([]wallets.Wallet, error) {
	return s.list, s.listErr
}
func (s *stubWalletClient) Create(_ context.Context, _ wallets.CreateArgs) (*wallets.Wallet, error) {
	return nil, nil
}
func (s *stubWalletClient) ForContext(_ context.Context) (*wallets.Wallet, error) {
	panic("unexpected")
}
func (s *stubWalletClient) Get(_ context.Context, _ string) (*wallets.Wallet, error) {
	panic("unexpected")
}
func (s *stubWalletClient) ListAll(_ context.Context, _ db.Pagination) ([]wallets.Wallet, error) {
	panic("unexpected")
}
func (s *stubWalletClient) SetWalletName(_ context.Context, _, _ string) (*wallets.Wallet, error) {
	panic("unexpected")
}
func (s *stubWalletClient) GetFromAddress(_ context.Context, _ string) (*wallets.Wallet, error) {
	panic("unexpected")
}
func (s *stubWalletClient) AddAddress(_ context.Context, _, _ string) (*wallets.Wallet, error) {
	panic("unexpected")
}
func (s *stubWalletClient) SetCountry(_ context.Context, _ string, _ country.Country) (*wallets.Wallet, error) {
	panic("unexpected")
}
func (s *stubWalletClient) SetExceededLimits(_ context.Context, _ string, _ bool) (*wallets.Wallet, error) {
	panic("unexpected")
}
