package grpc

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"gitlab.com/fynbos/backend/keys"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

type validateCreateConnectionArgs struct {
	ApplicationName string `validate:"required"`
	PublicKey       string `validate:"required"`
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
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	jsonJWK, err := base64.StdEncoding.DecodeString(req.GetPublicKey())
	if err != nil {
		return nil, toGRPCError(err)
	}

	parsedKey, err := jwk.ParseKey(jsonJWK)
	if err != nil {
		return nil, toGRPCError(err)
	}

	rawAlg, ok := parsedKey.Get("crv")
	if !ok {
		return nil, toGRPCError(errors.New("failed to parse jwk"))
	}

	if !strings.EqualFold(fmt.Sprintf("%+v", rawAlg), "Ed25519") {
		return nil, NewValidationError("PublicKey", "Must be an Ed25519 public key.")
	}

	n, _ := parsedKey.Get("x")
	nn, ok := n.([]byte)
	if !ok {
		return nil, toGRPCError(errors.New("failed to parse jwk"))
	}

	keyID := parsedKey.KeyID()
	if keyID == "" {
		return nil, NewValidationError("kid", "kid is required for JWK")
	}
	key, err := s.b.Keys().AddPublicKey(ctx, wallet.ID, base64.RawURLEncoding.EncodeToString(nn), req.GetApplicationName(), keyID)
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
