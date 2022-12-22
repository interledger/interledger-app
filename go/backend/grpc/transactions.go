package grpc

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"gitlab.com/fynbos/backend/db"

	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) ListTransactions(ctx context.Context, req *pb.PaginationRequest) (*pb.ListTransactionsResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	page := db.PaginationFromPB(req)

	txs, err := s.b.Transactions().ListTransactions(ctx, page, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	res := make([]*pb.Transaction, len(txs))
	for i, tx := range txs {
		trs := make([]*pb.Transfer, len(tx.Transfers))
		for y, tr := range tx.Transfers {
			trs[y] = &pb.Transfer{
				ForeignId:       tr.ForeignID,
				LinkedAccountId: tr.LinkedAccountID,
				Type:            string(tr.Type),
				State:           string(tr.State),
				Timestamp:       timestamppb.New(tr.Timestamp),
				Amount: &pb.Amount{
					Amount:     tr.Amount.Value,
					Asset:      tr.Amount.Asset,
					AssetScale: int32(tr.Amount.AssetScale),
				},
			}
		}

		res[i] = &pb.Transaction{
			ForeignId: tx.ForeignID,
			Type:      string(tx.Type),
			Amount: &pb.Amount{
				Amount:     tx.Amount.Value,
				Asset:      tx.Amount.Asset,
				AssetScale: int32(tx.Amount.AssetScale),
			},
			Source:      tx.Source,
			Destination: tx.Destination,
			Timestamp:   timestamppb.New(tx.Timestamp),
			State:       string(tx.State),
			Transfers:   trs,
		}
	}

	return &pb.ListTransactionsResponse{
		Transactions: res,
		Page:         page.ToPB(len(res)),
	}, nil
}
