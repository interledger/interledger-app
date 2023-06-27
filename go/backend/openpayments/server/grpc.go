package server

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gitlab.com/fynbos/backend/twilio"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
	"gitlab.com/fynbos/backend/openpayments/workflows"
	pb "gitlab.com/fynbos/proto/backend/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ pb.OpenPaymentServiceServer = &grpcServer{}

type grpcServer struct {
	b Backends
}

func NewGRPCServer(b Backends) pb.OpenPaymentServiceServer {
	return &grpcServer{b: b}
}

func (g *grpcServer) CreatePaymentPointer(ctx context.Context, req *pb.CreatePaymentPointerRequest) (*pb.Empty, error) {
	_, err := g.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("no login found")
	}

	wallet, err := g.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	pp := openpayments.PaymentPointer{
		URL:        ops.StandardisePaymentPointer(req.Url),
		WalletID:   wallet.ID,
		Alias:      req.GetAlias(),
		Asset:      currency.ParseCurrency(req.GetAsset()),
		AssetScale: int(req.GetAssetScale()),
	}

	err = g.b.Validator().Struct(pp)
	if err != nil {
		return nil, toGRPCError(err)
	}

	err = ops.CreatePaymentPointer(ctx, g.b, pp)
	if errors.Is(err, openpayments.ErrInvalidPointerPath) {
		return nil, NewValidationError("url", strings.TrimSpace(strings.TrimPrefix(err.Error(), openpayments.ErrInvalidPointerPath.Error())))
	}
	if err != nil {
		return nil, toGRPCError(err)
	}

	_, err = g.b.Wallets().SetWalletName(ctx, wallet.ID, req.GetAlias())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}

func (g *grpcServer) GetPaymentPointer(ctx context.Context, req *pb.GetPaymentPointerRequest) (*pb.PaymentPointer, error) {
	pp, err := ops.GetPaymentPointer(ctx, g.b, ops.StandardisePaymentPointer(req.Url))
	if err != nil {
		return nil, toGRPCError(err)
	}

	formattedPP, err := ops.FormattedPaymentPointer(pp.URL)
	if err != nil {
		return nil, toGRPCError(err)
	}

	legalName := ""
	kycInfo, err := g.b.KYC().GetIndividualDetails(ctx, pp.WalletID)
	if err == nil {
		legalName = kycInfo.FirstName + " " + kycInfo.LastName
	}

	return &pb.PaymentPointer{
		Url:        pp.URL,
		Asset:      pp.Asset.String(),
		AssetScale: int32(pp.AssetScale),
		Alias:      pp.Alias,
		WalletID:   pp.WalletID,
		Formatted:  formattedPP,
		LegalName:  legalName,
	}, nil
}

func (g *grpcServer) ListWalletPaymentPointers(ctx context.Context, _ *pb.Empty) (*pb.ListWalletPaymentPointersResponse, error) {
	_, err := g.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("no login found")
	}

	wallet, err := g.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	pp, err := ops.ListWalletPaymentPointers(ctx, g.b, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	legalName := ""
	kycInfo, err := g.b.KYC().GetIndividualDetails(ctx, wallet.ID)
	if err == nil {
		legalName = kycInfo.FirstName + " " + kycInfo.LastName
	}

	resp := make([]*pb.PaymentPointer, len(pp))
	for i, p := range pp {
		formattedPP, err := ops.FormattedPaymentPointer(p.URL)
		if err != nil {
			return nil, toGRPCError(err)
		}
		resp[i] = &pb.PaymentPointer{
			Url:        p.URL,
			Asset:      p.Asset.String(),
			AssetScale: int32(p.AssetScale),
			Alias:      p.Alias,
			WalletID:   p.WalletID,
			Formatted:  formattedPP,
			LegalName:  legalName,
		}
	}

	return &pb.ListWalletPaymentPointersResponse{Pointers: resp}, nil
}

