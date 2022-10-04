package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/kyc"

	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) UpdateIndividualKYC(ctx context.Context, req *pb.UpdateIndividualKYCRequest) (*pb.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	update := kyc.IndividualDetails{
		WalletID:    wallet.ID,
		FirstName:   req.GetFirstName(),
		LastName:    req.GetLastName(),
		CountryCode: req.GetCountryCode(),
		Gender:      kyc.Gender(req.GetGender()),
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
			Line1:            req.Address.GetLine1(),
			Line2:            req.Address.GetLine2(),
			Building:         req.Address.GetBuilding(),
			Apartment:        req.Address.GetApartment(),
			City:             req.Address.GetCity(),
			State:            req.Address.GetState(),
			ZipCode:          req.Address.GetZipCode(),
			CountryCode:      req.Address.GetCountryCode(),
			FormattedAddress: req.Address.GetFormattedAddress(),
			PlaceID:          req.Address.GetPlaceID(),
		}
		// Validate struct doesn't validate sub structs individually, so we do it manually
		err = s.b.Validator().Struct(update.Address)
		if err != nil {
			return nil, toGRPCError(err)
		}
	}

	_, err = s.b.KYC().UpdateIndividualDetails(ctx, update)

	return &pb.Empty{}, err
}
