package admin

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/healthcheck"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
	"google.golang.org/grpc"
)

func TestAccounts(s *testing.T) {
	ctx := context.Background()
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", "8443"))
	if err != nil {
		s.Fatal(err)
	}
	health, err := healthcheck.NewService()
	if err != nil {
		s.Fatal(err)
	}
	server := NewServer(health)
	go func() {
		if err := server.Serve(listener); err != nil {
			s.Fatal(err)
		}
	}()

	conn, err := grpc.Dial("127.0.0.1:8443", grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		s.Fatal(err)
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			s.Fatal(err)
		}
	}()
	client := backendv1.NewBackendServiceClient(conn)

	s.Run("can get user account by email", func(t *testing.T) {
		response, err := client.GetUserAccountByEmail(ctx, &backendv1.GetUserAccountByEmailRequest{
			Email: "test@fynbos.dev",
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "1234", response.GetId())
	})
}
