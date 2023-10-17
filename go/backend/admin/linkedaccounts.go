package admin

import (
	"context"
	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/env"
	pb "gitlab.com/fynbos/proto/backend/admin/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *AdminRpcService) ListLinkedAccounts(ctx context.Context, req *pb.ListLinkedAccountsRequest) (*pb.ListLinkedAccountsResponse, error) {
	lal, err := s.b.LinkedAccounts().ListByWalletId(ctx, req.WalletID)
	if err != nil {
		return nil, err
	}

	resp := make([]*pb.LinkedAccount, len(lal))
	for i, la := range lal {
		pbLa := transformLinkedAccount(la)
		resp[i] = pbLa
	}

	return &pb.ListLinkedAccountsResponse{Accounts: resp}, err
}

// create an rpc method for getting a linked account by id
func (s *AdminRpcService) GetLinkedAccount(ctx context.Context, req *pb.GetLinkedAccountRequest) (*pb.LinkedAccount, error) {
	la, err := s.b.LinkedAccounts().Get(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	ret := transformLinkedAccount(*la)

	return ret, nil
}

func (s *AdminRpcService) SeedLinkedAccounts(ctx context.Context, req *pb.SeedLinkedAccountsRequest) (*pb.Empty, error) {
	if env.IsProd() {
		return nil, NotFoundError("")
	}

	var las []linkedaccounts.LinkedAccount
	for _, la := range req.LinkedAccounts {
		las = append(las, linkedaccounts.LinkedAccount{
			ID:                  la.Id,
			WalletID:            la.WalletID,
			Name:                la.Name,
			Nickname:            la.Nickname,
			Mask:                la.Mask,
			Provider:            la.Provider,
			ProviderID:          la.ProviderID,
			Type:                la.Type,
			CanSend:             la.CanSend,
			CanReceive:          la.CanReceive,
			State:               linkedaccounts.State(la.State),
			SendCurrency:        currency.ParseCurrency(la.SendCurrencyCode),
			SendCountry:         country.ParseCountry(la.SendCurrencyCountryCode),
			SendAvailability:    la.SendAvailability,
			SendNetwork:         la.SendNetwork,
			ReceiveCountry:      country.ParseCountry(la.ReceiveCurrencyCountryCode),
			ReceiveCurrency:     currency.ParseCurrency(la.ReceiveCurrencyCode),
			ReceiveAvailability: la.ReceiveAvailability,
			ReceiveNetwork:      la.ReceiveNetwork,
		})
	}

	_, err := s.b.AdminLinkedAccounts().Seed(ctx, las)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}

func transformLinkedAccount(la linkedaccounts.LinkedAccount) *pb.LinkedAccount {
	ret := pb.LinkedAccount{
		Id:                         la.ID,
		WalletID:                   la.WalletID,
		Name:                       la.Name,
		Nickname:                   la.Nickname,
		Mask:                       la.Mask,
		Provider:                   la.Provider,
		ProviderID:                 la.ProviderID,
		Type:                       la.Type,
		State:                      string(la.State),
		CanSend:                    la.CanSend,
		CanReceive:                 la.CanReceive,
		SendCurrencyCode:           la.SendCurrency.String(),
		SendCurrencyCountryCode:    country.ParseCountry(la.SendCurrency.ISO4217()).String(),
		ReceiveCurrencyCode:        la.ReceiveCurrency.String(),
		ReceiveCurrencyCountryCode: country.ParseCountry(la.ReceiveCurrency.ISO4217()).String(),
		DefaultSend:                la.DefaultSend,
		DefaultReceive:             la.DefaultReceive,
		SendAvailability:           la.SendAvailability,
		SendNetwork:                la.SendNetwork,
		ReceiveAvailability:        la.ReceiveAvailability,
		ReceiveNetwork:             la.ReceiveNetwork,
	}
	if la.DeletedAt.Valid {
		ret.DeletedAt = timestamppb.New(la.DeletedAt.Time)
	}

	return &ret
}
