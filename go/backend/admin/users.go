package admin

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/user"

	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/kyc"
	adminv1 "gitlab.com/fynbos/proto/backend/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *AdminRpcService) ListWallets(ctx context.Context, req *adminv1.PaginationRequest) (*adminv1.ListWalletsResponse, error) {
	page := db.FromAdminPB(req)
	wallets, err := s.b.Users().ListAllWallets(ctx, page)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := make([]*adminv1.Wallet, len(wallets))
	for i, w := range wallets {
		users, err := s.b.Users().ListUsers(ctx, w.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		usersPB := make([]*adminv1.User, len(users))
		for y, u := range users {
			usersPB[y] = convertUser(u)
		}

		resp[i] = &adminv1.Wallet{
			WalletID:   w.ID,
			WalletName: w.Name,
			Users:      usersPB,
		}
	}

	return &adminv1.ListWalletsResponse{
		Wallets:       resp,
		NextPageToken: "", // TODO Need to actually populate this
	}, nil
}

func convertUser(input user.User) *adminv1.User {
	return &adminv1.User{
		Id:          input.ID,
		Email:       input.Email,
		PhoneNumber: input.PhoneNumber,
	}
}

func (s *AdminRpcService) GetWalletDetails(ctx context.Context, req *adminv1.GetWalletDetailsRequest) (*adminv1.WalletDetails, error) {
	users, err := s.b.Users().ListUsers(ctx, req.WalletID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if len(users) == 0 {
		return nil, status.Error(codes.NotFound, "no user found for walletID")
	}

	usersPB := make([]*adminv1.User, len(users))
	for i, u := range users {
		usersPB[i] = convertUser(u)
	}

	wallet, err := s.b.Users().GetWallet(ctx, users[0].ID, req.WalletID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	id, err := s.b.KYC().GetIndividualDetails(ctx, wallet.ID)
	if errors.Is(err, kyc.ErrNoKYCInfo) {
		return &adminv1.WalletDetails{
			Users:    usersPB,
			WalletID: req.WalletID,
		}, nil
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var address string
	if id.Address != nil {
		address = id.Address.FormattedAddress
	}

	return &adminv1.WalletDetails{
		WalletID:    req.WalletID,
		FirstName:   id.FirstName,
		LastName:    id.LastName,
		CountryCode: id.CountryCode,
		Gender:      int32(id.Gender),
		DateOfBirth: timestamppb.New(id.DateOfBirth),
		Address:     address,
		Users:       usersPB,
	}, nil
}
