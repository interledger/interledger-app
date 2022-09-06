package grpc

import (
	"context"

	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) GetIdentity(ctx context.Context, empty *pb.Empty) (*pb.UserIdentity, error) {
	user, err := s.userService.ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	id, err := s.identityService.Get(ctx, user.ID)
	if err != nil {
		return nil, grpcError(err)
	}

	return &pb.UserIdentity{
		Id:           id.ID,
		FirstName:    id.FirstName,
		LastName:     id.LastName,
		MobileNumber: id.MobileNumber,
		Email:        id.Email,
		DateOfBirth:  id.DateOfBirth,
		CountryCode:  id.Country,
	}, nil
}
