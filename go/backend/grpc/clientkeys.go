package grpc

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/authorisation"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/limits"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

type validateCreatePublicKeyArgs struct {
	ApplicationName string `validate:"required"`
	PublicKey       string `validate:"required"`
	DailyLimit      currency.Amount
	MonthlyLimit    currency.Amount
	OverallLimit    currency.Amount
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
		DailyLimit:      currency.FromPB(req.GetDailyLimit()),
		MonthlyLimit:    currency.FromPB(req.GetMonthlyLimit()),
		OverallLimit:    currency.FromPB(req.GetOverallLimit()),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	// using payment pointer as client url for now
	pp, err := s.b.OpenPayments().GetWalletPaymentPointer(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}
	key, err := s.b.Authorisation().AddPublicKey(ctx, pp.URL, authorisation.Jwk{
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

	err = s.b.Limits().UpdatePublicKeyLimits(ctx, wallet.ID, key.ID, limits.Limit{
		Daily:   currency.FromPB(req.GetDailyLimit()),
		Monthly: currency.FromPB(req.GetMonthlyLimit()),
		Overall: currency.FromPB(req.GetOverallLimit()),
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
	if errors.Is(err, authorisation.ErrNotFound) {
		return &backendv1.ListPublicKeysResponse{
			Keys: nil, // no keys to return
		}, nil
	}
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
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	lims, err := s.b.Limits().List(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	var keyLimit *limits.Limit
	for _, l := range lims {
		if l.ForeignType == limits.FKTypeClientPublicKey && l.ForeignID == req.GetId() {
			keyLimit = &l.Limit
			break
		}
	}
	if keyLimit == nil {
		return nil, NotFoundError("public key not found")
	}

	return &backendv1.PublicKeyLimits{
		Daily:   keyLimit.Daily.ToPB(),
		Monthly: keyLimit.Monthly.ToPB(),
		Overall: keyLimit.Overall.ToPB(),
	}, nil
}

type validateUpdatePublicKeyLimitArgs struct {
	ID           string
	DailyLimit   currency.Amount
	MonthlyLimit currency.Amount
	OverallLimit currency.Amount
}

func (s *rpcService) UpdatePublicKeyLimit(ctx context.Context, req *backendv1.UpdatePublicKeyLimitsRequest) (*backendv1.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	err = s.b.Validator().StructCtx(ctx, validateUpdatePublicKeyLimitArgs{
		ID:           req.GetId(),
		DailyLimit:   currency.FromPB(req.GetDaily()),
		MonthlyLimit: currency.FromPB(req.GetMonthly()),
		OverallLimit: currency.FromPB(req.GetOverall()),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	err = s.b.Limits().UpdatePublicKeyLimits(ctx, wallet.ID, req.GetId(), limits.Limit{
		Daily:   currency.FromPB(req.GetDaily()),
		Monthly: currency.FromPB(req.GetMonthly()),
		Overall: currency.FromPB(req.GetOverall()),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.Empty{}, nil
}

func (s *rpcService) DeletePublicKey(ctx context.Context, req *backendv1.DeletePublicKeyRequest) (*backendv1.Empty, error) {
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

	err = s.b.Authorisation().DeletePublicKey(ctx, pp.URL, req.GetId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.Empty{}, nil
}
