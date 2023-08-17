package grpc

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/payments"

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
	w, err := s.b.Wallets().ForContext(ctx)
	if err == nil {
		walletID = w.ID
	}

	if source == "fynbos" {
		wallet, err := s.b.Wallets().GetFromAddress(ctx, add)
		if err != nil {
			return nil, toGRPCError(err)
		}

		canSendToAddress, err := canSendToWallet(ctx, s.b, walletID, wallet.ID)
		if err != nil {
			return nil, toGRPCError(err)
		}

		return &pb.GetPaymentAddressResponse{
			WalletUrl:        wallet.AddressString(),
			Type:             "wallet",
			Handle:           wallet.AddressShortString(),
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

		canSendToAddress, err := canSendToWallet(ctx, s.b, walletID, id.WalletID)
		if err != nil {
			return nil, toGRPCError(err)
		}

		receiverWallet, err := s.b.Wallets().Get(ctx, id.WalletID)
		if err != nil {
			return nil, toGRPCError(err)
		}

		return &pb.GetPaymentAddressResponse{
			WalletUrl:        receiverWallet.AddressString(),
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

// canSendToWallet returns false if
// 1) sending to own wallet
// 2) wallet doesn't have any linked accounts that can receive
func canSendToWallet(ctx context.Context, b Backends, fromWalletID string, toWalletID string) (bool, error) {

	if toWalletID == fromWalletID {
		return false, nil
	}

	las, err := b.LinkedAccounts().ListByWalletId(ctx, toWalletID)
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

func (s *rpcService) CreatePayment(ctx context.Context, req *pb.CreatePaymentRequest) (*pb.Payment, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	args := payments.CreateArgs{
		Sender: payments.Identity{
			Type:       payments.IdentityTypeWalletID,
			Identifier: w.ID,
		},
		Receiver: payments.Identity{
			Type:       payments.IdentityType(req.ReceiverIdentityType),
			Identifier: req.ReceiverIdentity,
		},
		SenderAmount:    currency.FromPB(req.SenderAmount),
		SenderAccount:   req.GetSenderAccount(),
		ReceiverAmount:  currency.FromPB(req.ReceiverAmount),
		ReceiverAccount: req.GetReceiverAccount(),
		Note:            req.GetNote(),
	}

	p, err := s.b.Payments().Create(ctx, args)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformPayment(p), nil
}

func (s *rpcService) UpdatePayment(ctx context.Context, req *pb.UpdatePaymentRequest) (*pb.Payment, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	_, err = s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	args := payments.UpdateArgs{
		ID:              req.Id,
		SenderAmount:    currency.FromPB(req.GetSenderAmount()),
		SenderAccount:   req.GetSenderAccount(),
		ReceiverAmount:  currency.FromPB(req.GetReceiverAmount()),
		ReceiverAccount: req.GetReceiverAccount(),
		Note:            req.GetNote(),
		ThreeDSID:       req.GetThreeDSID(),
	}

	p, err := s.b.Payments().Update(ctx, args)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformPayment(p), nil
}

func (s *rpcService) GetPayment(ctx context.Context, req *pb.GetPaymentRequest) (*pb.Payment, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	_, err = s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	p, err := s.b.Payments().Lookup(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformPayment(p), nil
}

func transformPayment(p *payments.Payment) *pb.Payment {
	var requiredActions []int32
	for _, ra := range p.RequiredActions {
		requiredActions = append(requiredActions, int32(ra))
	}

	return &pb.Payment{
		Id:                   p.ID,
		PublicID:             p.PublicID,
		State:                int32(p.State),
		SenderIdentity:       p.Sender.Identifier,
		SenderIdentityType:   int32(p.Sender.Type),
		ReceiverIdentity:     p.Receiver.Identifier,
		ReceiverIdentityType: int32(p.Receiver.Type),
		SenderAmount:         p.SenderAmount.ToPB(),
		ReceiverAmount:       p.ReceiverAmount.ToPB(),
		SenderAccount:        p.SenderAccount,
		ReceiverAccount:      p.ReceiverAccount,
		Note:                 p.Note,
		RequiredActions:      requiredActions,
	}
}
