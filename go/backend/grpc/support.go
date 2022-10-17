package grpc

import (
	"context"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"

	"gitlab.com/fynbos/backend/supporttickets"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) CreateSupportTicket(ctx context.Context, req *pb.CreateSupportTicketRequest) (*pb.Empty, error) {
	err := s.b.SupportTickets().CreateTicket(ctx, supporttickets.CreateTicketArgs{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       req.Email,
		Description: req.Description,
	})

	if err != nil {
		log.Error("Error creating support ticket", zap.Error(err))
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}
