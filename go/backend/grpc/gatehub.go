package grpc

import (
	"context"

	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (r *rpcService) GetGatehubOnboardingWidget(
	ctx context.Context, req *pb.Empty,
) (*pb.GatehubOnboardingWidget, error) {
	_, err := r.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := r.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	widget, err := r.b.Gatehub().GetOnboardingWidget(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.GatehubOnboardingWidget{
		WidgetUrl: widget,
	}, nil
}
