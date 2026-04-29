package grpc

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/accountdeletion"
	"gitlab.com/fynbos/backend/user"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	"gitlab.com/fynbos/backend/wallets"
	pb "gitlab.com/fynbos/proto/backend/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRequestAccountDeletion_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	u := &user.User{ID: uuid.NewString()}
	w := wallets.Wallet{ID: uuid.NewString(), Name: "test-wallet"}

	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{w}, nil).AnyTimes()

	totpURL := fmt.Sprintf("otpauth://totp/test:%s?algorithm=SHA1&digits=6&issuer=test&period=30&secret=EGO3DEBFSF6Q3RKNRENIQ7XT7JO76MFA", u.ID)
	userMock := c.UserService.(*user_mock.MockClient)
	userMock.MapUserTotpURL(context.Background(), u.ID, totpURL)

	now := time.Now()
	key, err := otp.NewKeyFromURL(totpURL)
	require.NoError(t, err)
	code, err := totp.GenerateCode(key.Secret(), now)
	require.NoError(t, err)

	c.AccountDeletionClient.EXPECT().Request(gomock.Any(), u.ID).Return(nil)
	c.EmailClient.EXPECT().NotifyAccountDeletionRequested(gomock.Any(), u.ID).Return(nil)

	_, err = client.RequestAccountDeletion(user_mock.ActingAsContext(t, context.Background(), u), &pb.RequestAccountDeletionRequest{TotpCode: code})
	require.NoError(t, err)
}

func TestRequestAccountDeletion_AlreadyRequested(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	u := &user.User{ID: uuid.NewString()}
	w := wallets.Wallet{ID: uuid.NewString(), Name: "test-wallet"}

	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{w}, nil).AnyTimes()

	totpURL := fmt.Sprintf("otpauth://totp/test:%s?algorithm=SHA1&digits=6&issuer=test&period=30&secret=EGO3DEBFSF6Q3RKNRENIQ7XT7JO76MFA", u.ID)
	userMock := c.UserService.(*user_mock.MockClient)
	userMock.MapUserTotpURL(context.Background(), u.ID, totpURL)

	now := time.Now()
	key, err := otp.NewKeyFromURL(totpURL)
	require.NoError(t, err)
	code, err := totp.GenerateCode(key.Secret(), now)
	require.NoError(t, err)

	c.AccountDeletionClient.EXPECT().Request(gomock.Any(), u.ID).Return(accountdeletion.ErrAlreadyRequested)
	// No EXPECT for NotifyAccountDeletionRequested — gomock fails the test if it is called.

	_, err = client.RequestAccountDeletion(user_mock.ActingAsContext(t, context.Background(), u), &pb.RequestAccountDeletionRequest{TotpCode: code})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.AlreadyExists, st.Code())
}

// Regression: email failure after a successful insert must roll the row back.
func TestRequestAccountDeletion_EmailFailureRollsBack(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	u := &user.User{ID: uuid.NewString()}
	w := wallets.Wallet{ID: uuid.NewString(), Name: "test-wallet"}

	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{w}, nil).AnyTimes()

	totpURL := fmt.Sprintf("otpauth://totp/test:%s?algorithm=SHA1&digits=6&issuer=test&period=30&secret=EGO3DEBFSF6Q3RKNRENIQ7XT7JO76MFA", u.ID)
	userMock := c.UserService.(*user_mock.MockClient)
	userMock.MapUserTotpURL(context.Background(), u.ID, totpURL)

	now := time.Now()
	key, err := otp.NewKeyFromURL(totpURL)
	require.NoError(t, err)
	code, err := totp.GenerateCode(key.Secret(), now)
	require.NoError(t, err)

	c.AccountDeletionClient.EXPECT().Request(gomock.Any(), u.ID).Return(nil)
	c.EmailClient.EXPECT().NotifyAccountDeletionRequested(gomock.Any(), u.ID).Return(errors.New("sendgrid unavailable"))
	c.AccountDeletionClient.EXPECT().Delete(gomock.Any(), u.ID).Return(nil)

	_, err = client.RequestAccountDeletion(user_mock.ActingAsContext(t, context.Background(), u), &pb.RequestAccountDeletionRequest{TotpCode: code})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Internal, st.Code())
}

func TestRequestAccountDeletion_InvalidTotp(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	u := &user.User{ID: uuid.NewString()}
	w := wallets.Wallet{ID: uuid.NewString(), Name: "test-wallet"}

	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{w}, nil).AnyTimes()

	totpURL := fmt.Sprintf("otpauth://totp/test:%s?algorithm=SHA1&digits=6&issuer=test&period=30&secret=EGO3DEBFSF6Q3RKNRENIQ7XT7JO76MFA", u.ID)
	userMock := c.UserService.(*user_mock.MockClient)
	userMock.MapUserTotpURL(context.Background(), u.ID, totpURL)

	_, err := client.RequestAccountDeletion(user_mock.ActingAsContext(t, context.Background(), u), &pb.RequestAccountDeletionRequest{TotpCode: "000000"})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGetAccountDeletionStatus_NoPending(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	u := &user.User{ID: uuid.NewString()}
	w := wallets.Wallet{ID: uuid.NewString(), Name: "test-wallet"}

	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{w}, nil).AnyTimes()
	c.AccountDeletionClient.EXPECT().GetForUser(gomock.Any(), u.ID).Return(nil, nil)

	resp, err := client.GetAccountDeletionStatus(user_mock.ActingAsContext(t, context.Background(), u), &pb.Empty{})
	require.NoError(t, err)
	require.Nil(t, resp.RequestedAt)
}

func TestGetAccountDeletionStatus_Pending(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	u := &user.User{ID: uuid.NewString()}
	w := wallets.Wallet{ID: uuid.NewString(), Name: "test-wallet"}
	requestedAt := time.Now().Add(-time.Hour).UTC()

	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{w}, nil).AnyTimes()
	c.AccountDeletionClient.EXPECT().GetForUser(gomock.Any(), u.ID).Return(&accountdeletion.Request{
		ID:          uuid.NewString(),
		UserID:      u.ID,
		RequestedAt: requestedAt,
	}, nil)

	resp, err := client.GetAccountDeletionStatus(user_mock.ActingAsContext(t, context.Background(), u), &pb.Empty{})
	require.NoError(t, err)
	require.NotNil(t, resp.RequestedAt)
	require.WithinDuration(t, requestedAt, resp.RequestedAt.AsTime(), time.Second)
}
