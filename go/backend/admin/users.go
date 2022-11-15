package admin

import (
	"context"
	"errors"

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
		wallets, err := s.b.Users().ListWallets(ctx, u.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		walletIDs := make([]string, len(wallets))
		for y, w := range wallets {
			walletIDs[y] = w.ID
		}

		resp[i] = &adminv1.User{
			Id:          u.ID,
			Email:       u.Email,
			PhoneNumber: u.PhoneNumber,
			Wallets:     walletIDs,
		}
	}

	return &adminv1.ListUsersResponse{
		Users: resp,
		Page:  page.ToAdminPB(len(resp)),
	}, nil
}

func (s *AdminRpcService) GetUserDetails(ctx context.Context, req *adminv1.GetUserDetailsRequest) (*adminv1.GetUserDetailsResponse, error) {

	u, err := s.b.Users().GetUser(ctx, req.UserID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	wallets, err := s.b.Users().ListWallets(ctx, u.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if len(wallets) == 0 {
		return nil, status.Error(codes.NotFound, "user has no wallets")
	}

	// TODO: Do we want this API to work with users or wallets?
	id, err := s.b.KYC().GetIndividualDetails(ctx, wallets[0].ID)
	if errors.Is(err, kyc.ErrNoKYCInfo) {
		return &adminv1.GetUserDetailsResponse{
			User: &adminv1.UserDetails{
				UserID:      u.ID,
				Email:       u.Email,
				PhoneNumber: u.PhoneNumber,
			},
		}, nil
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var address string
	if id.Address != nil {
		address = id.Address.FormattedAddress
	}

	return &adminv1.GetUserDetailsResponse{User: &adminv1.UserDetails{
		UserID:      u.ID,
		Email:       u.Email,
		PhoneNumber: u.PhoneNumber,
		FirstName:   id.FirstName,
		LastName:    id.LastName,
		CountryCode: id.CountryCode,
		Gender:      int32(id.Gender),
		DateOfBirth: timestamppb.New(id.DateOfBirth),
		Address:     address,
	}}, nil

}

func (s *AdminRpcService) ListUserTransactions(ctx context.Context, req *adminv1.ListUserTransactionsRequest) (*adminv1.ListUserTransactionsResponse, error) {
	page := db.FromAdminPB(req.Page)

	wallets, err := s.b.Users().ListWallets(ctx, req.UserID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if len(wallets) == 0 {
		return nil, status.Error(codes.NotFound, "user has no wallets")
	}

	trxs, err := s.b.OpenPayments().ListTransactions(ctx, wallets[0].ID, db.Pagination{})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := make([]*adminv1.Transaction, len(trxs))
	for i, tx := range trxs {
		res[i] = &adminv1.Transaction{
			WalletID:    wallets[0].ID,
			Id:          tx.ID,
			Type:        string(tx.Type),
			Asset:       tx.Amount.Asset,
			Amount:      tx.Amount.FormatAmount(),
			Source:      tx.Source,
			Destination: tx.Destination,
			Timestamp:   timestamppb.New(tx.Timestamp),
		}
	}

	return &adminv1.ListUserTransactionsResponse{Transactions: res, Page: page.ToAdminPB(len(res))}, nil
}
