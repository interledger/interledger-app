package grpc

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
	"gitlab.com/fynbos/backend/paymentpointers"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

/*
GetPaymentAddress
This function will handle the following cases
https://fynbos.dev/matt , fynbos.dev/matt (return fynbos)
@matbuddhabumb, https://twitter.com/matbuddhabum (returns twitter)
*/
func (s *rpcService) GetPaymentAddress(ctx context.Context, req *pb.GetPaymentAddressRequest) (*pb.GetPaymentAddressResponse, error) {
	add := req.GetAddress()

	source := identifySource(add)

	var walletID string
	w, err := s.b.Users().WalletForContext(ctx)
	if err == nil {
		walletID = w.ID
	}

	if source == "fynbos" {
		pp, err := ops.GetPaymentPointer(ctx, s.b, ops.StandardisePaymentPointer(add))
		if err != nil {
			return nil, toGRPCError(err)
		}
		parsedPP, err := paymentpointers.Parse(pp.URL)
		if err != nil {
			return nil, toGRPCError(err)
		}

		canSendToAddress, err := canSendToPaymentPointer(ctx, s.b, walletID, pp)
		if err != nil {
			return nil, toGRPCError(err)
		}

		return &pb.GetPaymentAddressResponse{
			WalletUrl:        pp.URL,
			Type:             "wallet",
			Handle:           parsedPP.ShortString(),
			CanSendToAddress: canSendToAddress,
		}, nil
	}

	if source == "twitter" {
		twitterIdentitifer, err := getTwitterHandle(add)
		if err != nil {
			return nil, toGRPCError(err)
		}
		id, err := s.b.Identities().GetByIdentifier(ctx, twitterIdentitifer)
		if err != nil {
			return nil, toGRPCError(err)
		}

		pp, err := s.b.OpenPayments().GetWalletPaymentPointer(ctx, id.WalletID)
		if err != nil {
			return nil, toGRPCError(err)
		}

		canSendToAddress, err := canSendToPaymentPointer(ctx, s.b, walletID, pp)
		if err != nil {
			return nil, toGRPCError(err)
		}

		return &pb.GetPaymentAddressResponse{
			WalletUrl:        pp.URL,
			Type:             "twitter",
			Handle:           "@" + twitterIdentitifer,
			CanSendToAddress: canSendToAddress,
		}, nil
	}

	return nil, NotFoundError("address not found")
}

func identifySource(input string) string {
	u, err := url.ParseRequestURI(input)
	if err != nil {
		if strings.HasPrefix(input, "@") {
			return "twitter"
		}
		parts := strings.Split(input, "/")
		hostParts := strings.Split(parts[0], ".")
		if strings.Contains(parts[0], "fynbos.me") {
			return "fynbos"
		}
		if strings.Contains(parts[0], "twitter.com") {
			return "twitter"
		}
		return hostParts[0]
	}

	if strings.Contains(u.Hostname(), "fynbos.me") {
		return "fynbos"
	}
	if strings.Contains(u.Hostname(), "twitter.com") {
		return "twitter"
	}

	return "unknown"
}

func getTwitterHandle(input string) (string, error) {
	if strings.HasPrefix(input, "@") {
		return strings.TrimPrefix(input, "@"), nil
	}

	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		input = "https://" + input
	}

	u, err := url.ParseRequestURI(input)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	pathParts := strings.Split(u.Path, "/")
	if len(pathParts) < 2 {
		return "", fmt.Errorf("no twitter handle found in URL")
	}

	return pathParts[1], nil
}

// Returns false if
// 1) sending to own wallet
// 2) payment pointer doesn't have any linked accounts that can receive
func canSendToPaymentPointer(ctx context.Context, b Backends, fromWalletID string, pp *openpayments.PaymentPointer) (bool, error) {
	if pp == nil {
		return false, nil
	}

	if pp.WalletID == fromWalletID {
		return false, nil
	}

	las, err := b.LinkedAccounts().ListByWalletId(ctx, pp.WalletID)
	if err != nil {
		return false, err
	}

	var ppCanReceive bool
	for _, la := range las {
		if la.CanReceive {
			ppCanReceive = true
			break
		}
	}
	if !ppCanReceive {
		return false, nil
	}

	return true, nil
}
