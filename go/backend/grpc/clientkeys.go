package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/authorisation"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) CreateClientPublicKey(
	ctx context.Context, req *backendv1.JWK,
) (*backendv1.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	key := authorisation.Jwk{
		Kty: req.GetKty(),
		Kid: req.GetKty(),
		Alg: req.GetAlg(),
		Crv: req.GetCrv(),
		X:   req.GetX(),
		Use: req.GetUse(),
	}
	if !key.IsEdDSAPublicKey() {
		return nil, NewValidationError("", "Must be a EdDSA public key.")
	}

	// using payment pointer as client url for now
	pp, err := s.b.OpenPayments().GetWalletPaymentPointer(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	err = s.b.Authorisation().AddPublicKey(ctx, pp.URL, key)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.Empty{}, nil
}
