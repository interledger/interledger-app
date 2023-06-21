package admin

import (
	"context"

	"gitlab.com/fynbos/backend/db"
	pb "gitlab.com/fynbos/proto/backend/admin/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *AdminRpcService) ListIncompleteLinkedAccountReviews(ctx context.Context, req *pb.PaginationRequest) (*pb.LinkedAccountReviews, error) {
	reviews, err := s.b.LinkedAccounts().ListIncompleteReviews(ctx, db.FromAdminPB(req))
	if err != nil {
		return nil, toGRPCError(err)
	}

	pbReviews := make([]*pb.LinkedAccountReview, len(reviews))
	for i, r := range reviews {
		pbReviews[i] = &pb.LinkedAccountReview{
			Id:              r.ID,
			LinkedAccountID: r.LinkedAccountID,
			State:           string(r.State),
			NewState:        string(r.NewState),
			ReviewedBy:      r.ReviewedBy,
			Reason:          r.Reason,
			CreatedAt:       timestamppb.New(r.CreatedAt),
			CompletedAt:     timestamppb.New(r.CompletedAt),
		}
	}

	return &pb.LinkedAccountReviews{
		Reviews: pbReviews,
	}, nil
}

func (s *AdminRpcService) GetLinkedAccountReview(ctx context.Context, req *pb.GetLinkedAccountReviewRequest) (*pb.LinkedAccountReview, error) {
	r, err := s.b.LinkedAccounts().GetReview(ctx, req.GetId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.LinkedAccountReview{
		Id:              r.ID,
		State:           string(r.State),
		NewState:        string(r.NewState),
		LinkedAccountID: r.LinkedAccountID,
		ReviewedBy:      r.ReviewedBy,
		Reason:          r.Reason,
		CreatedAt:       timestamppb.New(r.CreatedAt),
		CompletedAt:     timestamppb.New(r.CompletedAt),
	}, nil
}

func (s *AdminRpcService) CompleteLinkedAccountReview(ctx context.Context, req *pb.CompleteLinkedAccountReviewRequest) (*pb.LinkedAccountReview, error) {
	reviewer, err := s.b.AdminAuth().ForContext(ctx)
	if err != nil {
		return nil, toGRPCError(err)
	}

	r, err := s.b.LinkedAccounts().CompleteReview(ctx, req.GetId(), reviewer.Email)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.LinkedAccountReview{
		Id:              r.ID,
		State:           string(r.State),
		NewState:        string(r.NewState),
		LinkedAccountID: r.LinkedAccountID,
		ReviewedBy:      r.ReviewedBy,
		Reason:          r.Reason,
		CreatedAt:       timestamppb.New(r.CreatedAt),
		CompletedAt:     timestamppb.New(r.CompletedAt),
	}, nil
}
