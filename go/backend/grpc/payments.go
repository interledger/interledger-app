package grpc

import (
	"context"
	"fmt"
	"gitlab.com/fynbos/backend/openpayments/ops"
	"gitlab.com/fynbos/backend/paymentpointers"
	pb "gitlab.com/fynbos/proto/backend/v1"
	"net/url"
	"strings"
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

	if source == "fynbos" {
		pp, err := ops.GetPaymentPointer(ctx, s.b, ops.StandardisePaymentPointer(add))
		if err != nil {
			return nil, toGRPCError(err)
		}
		parsedPP, err := paymentpointers.Parse(pp.URL)
		if err != nil {
			return nil, toGRPCError(err)
		}

		return &pb.GetPaymentAddressResponse{
			WalletUrl: pp.URL,
			Type:      "wallet",
			Handle:    parsedPP.ShortString(),
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

		return &pb.GetPaymentAddressResponse{
			WalletUrl: pp.URL,
			Type:      "twitter",
			Handle:    "@" + twitterIdentitifer,
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
