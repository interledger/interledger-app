package grpc

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func TestJoinWaitlist(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	c.WaitlistClient.EXPECT().Add(gomock.Any(), "join@fynbos.dev", "ZA", "Bob").Return(nil).Times(1)

	_, err := client.JoinWaitlist(context.Background(), &pb.JoinWaitlistRequest{
		Email:       "join@fynbos.dev",
		CountryCode: "ZA",
		FullName:    "Bob",
	})
	assert.NoError(t, err)
}
