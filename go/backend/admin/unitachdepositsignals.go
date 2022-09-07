package admin

import (
	"context"

	unit_external "gitlab.com/fynbos/backend/providers/unit/external"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *AdminRpcService) SignalUnitAchDepositEvent(
	ctx context.Context,
	req *backendv1.SignalUnitAchDepositEventRequest,
) (*backendv1.Empty, error) {
	// TODO: enable auth
	// adminUser, err := s.AuthService.ForContext(ctx)
	// if err != nil {
	// 	return nil, status.Error(codes.Unauthenticated, "Unauthenticated.")
	// }
	// isAdmin := authorizeAdmin(adminUser.Email)
	// if !isAdmin {
	// 	return nil, status.Error(codes.PermissionDenied, "Forbidden.")
	// }
	if req.GetEventType() != string(unit_external.PAYMENT_SENT) &&
		req.GetEventType() != string(unit_external.PAYMENT_REJECTED) &&
		req.GetEventType() != string(unit_external.PAYMENT_RETURNED) &&
		req.GetEventType() != string(unit_external.PAYMENT_CREATED) &&
		req.GetEventType() != string(unit_external.PAYMENT_CLEARING) &&
		req.GetEventType() != string(unit_external.PAYMENT_CANCELED) &&
		req.GetEventType() != string(unit_external.PAYMENT_PENDING_REVIEW) {
		return nil, status.Error(codes.InvalidArgument, "Invalid unit ach payment event type.")
	}

	// TODO: call out to unit to make sure ach payment exists.
	err := s.Temporal.SignalWorkflow(
		ctx,
		"deposit_"+req.GetDepositId(),
		"",
		"unit-user-ach-deposit",
		req.GetEventType(),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &backendv1.Empty{}, nil
}