func (g *grpcServer) PaymentPointerExists(ctx context.Context, req *pb.PaymentPointerExistsRequest) (*pb.PaymentPointerExistsResponse, error) {
	exists, err := ops.PaymentPointerExists(ctx, g.b, ops.StandardisePaymentPointer(req.Url))
	if errors.Is(err, openpayments.ErrInvalidPointerPath) {
		return nil, NewValidationError("url", strings.TrimSpace(strings.TrimPrefix(err.Error(), openpayments.ErrInvalidPointerPath.Error())))
	}
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.PaymentPointerExistsResponse{
		Exists: exists,
	}, nil
}

func (g *grpcServer) CreateQuote(ctx context.Context, req *pb.CreateQuoteRequest) (*pb.Quote, error) {

	_, err := g.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("no login found")
	}

	wallet, err := g.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	ppl, err := ops.ListWalletPaymentPointers(ctx, g.b, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	amount := currency.FromPB(req.Amount)
	exceedsLimit, limitType, err := g.b.Limits().ExceedsKYCLimits(ctx, wallet.ID, amount)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if exceedsLimit {
		return nil, FailedPreconditionError(string(limitType))
	}

	args := openpayments.CreateQuoteArgs{
		SendPaymentPointer:      ops.StandardisePaymentPointer(req.SendPaymentPointer),
		ReceivePaymentPointer:   ops.StandardisePaymentPointer(req.ReceivePaymentPointer),
		ExpiresAt:               req.ExpiresAt.AsTime(),
		SendAmount:              amount,
		Reference:               "",
		Description:             req.Description,
		LinkedAccID:             req.GetSendLinkedAccount(),
		CreatedBy:               ops.StandardisePaymentPointer(req.SendPaymentPointer),
		DestinationIdentity:     req.GetIdentity(),
		DestinationIdentityType: req.GetIdentityType(),
	}

	var found bool
	for _, pp := range ppl {
		if strings.EqualFold(pp.URL, args.SendPaymentPointer) {
			found = true
		}
	}
	if !found {
		// Signed in user doesn't own the payment pointer it's trying to send from
		return nil, ForbiddenError("Unauthenticated")
	}

	err = g.b.Validator().Struct(args)
	if err != nil {
		return nil, toGRPCError(err)
	}

	q, err := ops.CreateQuote(ctx, g.b, args)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Quote{
		Id:             q.ID,
		PaymentPointer: q.PaymentPointer,
		Receiver:       q.IncomingPayment,
		SendAmount:     q.SendAmount.ToPB(),
		ReceiveAmount:  q.ReceiveAmount.ToPB(),
		ExpiresAt:      timestamppb.New(q.ExpiresAt),
		CreatedAt:      timestamppb.New(q.CreatedAt),
		RequiresOTP:    q.RequiresOTP,
		OtpComplete:    q.OTPValidated,
	}, nil
}

func (g *grpcServer) LookupQuote(ctx context.Context, req *pb.LookupQuoteRequest) (*pb.Quote, error) {
	_, err := g.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("no login found")
	}

	wallet, err := g.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	q, err := ops.GetWalletQuote(ctx, g.b, wallet.ID, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Quote{
		Id:             q.ID,
		PaymentPointer: q.PaymentPointer,
		Receiver:       q.IncomingPayment,
		SendAmount:     q.SendAmount.ToPB(),
		ReceiveAmount:  q.ReceiveAmount.ToPB(),
		ExpiresAt:      timestamppb.New(q.ExpiresAt),
		CreatedAt:      timestamppb.New(q.CreatedAt),
		RequiresOTP:    q.RequiresOTP,
		OtpComplete:    q.OTPValidated,
	}, nil
}

