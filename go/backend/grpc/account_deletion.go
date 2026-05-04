package grpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"gitlab.com/fynbos/backend/accountdeletion"
	"gitlab.com/fynbos/backend/slack"
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

	if emailErr := s.b.Email().SendAccountDeletionRequested(ctx, u.ID); emailErr != nil {
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

	wallets, err := s.b.Wallets().List(ctx, u.ID)
	if err == nil {
		walletIDs := make([]string, 0, len(wallets))
		for _, w := range wallets {
			walletIDs = append(walletIDs, w.ID)
		}
		slack.SendToChannel(ctx, slack.ChannelNotifyEvents, "wallet-info-bot",
			fmt.Sprintf("Account deletion requested\nUserID: %s\nEmail: %s\nWalletIDs: %v", u.ID, strings.TrimSpace(u.Email), walletIDs))
	} else {
		log.Warn("account deletion: skipping Slack notification because wallet lookup failed",
			zap.String("user_id", u.ID),
			zap.Error(err))
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
		resp.Status = toPBAccountDeletionStatus(r.Status)
		resp.CreatedAt = timestamppb.New(r.CreatedAt)
		resp.UpdatedAt = timestamppb.New(r.UpdatedAt)
	}
	return resp, nil
}

func toPBAccountDeletionStatus(s accountdeletion.Status) pb.AccountDeletionRequestStatus {
	switch s {
	case accountdeletion.StatusPending:
		return pb.AccountDeletionRequestStatus_ACCOUNT_DELETION_REQUEST_STATUS_PENDING
	case accountdeletion.StatusInProgress:
		return pb.AccountDeletionRequestStatus_ACCOUNT_DELETION_REQUEST_STATUS_IN_PROGRESS
	case accountdeletion.StatusCompleted:
		return pb.AccountDeletionRequestStatus_ACCOUNT_DELETION_REQUEST_STATUS_COMPLETED
	default:
		return pb.AccountDeletionRequestStatus_ACCOUNT_DELETION_REQUEST_STATUS_UNSPECIFIED
	}
}
