package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/log"
	pb "gitlab.com/fynbos/proto/backend/v1"
	"go.uber.org/zap"
)

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
						Type:        pb.CustomerDeliveryAddressType_CUSTOMER_DELIVERY_ADDRESS_PERMANENT_RESIDENCE,
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

	addrs, err := s.b.Gatehub().ListDeliveryAddresses(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	var res = []*pb.CustomerDeliveryAddress{}
	for _, a := range addrs {
		addr := newDeliveryAddress(a)
		res = append(res, &addr)
	}

	return &pb.GetCustomerDeliveryAddressesResponse{DeliveryAddresses: res}, nil
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
		card := newCard(c)
		res = append(res, &card)
	}

	return &pb.ListCardsResponse{
		Cards: res,
	}, nil
}

func (s *rpcService) OrderCard(ctx context.Context, req *pb.OrderCardRequest) (*pb.Empty, error) {
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

	args := gatehub.OrderCardArgs{
		WalletID: wallet.ID,
	}
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

	err = s.b.Gatehub().OrderCard(ctx, args)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}

func newCard(c gatehub.Card) pb.Card {
	var status pb.CardStatus
	statusReasonCode := pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_UNKNOWN
	lockLevel := pb.CardLockLevel_CARD_LOCK_LEVEL_UNKNOWN

	switch c.Status {
	case "Active":
		status = pb.CardStatus_CARD_STATUS_ACTIVE
	case "Blocked":
		status = pb.CardStatus_CARD_STATUS_BLOCKED
	case "TemporaryBlocked":
		status = pb.CardStatus_CARD_STATUS_TEMPORARY_BLOCKED
	case "Replaced":
		status = pb.CardStatus_CARD_STATUS_REPLACED
	case "SoftDelete":
		status = pb.CardStatus_CARD_STATUS_SOFT_DELETE
	case "AccountBlocked":
		status = pb.CardStatus_CARD_STATUS_ACCOUNT_BLOCKED
	case "InCreation":
		status = pb.CardStatus_CARD_STATUS_IN_CREATION
	default:
		status = pb.CardStatus_CARD_STATUS_UNKNOWN
	}

	if c.StatusReasonCode != nil {
		switch *c.StatusReasonCode {
		case "ClientRequestLock":
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_CLIENT_REQUESTED_LOCK
		case "LostCard":
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_LOST_CARD
		case "StolenCard":
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_STOLEN_CARD
		case "IssuerRequestGeneral":
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_ISSUER_REQUEST_GENERAL
		case "IssuerRequestFraud":
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_ISSUER_REQUEST_FRAUD
		case "IssuerRequestLegal":
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_ISSUER_REQUEST_LEGAL
		case "IssuerRequestIncorrectOpening":
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_ISSUER_REQUEST_INCORRECT_OPENING
		case "CardDamagedOrNotWorking":
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_CARD_DAMAGED_OR_NOT_WORKING
		case "UserRequest":
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_USER_REQUEST
		case "IssuerRequestCustomerDeceased":
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_ISSUER_REQUEST_CUSTOMER_DECEASED
		case "ProductDoesNotRenew":
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_PRODUCT_DOES_NOT_RENEW
		case "ProductChange":
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_PRODUCT_CHANGE
		case "Renewed":
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_RENEWED
		default:
			log.Warn("received unknown card status reason code", zap.String("statusReasonCode", *c.StatusReasonCode))
		}
	}

	if c.LockLevel != nil {
		switch *c.LockLevel {
		case "Client":
			lockLevel = pb.CardLockLevel_CARD_LOCK_LEVEL_CLIENT
		case "Admin":
			lockLevel = pb.CardLockLevel_CARD_LOCK_LEVEL_ADMIN
		default:
			log.Warn("received unknown card lock level", zap.String("lockLevel", *c.LockLevel))
		}
	}

	return pb.Card{
		Id:               c.ID,
		NameOnCard:       c.NameOnCard,
		MaskedPan:        c.MaskedPan,
		Status:           status,
		ExpiryDate:       c.ExpiryDate,
		StatusReasonCode: &statusReasonCode,
		LockLevel:        &lockLevel,
	}
}

func newDeliveryAddress(da gatehub.CustomerDeliveryAddress) pb.CustomerDeliveryAddress {
	var daType pb.CustomerDeliveryAddressType

	switch da.Type {
	case "PermanentResidence":
		{
			daType = pb.CustomerDeliveryAddressType_CUSTOMER_DELIVERY_ADDRESS_PERMANENT_RESIDENCE
		}
	case "TemporaryResidence":
		{
			daType = pb.CustomerDeliveryAddressType_CUSTOMER_DELIVERY_ADDRESS_TEMPORARY_RESIDENCE
		}
	case "Work":
		{
			daType = pb.CustomerDeliveryAddressType_CUSTOMER_DELIVERY_ADDRESS_WORK
		}
	case "Other":
		{
			daType = pb.CustomerDeliveryAddressType_CUSTOMER_DELIVERY_ADDRESS_TYPE_OTHER
		}
	}

	return pb.CustomerDeliveryAddress{
		Id: da.ID,
		Details: &pb.CustomerDeliveryAddressBase{
			Type:        daType,
			CountryCode: da.CountryCode,
			Line1:       da.Line1,
			Line2:       da.Line2,
			Line3:       da.Line3,
			PostOffice:  da.PostOffice,
			City:        da.City,
			ZipCode:     da.ZipCode,
		},
	}
}
