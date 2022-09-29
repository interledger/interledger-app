package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/kyc"

	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) UpdateUserKYC(ctx context.Context, req *pb.UpdateUserKYCRequest) (*pb.Empty, error) {
	u, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	update := kyc.UserDetails{
		UserID:      u.ID,
		FirstName:   *req.FirstName,
		LastName:    *req.LastName,
		CountryCode: *req.CountryCode,
		Gender:      kyc.Gender(*req.Gender),
	}
	if req.DateOfBirth.IsValid() {
		update.DateOfBirth = req.DateOfBirth.AsTime()
	}
	err = s.b.Validator().Struct(update)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if req.Address != nil {
		update.Address = &kyc.Address{
			Line1:       *req.Address.Line1,
			Line2:       *req.Address.Line2,
			Building:    *req.Address.Building,
			Apartment:   *req.Address.Apartment,
			City:        *req.Address.City,
			State:       *req.Address.State,
			ZipCode:     *req.Address.ZipCode,
			CountryCode: *req.Address.CountryCode,
		}
		// Validate struct doesn't validate sub structs individually, so we do it manually
		err = s.b.Validator().Struct(update.Address)
		if err != nil {
			return nil, toGRPCError(err)
		}
	}

	_, err = s.b.KYC().UpdateUserDetails(ctx, update)

	return &pb.Empty{}, err
}
