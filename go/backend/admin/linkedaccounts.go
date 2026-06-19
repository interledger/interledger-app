package admin

import (
	"context"
	"strconv"

	"github.com/interledger/interledger-app/go/backend/country"
	pb "github.com/interledger/interledger-app/go/proto/backend/admin/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *AdminRpcService) ListLinkedAccounts(ctx context.Context, req *pb.ListLinkedAccountsRequest) (*pb.ListLinkedAccountsResponse, error) {
	lal, err := s.b.LinkedAccounts().ListByWalletId(ctx, req.WalletID)
	if err != nil {
		return nil, err
	}

	resp := make([]*pb.LinkedAccount, len(lal))
	for i, la := range lal {
		resp[i] = &pb.LinkedAccount{
			Id:                         la.ID,
			WalletID:                   la.WalletID,
			Name:                       la.Name,
			Nickname:                   la.Nickname,
			Mask:                       la.Mask,
			Provider:                   la.Provider,
			ProviderID:                 la.ProviderID,
			Type:                       la.Type,
			State:                      string(la.State),
			CanSend:                    strconv.FormatBool(la.CanSend),
			CanReceive:                 strconv.FormatBool(la.CanReceive),
			SendCurrencyCode:           la.SendCurrency.String(),
			SendCurrencyCountryCode:    country.ParseCountry(la.SendCurrency.ISO4217()).String(),
			ReceiveCurrencyCode:        la.ReceiveCurrency.String(),
			ReceiveCurrencyCountryCode: country.ParseCountry(la.ReceiveCurrency.ISO4217()).String(),
			DefaultSend:                la.DefaultSend,
			DefaultReceive:             la.DefaultReceive,
		}
		if la.DeletedAt.Valid {
			resp[i].DeletedAt = timestamppb.New(la.DeletedAt.Time)
		}
	}

	return &pb.ListLinkedAccountsResponse{Accounts: resp}, err
}

// create an rpc method for getting a linked account by id
func (s *AdminRpcService) GetLinkedAccount(ctx context.Context, req *pb.GetLinkedAccountRequest) (*pb.LinkedAccount, error) {
	la, err := s.b.LinkedAccounts().Get(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	ret := &pb.LinkedAccount{
		Id:                         la.ID,
		WalletID:                   la.WalletID,
		Name:                       la.Name,
		Nickname:                   la.Nickname,
		Mask:                       la.Mask,
		Provider:                   la.Provider,
		ProviderID:                 la.ProviderID,
		Type:                       la.Type,
		State:                      string(la.State),
		CanSend:                    strconv.FormatBool(la.CanSend),
		CanReceive:                 strconv.FormatBool(la.CanReceive),
		SendCurrencyCode:           la.SendCurrency.String(),
		SendCurrencyCountryCode:    country.ParseCountry(la.SendCurrency.ISO4217()).String(),
		ReceiveCurrencyCode:        la.ReceiveCurrency.String(),
		ReceiveCurrencyCountryCode: country.ParseCountry(la.ReceiveCurrency.ISO4217()).String(),
		DefaultSend:                la.DefaultSend,
		DefaultReceive:             la.DefaultReceive,
	}
	if la.DeletedAt.Valid {
		ret.DeletedAt = timestamppb.New(la.DeletedAt.Time)
	}

	return ret, nil
}
