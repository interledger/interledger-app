package grpc

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/gatehub/external"
	"gitlab.com/fynbos/log"
	pb "gitlab.com/fynbos/proto/backend/v1"
	"go.uber.org/zap"
)

func (s *rpcService) GetCardOrderOptions(ctx context.Context, req *pb.Empty) (*pb.GetCardOrderOptionsResponse, error) {
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

	externalIDs, err := s.b.Gatehub().GetExternalIDs(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if externalIDs.IsPendingExternalCreation() {
		return &pb.GetCardOrderOptionsResponse{
			IsWaitingForCreation: true,
			Products:             []*pb.CardApplicationProduct{},
			Addresses:            []*pb.CustomerDeliveryAddress{},
			Countries:            []*pb.Country{},
		}, nil
	}

	ret := &pb.GetCardOrderOptionsResponse{
		Countries: country.GetGRPCCountries(),
	}

	var (
		ghUser   *gatehub.User
		products []external.CardApplicationProduct
		addrs    []external.CustomerDeliveryAddress
	)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		wg       sync.WaitGroup
		once     sync.Once
		firstErr error
	)

	fail := func(e error) {
		if e == nil {
			return
		}
		once.Do(func() {
			firstErr = e
			cancel() // cancel other goroutines immediately
		})
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		p, e := s.b.Gatehub().GetCardApplicationProducts(ctx)
		if e != nil {
			fail(e)
			return
		}
		products = p
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if !externalIDs.IsCustomerCreated() {
			u, e := s.b.Gatehub().GetUser(ctx, wallet.ID)
			if e != nil {
				fail(e)
				return
			}
			ghUser = u
		} else {
			addrsList, e := s.b.Gatehub().ListDeliveryAddresses(ctx, wallet.ID)
			if e != nil {
				fail(e)
				return
			}
			addrs = addrsList
		}
	}()

	wg.Wait()

	if firstErr != nil {
		return nil, toGRPCError(err)
	}

	if !externalIDs.IsCustomerCreated() && ghUser != nil {
		ret.Addresses = []*pb.CustomerDeliveryAddress{
			{
				Id: gatehub.KycAddressID,
				Details: &pb.CustomerDeliveryAddressBase{
					Type:        pb.CustomerDeliveryAddressType_CUSTOMER_DELIVERY_ADDRESS_PERMANENT_RESIDENCE,
					CountryCode: ghUser.Profile.AddressCountryCode,
					Line1:       ghUser.Profile.AddressStreet1,
					Line2:       &ghUser.Profile.AddressStreet2,
					Line3:       nil,
					PostOffice:  nil,
					City:        ghUser.Profile.AddressCity,
					ZipCode:     ghUser.Profile.AddressPostalCode,
				},
			},
		}
	} else {
		for _, a := range addrs {
			addr := newDeliveryAddress(a)
			ret.Addresses = append(ret.Addresses, &addr)
		}
	}

	for _, p := range products {
		ret.Products = append(ret.Products, &pb.CardApplicationProduct{
			Name: p.Name,
			Code: p.Code,
		})
	}

	return ret, nil
}

