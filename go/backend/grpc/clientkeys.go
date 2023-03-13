package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/authorisation"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/limits"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

type validateLimit struct {
	Value    uint64 `validate:"required"`
	Currency string `validate:"oneof=USD"`
}

type validateCreatePublicKeyArgs struct {
	ApplicationName string `validate:"required"`
	PublicKey       string `validate:"required"`
	DailyLimit      validateLimit
	MonthlyLimit    validateLimit
	OverallLimit    validateLimit
}

func (s *rpcService) CreatePublicKey(
	ctx context.Context, req *backendv1.CreatePublicKeyRequest,
) (*backendv1.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	err = s.b.Validator().StructCtx(ctx, validateCreatePublicKeyArgs{
		ApplicationName: req.GetApplicationName(),
		PublicKey:       req.GetPublicKey(),
		DailyLimit: validateLimit{
			Value:    req.GetDailyLimit().GetAmount(),
			Currency: req.GetDailyLimit().GetCurrency(),
		},
		MonthlyLimit: validateLimit{
			Value:    req.GetMonthlyLimit().GetAmount(),
			Currency: req.GetMonthlyLimit().GetCurrency(),
		},
		OverallLimit: validateLimit{
			Value:    req.GetOverallLimit().GetAmount(),
			Currency: req.GetOverallLimit().GetCurrency(),
		},
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	// using payment pointer as client url for now
	pp, err := s.b.OpenPayments().GetWalletPaymentPointer(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	err = s.b.Authorisation().AddPublicKey(ctx, pp.URL, authorisation.Jwk{
		Kty: "OKP",
		Kid: req.GetApplicationName(),
		Alg: "EdDSA",
		Crv: "Ed25519",
		X:   req.GetPublicKey(),
		Use: "sign",
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	err = s.b.Limits().UpdateClientLimits(ctx, wallet.ID, pp.URL, limits.Limit{
		Daily: currency.Amount{
			Value:    req.GetDailyLimit().GetAmount(),
			Currency: currency.Currency(req.GetDailyLimit().GetCurrency()),
		},
		Monthly: currency.Amount{
			Value:    req.GetMonthlyLimit().GetAmount(),
			Currency: currency.Currency(req.GetMonthlyLimit().GetCurrency()),
		},
		Overall: currency.Amount{
			Value:    req.GetOverallLimit().GetAmount(),
			Currency: currency.Currency(req.GetOverallLimit().GetCurrency()),
		},
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.Empty{}, nil
}

func (s *rpcService) ListPublicKeys(ctx context.Context, req *backendv1.Empty) (*backendv1.ListPublicKeysResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	pp, err := s.b.OpenPayments().GetWalletPaymentPointer(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	keys, err := s.b.Authorisation().ListKeys(ctx, pp.URL)
	if err != nil {
		return nil, toGRPCError(err)
	}

	keyset := make([]*backendv1.PublicKey, len(keys))
	for i, k := range keys {
		keyset[i] = &backendv1.PublicKey{
			Id:              k.ID,
			ApplicationName: k.Kid,
			PublicKey:       k.X,
			CreatedAt:       k.CreatedAt.Format("Jan 02, 2006"),
			LastUsedAt:      "",
		}
	}

	return &backendv1.ListPublicKeysResponse{
		Keys: keyset,
	}, nil
}

func (s *rpcService) GetPublicKey(ctx context.Context, req *backendv1.GetPublicKeyRequest) (*backendv1.PublicKey, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	pp, err := s.b.OpenPayments().GetWalletPaymentPointer(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	key, err := s.b.Authorisation().GetPublicKeyByID(ctx, pp.URL, req.GetId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.PublicKey{
		Id:              key.ID,
		ApplicationName: key.Kid,
		PublicKey:       key.X,
		CreatedAt:       key.CreatedAt.Format("Jan 02, 2006"),
		LastUsedAt:      "",
	}, nil
}

func (s *rpcService) GetPublicKeyLimits(ctx context.Context, req *backendv1.GetPublicKeyLimitsRequest) (*backendv1.PublicKeyLimits, error) {
	panic("TODO: implement me")
}

func (s *rpcService) UpdatePublicKeyLimit(ctx context.Context, req *backendv1.UpdatePublicKeyLimitsRequest) (*backendv1.Empty, error) {
	panic("TODO: implement me")
}

func (s *rpcService) DeletePublicKey(ctx context.Context, req *backendv1.DeletePublicKeyRequest) (*backendv1.Empty, error) {
	panic("TODO: implement me")
}
