package grpc

import (
	"context"

	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"

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
		return nil, UnauthenticatedError("Unauthenticated.")
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
		WalletID:     wallet.ID,
		FirstName:    req.GetFirstName(),
		LastName:     req.GetLastName(),
		CountryCode:  req.GetCountryCode(),
		Gender:       kyc.Gender(req.GetGender()),
		IPAddress:    req.GetIpAddress(),
		Nationality:  req.GetNationality(),
		PlaceOfBirth: req.GetPlaceOfBirth(),
	}
	if req.DateOfBirth.IsValid() {
		update.DateOfBirth = req.DateOfBirth.AsTime()
	}
	err = s.b.Validator().Struct(update)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if req.Address != nil {
		update.Address = addressFromPB(req.Address)
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

func addressFromPB(address *pb.Address) *kyc.Address {
	if address == nil {
		return nil
	}

	return &kyc.Address{
		Line1:       address.GetLine1(),
		Line2:       address.GetLine2(),
		Building:    address.GetBuilding(),
		Apartment:   address.GetApartment(),
		City:        address.GetCity(),
		State:       address.GetState(),
		ZipCode:     address.GetZipCode(),
		CountryCode: address.GetCountryCode(),
		PlaceID:     address.GetPlaceID(),
	}
}

func (s *rpcService) IsUSPSAddress(ctx context.Context, req *pb.Address) (*pb.IsUSPSAddressResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	if env.IsLocal() {
		return &pb.IsUSPSAddressResponse{
			Valid: true,
		}, nil
	}
	address := addressFromPB(req)
	// Validate struct doesn't validate sub structs individually, so we do it manually
	err = s.b.Validator().Struct(address)
	if err != nil {
		return nil, toGRPCError(err)
	}

	valid, err := s.b.KYC().IsUSPSAddress(ctx, *address)

	return &pb.IsUSPSAddressResponse{
		Valid: valid,
	}, toGRPCError(err)
}

func (s *rpcService) GetIndividualKYC(ctx context.Context, req *pb.Empty) (*pb.IndividualKYCResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	details, err := s.b.KYC().GetIndividualDetails(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	resp := &pb.IndividualKYCResponse{
		FirstName:    details.FirstName,
		LastName:     details.LastName,
		CountryCode:  details.CountryCode,
		Gender:       int32(details.Gender),
		DateOfBirth:  timestamppb.New(details.DateOfBirth),
		Nationality:  details.Nationality,
		PlaceOfBirth: details.PlaceOfBirth,
	}

	if details.Address != nil {
		address := details.Address.String()
		resp.Address = &pb.Address{
			Line1:            &details.Address.Line1,
			Line2:            &details.Address.Line2,
			Building:         &details.Address.Building,
			Apartment:        &details.Address.Apartment,
			City:             &details.Address.City,
			State:            &details.Address.State,
			ZipCode:          &details.Address.ZipCode,
			CountryCode:      &details.Address.CountryCode,
			PlaceID:          &details.Address.PlaceID,
			FormattedAddress: &address,
		}
	}

	return resp, nil
}

func (s *rpcService) KYCStatus(ctx context.Context, req *pb.Empty) (*pb.KYCStatusResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	status, err := s.b.KYC().GetKYCStatus(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.KYCStatusResponse{
		KycStatus: status.ToInt32(),
	}, nil
}

func (s *rpcService) StartKYC(ctx context.Context, _ *pb.Empty) (*pb.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	err = s.b.KYC().StartKYC(ctx, wallet.ID)
	if err != nil {
		log.Error("error starting kyc", zap.Error(err))
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}

func (s *rpcService) GetPersonaInquiry(ctx context.Context, req *pb.KYCPersonaInquiryRequest) (*pb.KYCPersonaInquiryResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	inq, err := s.b.KYC().GetPersonaInquiry(ctx, wallet.ID, req.GetIdempotencyKey())
	if err != nil {
		return nil, toGRPCError(err)
	}

	resp := &pb.KYCPersonaInquiryResponse{
		Id: inq.ID,
	}
	if inq.SessionToken != "" {
		resp.SessionToken = &inq.SessionToken
	}

	return resp, nil
}
