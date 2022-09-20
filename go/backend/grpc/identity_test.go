package grpc

import (
	"context"
	"testing"

	user_mock "gitlab.com/fynbos/backend/user/client/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/user"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func TestGetIdentity(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	// Receive an error when there is no log in information in the context
	_, err := client.GetIdentity(context.Background(), &pb.Empty{})
	require.Error(t, err)

	c.IdentityService.EXPECT().Get(gomock.Any(), "user_uuid").Return(&identity.Identity{
		ID:           "user_uuid",
		FirstName:    "FirstName",
		LastName:     "LastName",
		MobileNumber: "+276666666",
		Email:        "fake@fynbos.dev",
		Country:      "ZA",
	}, nil)

	id, err := client.GetIdentity(user_mock.ActingAsContext(t, context.Background(), &user.User{
		ID: "user_uuid",
	}), &pb.Empty{})
	require.NoError(t, err)
	require.NotNil(t, id)
	assert.Equal(t, "user_uuid", id.Id)
	assert.Equal(t, "FirstName", id.FirstName)
	assert.Equal(t, "LastName", id.LastName)
	assert.Equal(t, "ZA", id.CountryCode)
	assert.Equal(t, "+276666666", id.MobileNumber)
	assert.Equal(t, "fake@fynbos.dev", id.Email)
}
