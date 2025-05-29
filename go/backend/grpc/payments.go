package grpc

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"gitlab.com/fynbos/backend/limits"
	"gitlab.com/fynbos/backend/linkedaccounts"

	"gitlab.com/fynbos/backend/twilio"

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

	if source == "interledger" {
		wallet, err := s.b.Wallets().GetFromAddress(ctx, add)
		if err != nil {
			return nil, toGRPCError(err)
		}

		canSendToAddress, err := s.b.LinkedAccounts().CanSendToWallet(ctx, walletID, wallet.ID)
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

		canSendToAddress, err := s.b.LinkedAccounts().CanSendToWallet(ctx, walletID, id.WalletID)
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
		if strings.Contains(parts[0], "ilp.link") {
			return "interledger"
		}
		if strings.Contains(parts[0], "twitter.com") {
			return "twitter"
		}
		return hostParts[0]
	}

	if strings.Contains(u.Hostname(), "ilp.link") {
		return "interledger"
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

func (s *rpcService) CreatePayment(ctx context.Context, req *pb.CreatePaymentRequest) (*pb.Payment, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	// check that does not exceed kyc limits.
	exceedsLimits, limitType, err := s.b.Limits().ExceedsKYCLimits(ctx, w.ID, currency.FromPB(req.SenderAmount))
	if err != nil {
		return nil, toGRPCError(err)
	}
	if exceedsLimits {
		var description string
		switch limitType {
		case limits.LimitTypeTransaction:
			description = "Exceeds per transaction limit."
		case limits.LimitTypeDaily:
			description = "Exceeds daily limit."
		case limits.LimitTypeMonthly:
			description = "Exceeds monthly limit."
		case limits.LimitType6Monthly:
			description = "Exceeds 6 monthly limit."
		case limits.LimitTypeYearly:
			description = "Exceeds yearly limit."
		default:
			description = "Exceeds account limit."
		}
		return nil, NewValidationError("amount", description)
	}

	senderAccount := req.GetSenderAccount()
	sendAmount := currency.FromPB(req.SenderAmount)
	args := payments.CreateArgs{
		Sender: payments.Identity{
			Type:       payments.IdentityTypeWalletID,
			Identifier: w.ID,
		},
		Receiver: payments.Identity{
			Type:       payments.IdentityType(req.ReceiverIdentityType),
			Identifier: req.ReceiverIdentity,
		},
		SenderAmount:    sendAmount,
		SenderAccount:   senderAccount,
		ReceiverAmount:  currency.FromPB(req.GetReceiverAmount()),
		ReceiverAccount: req.GetReceiverAccount(),
		Note:            req.GetNote(),
		IPAddress:       req.GetIpAddress(),
	}

	p, err := s.b.Payments().Create(ctx, args)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformPayment(ctx, s.b, p)
}

func (s *rpcService) UpdatePayment(ctx context.Context, req *pb.UpdatePaymentRequest) (*pb.Payment, error) {
	u, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}
	p, err := s.b.Payments().Lookup(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if p.Sender.WalletID != w.ID {
		return nil, NotFoundError("payment not found")
	}

	// check that does not exceed kyc limits.
	if req.SenderAmount != nil {
		exceedsLimits, limitType, err := s.b.Limits().ExceedsKYCLimits(ctx, w.ID, currency.FromPB(req.GetSenderAmount()))
		if err != nil {
			return nil, toGRPCError(err)
		}
		if exceedsLimits {
			var description string
			switch limitType {
			case limits.LimitTypeTransaction:
				description = "Exceeds per transaction limit."
			case limits.LimitTypeDaily:
				description = "Exceeds daily limit."
			case limits.LimitTypeMonthly:
				description = "Exceeds monthly limit."
			case limits.LimitType6Monthly:
				description = "Exceeds 6 monthly limit."
			case limits.LimitTypeYearly:
				description = "Exceeds yearly limit."
			default:
				description = "Exceeds account limit."
			}
			return nil, NewValidationError("amount", description)
		}
	}

	if req.GetOtp() != "" {
		vc, err := s.b.Twilio().CheckVerificationCode(ctx, &twilio.CheckVerificationCodeArgs{
			PhoneNumber: u.PhoneNumber,
			Code:        req.GetOtp(),
		})
		if err != nil {
			return nil, toGRPCError(err)
		}
		if !vc.IsValid() {
			return nil, NewValidationError("otp", "Invalid OTP")
		}
	}

	args := payments.UpdateArgs{
		ID:             req.Id,
		SenderAmount:   currency.FromPB(req.GetSenderAmount()),
		SenderAccount:  req.GetSenderAccount(),
		ReceiverAmount: currency.FromPB(req.GetReceiverAmount()),
		Receiver: payments.Identity{
			Type:       payments.IdentityType(req.GetReceiverIdentityType()),
			Identifier: req.GetReceiverIdentity(),
		},
		ReceiverAccount: req.GetReceiverAccount(),
		Note:            req.GetNote(),
		ThreeDSID:       req.GetThreeDSID(),
		OTP:             req.GetOtp(),
		IPAddress:       req.GetIpAddress(),
	}

	p, err = s.b.Payments().Update(ctx, args)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformPayment(ctx, s.b, p)
}

func (s *rpcService) GetPayment(ctx context.Context, req *pb.GetPaymentRequest) (*pb.Payment, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	p, err := s.b.Payments().Lookup(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if p.Sender.WalletID != w.ID {
		return nil, NotFoundError("payment not found")
	}

	return transformPayment(ctx, s.b, p)
}

func (s *rpcService) ConfirmPayment(ctx context.Context, req *pb.ConfirmPaymentRequest) (*pb.Payment, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	p, err := s.b.Payments().Lookup(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if p.Sender.WalletID != w.ID {
		return nil, NotFoundError("payment not found")
	}

	// check that does not exceed kyc limits.
	exceedsLimits, limitType, err := s.b.Limits().ExceedsKYCLimits(ctx, w.ID, p.SenderAmount)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if exceedsLimits {
		var description string
		switch limitType {
		case limits.LimitTypeTransaction:
			description = "Exceeds per transaction limit."
		case limits.LimitTypeDaily:
			description = "Exceeds daily limit."
		case limits.LimitTypeMonthly:
			description = "Exceeds monthly limit."
		case limits.LimitType6Monthly:
			description = "Exceeds 6 monthly limit."
		case limits.LimitTypeYearly:
			description = "Exceeds yearly limit."
		default:
			description = "Exceeds account limit."
		}
		return nil, NewValidationError("amount", description)
	}

	p, requiredActions, err := s.b.Payments().Confirm(ctx, req.Id)
	if errors.Is(err, payments.ErrRequiredActions) {
		return nil, PaymentPreconditionError(requiredActions)
	}
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformPayment(ctx, s.b, p)
}

func transformPayment(ctx context.Context, b Backends, p *payments.Payment) (*pb.Payment, error) {
	var requiredActions []int32
	for _, ra := range p.RequiredActions {
		requiredActions = append(requiredActions, int32(ra))
	}

	var receiveWalletAddress string
	if p.Receiver.WalletID != "" {
		receiveWallet, err := b.Wallets().Get(ctx, p.Receiver.WalletID)
		if err != nil {
			return nil, toGRPCError(err)
		}
		receiveWalletAddress = receiveWallet.AddressString()
	}

	// hard-coded to be 0 for now
	fees := currency.FromUInt64(0, p.SenderAmount.Currency)
	ret := &pb.Payment{
		Id:                   p.ID,
		PublicID:             p.PublicID,
		State:                int32(p.State),
		ReceiverWalletUrl:    receiveWalletAddress,
		ReceiverIdentity:     p.Receiver.Identifier,
		ReceiverIdentityType: int32(p.Receiver.Type),
		SenderAmount:         p.SenderAmount.ToPB(),
		SenderAccount:        p.SenderAccount,
		TotalSendAmount:      p.SenderAmount.Format(),
		Note:                 p.Note,
		RequiredActions:      requiredActions,
		FxRate:               fmt.Sprintf("%6f", p.FXRate),
		ReceiverAmount:       p.ReceiverAmount.ToPB(),
		FormattedFees:        fees.Format(),
	}

	if p.Type == payments.TypeWithdrawal {
		ret.ReceiverAccount = p.ReceiverAccount
	}

	return ret, nil
}

func (s *rpcService) GetLinkedAccountsForPayment(ctx context.Context, req *pb.GetLinkedAccountsForPaymentRequest) (*pb.GetLinkedAccountsForPaymentResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	p, err := s.b.Payments().Lookup(ctx, req.GetPaymentId())
	if err != nil || p.Sender.WalletID != w.ID {
		return nil, NotFoundError("")
	}

	sendAccounts, err := s.b.LinkedAccounts().ListByWalletId(ctx, w.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	recvAccounts, err := s.b.LinkedAccounts().ListByWalletId(ctx, p.Receiver.WalletID)
	if err != nil && !errors.Is(err, linkedaccounts.ErrNotFound) {
		return nil, toGRPCError(err)
	}

	var las []*pb.LinkedAccountForPayment
	for _, sendAcc := range sendAccounts {
		if sendAcc.DeletedAt.Valid {
			continue
		}

		acc := &pb.LinkedAccountForPayment{
			Details: transformLinkedAccount(sendAcc),
			Enabled: recvAccounts == nil, // could be paying to non-signed up user
		}
		for _, recvAcc := range recvAccounts {
			if sendAcc.CanPay(recvAcc) {
				acc.Enabled = true
				break
			}
		}

		las = append(las, acc)
	}

	return &pb.GetLinkedAccountsForPaymentResponse{
		LinkedAccounts: las,
	}, nil
}
