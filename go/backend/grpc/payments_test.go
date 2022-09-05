package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/user"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func TestInitiateOutgoingPayment(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	c.PaymentsClient.EXPECT().InitiateOutgoingPayment(gomock.Any(), payments.InitiateOutgoingPaymentArgs{
		UserID: "something_darkside",
		Amount: 100,
		To:     "some_payment_pointer",
		OTP:    "123",
	}).Return(&payments.OutgoingPayment{
		ID: "some_uuid",
	}, nil).Times(1)

	resp, err := client.InitiateOutgoingPayment(user.ActingAsContext(t, context.Background(), &user.User{
		ID: "something_darkside",
	}), &pb.InitiateOutgoingPaymentRequest{
		Amount: 100,
		To:     "some_payment_pointer",
		Otp:    "123",
	})
	require.NoError(t, err)
	assert.Equal(t, "some_uuid", resp.Id)
}
