package grpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/rafiki"
	rafikiops "gitlab.com/fynbos/backend/rafiki/ops"
	pb "gitlab.com/fynbos/proto/backend/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *rpcService) CreateIncomingPaymentRequest(
	ctx context.Context,
	req *pb.CreateIncomingPaymentRequestInput,
) (*pb.IncomingPaymentRequest, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	if req.GetValue() == 0 {
		return nil, InvalidArgumentError("value must be greater than zero")
	}
	expiresAt := req.GetExpiresAt().AsTime()
	if expiresAt.Before(time.Now()) {
		return nil, InvalidArgumentError("expiresAt must be in the future")
	}

	// Safety net: ensure the wallet has a Rafiki payment pointer before
	// GetWalletAddress tries to look one up. CreatePaymentPointer is
	// idempotent (no-op when rafiki_payment_pointers already has a row),
	// so this only does work for wallets that somehow lack a pp. The
	// KYC gate still runs inside GetWalletAddress below.
	if _, err := s.b.Rafiki().CreatePaymentPointer(ctx, *wallet); err != nil {
		return nil, toGRPCError(err)
	}

	walletAddr, err := s.b.Rafiki().GetWalletAddress(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	// Non-Gatehub wallets would never settle — see isGatehubIncomingWebhook.
	if _, err := rafikiops.GetGatehubBalanceAccount(ctx, s.b, wallet.ID); err != nil {
		return nil, toGRPCError(err)
	}

	description := strings.TrimSpace(req.GetDescription())
	metadata := map[string]interface{}{}
	if description != "" {
		metadata["description"] = description
	}

	amount := currency.Amount{
		Value:    int64(req.GetValue()),
		Currency: currency.ParseCurrency(walletAddr.AssetCode),
		Scale:    int(walletAddr.AssetScale),
	}

	ip, err := s.b.Rafiki().CreateIncomingPayment(ctx, rafiki.CreateIncomingPaymentArgs{
		WalletAddressID: walletAddr.ID,
		IncomingAmount:  amount,
		ExpiresAt:       expiresAt,
		Metadata:        metadata,
		IdempotencyKey:  uuid.New().String(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformIncomingPaymentRequest(ip, walletAddr), nil
}

func transformIncomingPaymentRequest(
	ip *rafiki.IncomingPayment,
	walletAddr *rafiki.WalletAddress,
) *pb.IncomingPaymentRequest {
	out := &pb.IncomingPaymentRequest{
		Id:          ip.ID,
		Url:         incomingPaymentURL(walletAddr.URL, ip.ID),
		Description: metadataDescription(ip.Metadata),
		State:       string(ip.State),
	}
	if ip.IncomingAmount != nil {
		out.IncomingAmount = ip.IncomingAmount.ToPB()
	}
	if expiresAt, ok := parseRafikiTime(ip.ExpiresAt); ok {
		out.ExpiresAt = timestamppb.New(expiresAt)
	}
	if createdAt, ok := parseRafikiTime(ip.CreatedAt); ok {
		out.CreatedAt = timestamppb.New(createdAt)
	}
	return out
}

func metadataDescription(md map[string]interface{}) string {
	if md == nil {
		return ""
	}
	v, ok := md["description"].(string)
	if !ok {
		return ""
	}
	return v
}

func incomingPaymentURL(walletAddrURL, id string) string {
	if walletAddrURL == "" {
		return id
	}
	return fmt.Sprintf("%s/incoming-payments/%s", strings.TrimRight(walletAddrURL, "/"), id)
}

// Returns (_, false) so callers can leave the proto timestamp unset.
func parseRafikiTime(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}
