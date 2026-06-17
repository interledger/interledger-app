package admin

import (
	"context"
	"errors"
	"net/mail"
	"sort"
	"time"

	"github.com/interledger/interledger-app/go/backend/country"
	"github.com/interledger/interledger-app/go/backend/user"
	"github.com/interledger/interledger-app/go/log"
	"go.uber.org/zap"

	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/interledger/interledger-app/go/backend/kyc"
	adminv1 "github.com/interledger/interledger-app/go/proto/backend/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *AdminRpcService) ListWallets(ctx context.Context, req *adminv1.PaginationRequest) (*adminv1.ListWalletsResponse, error) {
	page := db.FromAdminPB(req)

	// Email addresses are stored in Kratos, not in the wallets table. When the
	// search term parses as a valid RFC 5322 address, resolve it to a wallet ID
	// first so that the normal name/ID filter path can handle it.
	if _, err := mail.ParseAddress(page.Search); err == nil {
		walletID, err := s.b.Users().FindWalletIDByEmail(ctx, page.Search)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if walletID == "" {
			return &adminv1.ListWalletsResponse{}, nil
		}
		page.Search = walletID
	}

	wallets, err := s.b.Wallets().ListAll(ctx, page)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	hasNextPage := len(wallets) > page.PageSize
	if hasNextPage {
		wallets = wallets[:page.PageSize]
	}

	resp := make([]*adminv1.Wallet, len(wallets))
	for i, w := range wallets {
		users, err := s.b.Users().ListUsers(ctx, w.ID)
		if err != nil {
			log.Error("Failed to fetch Kratos users", zap.Error(err), zap.String("walletID", w.ID))
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

	nextPageToken := ""
	if hasNextPage && len(resp) > 0 {
		nextPageToken = resp[len(resp)-1].WalletID
	}

	return &adminv1.ListWalletsResponse{
		Wallets:       resp,
		NextPageToken: nextPageToken,
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

	wallet, err := s.b.Wallets().Get(ctx, req.WalletID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	id, err := s.b.KYC().GetIndividualDetails(ctx, wallet.ID)
	if errors.Is(err, kyc.ErrNoKYCInfo) {
		return &adminv1.WalletDetails{
			Users:       usersPB,
			WalletID:    req.WalletID,
			CountryCode: wallet.Country.String(),
		}, nil
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	kycStatus, err := s.b.KYC().GetKYCStatus(ctx, req.WalletID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var address string
	if id.Address != nil {
		address = id.Address.String()
	}

	return &adminv1.WalletDetails{
		WalletID:     req.WalletID,
		WalletName:   wallet.Name,
		FirstName:    id.FirstName,
		LastName:     id.LastName,
		CountryCode:  wallet.Country.String(),
		PlaceOfBirth: id.PlaceOfBirth,
		Nationality:  id.Nationality,
		Gender:       int32(id.Gender),
		DateOfBirth:  timestamppb.New(id.DateOfBirth),
		Address:      address,
		Users:        usersPB,
		KycStatus:    kycStatus.String(),
	}, nil
}

func (s *AdminRpcService) ListAudit(ctx context.Context, req *adminv1.ListAuditRequest) (*adminv1.ListAuditResponse, error) {
	type Operation struct {
		AdminUser  string    `db:"admin_user"`
		WalletID   string    `db:"wallet_id"`
		Operation  string    `db:"operation"`
		Parameters string    `db:"parameters"`
		Date       time.Time `db:"created_at"`
	}

	var ops []Operation
	err := s.b.DB().SelectContext(ctx, &ops,
		"SELECT admin_user, wallet_id, operation, parameters, created_at FROM admin_audit_log WHERE wallet_id=$1 ORDER BY created_at DESC ",
		req.WalletID)
	if err != nil {
		return nil, err
	}

	resp := make([]*adminv1.AuditOperation, len(ops))
	for i, o := range ops {
		resp[i] = &adminv1.AuditOperation{
			AdminUser:  o.AdminUser,
			WalletID:   o.WalletID,
			Operation:  o.Operation,
			Parameters: o.Parameters,
			Timestamp:  timestamppb.New(o.Date),
		}
	}

	return &adminv1.ListAuditResponse{Operations: resp}, nil
}

func (s *AdminRpcService) SetWalletCountry(ctx context.Context, req *adminv1.SetWalletCountryRequest) (*adminv1.Empty, error) {
	_, err := s.b.Wallets().SetCountry(ctx, req.GetId(), country.ParseCountry(req.GetCountryCode()))
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &adminv1.Empty{}, nil
}

func (s *AdminRpcService) ListCountries(ctx context.Context, req *adminv1.Empty) (*adminv1.ListCountriesResponse, error) {
	var countries = []*adminv1.Country{}
	for c, d := range country.Details {
		countries = append(countries, &adminv1.Country{
			Code:      c.String(),
			Name:      d.Name,
			Numeric:   d.Numeric,
			Supported: d.Supported,
		})
	}

	sort.Slice(countries, func(i, j int) bool {
		return countries[i].Name < countries[j].Name
	})

	return &adminv1.ListCountriesResponse{
		Countries: countries,
	}, nil
}

func (s *AdminRpcService) CheckUserTotpEnabled(ctx context.Context, req *adminv1.CheckUserTotpEnabledRequest) (*adminv1.CheckUserTotpEnabledResponse, error) {
	identityID := req.GetIdentityId()
	if identityID == "" {
		return nil, status.Error(codes.FailedPrecondition, "identityID not provided in request")
	}

	isEnabled, err := s.b.Users().CheckUserTotpEnabled(ctx, identityID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &adminv1.CheckUserTotpEnabledResponse{IsEnabled: isEnabled}, nil
}

func (s *AdminRpcService) Delete2FATotpEnrollment(ctx context.Context, req *adminv1.Delete2FATotpEnrollmentRequest) (*adminv1.Empty, error) {
	identityID := req.GetIdentityId()
	if identityID == "" {
		return nil, status.Error(codes.FailedPrecondition, "identityID not provided in request")
	}

	err := s.b.Users().Delete2FATotpEnrollment(ctx, identityID)
	if err != nil {
		return nil, toGRPCError(err)
	}
	log.Info("admin reset TOTP enrollment", zap.String("identityID", identityID), zap.String("walletID", req.GetWalletID()))

	walletID := req.GetWalletID()
	if walletID != "" {
		s.b.Email().SendAuthenticatorResetEmail(ctx, walletID)
		log.Info("sent authenticator reset email", zap.String("walletID", walletID))
	}

	return &adminv1.Empty{}, nil
}