func (s *rpcService) ListCards(ctx context.Context, req *pb.Empty) (*pb.ListCardsResponse, error) {
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

	externalIDs, err := s.b.Gatehub().GetExternalIDs(ctx, wallet.ID)
	if err != nil {
		return nil, FailedPreconditionError("Could not retrieve external IDs")
	}

	if externalIDs.IsPendingExternalCreation() {
		return &pb.ListCardsResponse{
			IsWaitingForCreation: true,
			Cards:                []*pb.Card{},
		}, nil
	}

	cards, err := s.b.Gatehub().ListCards(ctx, *externalIDs)
	if err != nil {
		return nil, toGRPCError(err)
	}

	var res = []*pb.Card{}

	// By default, the cards are ordered by the created time in ASC order.
	// Here we reverse, to have the oldest cards first in the list.
	// Also, we hide the cards that are blocked or that are being deleted.
	for _, c := range slices.Backward(cards) {
		if c.Status != gatehub.CardStatusBlocked && c.Status != gatehub.CardStatusSoftDelete {
			card := newCard(c)
			res = append(res, &card)
		}
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

	externalIDs, err := s.b.Gatehub().GetExternalIDs(ctx, wallet.ID)
	if err != nil {
		return nil, FailedPreconditionError("Could not retrieve external IDs")
	}

	if externalIDs.IsPendingExternalCreation() {
		return nil, FailedPreconditionError("Attempted to order a new card for a pending customer")
	}

	args := gatehub.OrderCardArgs{
		Wallet:          *wallet,
		ExternalIDs:     *externalIDs,
		CardProductCode: req.CardProductCode,
	}

	err = s.b.Gatehub().ValidateCardProductCode(ctx, req.CardProductCode)
	if err != nil {
		return nil, toGRPCError(err)
	}

	switch req.Type {
	case pb.CardType_CARD_TYPE_PHYSICAL:
		args.ShouldOrderPlastic = true
	case pb.CardType_CARD_TYPE_VIRTUAL:
		args.ShouldOrderPlastic = false
	default:
		return nil, toGRPCError(errors.New("received unknown card type"))
	}

	// We only need to process the delivery address if the user wants a physical card
	if args.ShouldOrderPlastic {
		addrID := req.GetDeliveryAddressId()
		newAddr := req.GetNewDeliveryAddress()

		if addrID == "" && newAddr == nil {
			return nil, toGRPCError(errors.New("no delivery address specified for physical card"))
		}

		if addrID != "" && newAddr != nil {
			return nil, toGRPCError(errors.New("received delivery address id and new delivery address. only one should be present"))
		}

		if !externalIDs.IsCustomerCreated() && addrID != "" && addrID != gatehub.KycAddressID {
			return nil, toGRPCError(errors.New("attempted to order card for non-customer with a delivery address id"))
		}

		if newAddr != nil {
			if newAddr.Details.Type != pb.CustomerDeliveryAddressType_CUSTOMER_DELIVERY_ADDRESS_OTHER && newAddr.Details.Type != pb.CustomerDeliveryAddressType_CUSTOMER_DELIVERY_ADDRESS_WORK {
				return nil, toGRPCError(errors.New("invalid delivery address type. expected other or work"))
			}

			countryCode := country.ToAlpha3(newAddr.Details.CountryCode)

			if countryCode == "" {
				return nil, toGRPCError(errors.New("could not find ISO3166 alpha-3 country code"))
			}

			args.NewDeliveryAddress = &external.CreateCustomerDeliveryAddressArgs{
				CountryCode: countryCode,
				Line1:       newAddr.Details.Line1,
				Line2:       newAddr.Details.Line2,
				Line3:       newAddr.Details.Line3,
				City:        newAddr.Details.City,
				PostOffice:  newAddr.Details.PostOffice,
				ZipCode:     newAddr.Details.ZipCode,
				Reason:      newAddr.Reason,
			}

			switch newAddr.Details.Type {
			case pb.CustomerDeliveryAddressType_CUSTOMER_DELIVERY_ADDRESS_WORK:
				args.NewDeliveryAddress.Type = gatehub.DeliveryAddressWork
			case pb.CustomerDeliveryAddressType_CUSTOMER_DELIVERY_ADDRESS_OTHER:
				args.NewDeliveryAddress.Type = gatehub.DeliveryAddressOther
			}

			err := s.b.Validator().StructCtx(ctx, args.NewDeliveryAddress)

			if err != nil {
				return nil, toGRPCError(err)
			}
		} else if addrID != gatehub.KycAddressID {
			args.DeliveryAddressID = &addrID
		}

	}

	err = s.b.Gatehub().OrderCard(ctx, args)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}

func (s *rpcService) GetCardToken(ctx context.Context, req *pb.GetCardTokenRequest) (*pb.GetCardTokenResponse, error) {
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

	externalIDs, err := s.b.Gatehub().GetExternalIDs(ctx, wallet.ID)
	if err != nil {
		return nil, FailedPreconditionError("Could not retrieve external IDs")
	}

	args := gatehub.GetCardTokenArgs{
		UserID: externalIDs.UserID,
		CardID: req.GetCardId(),
	}

	switch req.GetTokenType() {
	case pb.CardTokenType_CARD_TOKEN_TYPE_CARD_DATA:
		args.TokenType = gatehub.CardTokenTypeCardData
	case pb.CardTokenType_CARD_TOKEN_TYPE_PIN:
		args.TokenType = gatehub.CardTokenTypePin
	case pb.CardTokenType_CARD_TOKEN_TYPE_APPLE_PROVISIONING:
		args.TokenType = gatehub.CardTokenTypeAppleProvisioning
	case pb.CardTokenType_CARD_TOKEN_TYPE_GOOGLE_PROVISIONING:
		args.TokenType = gatehub.CardTokenTypeGoogleProvisioning
	case pb.CardTokenType_CARD_TOKEN_TYPE_DIGITALIZATION_PAI:
		args.TokenType = gatehub.CardTokenTypeDigitalizationPAI
	case pb.CardTokenType_CARD_TOKEN_TYPE_PIN_CHANGE:
		args.TokenType = gatehub.CardTokenTypePinChange
	}

	if args.TokenType == gatehub.CardTokenTypePinChange {
		// Public Key is not needed for PIN change
		args.PublicKey = nil
	} else {
		publicKey := req.GetPublicKey()
		if publicKey == "" {
			return nil, errors.New("publicKey is required for this token type")
		}
		args.PublicKey = &publicKey
	}

	res, err := s.b.Gatehub().GetCardToken(ctx, args)
	if err != nil {
		return nil, toGRPCError(err)
	}

	var links = []*pb.TokenLink{}
	for _, l := range res.Links {
		link := newLink(l)
		links = append(links, &link)
	}

	return &pb.GetCardTokenResponse{
		Token: res.Token,
		Links: links,
	}, nil
}

func (s *rpcService) FreezeCard(ctx context.Context, req *pb.FreezeCardRequest) (*pb.Empty, error) {
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

	externalIDs, err := s.b.Gatehub().GetExternalIDs(ctx, wallet.ID)
	if err != nil {
		return nil, FailedPreconditionError("Could not retrieve external IDs")
	}

	err = s.b.Gatehub().FreezeCard(ctx, gatehub.FreezeCardArgs{
		UserID:     externalIDs.UserID,
		CardID:     req.CardId,
		ReasonCode: gatehub.CardStatusReasonCodeClientRequestedLock,
		Note:       nil,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}

func (s *rpcService) UnfreezeCard(ctx context.Context, req *pb.UnfreezeCardRequest) (*pb.Empty, error) {
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

	externalIDs, err := s.b.Gatehub().GetExternalIDs(ctx, wallet.ID)
	if err != nil {
		return nil, FailedPreconditionError("Could not retrieve external IDs")
	}

	err = s.b.Gatehub().UnfreezeCard(ctx, gatehub.UnfreezeCardArgs{
		UserID: externalIDs.UserID,
		CardID: req.CardId,
		Note:   nil,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}

func (s *rpcService) BlockCard(ctx context.Context, req *pb.BlockCardRequest) (*pb.Empty, error) {
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

	externalIDs, err := s.b.Gatehub().GetExternalIDs(ctx, wallet.ID)
	if err != nil {
		return nil, FailedPreconditionError("Could not retrieve external IDs")
	}

	err = s.b.Gatehub().BlockCard(ctx, gatehub.BlockCardArgs{
		UserID:     externalIDs.UserID,
		CardID:     req.CardId,
		ReasonCode: gatehub.CardStatusReasonCodeUserRequest,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}

func (s *rpcService) GetPendingThreeDSConfirmations(ctx context.Context, req *pb.Empty) (*pb.GetPendingThreeDSConfirmationsResponse, error) {
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

	externalIDs, err := s.b.Gatehub().GetExternalIDs(ctx, wallet.ID)
	if err != nil {
		return nil, FailedPreconditionError("Could not retrieve external IDs")
	}

	confirmations, err := s.b.Gatehub().GetPendingThreeDSConfirmations(ctx, externalIDs.UserID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	var res pb.GetPendingThreeDSConfirmationsResponse

	for _, c := range confirmations {
		cPB := pb.PendingThreeDSConfirmation{
			TransactionId:    c.TransactionID,
			MerchantName:     c.MerchantName,
			PurchaseAmount:   c.PurchaseAmount,
			PurchaseCurrency: c.PurchaseCurrency,
			PurchaseDate:     c.PurchaseDate,
			Timeout:          c.Timeout,
		}

		res.Confirmations = append(res.Confirmations, &cPB)
	}

	return &res, nil
}

func (s *rpcService) ThreeDSPaymentConfirmation(ctx context.Context, req *pb.ThreeDSPaymentConfirmationRequest) (*pb.Empty, error) {
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

	externalIDs, err := s.b.Gatehub().GetExternalIDs(ctx, wallet.ID)
	if err != nil {
		return nil, FailedPreconditionError("Could not retrieve external IDs")
	}

	err = s.b.Gatehub().ThreeDSPaymentConfirmation(ctx, externalIDs.UserID, req.TransactionId, req.Confirmed)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}

func newCard(c gatehub.Card) pb.Card {
	var status pb.CardStatus
	var cardType pb.CardType
	statusReasonCode := pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_NONE
	lockLevel := pb.CardLockLevel_CARD_LOCK_LEVEL_NONE

	switch c.Status {
	case gatehub.CardStatusActive:
		status = pb.CardStatus_CARD_STATUS_ACTIVE
	case gatehub.CardStatusTemporaryBlocked:
		status = pb.CardStatus_CARD_STATUS_TEMPORARY_BLOCKED
	case gatehub.CardStatusReplaced:
		status = pb.CardStatus_CARD_STATUS_REPLACED
	case gatehub.CardStatusAccountBlocked:
		status = pb.CardStatus_CARD_STATUS_ACCOUNT_BLOCKED
	case gatehub.CardStatusInCreation:
		status = pb.CardStatus_CARD_STATUS_IN_CREATION
	default:
		status = pb.CardStatus_CARD_STATUS_UNKNOWN
	}

	if c.StatusReasonCode != nil {
		switch *c.StatusReasonCode {
		case gatehub.CardStatusReasonCodeClientRequestedLock:
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_CLIENT_REQUESTED_LOCK
		case gatehub.CardStatusReasonCodeLostCard:
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_LOST_CARD
		case gatehub.CardStatusReasonCodeStolenCard:
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_STOLEN_CARD
		case gatehub.CardStatusReasonCodeIssuerRequestGeneral:
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_ISSUER_REQUEST_GENERAL
		case gatehub.CardStatusReasonCodeIssuerRequestFraud:
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_ISSUER_REQUEST_FRAUD
		case gatehub.CardStatusReasonCodeIssuerRequestLegal:
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_ISSUER_REQUEST_LEGAL
		case gatehub.CardStatusReasonCodeIssuerRequestIncorrectOpening:
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_ISSUER_REQUEST_INCORRECT_OPENING
		case gatehub.CardStatusReasonCodeCardDamagedOrNotWorking:
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_CARD_DAMAGED_OR_NOT_WORKING
		case gatehub.CardStatusReasonCodeUserRequest:
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_USER_REQUEST
		case gatehub.CardStatusReasonCodeIssuerRequestCustomerDeceased:
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_ISSUER_REQUEST_CUSTOMER_DECEASED
		case gatehub.CardStatusReasonCodeProductDoesNotRenew:
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_PRODUCT_DOES_NOT_RENEW
		case gatehub.CardStatusReasonCodeProductChange:
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_PRODUCT_CHANGE
		case gatehub.CardStatusReasonCodeRenewed:
			statusReasonCode = pb.CardStatusReasonCode_CARD_STATUS_REASON_CODE_RENEWED
		default:
			log.Warn("received unknown card status reason code", zap.String("statusReasonCode", *c.StatusReasonCode))
		}
	}

	if c.LockLevel != nil {
		switch *c.LockLevel {
		case gatehub.CardLockLevelClient:
			lockLevel = pb.CardLockLevel_CARD_LOCK_LEVEL_CLIENT
		case gatehub.CardLockLevelAdmin:
			lockLevel = pb.CardLockLevel_CARD_LOCK_LEVEL_ADMIN
		default:
			log.Warn("received unknown card lock level", zap.String("lockLevel", *c.LockLevel))
		}
	}

	if c.PlasticCreated {
		cardType = pb.CardType_CARD_TYPE_PHYSICAL
	} else {
		cardType = pb.CardType_CARD_TYPE_VIRTUAL
	}

	return pb.Card{
		Id:               c.ID,
		NameOnCard:       strings.Replace(c.NameOnCard, gatehub.DollarSignPlaceholder, gatehub.DollarSign, 1),
		Type:             cardType,
		MaskedPan:        c.MaskedPan,
		Status:           status,
		ExpiryDate:       c.ExpiryDate,
		StatusReasonCode: statusReasonCode,
		LockLevel:        lockLevel,
		ProductCode:      c.ProductCode,
	}
}

func newDeliveryAddress(da gatehub.CustomerDeliveryAddress) pb.CustomerDeliveryAddress {
	var daType pb.CustomerDeliveryAddressType

	switch da.Type {
	case gatehub.DeliveryAddressPermanentResidence:
		daType = pb.CustomerDeliveryAddressType_CUSTOMER_DELIVERY_ADDRESS_PERMANENT_RESIDENCE
	case gatehub.DeliveryAddressTemporaryResidence:
		daType = pb.CustomerDeliveryAddressType_CUSTOMER_DELIVERY_ADDRESS_TEMPORARY_RESIDENCE
	case gatehub.DeliveryAddressWork:
		daType = pb.CustomerDeliveryAddressType_CUSTOMER_DELIVERY_ADDRESS_WORK
	case gatehub.DeliveryAddressOther:
		daType = pb.CustomerDeliveryAddressType_CUSTOMER_DELIVERY_ADDRESS_OTHER
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

func newLink(l external.TokenLink) pb.TokenLink {
	return pb.TokenLink{
		Href:   l.Href,
		Rel:    l.Rel,
		Method: l.Method,
	}
}
