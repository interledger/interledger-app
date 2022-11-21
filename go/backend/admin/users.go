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

func (s *AdminRpcService) ListUsers(ctx context.Context, req *adminv1.PaginationRequest) (*adminv1.ListUsersResponse, error) {
	page := db.FromAdminPB(req)
	users, err := s.b.Users().ListAllUsers(ctx, page)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := make([]*adminv1.User, len(users))
	for i, u := range users {
		resp[i], err = convertUser(ctx, s.b, u)
		if err != nil {
			return nil, err
		}
	}

	return &adminv1.ListUsersResponse{
		Users: resp,
		Page:  page.ToAdminPB(len(resp)),
	}, nil
}

func convertUser(ctx context.Context, b Backends, input user.User) (*adminv1.User, error) {
	wallets, err := b.Users().ListWallets(ctx, input.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	walletIDs := make([]string, len(wallets))
	for y, w := range wallets {
		walletIDs[y] = w.ID
	}

	return &adminv1.User{
		Id:          input.ID,
		Email:       input.Email,
		PhoneNumber: input.PhoneNumber,
		Wallets:     walletIDs,
	}, nil
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
		usersPB[i], err = convertUser(ctx, s.b, u)
		if err != nil {
			return nil, err
		}
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

func (s *AdminRpcService) ListWalletTransactions(ctx context.Context, req *adminv1.ListWalletTransactionsRequest) (*adminv1.ListWalletTransactionsResponse, error) {
	page := db.FromAdminPB(req.Page)

	trxs, err := s.b.OpenPayments().ListTransactions(ctx, req.WalletID, db.Pagination{})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := make([]*adminv1.Transaction, len(trxs))
	for i, tx := range trxs {
		res[i] = &adminv1.Transaction{
			WalletID:    req.WalletID,
			Id:          tx.ID,
			Type:        string(tx.Type),
			Asset:       tx.Amount.Asset,
			Amount:      tx.Amount.FormatAmount(),
			Source:      tx.Source,
			Destination: tx.Destination,
			Timestamp:   timestamppb.New(tx.Timestamp),
		}
	}

	return &adminv1.ListWalletTransactionsResponse{Transactions: res, Page: page.ToAdminPB(len(res))}, nil
}
