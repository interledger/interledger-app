package grpc

import (
	"context"

	"github.com/interledger/interledger-app/go/backend/country"
	pb "github.com/interledger/interledger-app/go/proto/backend/v1"
)

func (r *rpcService) GetGatehubOnboardingWidget(
	ctx context.Context, req *pb.Empty,
) (*pb.GatehubWidget, error) {
	_, err := r.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := r.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	if !country.EUCountries[wallet.Country] {
		return nil, toGRPCError(FailedPreconditionError("Wallet not in the EU region"))
	}

	widget, err := r.b.Gatehub().GetOnboardingWidget(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.GatehubWidget{
		WidgetUrl: widget,
	}, nil
}

func (r *rpcService) GetGatehubDepositWidget(
	ctx context.Context, req *pb.Empty,
) (*pb.GatehubWidget, error) {
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

	widget, err := r.b.Gatehub().GetOnOffRampWidget(ctx, wallet.ID, true)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.GatehubWidget{
		WidgetUrl: widget,
	}, nil
}

func (r *rpcService) GetGatehubWithdrawalWidget(
	ctx context.Context, req *pb.Empty,
) (*pb.GatehubWidget, error) {
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

	widget, err := r.b.Gatehub().GetOnOffRampWidget(ctx, wallet.ID, false)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.GatehubWidget{
		WidgetUrl: widget,
	}, nil
}

func (r *rpcService) CreateGatehubWithdrawal(ctx context.Context, req *pb.CreateGatehubWithdrawalRequest) (*pb.CreateGatehubWithdrawalResponse, error) {
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

	txID, err := r.b.Gatehub().CreateWithdrawal(ctx, wallet.ID, req.GetExternalTransactionId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.CreateGatehubWithdrawalResponse{
		TransactionId: txID,
	}, nil
}
