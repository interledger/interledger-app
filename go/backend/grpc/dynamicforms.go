package grpc

import (
	"context"
	"gitlab.com/fynbos/backend/dynamicforms"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) SubmitForm(ctx context.Context, req *pb.SubmitFormRequest) (*pb.Empty, error) {
	var walletID string
	_, _ = s.b.Users().UserForContext(ctx)
	w, _ := s.b.Wallets().ForContext(ctx)
	if w != nil {
		walletID = w.ID
	}

	_, err := s.b.DynamicForms().Submit(ctx, &dynamicforms.SubmitArgs{
		FormID:   req.GetFormId(),
		Data:     req.GetData(),
		WalletID: walletID,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}
