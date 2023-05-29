package grpc

import (
	"context"
	"encoding/base64"
	"errors"
	"gitlab.com/fynbos/backend/identities"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) ListIdentities(ctx context.Context, _ *pb.Empty) (*pb.ListIdentitiesResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	ids, err := s.b.Identities().List(ctx, w.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	resp := make([]*pb.Identity, len(ids))
	for i, id := range ids {
		resp[i] = identityToPB(&id)
	}

	return &pb.ListIdentitiesResponse{Identities: resp}, nil
}

func (s *rpcService) ListPublicIdentities(ctx context.Context, req *pb.ListPublicIdentitiesRequest) (*pb.ListIdentitiesResponse, error) {
	ids, err := s.b.Identities().ListPublic(ctx, req.WalletId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	resp := make([]*pb.Identity, len(ids))
	for i, id := range ids {
		resp[i] = identityToPB(&id)
	}

	return &pb.ListIdentitiesResponse{Identities: resp}, nil
}

func (s *rpcService) DeleteIdentity(ctx context.Context, req *pb.DeleteIdentityRequest) (*pb.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	err = s.b.Identities().Delete(ctx, req.Id, w.ID)

	return &pb.Empty{}, toGRPCError(err)
}

func (s *rpcService) SetIdentityPublic(ctx context.Context, req *pb.SetIdentityPublicRequest) (*pb.Identity, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	id, err := s.b.Identities().SetPublic(ctx, req.Id, w.ID, req.Public)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return identityToPB(id), nil
}

func (s *rpcService) GetIdentity(ctx context.Context, req *pb.GetIdentityRequest) (*pb.GetIdentityResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	id, err := s.b.Identities().Get(ctx, req.Id)
	if err != nil {
		if errors.Is(err, identities.ErrNotFound) {
			return nil, NotFoundError("identity not found.")
		}
		return nil, toGRPCError(err)
	}

	if id.WalletID != w.ID {
		return nil, NotFoundError("identity not found.")
	}

	return &pb.GetIdentityResponse{
		Identity: identityToPB(id),
	}, nil
}

func (s *rpcService) GetIdentityBySignatureHash(ctx context.Context, req *pb.GetIdentityBySignatureHashRequest) (*pb.GetIdentityResponse, error) {
	sigHashBase64 := req.GetSignatureHash()
	sigHash, err := base64.URLEncoding.DecodeString(sigHashBase64)
	if err != nil {
		return nil, toGRPCError(err)
	}

	id, err := s.b.Identities().GetBySignatureHash(ctx, sigHash)
	if err != nil {
		if errors.Is(err, identities.ErrNotFound) {
			return nil, NotFoundError("identity not found.")
		}
		return nil, toGRPCError(err)
	}

	return &pb.GetIdentityResponse{
		Identity: identityToPB(id),
	}, nil
}

func identityToPB(identity *identities.Identity) *pb.Identity {
	base64Signature := base64.URLEncoding.EncodeToString(identity.Signature)
	base64SignatureHash := base64.URLEncoding.EncodeToString(identity.SignatureHash)

	return &pb.Identity{
		Id:            identity.ID,
		Wallet:        "",
		Platform:      string(identity.Platform),
		Identifier:    identity.Identifier,
		State:         string(identity.State),
		KeyId:         identity.KeyID,
		Signature:     string(base64Signature),
		SignatureHash: string(base64SignatureHash),
		Proof:         identity.VerificationProof,
		Ctime:         identity.CreatedAt.String(),
		VerifiedAt:    timestamppb.New(identity.VerifiedAt.Time),
		Public:        identity.Public,
	}
}
