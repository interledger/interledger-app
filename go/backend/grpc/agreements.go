package grpc

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/agreements"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) GetAgreement(
	ctx context.Context,
	req *backendv1.GetAgreementRequest,
) (*backendv1.Agreement, error) {
	if err := s.b.Validator().VarCtx(ctx, req.GetId(), "required"); err != nil {
		return nil, toGRPCError(err)
	}

	agreement, err := s.b.Agreements().Get(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, agreements.ErrNotFound) {
			return nil, NotFoundError("Failed to find agreement.")
		}
		return nil, InternalError("Get agreement.")
	}

	return &backendv1.Agreement{
		Content: agreement.Content,
	}, nil
}

type validateSignAgreements struct {
	AgreementIDs []string `validate:"required,min=1"`
	UserID       string   `validate:"required,uuid"`
}

func (s *rpcService) SignAgreements(
	ctx context.Context,
	req *backendv1.SignAgreementsRequest,
) (*backendv1.SignAgreementsResponse, error) {
	if err := s.b.Validator().StructCtx(ctx, &validateSignAgreements{
		AgreementIDs: req.GetAgreementIds(),
		UserID:       req.UserId,
	}); err != nil {
		return nil, toGRPCError(err)
	}

	err := s.b.Agreements().Sign(ctx, &agreements.SignArgs{
		AgreementIDs: req.GetAgreementIds(),
		UserID:       req.UserId,
	})
	if err != nil {
		if errors.Is(err, agreements.ErrNotFound) {
			return nil, NotFoundError("Agreement not found.")
		}
		return nil, InternalError("Sign agreements.")
	}

	return &backendv1.SignAgreementsResponse{
		Signed: true,
	}, nil
}
