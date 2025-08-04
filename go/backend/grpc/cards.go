package grpc

import (
	"context"

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
	lockLevel := strPointerToEnum(c.LockLevel, pb.CardLockLevel_value, pb.CardLockLevel_CARD_LOCK_LEVEL_UNSPECIFIED)
	statusReasonCode := strPointerToEnum(c.StatusReasonCode, pb.CardStatusReasonCode_value, pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_UNSPECIFIED)
	status := strPointerToEnum(&c.Status, pb.CardStatus_value, pb.CardStatus_CARD_STATUS_UNSPECIFIED)
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

	feats, err := s.b.Features().Features(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if !feats.ManageWalletCardsEnabled {
		return nil, ForbiddenError("Wallet cards not enabled")
	}

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
			DeliveryAddresses: []*pb.CustomerDeliveryAddress{
				{
					Id: "kyc-address",
					Details: &pb.CustomerDeliveryAddressBase{
						Type:        pb.CustomerDeliveryAddressType_PERMANENT_RESIDENCE,
						CountryCode: user.Profile.AddressCountryCode,
						Line1:       user.Profile.AddressStreet1,
						Line2:       &user.Profile.AddressStreet2,
						Line3:       nil,
						PostOffice:  nil,
						City:        user.Profile.AddressCity,
						ZipCode:     user.Profile.AddressPostalCode,
					},
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
	products, err := s.b.Gatehub().GetCardApplicationProducts(ctx)
	if err != nil {
		return nil, toGRPCError(err)
	}

	var res = []*pb.CardApplicationProduct{}
	for _, p := range products {
		res = append(res, &pb.CardApplicationProduct{Name: p.Name, Code: p.Code})
	}

	return &pb.GetCardApplicationProductsResponse{
		Products: res,
	}, nil
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

	feats, err := s.b.Features().Features(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if !feats.ManageWalletCardsEnabled {
		return nil, ForbiddenError("Wallet cards not enabled")
	}

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

func (s *rpcService) OrderCard(ctx context.Context, req *pb.OrderCardRequest) (*pb.Empty, error) {
	// _, err := s.b.Users().UserForContext(ctx)
	// if err != nil {
	// 	return nil, UnauthenticatedError("Unauthenticated.")
	// }
	//
	// wallet, err := s.b.Wallets().ForContext(ctx)
	// if err != nil {
	// 	return nil, UnauthenticatedError("Unauthenticated.")
	// }
	//
	// _, isEU := country.EUCountries[wallet.Country]
	// if !isEU {
	// 	return nil, FailedPreconditionError("Wallet not in the EU region")
	// }
	//
	// args := gatehub.OrderCardArgs{
	// 	WalletID: wallet.ID,
	// }
	//
	// if req.GetDeliveryAddressId() != "" && req.GetNewDeliveryAddress() != nil {
	// 	return nil, toGRPCError(errors.New("please only provide the delivery address or a new delivery address"))
	// }
	//
	// if req.GetNewDeliveryAddress() != nil {
	// 	args.NewDeliveryAddress = &gatehub.NewCustomerDeliveryAddressArgs{
	// 		Type:        req.NewDeliveryAddress.Type.String(),
	// 		CountryCode: req.NewDeliveryAddress.CountryCode,
	// 		Line1:       req.NewDeliveryAddress.Line1,
	// 		Line2:       req.NewDeliveryAddress.Line2,
	// 		Line3:       req.NewDeliveryAddress.Line3,
	// 		City:        req.NewDeliveryAddress.City,
	// 		PostOffice:  req.NewDeliveryAddress.PostOffice,
	// 		ZipCode:     req.NewDeliveryAddress.ZipCode,
	// 		Reason:      req.NewDeliveryAddress.Reason,
	// 	}
	// }
	//
	// err = s.b.Gatehub().OrderCard(ctx, args)
	// if err != nil {
	// 	return nil, toGRPCError(err)
	// }

	return &pb.Empty{}, nil
}
