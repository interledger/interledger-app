package grpc

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/keys"
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
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
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

	pemBlock, _ := pem.Decode([]byte(req.GetPublicKey()))
	if pemBlock.Type != "PUBLIC KEY" {
		return nil, NewValidationError("PublicKey", "Must be an Ed25519 public key in pem format.")
	}

	pub, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)
	if err != nil {
		return nil, toGRPCError(err)
	}

	ed25519PubKey, ok := pub.(ed25519.PublicKey)
	if !ok {
		return nil, NewValidationError("PublicKey", "Must be an Ed25519 public key in pem format.")
	}

	key, err := s.b.Keys().AddPublicKey(ctx, wallet.ID, base64.StdEncoding.EncodeToString(ed25519PubKey), req.GetApplicationName())
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

	err = s.b.Rafiki().CreatePaymentPointerKey(ctx, key.ID, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.Empty{}, nil
}

func (s *rpcService) ListConnections(ctx context.Context, req *backendv1.Empty) (*backendv1.ListConnectionsResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	keys, err := s.b.Keys().List(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	keyset := make([]*backendv1.Connection, len(keys))
	for i, k := range keys {
		fingerprint, err := k.Fingerprint()
		if err != nil {
			return nil, toGRPCError(err)
		}

		keyset[i] = &backendv1.Connection{
			Id:                   k.ID,
			ApplicationName:      k.Name,
			PublicKeyFingerprint: fingerprint,
			CreatedAt:            k.CreatedAt.Format("Jan 02, 2006"),
			LastUsedAt:           "",
		}
	}

	return &backendv1.ListConnectionsResponse{
		Keys: keyset,
	}, nil
}

func (s *rpcService) GetConnection(ctx context.Context, req *backendv1.GetConnectionRequest) (*backendv1.Connection, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	keyset, err := s.b.Keys().List(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	var key *keys.Key
	for _, k := range keyset {
		if k.ID == req.GetId() {
			key = &k
			break
		}
	}
	if key == nil {
		return nil, NotFoundError("Not found.")
	}

	fingerprint, err := key.Fingerprint()
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.Connection{
		Id:                   key.ID,
		ApplicationName:      key.Name,
		PublicKeyFingerprint: fingerprint,
		CreatedAt:            key.CreatedAt.Format("Jan 02, 2006"),
		LastUsedAt:           "",
	}, nil
}

func (s *rpcService) GetConnectionLimits(ctx context.Context, req *backendv1.GetConnectionLimitsRequest) (*backendv1.ConnectionLimits, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	keyLimit, err := s.b.Limits().GetPublicKeyLimit(ctx, wallet.ID, req.GetId())
	if err != nil {
		return nil, toGRPCError(err)
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
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
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
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	keyset, err := s.b.Keys().List(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	var key *keys.Key
	for _, k := range keyset {
		if k.ID == req.GetId() {
			key = &k
			break
		}
	}
	if key == nil {
		return &backendv1.Empty{}, nil
	}

	err = s.b.Keys().DeletePublicKey(ctx, key.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	err = s.b.Rafiki().RevokePaymentPointerKey(ctx, key.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.Empty{}, nil
}
