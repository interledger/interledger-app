package grpc

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func TestJoinWaitlist(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	c.WaitlistClient.EXPECT().Add(gomock.Any(), "join@interledger.test", "ZA", "Bob", "", false).Return(nil).Times(1)

	_, err := client.JoinWaitlist(context.Background(), &pb.JoinWaitlistRequest{
		Email:       "join@interledger.test",
		CountryCode: "ZA",
		FullName:    "Bob",
		BetaOptIn:   false,
	})
	assert.NoError(t, err)
}

func TestJoinWaitlistWithBetaOptIn(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	c.WaitlistClient.EXPECT().Add(gomock.Any(), "join@interledger.test", "ZA", "Bob", "", true).Return(nil).Times(1)

	_, err := client.JoinWaitlist(context.Background(), &pb.JoinWaitlistRequest{
		Email:       "join@interledger.test",
		CountryCode: "ZA",
		FullName:    "Bob",
		BetaOptIn:   true,
	})
	assert.NoError(t, err)
}

func TestCanSignup(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	signupId := uuid.NewString()

	c.WaitlistClient.EXPECT().CanSignup(gomock.Any(), signupId).Return(true, nil).Times(1)

	resp, err := client.CanSignup(context.Background(), &pb.CanSignupRequest{
		Id: signupId,
	})

	assert.NoError(t, err)
	assert.True(t, resp.CanSignup)
}

func TestSetSignupComplete(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	signupId := uuid.NewString()
	userId := uuid.NewString()

	c.WaitlistClient.EXPECT().SetSignupComplete(gomock.Any(), signupId, userId).Return(nil).Times(1)

	_, err := client.SetSignupComplete(context.Background(), &pb.SetSignupCompleteRequest{
		Id:     signupId,
		UserId: userId,
	})

	assert.NoError(t, err)
}
