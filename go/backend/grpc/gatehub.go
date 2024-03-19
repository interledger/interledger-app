package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/country"
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

	_, isEU := country.EUCountries[wallet.Country]
	if !isEU {
		return nil, toGRPCError(FailedPreconditionError("Wallet not in the EU region"))
	}

	widget, err := r.b.Gatehub().GetOnboardingWidget(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.GatehubOnboardingWidget{
		WidgetUrl: widget,
	}, nil
}
