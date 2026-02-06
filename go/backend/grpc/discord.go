package grpc

import (
	"context"

	pb "gitlab.com/fynbos/proto/backend/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *rpcService) CreateDiscordAuthURL(ctx context.Context, _ *pb.Empty) (*pb.CreateDiscordAuthURLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "Discord integration has been removed")
}

func (s *rpcService) DiscordCallback(ctx context.Context, _ *pb.DiscordCallbackRequest) (*pb.DiscordCallbackResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "Discord integration has been removed")
}