func (g *grpcServer) SetQuoteOTP(ctx context.Context, req *pb.SetQuoteOTPRequest) (*pb.Quote, error) {
	u, err := g.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("no login found")
	}

	_, err = g.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	vc, err := g.b.Twilio().CheckVerificationCode(ctx, &twilio.CheckVerificationCodeArgs{
		PhoneNumber: u.PhoneNumber,
		Code:        req.Otp,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	if !vc.IsValid() {
		return nil, NewValidationError("otp", "Invalid OTP")
	}

	q, err := ops.SetQuoteOTPValidated(ctx, g.b, req.QuoteID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Quote{
		Id:             q.ID,
		PaymentPointer: q.PaymentPointer,
		Receiver:       q.IncomingPayment,
		SendAmount:     q.SendAmount.ToPB(),
		ReceiveAmount:  q.ReceiveAmount.ToPB(),
		ExpiresAt:      timestamppb.New(q.ExpiresAt),
		CreatedAt:      timestamppb.New(q.CreatedAt),
		RequiresOTP:    q.RequiresOTP,
		OtpComplete:    q.OTPValidated,
	}, nil
}

func (g *grpcServer) SendQuoteOTP(ctx context.Context, req *pb.SendQuoteOTPRequest) (*pb.Empty, error) {
	_, err := g.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("no login found")
	}

	_, err = g.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	err = ops.SendQuoteOTP(ctx, g.b, req.QuoteID)

	return &pb.Empty{}, toGRPCError(err)
}

func (g *grpcServer) CreateIncomingPayment(ctx context.Context, req *pb.CreateIncomingPaymentRequest) (*pb.IncomingPayment, error) {

	_, err := g.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("no login found")
	}

	wallet, err := g.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	ppl, err := ops.ListWalletPaymentPointers(ctx, g.b, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	args := openpayments.CreateIncomingPaymentArgs{
		PaymentPointer: ops.StandardisePaymentPointer(req.PaymentPointer),
		ExternalRef:    req.Reference,
		ExpiresAt:      req.ExpiresAt.AsTime(),
		CreatedBy:      ops.StandardisePaymentPointer(req.PaymentPointer),
	}
	if req.Amount != nil {
		amt := currency.FromPB(req.GetAmount())
		args.IncomingAmount = &amt
	}

	// TODO: Can you only create incoming payments for yourself...
	// Probably not long term, but for now :shrug:
	var found bool
	for _, pp := range ppl {
		if strings.EqualFold(pp.URL, args.PaymentPointer) {
			found = true
		}
	}
	if !found {
		// Signed in user doesn't own the payment pointer it's trying to send from
		return nil, ForbiddenError("Unauthenticated")
	}

	err = g.b.Validator().Struct(args)
	if err != nil {
		return nil, toGRPCError(err)
	}

	ip, err := ops.CreateIncomingPayment(ctx, g.b, args)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.IncomingPayment{
		Id:                 ip.ID,
		PaymentPointer:     ip.PaymentPointer,
		FromPaymentPointer: ip.FromPaymentPointer,
		IncomingAmount:     ip.IncomingAmount.ToPB(),
		ReceivedAmount:     ip.ReceivedAmount.ToPB(),
		Completed:          ip.Completed,
		ExternalRef:        ip.ExternalRef,
		ExpiresAt:          timestamppb.New(ip.ExpiresAt),
		CreatedAt:          timestamppb.New(ip.CreatedAt),
		UpdatedAt:          timestamppb.New(ip.UpdatedAt),
	}, nil
}

func (g *grpcServer) LookupIncomingPayment(ctx context.Context, req *pb.LookupIncomingPaymentRequest) (*pb.IncomingPayment, error) {
	_, err := g.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("no login found")
	}

	_, err = g.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	ip, err := ops.GetIncomingPayment(ctx, g.b, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.IncomingPayment{
		Id:                 ip.ID,
		PaymentPointer:     ip.PaymentPointer,
		FromPaymentPointer: ip.FromPaymentPointer,
		IncomingAmount:     ip.IncomingAmount.ToPB(),
		ReceivedAmount:     ip.ReceivedAmount.ToPB(),
		Completed:          ip.Completed,
		ExternalRef:        ip.ExternalRef,
		ExpiresAt:          timestamppb.New(ip.ExpiresAt),
		CreatedAt:          timestamppb.New(ip.CreatedAt),
		UpdatedAt:          timestamppb.New(ip.UpdatedAt),
	}, nil
}

func (g *grpcServer) PreCheckOutgoingPayment(ctx context.Context, req *pb.PreCheckOutgoingPaymentRequest) (*pb.PreCheckOutgoingPaymentResponse, error) {
	_, err := g.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("no login found")
	}

	wallet, err := g.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	q, err := ops.GetWalletQuote(ctx, g.b, wallet.ID, req.QuoteID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	// check that does not exceed kyc limits.
	exceedsLimits, limitType, err := g.b.Limits().ExceedsKYCLimits(ctx, wallet.ID, q.SendAmount)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.PreCheckOutgoingPaymentResponse{
		ExceedsLimits:       exceedsLimits,
		LimitType:           string(limitType),
		InsufficientBalance: false,
	}, nil
}

func (g *grpcServer) CreateOutgoingPayment(ctx context.Context, req *pb.CreateOutgoingPaymentRequest) (*pb.OutgoingPayment, error) {

	_, err := g.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("no login found")
	}

	wallet, err := g.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	args := openpayments.CreateOutgoingPaymentArgs{
		QuoteID:   req.QuoteID,
		IPAddress: req.IpAddress,
		ThreeDSID: req.GetThreeDSID(),
	}

	err = g.b.Validator().Struct(args)
	if err != nil {
		return nil, toGRPCError(err)
	}

	q, err := ops.GetWalletQuote(ctx, g.b, wallet.ID, args.QuoteID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if q.RequiresOTP && !q.OTPValidated {
		return nil, FailedPreconditionError("OTP for quote has not been validated")
	}

	if args.ThreeDSID != "" {
		session3DS, err := g.b.Tabapay().Get3DSSession(ctx, args.ThreeDSID)
		if err != nil {
			return nil, toGRPCError(err)
		}

		orderID := q.ID
		idxSlash := strings.LastIndex(q.ID, "/")
		if idxSlash > 0 {
			orderID = orderID[idxSlash+1:]
		}
		if session3DS.OrderID != orderID {
			err = fmt.Errorf("%w 3DS session invalid.", openpayments.ErrInternal)
			return nil, toGRPCError(err)
		}
	}

	// Assume that the quote and the outgoing payment have the same created by.
	args.CreatedBy = q.CreatedBy

	op, err := workflows.StartOutgoingPayment(ctx, g.b, args)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.OutgoingPayment{
		Id:               op.ID,
		PaymentPointer:   op.PaymentPointer,
		ToPaymentPointer: op.ToPaymentPointer,
		Failed:           op.Failed,
		Receiver:         op.Receiver,
		SendAmount:       op.SendAmount.ToPB(),
		ReceiveAmount:    op.ReceiveAmount.ToPB(),
		SentAmount:       op.SentAmount.ToPB(),
		Description:      op.Description,
		CreatedAt:        timestamppb.New(op.CreatedAt),
		UpdatedAt:        timestamppb.New(op.UpdatedAt),
	}, nil
}

func (g *grpcServer) LookupOutgoingPayment(ctx context.Context, req *pb.LookupOutgoingPaymentRequest) (*pb.OutgoingPayment, error) {
	op, err := ops.GetOutgoingPayment(ctx, g.b, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.OutgoingPayment{
		Id:               op.ID,
		PaymentPointer:   op.PaymentPointer,
		ToPaymentPointer: op.ToPaymentPointer,
		Failed:           op.Failed,
		Receiver:         op.Receiver,
		SendAmount:       op.SendAmount.ToPB(),
		ReceiveAmount:    op.ReceiveAmount.ToPB(),
		SentAmount:       op.SentAmount.ToPB(),
		Description:      op.Description,
		CreatedAt:        timestamppb.New(op.CreatedAt),
		UpdatedAt:        timestamppb.New(op.UpdatedAt),
	}, nil
}

func (g *grpcServer) CanSendToPaymentPointer(ctx context.Context, req *pb.CanSendToPaymentPointerRequest) (*pb.CanSendToPaymentPointerResponse, error) {
	var walletID string

	wallet, err := g.b.Wallets().ForContext(ctx)
	if err == nil {
		// Ignore the error. If the request is unauthenticated we just don't check the users PP
		walletID = wallet.ID
	}

	canSend, err := ops.ValidateCanSend(ctx, g.b, walletID, req.PaymentPointer)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.CanSendToPaymentPointerResponse{CanSend: canSend}, nil
}
