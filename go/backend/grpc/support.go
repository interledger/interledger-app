package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/supporttickets"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) CreateSupportTicket(ctx context.Context, req *pb.CreateSupportTicketRequest) (*pb.Empty, error) {
	err := s.ticketsClient.CreateTicket(ctx, supporttickets.CreateTicketArgs{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       req.Email,
		Description: req.Description,
	})

	if err != nil {
		return nil, grpcError(err)
	}

	return &pb.Empty{}, nil
}
