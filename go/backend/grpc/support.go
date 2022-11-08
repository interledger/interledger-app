package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/supporttickets"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) CreateSupportTicket(ctx context.Context, req *pb.CreateSupportTicketRequest) (*pb.Empty, error) {
	args := supporttickets.CreateTicketArgs{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       req.Email,
		Description: req.Description,
	}
	if err := s.b.Validator().StructCtx(ctx, args); err != nil {
		return nil, toGRPCError(err)
	}

	err := s.b.SupportTickets().CreateTicket(ctx, args)

	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}
