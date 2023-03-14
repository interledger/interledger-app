package grpc

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/authorisation"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/limits"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

type validateCreateConnectionArgs struct {
	ApplicationName string `validate:"required"`
	PublicKey       string `validate:"required"`
	DailyLimit      currency.Amount
	MonthlyLimit    currency.Amount
	OverallLimit    currency.Amount
}

func (s *rpcService) CreateConnection(
	ctx context.Context, req *backendv1.CreateConnectionRequest,
) (*backendv1.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	err = s.b.Validator().StructCtx(ctx, validateCreateConnectionArgs{
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

func (s *rpcService) ListConnections(ctx context.Context, req *backendv1.Empty) (*backendv1.ListConnectionsResponse, error) {
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
		return &backendv1.ListConnectionsResponse{
			Keys: nil, // no keys to return
		}, nil
	}
	if err != nil {
		return nil, toGRPCError(err)
	}

	keyset := make([]*backendv1.Connection, len(keys))
	for i, k := range keys {
		keyset[i] = &backendv1.Connection{
			Id:              k.ID,
			ApplicationName: k.Kid,
			PublicKey:       k.X,
			CreatedAt:       k.CreatedAt.Format("Jan 02, 2006"),
			LastUsedAt:      "",
		}
	}

	return &backendv1.ListConnectionsResponse{
		Keys: keyset,
	}, nil
}

func (s *rpcService) GetConnection(ctx context.Context, req *backendv1.GetConnectionRequest) (*backendv1.Connection, error) {
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

	return &backendv1.Connection{
		Id:              key.ID,
		ApplicationName: key.Kid,
		PublicKey:       key.X,
		CreatedAt:       key.CreatedAt.Format("Jan 02, 2006"),
		LastUsedAt:      "",
	}, nil
}

func (s *rpcService) GetConnectionLimits(ctx context.Context, req *backendv1.GetConnectionLimitsRequest) (*backendv1.ConnectionLimits, error) {
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

	return &backendv1.ConnectionLimits{
		Daily:   keyLimit.Daily.ToPB(),
		Monthly: keyLimit.Monthly.ToPB(),
		Overall: keyLimit.Overall.ToPB(),
	}, nil
}

type validateUpdateConnectionLimitArgs struct {
	ID           string
	DailyLimit   currency.Amount
	MonthlyLimit currency.Amount
	OverallLimit currency.Amount
}

func (s *rpcService) UpdateConnectionLimits(ctx context.Context, req *backendv1.UpdateConnectionLimitsRequest) (*backendv1.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	err = s.b.Validator().StructCtx(ctx, validateUpdateConnectionLimitArgs{
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

func (s *rpcService) DeleteConnection(ctx context.Context, req *backendv1.DeleteConnectionRequest) (*backendv1.Empty, error) {
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
