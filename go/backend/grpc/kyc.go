package grpc

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/kyc/persona"
	"gitlab.com/fynbos/backend/providers/chimoney"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/pti"

	"gitlab.com/fynbos/env"
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

func (s *rpcService) GetKYCProviderWidget(ctx context.Context, req *pb.GetKYCProviderWidgetRequest) (*pb.KYCProviderWidget, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	if _, isEU := country.EUCountries[wallet.Country]; isEU {
		onboardingWidget, err := s.b.Gatehub().GetOnboardingWidget(ctx, wallet.ID)
		if err != nil {
			return nil, toGRPCError(err)
		}

		return &pb.KYCProviderWidget{
			Provider: gatehub.ProviderName,
			GatehubWidget: &pb.GatehubWidget{
				WidgetUrl: onboardingWidget,
			},
		}, nil
	}
	if country.CA == wallet.Country {
		widget, err := s.b.Chimoney().GetKYCWidget(ctx, wallet.ID)
		if err != nil {
			return nil, toGRPCError(err)
		}

		return &pb.KYCProviderWidget{
			Provider:       chimoney.ProviderName,
			ChimoneyWidget: widget,
		}, nil
	}
	if country.US == wallet.Country {
		widget, err := s.b.PTI().GetWidget(ctx, wallet.ID)
		if err != nil {
			return nil, toGRPCError(err)
		}

		return &pb.KYCProviderWidget{
			Provider: pti.ProviderName,
			PtiWidget: &pb.PtiWidget{
				ScenarioId:        widget.ScenarioID,
				UserId:            widget.UserID,
				RequestId:         widget.RequestID,
				ClientId:          widget.ClientID,
				GenerateTokenPath: widget.GenerateTokenPath,
				SdkUrl:            widget.SdkUrl,
				FormsUrl:          widget.FormsUrl,
			},
		}, nil
	}

	operation := func() (*kyc.PersonaInquiry, error) {
		return s.b.KYC().GetPersonaInquiry(ctx, wallet.ID, req.GetIdempotencyKey())
	}

	// loop until we get a response
	wrappedOperation := retryWithBackoff(operation)
	inq, err := wrappedOperation()
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.KYCProviderWidget{
		Provider: "persona",
		PersonaInquiry: &pb.KYCPersonaInquiryResponse{
			Id: inq.ID,
		},
	}, nil
}

func (s *rpcService) UpdateIndividualKYC(ctx context.Context, req *pb.UpdateIndividualKYCRequest) (*pb.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
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

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	var details *kyc.IndividualDetails
	if country.EUCountries[wallet.Country] {
		details, err = getGatehubKYCDetails(ctx, s.b, wallet.ID)
	} else {
		details, err = s.b.KYC().GetIndividualDetails(ctx, wallet.ID)
	}
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

func getGatehubKYCDetails(ctx context.Context, b Backends, walletID string) (*kyc.IndividualDetails, error) {
	gatehubDetails, err := b.Gatehub().GetUser(ctx, walletID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	location, err := time.LoadLocation("UTC")
	if err != nil {
		return nil, toGRPCError(err)
	}

	gender := kyc.GenderUnknown
	if "male" == gatehubDetails.Profile.Gender {
		gender = kyc.GenderMale
	} else if "female" == gatehubDetails.Profile.Gender {
		gender = kyc.GenderFemale
	}

	details := &kyc.IndividualDetails{
		WalletID:     walletID,
		FirstName:    gatehubDetails.Profile.FirstName,
		LastName:     gatehubDetails.Profile.LastName,
		CountryCode:  gatehubDetails.Profile.AddressCountryCode,
		PlaceOfBirth: gatehubDetails.Profile.BirthCity,
		Nationality:  gatehubDetails.Profile.Citizenship,
		Gender:       gender,
		DateOfBirth:  time.Date(gatehubDetails.Profile.BirthYear, time.Month(gatehubDetails.Profile.BirthMonth), gatehubDetails.Profile.BirthDay, 0, 0, 0, 0, location),
		Address: &kyc.Address{
			Line1:       gatehubDetails.Profile.AddressStreet1,
			Line2:       gatehubDetails.Profile.AddressStreet2,
			City:        gatehubDetails.Profile.AddressCity,
			State:       gatehubDetails.Profile.AddressSubdivision,
			ZipCode:     gatehubDetails.Profile.AddressPostalCode,
			CountryCode: gatehubDetails.Profile.AddressCountryCode,
		},
	}

	return details, nil
}

func (s *rpcService) KYCStatus(ctx context.Context, req *pb.Empty) (*pb.KYCStatusResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	status, err := s.b.KYC().GetKYCStatus(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if status == kyc.StatusLevel1 || status == kyc.StatusLevel2 {
		status = kyc.StatusApproved
	}

	return &pb.KYCStatusResponse{
		KycStatus: status.ToInt32(),
	}, nil
}

func (s *rpcService) GetPersonaInquiry(ctx context.Context, req *pb.KYCPersonaInquiryRequest) (*pb.KYCPersonaInquiryResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	operation := func() (*kyc.PersonaInquiry, error) {
		return s.b.KYC().GetPersonaInquiry(ctx, wallet.ID, req.GetIdempotencyKey())
	}

	// loop until we get a response
	wrappedOperation := retryWithBackoff(operation)
	inq, err := wrappedOperation()
	if err != nil {
		return nil, toGRPCError(err)
	}

	resp := &pb.KYCPersonaInquiryResponse{
		Id: inq.ID,
	}

	return resp, nil
}

func (s *rpcService) SetKYCStatusPending(ctx context.Context, req *pb.Empty) (*pb.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	kycStatus, err := s.b.KYC().GetKYCStatus(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if kycStatus != kyc.StatusUnknown {
		return &pb.Empty{}, nil
	}

	err = s.b.KYC().SetKYCStatus(ctx, wallet.ID, kyc.StatusPending)
	if err != nil {
		return nil, err
	}

	if country.EUCountries[wallet.Country] {
		err := s.b.Gatehub().LinkUserToGatewayByWalletID(ctx, wallet.ID)
		if err != nil {
			return nil, toGRPCError(err)
		}
	}

	return &pb.Empty{}, nil
}

var maxRetries = 3
var baseDelay = 1 * time.Millisecond

// RetryFunc is a function that can be retried
type retryFunc func() (*kyc.PersonaInquiry, error)

// RetryWithBackoff retries the given operation with exponential backoff
func retryWithBackoff(operation retryFunc) retryFunc {
	return func() (*kyc.PersonaInquiry, error) {
		var lastError error

		for i := 0; i < maxRetries; i++ {
			val, err := operation()
			if err == nil {
				return val, nil
			}
			// only retry for idempotency errors
			if !errors.Is(err, persona.ErrIdempotencyDuplicate) {
				return nil, err
			}
			secRetry := math.Pow(2, float64(i))
			fmt.Printf("Retrying operation in %f seconds\n", secRetry)
			delay := time.Duration(secRetry) * baseDelay
			time.Sleep(delay)
			lastError = err
		}

		return nil, lastError
	}
}
