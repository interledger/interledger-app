package admin

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/identity"
	unit_external "gitlab.com/fynbos/backend/providers/unit/external"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *AdminRpcService) SignalUnitCustomerCreated(
	ctx context.Context,
	req *backendv1.SignalUnitCustomerCreatedRequest,
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

	user, err := s.IdentityService.GetByEmail(ctx, req.GetUserEmail())
	if errors.Is(err, identity.ErrNotFound) {
		return nil, status.Error(codes.Internal, "User not found for email.")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// TODO: call out to unit to make sure customer exists.
	err = s.Temporal.SignalWorkflow(
		ctx,
		"unit_onboarding_"+user.ID,
		"",
		"onboard-unit-customer-created",
		&unit_external.CustomerCreatedEvent{
			Relationships: unit_external.EventRelationships{
				Customer: unit_external.JsonCustomer{
					Data: unit_external.Data{
						ID:   req.GetCustomerId(),
						Type: req.GetType(),
					},
				},
			},
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &backendv1.Empty{}, nil
}
