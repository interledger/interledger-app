package admin

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/interledger/interledger-app/go/backend/db"

	pb "github.com/interledger/interledger-app/go/proto/backend/admin/v1"
)

func (s *AdminRpcService) ListTransactions(ctx context.Context, req *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
	txs, err := s.b.Transactions().List(ctx, db.FromAdminPB(req.Page), req.WalletID)
	if err != nil {
		return nil, err
	}

	resp := make([]*pb.Transaction, len(txs))
	var token string
	for i, tx := range txs {
		resp[i] = &pb.Transaction{
			WalletID:    req.WalletID,
			Id:          tx.ID,
			Type:        string(tx.Type),
			Asset:       tx.Amount.Currency.String(),
			Amount:      tx.Amount.Float64(),
			Source:      tx.Source,
			Destination: tx.Destination,
			Timestamp:   timestamppb.New(tx.Timestamp),
		}
		token = tx.ID
	}

	return &pb.ListTransactionsResponse{Transactions: resp, NextPageToken: token}, nil
}

func (s *AdminRpcService) GetTransactionDetails(ctx context.Context, req *pb.GetTransactionDetailsRequest) (*pb.GetTransactionDetailsResponse, error) {
	tx, err := s.b.Transactions().GetTransaction(ctx, req.WalletID, req.TransactionID)
	if err != nil {
		return nil, err
	}

	transf, err := s.b.Transactions().ListTransfers(ctx, tx.ID)
	if err != nil {
		return nil, err
	}

	transResp := make([]*pb.Transfer, len(transf))
	for i, t := range transf {
		la, err := s.b.LinkedAccounts().Get(ctx, t.LinkedAccountID)
		if err != nil {
			return nil, err
		}

		transResp[i] = &pb.Transfer{
			ID:                    t.ID,
			LinkedAccountID:       t.LinkedAccountID,
			LinkedAccountProvider: la.Provider,
			LinkedAccountType:     la.Type,
			Amount:                t.Amount.Float64(),
			Currency:              t.Amount.Currency.String(),
			State:                 string(t.State),
			ForeignID:             t.ForeignID,
			Timestamp:             timestamppb.New(t.Timestamp),
		}
	}

	return &pb.GetTransactionDetailsResponse{
		Transaction: &pb.Transaction{
			WalletID:    req.WalletID,
			Id:          tx.ID,
			Type:        string(tx.Type),
			Asset:       tx.Amount.Currency.String(),
			Amount:      tx.Amount.Float64(),
			Source:      tx.Source,
			Destination: tx.Destination,
			Timestamp:   timestamppb.New(tx.Timestamp),
		},
		Transfers: transResp,
	}, nil
}
