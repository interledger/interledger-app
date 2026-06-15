package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/accountdeletion"
	"gitlab.com/fynbos/backend/errcodes"
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
	c.UserService.EnrollTotp(u.ID)

	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{{ID: uuid.NewString(), Name: "test-wallet"}}, nil).AnyTimes()
	c.AccountDeletionClient.EXPECT().Request(gomock.Any(), u.ID).Return(nil)
	c.EmailClient.EXPECT().SendAccountDeletionRequested(gomock.Any(), u.ID).Return(nil)

	_, err := client.RequestAccountDeletion(user_mock.ActingAsContext(t, context.Background(), u), &pb.RequestAccountDeletionRequest{})
	require.NoError(t, err)
}

func TestRequestAccountDeletion_AlreadyRequested(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	u := &user.User{ID: uuid.NewString()}
	c.UserService.EnrollTotp(u.ID)

	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{{ID: uuid.NewString(), Name: "test-wallet"}}, nil).AnyTimes()
	c.AccountDeletionClient.EXPECT().Request(gomock.Any(), u.ID).Return(accountdeletion.ErrAlreadyRequested)

	_, err := client.RequestAccountDeletion(user_mock.ActingAsContext(t, context.Background(), u), &pb.RequestAccountDeletionRequest{})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.AlreadyExists, st.Code())
	appErr := statusFindDetail[*pb.AppError](st)
	require.NotNil(t, appErr)
	require.Equal(t, errcodes.ErrCodeAccountDeletionAlreadyRequested, appErr.ErrorCode)
}

// Regression: support notification failure after a successful insert must roll the row back.
func TestRequestAccountDeletion_SupportEmailFailureRollsBack(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	u := &user.User{ID: uuid.NewString()}
	c.UserService.EnrollTotp(u.ID)

	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{{ID: uuid.NewString(), Name: "test-wallet"}}, nil).AnyTimes()
	c.AccountDeletionClient.EXPECT().Request(gomock.Any(), u.ID).Return(nil)
	c.EmailClient.EXPECT().SendAccountDeletionRequested(gomock.Any(), u.ID).Return(errors.New("sendgrid unavailable"))
	c.AccountDeletionClient.EXPECT().Delete(gomock.Any(), u.ID).Return(nil)

	_, err := client.RequestAccountDeletion(user_mock.ActingAsContext(t, context.Background(), u), &pb.RequestAccountDeletionRequest{})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Internal, st.Code())
}

func TestRequestAccountDeletion_SlackLookupFailureDoesNotFailRequest(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	u := &user.User{ID: uuid.NewString()}
	c.UserService.EnrollTotp(u.ID)

	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return(nil, errors.New("wallets unavailable")).AnyTimes()
	c.AccountDeletionClient.EXPECT().Request(gomock.Any(), u.ID).Return(nil)
	c.EmailClient.EXPECT().SendAccountDeletionRequested(gomock.Any(), u.ID).Return(nil)

	_, err := client.RequestAccountDeletion(user_mock.ActingAsContext(t, context.Background(), u), &pb.RequestAccountDeletionRequest{})
	require.NoError(t, err)
}

func TestRequestAccountDeletion_TotpNotEnrolled(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	u := &user.User{ID: uuid.NewString()}

	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{{ID: uuid.NewString(), Name: "test-wallet"}}, nil).AnyTimes()

	_, err := client.RequestAccountDeletion(user_mock.ActingAsContext(t, context.Background(), u), &pb.RequestAccountDeletionRequest{})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.FailedPrecondition, st.Code())
	appErr := statusFindDetail[*pb.AppError](st)
	require.NotNil(t, appErr)
	require.Equal(t, errcodes.ErrCodeUserTotpNotConfigured, appErr.ErrorCode)
}

func TestGetAccountDeletionStatus_NoPending(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	u := &user.User{ID: uuid.NewString()}

	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{{ID: uuid.NewString(), Name: "test-wallet"}}, nil).AnyTimes()
	c.AccountDeletionClient.EXPECT().GetForUser(gomock.Any(), u.ID).Return(nil, nil)

	resp, err := client.GetAccountDeletionStatus(user_mock.ActingAsContext(t, context.Background(), u), &pb.Empty{})
	require.NoError(t, err)
	require.Equal(t, pb.AccountDeletionRequestStatus_ACCOUNT_DELETION_REQUEST_STATUS_UNSPECIFIED, resp.Status)
	require.Nil(t, resp.CreatedAt)
}

func TestGetAccountDeletionStatus_Pending(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	u := &user.User{ID: uuid.NewString()}
	createdAt := time.Now().Add(-time.Hour).UTC()

	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{{ID: uuid.NewString(), Name: "test-wallet"}}, nil).AnyTimes()
	c.AccountDeletionClient.EXPECT().GetForUser(gomock.Any(), u.ID).Return(&accountdeletion.Request{
		ID:        uuid.NewString(),
		UserID:    u.ID,
		Status:    accountdeletion.StatusPending,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}, nil)

	resp, err := client.GetAccountDeletionStatus(user_mock.ActingAsContext(t, context.Background(), u), &pb.Empty{})
	require.NoError(t, err)
	require.Equal(t, pb.AccountDeletionRequestStatus_ACCOUNT_DELETION_REQUEST_STATUS_PENDING, resp.Status)
	require.NotNil(t, resp.CreatedAt)
	require.WithinDuration(t, createdAt, resp.CreatedAt.AsTime(), time.Second)
	require.NotNil(t, resp.UpdatedAt)
}

func TestGetAccountDeletionStatus_InProgressAndCompleted(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   accountdeletion.Status
		want pb.AccountDeletionRequestStatus
	}{
		{"in_progress", accountdeletion.StatusInProgress, pb.AccountDeletionRequestStatus_ACCOUNT_DELETION_REQUEST_STATUS_IN_PROGRESS},
		{"completed", accountdeletion.StatusCompleted, pb.AccountDeletionRequestStatus_ACCOUNT_DELETION_REQUEST_STATUS_COMPLETED},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			c := NewTestContainer(t, ctrl)
			_, _, client := startTestServer(t, c)

			u := &user.User{ID: uuid.NewString()}
			now := time.Now().UTC()

			c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{{ID: uuid.NewString(), Name: "test-wallet"}}, nil).AnyTimes()
			c.AccountDeletionClient.EXPECT().GetForUser(gomock.Any(), u.ID).Return(&accountdeletion.Request{
				ID:        uuid.NewString(),
				UserID:    u.ID,
				Status:    tc.in,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil)

			resp, err := client.GetAccountDeletionStatus(user_mock.ActingAsContext(t, context.Background(), u), &pb.Empty{})
			require.NoError(t, err)
			require.Equal(t, tc.want, resp.Status)
		})
	}
}
