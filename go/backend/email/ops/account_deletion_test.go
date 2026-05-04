package ops

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/email/sendgrid"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
)

type sendTemplateCall struct {
	subject string
	to      []sendgrid.Email
}

type fakeSendgridClient struct {
	errs  []error
	calls []sendTemplateCall
}

func (f *fakeSendgridClient) SendTemplate(_ context.Context, subject string, to []sendgrid.Email, _ string, _ map[string]interface{}, _ []mail.Attachment) error {
	f.calls = append(f.calls, sendTemplateCall{
		subject: subject,
		to:      append([]sendgrid.Email(nil), to...),
	})
	if len(f.errs) == 0 {
		return nil
	}
	err := f.errs[0]
	f.errs = f.errs[1:]
	return err
}

type testUsersClient struct {
	getUser func(ctx context.Context, userID string) (*user.User, error)
}

func (t *testUsersClient) UserForCookie(_ context.Context, _ string) (*user.User, error) {
	return nil, errors.New("not implemented")
}
func (t *testUsersClient) UserForToken(_ context.Context, _ string) (*user.User, error) {
	return nil, errors.New("not implemented")
}
func (t *testUsersClient) UserForContext(_ context.Context) (*user.User, error) {
	return nil, errors.New("not implemented")
}
func (t *testUsersClient) GetUser(ctx context.Context, userID string) (*user.User, error) {
	return t.getUser(ctx, userID)
}
func (t *testUsersClient) ListUsers(_ context.Context, _ string) ([]user.User, error) {
	return nil, errors.New("not implemented")
}
func (t *testUsersClient) CheckUserTotpEnabled(_ context.Context, _ string) (bool, error) {
	return false, errors.New("not implemented")
}
func (t *testUsersClient) Delete2FATotpEnrollment(_ context.Context, _ string) error {
	return errors.New("not implemented")
}
func (t *testUsersClient) GetTotpURL(_ context.Context, _ string) (string, error) {
	return "", errors.New("not implemented")
}
func (t *testUsersClient) ValidateTotpCode(_ context.Context, _, _ string, _ time.Time) error {
	return errors.New("not implemented")
}
func (t *testUsersClient) GetUserIDForWallet(_ context.Context, _ string) (string, error) {
	return "", errors.New("not implemented")
}

type testWalletsClient struct {
	list func(ctx context.Context, userID string) ([]wallets.Wallet, error)
}

func (t *testWalletsClient) Create(_ context.Context, _ wallets.CreateArgs) (*wallets.Wallet, error) {
	return nil, errors.New("not implemented")
}
func (t *testWalletsClient) ForContext(_ context.Context) (*wallets.Wallet, error) {
	return nil, errors.New("not implemented")
}
func (t *testWalletsClient) Get(_ context.Context, _ string) (*wallets.Wallet, error) {
	return nil, errors.New("not implemented")
}
func (t *testWalletsClient) List(ctx context.Context, userID string) ([]wallets.Wallet, error) {
	return t.list(ctx, userID)
}
func (t *testWalletsClient) ListAll(_ context.Context, _ db.Pagination) ([]wallets.Wallet, error) {
	return nil, errors.New("not implemented")
}
func (t *testWalletsClient) SetWalletName(_ context.Context, _, _ string) (*wallets.Wallet, error) {
	return nil, errors.New("not implemented")
}
func (t *testWalletsClient) GetFromAddress(_ context.Context, _ string) (*wallets.Wallet, error) {
	return nil, errors.New("not implemented")
}
func (t *testWalletsClient) AddAddress(_ context.Context, _, _ string) (*wallets.Wallet, error) {
	return nil, errors.New("not implemented")
}
func (t *testWalletsClient) SetCountry(_ context.Context, _ string, _ country.Country) (*wallets.Wallet, error) {
	return nil, errors.New("not implemented")
}
func (t *testWalletsClient) SetExceededLimits(_ context.Context, _ string, _ bool) (*wallets.Wallet, error) {
	return nil, errors.New("not implemented")
}

type testBackends struct {
	external     sendgrid.Client
	users        user.Client
	wallets      wallets.Client
	supportEmail string
}

func (t *testBackends) External() sendgrid.Client { return t.external }
func (t *testBackends) OneTemplateID() string     { return "template-id" }
func (t *testBackends) SupportEmail() string      { return t.supportEmail }
func (t *testBackends) Users() user.Client        { return t.users }
func (t *testBackends) KYC() kyc.Client           { return nil }
func (t *testBackends) Wallets() wallets.Client   { return t.wallets }

func TestSendAccountDeletionRequestedEmail_SupportSendFailureReturnsError(t *testing.T) {
	sg := &fakeSendgridClient{errs: []error{errors.New("sendgrid down")}}
	b := &testBackends{
		external:     sg,
		supportEmail: "ops@interledger.app",
		users: &testUsersClient{getUser: func(_ context.Context, userID string) (*user.User, error) {
			return &user.User{ID: userID, Email: "user@example.com"}, nil
		}},
		wallets: &testWalletsClient{list: func(_ context.Context, _ string) ([]wallets.Wallet, error) {
			return []wallets.Wallet{{ID: "wallet-1"}}, nil
		}},
	}

	err := SendAccountDeletionRequestedEmail(context.Background(), b, "user-1")
	require.Error(t, err)
	require.ErrorIs(t, err, email.ErrInternal)
	require.Len(t, sg.calls, 1)
	require.Equal(t, "ops@interledger.app", sg.calls[0].to[0].Address)
}

func TestSendAccountDeletionRequestedEmail_UserConfirmationFailureIsNonFatal(t *testing.T) {
	sg := &fakeSendgridClient{errs: []error{nil, errors.New("second send failed")}}
	b := &testBackends{
		external:     sg,
		supportEmail: "ops@interledger.app",
		users: &testUsersClient{getUser: func(_ context.Context, userID string) (*user.User, error) {
			return &user.User{ID: userID, Email: "user@example.com", FirstName: "Test", LastName: "User"}, nil
		}},
		wallets: &testWalletsClient{list: func(_ context.Context, _ string) ([]wallets.Wallet, error) {
			return []wallets.Wallet{{ID: "wallet-1"}, {ID: "wallet-2"}}, nil
		}},
	}

	err := SendAccountDeletionRequestedEmail(context.Background(), b, "user-1")
	require.NoError(t, err)
	require.Len(t, sg.calls, 2)
	require.Equal(t, "ops@interledger.app", sg.calls[0].to[0].Address)
	require.Equal(t, "user@example.com", sg.calls[1].to[0].Address)
}

func TestSendAccountDeletionRequestedEmail_EmptyUserEmailSkipsConfirmation(t *testing.T) {
	sg := &fakeSendgridClient{}
	b := &testBackends{
		external:     sg,
		supportEmail: "ops@interledger.app",
		users: &testUsersClient{getUser: func(_ context.Context, userID string) (*user.User, error) {
			return &user.User{ID: userID, Email: "   "}, nil
		}},
		wallets: &testWalletsClient{list: func(_ context.Context, _ string) ([]wallets.Wallet, error) {
			return []wallets.Wallet{{ID: "wallet-1"}}, nil
		}},
	}

	err := SendAccountDeletionRequestedEmail(context.Background(), b, "user-1")
	require.NoError(t, err)
	require.Len(t, sg.calls, 1)
	require.Equal(t, "ops@interledger.app", sg.calls[0].to[0].Address)
}
