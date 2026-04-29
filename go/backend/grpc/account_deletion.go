package grpc

import (
	"context"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"gitlab.com/fynbos/log"
	pb "gitlab.com/fynbos/proto/backend/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *rpcService) RequestAccountDeletion(ctx context.Context, req *pb.RequestAccountDeletionRequest) (*pb.Empty, error) {
	u, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	code := strings.TrimSpace(req.GetTotpCode())
	if code == "" {
		return nil, NewValidationError("totp_code", "Enter your authenticator code.")
	}

	err = s.b.Users().ValidateTotpCode(ctx, u.ID, code, time.Now())
	if err != nil {
		return nil, toGRPCError(err)
	}

	// Persist before notifying so duplicates are deduped before support is paged.
	err = s.b.AccountDeletion().Request(ctx, u.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if emailErr := s.b.Email().NotifyAccountDeletionRequested(ctx, u.ID); emailErr != nil {
		// Roll back so a retry isn't blocked by ErrAlreadyRequested with no notification sent.
		if delErr := s.b.AccountDeletion().Delete(ctx, u.ID); delErr != nil {
			log.Error("account deletion: rollback failed after email error; pending row left for manual cleanup",
				zap.String("user_id", u.ID),
				zap.NamedError("email_err", emailErr),
				zap.NamedError("rollback_err", delErr))
			sentry.CaptureException(delErr)
		}
		return nil, toGRPCError(emailErr)
	}

	return &pb.Empty{}, nil
}

func (s *rpcService) GetAccountDeletionStatus(ctx context.Context, _ *pb.Empty) (*pb.AccountDeletionStatus, error) {
	u, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	r, err := s.b.AccountDeletion().GetForUser(ctx, u.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	resp := &pb.AccountDeletionStatus{}
	if r != nil {
		resp.RequestedAt = timestamppb.New(r.RequestedAt)
	}
	return resp, nil
}
