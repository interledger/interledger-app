package grpc

import (
	"context"
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/interledger/interledger-app/go/backend/accountdeletion"
	"github.com/interledger/interledger-app/go/backend/slack"
	"github.com/interledger/interledger-app/go/backend/user"
	"github.com/interledger/interledger-app/go/log"
	pb "github.com/interledger/interledger-app/go/proto/backend/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *rpcService) RequestAccountDeletion(ctx context.Context, _ *pb.RequestAccountDeletionRequest) (*pb.Empty, error) {
	u, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	// Kratos highest_available AAL lets non-enrolled users through on AAL1;
	// this guard ensures destructive deletion always requires TOTP.
	hasTotp, err := s.b.Users().CheckUserTotpEnabled(ctx, u.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if !hasTotp {
		return nil, toGRPCError(user.ErrTotpNotConfigured)
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
		// Email is intentionally omitted: SendToChannel logs the full message at info
		// level, so including PII here would leak it into application logs. Support is
		// already notified by email above.
		slack.SendToChannel(ctx, slack.ChannelSignupKYC, "wallet-info-bot",
			fmt.Sprintf("Account deletion requested\nUserID: %s\nWalletIDs: %v", u.ID, walletIDs))
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
