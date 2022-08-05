package grpc

import (
	"context"
	"testing"

	"gitlab.com/fynbos/backend/supporttickets"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func TestCreateSupportTicket(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	c.TicketClient.EXPECT().CreateTicket(gomock.Any(), supporttickets.CreateTicketArgs{
		FirstName:   "John",
		LastName:    "Faker",
		Email:       "john@faker.com",
		Description: "Some Very insightful query",
	}).Return(nil).Times(1)

	_, err := client.CreateSupportTicket(context.Background(), &pb.CreateSupportTicketRequest{
		FirstName:   "John",
		LastName:    "Faker",
		Email:       "john@faker.com",
		Description: "Some Very insightful query",
	})
	assert.NoError(t, err)
}
