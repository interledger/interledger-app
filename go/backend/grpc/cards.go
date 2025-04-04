package grpc

import (
	"context"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/env"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func strPointerToEnum[T ~int32](v *string, m map[string]int32, d T) T {
	if v == nil {
		return d
	}
	if val, ok := m[*v]; ok {
		return T(val)
	}
	return d
}

func transformCard(c gatehub.Card) *pb.Card {
	lockLevel := strPointerToEnum(c.LockLevel, pb.CardLockLevel_value, pb.CardLockLevel_UnknownLockLevel)
	statusReasonCode := strPointerToEnum(c.StatusReasonCode, pb.CardStatusReasonCode_value, pb.CardStatusReasonCode_UnknownStatusReason)
	status := strPointerToEnum(&c.Status, pb.CardStatus_value, pb.CardStatus_UnknownStatus)
	return &pb.Card{
		Id:               c.ID,
		NameOnCard:       c.NameOnCard,
		MaskedPan:        c.MaskedPan,
		Status:           status,
		ExpiryDate:       c.ExpiryDate,
		StatusReasonCode: &statusReasonCode,
		LockLevel:        &lockLevel,
	}
}

func (s *rpcService) GetCustomerDeliveryAddresses(ctx context.Context, req *pb.Empty) (*pb.GetCustomerDeliveryAddressesResponse, error) {
	if !env.IsLocal() {
		return nil, ForbiddenError("Unauthorized.")
	}

	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	_, isEU := country.EUCountries[wallet.Country]
	if !isEU {
		return nil, FailedPreconditionError("Wallet not in the EU region")
	}

	// TODO(@radu): Enable this check once the feature flag is in place;
	// feats, err := s.b.Features().Features(ctx, wallet.ID)
	// if err != nil {
	// 	return nil, toGRPCError(err)
	// }
	// if !feats.InterledgerCardsEnabled {
	// 	return nil, ForbiddenError("TODO(@radu): message")
	// }

	isCustomer, err := s.b.Gatehub().IsCustomer(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if !isCustomer {
		user, err := s.b.Gatehub().GetUser(ctx, wallet.ID)
		if err != nil {
			return nil, toGRPCError(err)
		}

		// TODO(@radu): How should we handle temporary residence?
		return &pb.GetCustomerDeliveryAddressesResponse{
			Payload: &pb.GetCustomerDeliveryAddressesResponse_KycAddress{
				KycAddress: &pb.CustomerDeliveryAddress{
					Id:          uuid.NewString(), // using a random ID - it should not be used
					Type:        pb.CustomerDeliveryAddressType_PermanentResidence,
					CountryCode: user.Profile.AddressCountryCode,
					Line1:       user.Profile.AddressStreet1,
					Line2:       &user.Profile.AddressStreet2,
					Line3:       nil,
					PostOffice:  nil,
					City:        user.Profile.AddressCity,
					ZipCode:     user.Profile.AddressPostalCode,
				},
			},
		}, nil
	}

	// addrs, err := s.b.Gatehub().ListDeliveryAddresses(ctx, wallet.ID)
	// if err != nil {
	// 	return nil, toGRPCError(err)
	// }
	//
	// var res = []*pb.CustomerDeliveryAddress{}

	// TOOD(@radu): Fetch delivery addresses from GateHub
	return &pb.GetCustomerDeliveryAddressesResponse{}, nil
}

func (s *rpcService) GetCardApplicationProducts(ctx context.Context, req *pb.Empty) (*pb.GetCardApplicationProductsResponse, error) {
	return &pb.GetCardApplicationProductsResponse{}, nil
}

func (s *rpcService) ListCards(ctx context.Context, req *pb.Empty) (*pb.ListCardsResponse, error) {
	if !env.IsLocal() {
		return nil, ForbiddenError("Unauthorized.")
	}

	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	_, isEU := country.EUCountries[wallet.Country]
	if !isEU {
		return nil, FailedPreconditionError("Wallet not in the EU region")
	}

	// TODO(@radu): Enable this check once the feature flag is in place;
	// feats, err := s.b.Features().Features(ctx, wallet.ID)
	// if err != nil {
	// 	return nil, toGRPCError(err)
	// }
	// if !feats.InterledgerCardsEnabled {
	// 	return nil, ForbiddenError("TODO(@radu): message")
	// }

	cards, err := s.b.Gatehub().ListCards(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	var res = []*pb.Card{}
	for _, c := range cards {
		res = append(res, transformCard(c))
	}

	return &pb.ListCardsResponse{
		Cards: res,
	}, nil
}
