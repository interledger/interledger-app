package admin

import (
	"context"
	"errors"
	"time"

	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub"
	pb "github.com/interledger/interledger-app/go/proto/backend/admin/v1"
)

func (s *AdminRpcService) CreateGatehubUser(
	ctx context.Context, req *pb.CreateGatehubUserRequest,
) (*pb.Empty, error) {
	await, err := s.b.Gatehub().CreateUser(ctx, req.GetWalletID())
	if err != nil {
		return nil, toGRPCError(err)
	}

	err = await(ctx, nil)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}

func (s *AdminRpcService) GetGatehubBalance(
	ctx context.Context, req *pb.GetGatehubBalanceRequest,
) (*pb.GetGatehubBalanceResponse, error) {
	balanceLas, err := s.b.LinkedAccounts().ListBalances(ctx, req.GetWalletID())
	if err != nil {
		return nil, toGRPCError(err)
	}

	var gatehubLa *linkedaccounts.LinkedAccount
	for _, balance := range balanceLas {
		if balance.Provider == gatehub.ProviderName && balance.Type == gatehub.AccTypeBalance {
			gatehubLa = &balance
			break
		}
	}
	if gatehubLa == nil {
		return nil, NotFoundError("balance not found")
	}

	bal, err := s.b.Gatehub().GetBalance(ctx, gatehubLa.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.GetGatehubBalanceResponse{
		Balance:   bal.Total.ToAdminPB(),
		Available: bal.Available.ToAdminPB(),
	}, nil
}

func (s *AdminRpcService) GetGatehubUser(
	ctx context.Context, req *pb.GetGatehubUserRequest,
) (*pb.GatehubUser, error) {
	externalUser, err := s.b.Gatehub().GetUser(ctx, req.GetWalletID())
	if errors.Is(err, gatehub.ErrNotFound) {
		return nil, NotFoundError("Gatehub user not found")
	}
	if err != nil {
		return nil, toGRPCError(err)
	}

	ret := &pb.GatehubUser{
		WalletID:          req.GetWalletID(),
		Email:             externalUser.Email,
		ExternalID:        externalUser.UUID,
		VerificationLevel: int32(externalUser.VerificationLevel),
		Profile: &pb.GatehubProfile{
			Uuid:               externalUser.Profile.UUID,
			BirthCity:          externalUser.Profile.BirthCity,
			BirthCountryCode:   externalUser.Profile.BirthCountryCode,
			TaxResidency:       externalUser.Profile.TaxResidency,
			BirthDay:           int32(externalUser.Profile.BirthDay),
			BirthMonth:         int32(externalUser.Profile.BirthMonth),
			BirthYear:          int32(externalUser.Profile.BirthYear),
			Gender:             externalUser.Profile.Gender,
			FirstName:          externalUser.Profile.FirstName,
			MiddleName:         externalUser.Profile.MiddleName,
			LastName:           externalUser.Profile.LastName,
			Citizenship:        externalUser.Profile.Citizenship,
			AddressPostalCode:  externalUser.Profile.AddressPostalCode,
			AddressSubdivision: externalUser.Profile.AddressSubdivision,
			AddressCountryCode: externalUser.Profile.AddressCountryCode,
			AddressCity:        externalUser.Profile.AddressCity,
			AddressStreet1:     externalUser.Profile.AddressStreet1,
			AddressStreet2:     externalUser.Profile.AddressStreet2,
			ExpectedVolume:     externalUser.Profile.ExpectedVolume,
		},
	}

	for _, doc := range externalUser.Documents {
		ret.Documents = append(ret.Documents, &pb.GatehubDocument{
			Id:         int32(doc.ID),
			Uuid:       doc.UUID,
			Type:       doc.Type,
			Subtype:    doc.Subtype,
			Value:      doc.Value,
			ProfileId:  int32(doc.ProfileID),
			UserId:     int32(doc.UserID),
			ExpiryDate: doc.ExpiryDate.Format(time.RFC3339),
		})
	}

	for _, verification := range externalUser.Verifications {
		ret.Verifications = append(ret.Verifications, &pb.GatehubVerification{
			Uuid:         verification.UUID,
			State:        int32(verification.State),
			Status:       int32(verification.Status),
			ProviderType: verification.ProviderType,
		})
	}

	return ret, nil
}
