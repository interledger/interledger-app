package admin

import (
	"bytes"
	"context"
	"gitlab.com/fynbos/backend/db"
	pb "gitlab.com/fynbos/proto/backend/admin/v1"
)

func (s *AdminRpcService) ListDynamicFormCounts(ctx context.Context, req *pb.PaginationRequest) (*pb.ListDynamicFormCountsResponse, error) {
	page := db.FromAdminPB(req)
	fcs, err := s.b.DynamicForms().ListFormCounts(ctx, page)
	if err != nil {
		return nil, toGRPCError(err)
	}

	pbFormCounts := make([]*pb.DynamicFormCount, len(fcs))
	for i, fc := range fcs {
		pbFormCounts[i] = &pb.DynamicFormCount{
			FormId:    fc.FormID,
			FormCount: fc.Count,
		}
	}

	return &pb.ListDynamicFormCountsResponse{
		DynamicFormCounts: pbFormCounts,
		NextPageToken:     "", // TODO: pagination
	}, nil
}

func (s *AdminRpcService) ExportDynamicForm(req *pb.ExportDynamicFormRequest, stream pb.Backend_ExportDynamicFormServer) error {
	ctx := stream.Context()
	buf := bytes.Buffer{}

	err := s.b.DynamicForms().ExportFormResults(ctx, req.GetFormId(), &buf)
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

		err = stream.Send(&pb.ExportDynamicFormResponse{
			Chunk: p[:n],
		})
		if err != nil {
			return toGRPCError(err)
		}
	}

	return nil
}
