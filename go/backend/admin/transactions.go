package admin

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/interledger/interledger-app/go/backend/db"

	adminv1 "github.com/interledger/interledger-app/go/proto/backend/admin/v1"
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

func (s *AdminRpcService) GetTransactionStats(ctx context.Context, _ *adminv1.Empty) (*adminv1.TransactionStats, error) {
	stats, err := s.b.Transactions().TransactionStats(ctx, time.Now())
	if err != nil {
		return nil, err
	}
	quarterly := make([]*adminv1.QuarterlyTransactionCount, 0, len(stats.ByQuarter))
	for i, count := range stats.ByQuarter {
		quarterly = append(quarterly, &adminv1.QuarterlyTransactionCount{
			Quarter: int32(i + 1),
			Count:   int32(count),
		})
	}

	byType := make([]*adminv1.TypeTransactionCount, 0, len(stats.ByType))
	for _, t := range stats.ByType {
		byType = append(byType, &adminv1.TypeTransactionCount{
			Type:  string(t.Type),
			Count: int32(t.Count),
		})
	}

	return &adminv1.TransactionStats{
		TotalTransactions:     int32(stats.Total),
		TransactionsThisYear:  int32(stats.ThisYear),
		QuarterlyTransactions: quarterly,
		TypeTransactions:      byType,
	}, nil
}
