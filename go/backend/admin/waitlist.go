package admin

import (
	"context"
	adminv1 "gitlab.com/fynbos/proto/backend/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *AdminRpcService) ListWaitlistSignups(
	ctx context.Context, request *emptypb.Empty,
) (*adminv1.ListWaitlistSignupsResponse, error) {
	//user, err := s.b.AdminAuth().GetAdminUser(ctx)
	//if err != nil {
	//	return nil, status.Error(codes.Internal, err.Error())
	//}

	//if !authorizeAdmin(user.Email) {
	//	return nil, status.Error(codes.PermissionDenied, "forbidden")
	//}

	signups, err := s.b.Waitlist().ListSignups(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	ret := make([]*adminv1.WaitlistSignup, len(signups))
	for i, signup := range signups {
		ret[i] = &adminv1.WaitlistSignup{
			Id:        signup.ID,
			Name:      signup.Name,
			Email:     signup.Email,
			BetaOptIn: signup.BetaOtpIn,
		}
	}

	return &adminv1.ListWaitlistSignupsResponse{
		Signups: ret,
	}, nil
}
