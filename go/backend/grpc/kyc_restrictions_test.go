package grpc

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/user"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	"gitlab.com/fynbos/backend/wallets"
	pb "gitlab.com/fynbos/proto/backend/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreatePayment_BlockedForDocumentsRequired(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	u := &user.User{ID: uuid.NewString()}
	w := wallets.Wallet{ID: uuid.NewString(), Name: "test-wallet"}

	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{w}, nil).AnyTimes()
	c.walletImpl.EXPECT().ForContext(gomock.Any()).Return(&w, nil).AnyTimes()
	c.KYCClient.EXPECT().GetKYCStatus(gomock.Any(), w.ID).Return(kyc.StatusDocumentsRequired, nil)

	_, err := client.CreatePayment(user_mock.ActingAsContext(t, context.Background(), u), &pb.CreatePaymentRequest{
		SenderAmount:         &pb.Amount{Amount: 100, Asset: "USD", AssetScale: 2},
		ReceiverAmount:       &pb.Amount{Amount: 100, Asset: "USD", AssetScale: 2},
		ReceiverIdentity:     "example.test/alice",
		ReceiverIdentityType: int32(3),
	})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.FailedPrecondition, st.Code())
}

func TestDepositBalance_BlockedForDocumentsRequired(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	u := &user.User{ID: uuid.NewString()}
	w := wallets.Wallet{ID: uuid.NewString(), Name: "test-wallet"}

	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{w}, nil).AnyTimes()
	c.walletImpl.EXPECT().ForContext(gomock.Any()).Return(&w, nil).AnyTimes()
	c.KYCClient.EXPECT().GetKYCStatus(gomock.Any(), w.ID).Return(kyc.StatusDocumentsRequired, nil)

	_, err := client.DepositBalance(user_mock.ActingAsContext(t, context.Background(), u), &pb.TransferBalanceRequest{})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.FailedPrecondition, st.Code())
}

func TestWithdrawBalance_BlockedForDocumentsRequired(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	u := &user.User{ID: uuid.NewString()}
	w := wallets.Wallet{ID: uuid.NewString(), Name: "test-wallet"}

	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{w}, nil).AnyTimes()
	c.walletImpl.EXPECT().ForContext(gomock.Any()).Return(&w, nil).AnyTimes()
	c.KYCClient.EXPECT().GetKYCStatus(gomock.Any(), w.ID).Return(kyc.StatusDocumentsRequired, nil)

	_, err := client.WithdrawBalance(user_mock.ActingAsContext(t, context.Background(), u), &pb.TransferBalanceRequest{})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.FailedPrecondition, st.Code())
}

func TestWithdrawXagoBalance_BlockedForDocumentsRequired(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	u := &user.User{ID: uuid.NewString()}
	w := wallets.Wallet{ID: uuid.NewString(), Name: "test-wallet"}

	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{w}, nil).AnyTimes()
	c.walletImpl.EXPECT().ForContext(gomock.Any()).Return(&w, nil).AnyTimes()
	c.KYCClient.EXPECT().GetKYCStatus(gomock.Any(), w.ID).Return(kyc.StatusDocumentsRequired, nil)

	_, err := client.WithdrawXagoBalance(user_mock.ActingAsContext(t, context.Background(), u), &pb.WithdrawXagoBalanceRequest{})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.FailedPrecondition, st.Code())
}

func TestConfirmPayment_BlockedForDocumentsRequired(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	u := &user.User{ID: uuid.NewString()}
	w := wallets.Wallet{ID: uuid.NewString(), Name: "test-wallet"}

	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{w}, nil).AnyTimes()
	c.walletImpl.EXPECT().ForContext(gomock.Any()).Return(&w, nil).AnyTimes()
	c.KYCClient.EXPECT().GetKYCStatus(gomock.Any(), w.ID).Return(kyc.StatusDocumentsRequired, nil)

	_, err := client.ConfirmPayment(user_mock.ActingAsContext(t, context.Background(), u), &pb.ConfirmPaymentRequest{Id: "payment-id"})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.FailedPrecondition, st.Code())
}
