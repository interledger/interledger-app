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
		Daily: currency.Amount{
			Value:    req.GetDailyLimit().Amount,
			Currency: currency.Currency(req.GetDailyLimit().GetCurrency()),
		},
		Monthly: currency.Amount{
			Value:    req.GetMonthlyLimit().Amount,
			Currency: currency.Currency(req.GetMonthlyLimit().GetCurrency()),
		},
		Overall: currency.Amount{
			Value:    req.GetOverallLimit().Amount,
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
		Daily: &backendv1.PublicKeyLimit{
			Currency: string(keyLimit.Daily.Currency),
			Amount:   keyLimit.Daily.Value,
		},
		Monthly: &backendv1.PublicKeyLimit{
			Currency: string(keyLimit.Monthly.Currency),
			Amount:   keyLimit.Monthly.Value,
		},
		Overall: &backendv1.PublicKeyLimit{
			Currency: string(keyLimit.Overall.Currency),
			Amount:   keyLimit.Overall.Value,
		},
	}, nil
}

type validateUpdatePublicKeyLimitArgs struct {
	ID           string
	DailyLimit   validateLimit
	MonthlyLimit validateLimit
	OverallLimit validateLimit
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
		ID: req.GetId(),
		DailyLimit: validateLimit{
			Value:    req.GetDaily().GetAmount(),
			Currency: req.GetDaily().GetCurrency(),
		},
		MonthlyLimit: validateLimit{
			Value:    req.GetMonthly().GetAmount(),
			Currency: req.GetMonthly().GetCurrency(),
		},
		OverallLimit: validateLimit{
			Value:    req.GetOverall().GetAmount(),
			Currency: req.GetOverall().GetCurrency(),
		},
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	err = s.b.Limits().UpdatePublicKeyLimits(ctx, wallet.ID, req.GetId(), limits.Limit{
		Daily: currency.Amount{
			Value:    req.Daily.Amount,
			Currency: currency.Currency(req.Daily.Currency),
		},
		Monthly: currency.Amount{
			Value:    req.Monthly.Amount,
			Currency: currency.Currency(req.Monthly.Currency),
		},
		Overall: currency.Amount{
			Value:    req.Overall.Amount,
			Currency: currency.Currency(req.Overall.Currency),
		},
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
