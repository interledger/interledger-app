package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/identities"

	pb "gitlab.com/fynbos/proto/backend/v1"
)

// FIXME: fix the structure of the protobuf
func (s *rpcService) ListIdentities(ctx context.Context, _ *pb.Empty) (*pb.ListIdentitiesResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
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
		resp[i] = &pb.Identity{
			Id:                id.ID,
			Platform:          string(id.Platform),
			Handle:            id.Identifier,
			State:             string(id.State),
			VerificationCode:  "",
			VerificationProof: id.VerificationProof,
			Public:            id.Public,
		}
	}

	return &pb.ListIdentitiesResponse{Identities: resp}, nil
}

// FIXME: fix the structure of the protobuf
func (s *rpcService) ListPublicIdentities(ctx context.Context, req *pb.ListPublicIdentitiesRequest) (*pb.ListIdentitiesResponse, error) {
	ids, err := s.b.Identities().ListPublic(ctx, req.WalletId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	resp := make([]*pb.Identity, len(ids))
	for i, id := range ids {
		resp[i] = &pb.Identity{
			Id:                id.ID,
			Platform:          string(id.Platform),
			Handle:            id.Identifier,
			State:             string(id.State),
			VerificationCode:  "",
			VerificationProof: id.VerificationProof,
			Public:            id.Public,
		}
	}

	return &pb.ListIdentitiesResponse{Identities: resp}, nil
}

// Deprecate
func (s *rpcService) AddIdentity(ctx context.Context, req *pb.AddIdentityRequest) (*pb.IdentityVerificationInstructions, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	w, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	_, err = s.b.Identities().Add(ctx, identities.AddArgs{
		WalletID:   w.ID,
		Platform:   identities.Platform(req.Platform),
		Identifier: req.Handle,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.IdentityVerificationInstructions{
		IdentityId:   "",
		Code:         "",
		Instructions: "",
	}, nil

}

func (s *rpcService) DeleteIdentity(ctx context.Context, req *pb.DeleteIdentityRequest) (*pb.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
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
		return nil, ForbiddenError("Unauthenticated.")
	}

	w, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	id, err := s.b.Identities().SetPublic(ctx, req.Id, w.ID, req.Public)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Identity{
		Id:                id.ID,
		Platform:          string(id.Platform),
		Handle:            id.Identifier,
		State:             string(id.State),
		VerificationCode:  "",
		VerificationProof: id.VerificationProof,
		Public:            id.Public,
	}, nil
}

// FIXME: Deprecate
func (s *rpcService) StartIdentityVerification(ctx context.Context, req *pb.StartIdentityVerificationRequest) (*pb.Identity, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	_, err = s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	id, err := s.b.Identities().StartVerification(ctx, req.Id, req.Proof)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Identity{
		Id:                id.ID,
		Platform:          string(id.Platform),
		Handle:            id.Identifier,
		State:             string(id.State),
		VerificationCode:  "",
		VerificationProof: id.VerificationProof,
		Public:            id.Public,
	}, nil
}
