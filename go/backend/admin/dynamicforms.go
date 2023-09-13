package admin

import (
	"bytes"
	"context"
	"gitlab.com/fynbos/backend/db"
	pb "gitlab.com/fynbos/proto/backend/admin/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strconv"
)

func (s *AdminRpcService) ListFormSubmissionCounts(ctx context.Context, req *pb.PaginationRequest) (*pb.ListFormSubmissionCountsResponse, error) {
	page := db.FromAdminPB(req)
	fcs, err := s.b.DynamicForms().ListSubmissionCounts(ctx, page)
	if err != nil {
		return nil, toGRPCError(err)
	}

	pbFormCounts := make([]*pb.FormSubmissionCount, len(fcs))
	for i, fc := range fcs {
		pbFormCounts[i] = &pb.FormSubmissionCount{
			FormId:          fc.FormID,
			SubmissionCount: fc.Count,
		}
	}

	return &pb.ListFormSubmissionCountsResponse{
		FormSubmissionCounts: pbFormCounts,
		NextPageToken:        "", // TODO: pagination
	}, nil
}

func (s *AdminRpcService) ExportFormSubmissions(req *pb.ExportFormSubmissionsRequest, stream pb.Backend_ExportFormSubmissionsServer) error {
	ctx := stream.Context()
	buf := bytes.Buffer{}

	err := s.b.DynamicForms().ExportSubmissions(ctx, req.GetFormId(), &buf)
	if err != nil {
		return toGRPCError(err)
	}

	p := make([]byte, 1024)
	for {
		n, err := buf.Read(p)
		if err != nil {
			return toGRPCError(err)
		}
		if n == 0 {
			break
		}

		err = stream.Send(&pb.ExportFormSubmissionsResponse{
			Chunk: p[:n],
		})
		if err != nil {
			return toGRPCError(err)
		}
	}

	return nil
}

func (s *AdminRpcService) ListFormSubmissions(ctx context.Context, req *pb.ListFormSubmissionsRequest) (*pb.ListFormSubmissionsResponse, error) {
	subs, err := s.b.DynamicForms().ListSubmissions(ctx, req.GetFormId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	var pbSubs []*pb.FormSubmission
	for _, sub := range subs {
		timestamp := timestamppb.New(sub.CreatedAt)

		pbSubs = append(pbSubs, &pb.FormSubmission{
			Id:        sub.ID,
			FormId:    sub.FormID,
			Timestamp: timestamp,
		})
	}

	return &pb.ListFormSubmissionsResponse{
		FormSubmissions: pbSubs,
	}, nil
}

func (s *AdminRpcService) GetFormSubmissionDetails(ctx context.Context, req *pb.GetFormSubmissionDetailsRequest) (*pb.FormSubmissionDetails, error) {
	sub, err := s.b.DynamicForms().GetSubmission(ctx, req.GetId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	jsonString, err := strconv.Unquote(sub.Data)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.FormSubmissionDetails{
		Id:        sub.ID,
		FormId:    sub.FormID,
		WalletId:  &sub.WalletID.String,
		Data:      jsonString,
		Timestamp: timestamppb.New(sub.CreatedAt),
	}, nil
}
