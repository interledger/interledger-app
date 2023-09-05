package grpc

import (
	"context"
	"gitlab.com/fynbos/backend/dynamicforms"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) CreateDynamicForm(ctx context.Context, req *pb.CreateDynamicFormRequest) (*pb.Empty, error) {
	var walletID string
	_, _ = s.b.Users().UserForContext(ctx)
	w, _ := s.b.Wallets().ForContext(ctx)
	if w != nil {
		walletID = w.ID
	}

	_, err := s.b.DynamicForms().Create(ctx, &dynamicforms.CreateFormArgs{
		FormID:   req.GetFormId(),
		FormData: req.GetData(),
		WalletID: walletID,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}
