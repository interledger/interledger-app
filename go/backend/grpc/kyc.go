package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/kyc"

	pb "gitlab.com/fynbos/proto/backend/v1"
)

type validateIndividualKYC struct {
	CountryCode        string `validate:"omitempty,iso3166_1_alpha2"`
	Gender             int32  `validate:"omitempty,gte=0,lt=4"`
	IpAddress          string `validate:"ip_addr"`
	State              string `validate:"omitempty,iso3166_2"`
	AddressCountryCode string `validate:"omitempty,iso3166_1_alpha2"`
	AddressState       string `validate:"omitempty,iso3166_2"`
}

func (s *rpcService) UpdateIndividualKYC(ctx context.Context, req *pb.UpdateIndividualKYCRequest) (*pb.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	args := validateIndividualKYC{
		CountryCode: req.GetCountryCode(),
		Gender:      req.GetGender(),
		IpAddress:   req.GetIpAddress(),
	}
	if req.Address != nil {
		args.AddressState = req.GetAddress().GetState()
		args.AddressCountryCode = req.GetAddress().GetCountryCode()
	}
	if err := s.b.Validator().StructCtx(ctx, args); err != nil {
		return nil, toGRPCError(err)
	}

	update := kyc.IndividualDetails{
		WalletID:    wallet.ID,
		FirstName:   req.GetFirstName(),
		LastName:    req.GetLastName(),
		CountryCode: req.GetCountryCode(),
		Gender:      kyc.Gender(req.GetGender()),
		IPAddress:   req.GetIpAddress(),
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
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}
