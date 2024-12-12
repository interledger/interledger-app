package grpc

import (
	"context"
	"fmt"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/pti"

	pb "gitlab.com/fynbos/proto/backend/v1"
)

func transformLinkedAccount(la linkedaccounts.LinkedAccount) *pb.LinkedAccount {
	title := la.Nickname
	if la.Nickname == "" {
		title = la.Mask
	}

	// EUR is a special case where we want to show the Euro flag on the frontend
	sendCurrencyCountryCode := country.ParseCountry(la.SendCurrency.ISO4217()).String()
	if _, isEU := country.EUCountries[la.SendCountry]; isEU {
		sendCurrencyCountryCode = "EU"
	}

	recvCurrencyCountryCode := country.ParseCountry(la.ReceiveCurrency.ISO4217()).String()
	if _, isEU := country.EUCountries[la.ReceiveCountry]; isEU {
		recvCurrencyCountryCode = "EU"
	}

	return &pb.LinkedAccount{
		Id:                         la.ID,
		Type:                       la.Type,
		Name:                       la.Name,
		Mask:                       la.Mask,
		Nickname:                   la.Nickname,
		CanSend:                    la.CanSend,
		CanReceive:                 la.CanReceive,
		Title:                      title,
		SendCurrencyCode:           la.SendCurrency.String(),
		SendCurrencyCountryCode:    sendCurrencyCountryCode,
		ReceiveCurrencyCode:        la.ReceiveCurrency.String(),
		ReceiveCurrencyCountryCode: recvCurrencyCountryCode,
		DefaultSend:                la.DefaultSend,
		DefaultReceive:             la.DefaultReceive,
		State:                      string(la.State),
	}
}
func (s *rpcService) GetLinkedAccounts(
	ctx context.Context, _ *pb.Empty,
) (*pb.GetLinkedAccountsResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	linkedAccounts, err := s.b.LinkedAccounts().ListByWalletId(ctx, wallet.ID)
	if err != nil {
		return nil, InternalError("Unable to get linked accounts.")
	}

	var ret = []*pb.LinkedAccount{}
	for _, fs := range linkedAccounts {
		if fs.DeletedAt.Valid {
			continue
		}
		ret = append(ret, transformLinkedAccount(fs))
	}

	return &pb.GetLinkedAccountsResponse{
		LinkedAccounts: ret,
	}, nil
}

func (s *rpcService) GetLinkedAccount(ctx context.Context, req *pb.GetLinkedAccountRequest) (*pb.LinkedAccount, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	la, err := s.b.LinkedAccounts().Get(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if la.WalletID != wallet.ID || la.DeletedAt.Valid {
		return nil, toGRPCError(linkedaccounts.ErrNotFound)
	}

	return transformLinkedAccount(*la), nil

}

func (s *rpcService) DeleteLinkedAccount(ctx context.Context, req *pb.DeleteLinkedAccountRequest) (*pb.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	la, err := s.b.LinkedAccounts().Get(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if la.WalletID != wallet.ID {
		return nil, toGRPCError(linkedaccounts.ErrNotFound)
	}

	err = s.b.LinkedAccounts().Delete(ctx, la.ID)

	return &pb.Empty{}, toGRPCError(err)
}

func (s *rpcService) SetNicknameLinkedAccount(ctx context.Context, req *pb.SetNicknameLinkedAccountRequest) (*pb.LinkedAccount, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	// ACL check if user and wallet owns linked account
	la, err := s.b.LinkedAccounts().Get(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if la.WalletID != wallet.ID || la.DeletedAt.Valid {
		return nil, toGRPCError(linkedaccounts.ErrNotFound)
	}

	la, err = s.b.LinkedAccounts().SetNickname(ctx, req.Id, req.Nickname)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformLinkedAccount(*la), nil
}

func (s *rpcService) GetCardDetails(ctx context.Context, req *pb.GetCardDetailsRequest) (*pb.CardDetails, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	la, err := s.b.LinkedAccounts().Get(ctx, req.GetId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	if la.WalletID != wallet.ID || la.Type != pti.TypeCard || la.DeletedAt.Valid {
		return nil, NotFoundError("ErrNotFound")
	}

	card, err := s.b.PTI().GetLinkedAccountCardDetails(ctx, la.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.CardDetails{
		Id:         card.ID,
		Network:    card.CreditCardType,
		Bin:        card.CreditCardBin,
		Expiration: fmt.Sprintf("%s/%s", card.ExpirationMonth, card.ExpirationYear),
		Last4:      la.Mask,
		Nickname:   la.Nickname,
		State:      string(la.State),
		CanSend:    la.CanSend,
		CanReceive: la.CanReceive,
	}, nil
}

func (s *rpcService) SetDefaultSendLinkedAccount(ctx context.Context, req *pb.SetDefaultSendLinkedAccountRequest) (*pb.LinkedAccount, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	la, err := s.b.LinkedAccounts().Get(ctx, req.GetId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	if la.WalletID != wallet.ID {
		return nil, NotFoundError("ErrNotFound")
	}

	la, err = s.b.LinkedAccounts().SetDefaultSend(ctx, la.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformLinkedAccount(*la), nil
}

func (s *rpcService) SetDefaultReceiveLinkedAccount(ctx context.Context, req *pb.SetDefaultReceiveLinkedAccountRequest) (*pb.LinkedAccount, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	la, err := s.b.LinkedAccounts().Get(ctx, req.GetId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	if la.WalletID != wallet.ID {
		return nil, NotFoundError("ErrNotFound")
	}

	la, err = s.b.LinkedAccounts().SetDefaultReceive(ctx, la.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformLinkedAccount(*la), nil
}
