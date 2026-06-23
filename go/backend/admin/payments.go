package admin

import (
	"context"

	pb "github.com/interledger/interledger-app/go/proto/backend/admin/v1"
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
