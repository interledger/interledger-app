package admin

import (
	"context"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/env"
	pb "gitlab.com/fynbos/proto/backend/admin/v1"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *AdminRpcService) ListPaymentsAwaitingSignal(ctx context.Context, _ *emptypb.Empty) (*pb.ListPaymentsAwaitingSignalResponse, error) {
	pl, err := s.b.Payments().AdminListAwaitingSignal(ctx)
	if err != nil {
		return nil, toGRPCError(err)
	}

	resp := make([]*pb.Payment, len(pl))

	for i, p := range pl {
		var requiredActions []string
		for _, ra := range p.RequiredActions {
			requiredActions = append(requiredActions, ra.String())
		}

		var receiveWalletAddress string
		if p.Receiver.WalletID != "" {
			receiveWallet, err := s.b.Wallets().Get(ctx, p.Receiver.WalletID)
			if err != nil {
				return nil, toGRPCError(err)
			}
			receiveWalletAddress = receiveWallet.AddressString()
		}

		senderWallet, err := s.b.Wallets().Get(ctx, p.Sender.WalletID)
		if err != nil {
			return nil, toGRPCError(err)
		}

		resp[i] = &pb.Payment{
			Id:                   p.ID,
			PublicID:             p.PublicID,
			State:                p.State.String(),
			ReceiverWalletUrl:    receiveWalletAddress,
			ReceiverIdentity:     p.Receiver.Identifier,
			ReceiverIdentityType: p.Receiver.Type.String(),
			SenderWalletUrl:      senderWallet.AddressString(),
			SenderAmount:         p.SenderAmount.Format(),
			SenderAccount:        p.SenderAccount,
			Note:                 p.Note,
			RequiredActions:      requiredActions,
			UpdatedAt:            timestamppb.New(p.UpdatedAt),
		}
	}

	return &pb.ListPaymentsAwaitingSignalResponse{Payments: resp}, nil
}

func (s *AdminRpcService) SeedPayments(ctx context.Context, req *pb.SeedPaymentsRequest) (*pb.Empty, error) {
	if env.IsProd() {
		return nil, UnimplementedError("")
	}

	var seedPs []payments.SeedPayment
	for _, p := range req.Payments {
		seedPs = append(seedPs, payments.SeedPayment{
			ID:                      p.Id,
			PublicID:                p.PublicId,
			State:                   payments.State(p.State),
			ThreeDSRequired:         p.ThreeDsRequired,
			ThreeDSID:               p.ThreeDsId,
			SenderID:                p.SenderId,
			SenderIDType:            payments.IdentityType(p.SenderIdType),
			SenderAmount:            p.SenderAmount,
			SenderCurrency:          p.SenderCurrency,
			SenderAccount:           p.SenderAccount,
			ReceiverID:              p.ReceiverId,
			ReceiverIDType:          payments.IdentityType(p.ReceiverIdType),
			ReceiverAmount:          p.ReceiverAmount,
			ReceiverCurrency:        p.ReceiverCurrency,
			ReceiverAccount:         p.ReceiverAccount,
			SendTransactionID:       p.SendTransactionId,
			ReceiveTransactionID:    p.ReceiveTransactionId,
			Note:                    p.Note,
			OTPRequired:             p.OtpRequired,
			OTP:                     p.Otp,
			IPAddress:               p.IpAddress,
			Type:                    payments.Type(p.Type),
			FXRate:                  p.FxRate,
			FXFeePercentage:         p.FxFeePercentage,
			ProtectionFeePercentage: p.ProtectionFeePercentage,
		})
	}

	_, err := s.b.AdminPayments().Seed(ctx, seedPs)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}
