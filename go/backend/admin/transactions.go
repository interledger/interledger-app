package admin

import (
	"context"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/env"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "gitlab.com/fynbos/proto/backend/admin/v1"
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

func (s *AdminRpcService) SeedTransactions(ctx context.Context, req *pb.SeedTransactionsRequest) (*pb.Empty, error) {
	if env.IsProd() {
		return nil, UnimplementedError("")
	}

	var txs []transactions.Transaction
	for _, tx := range req.Transactions {
		txs = append(txs, transactions.Transaction{
			ID:                             tx.Id,
			ForeignID:                      tx.ForeignId,
			Source:                         tx.Source,
			Destination:                    tx.Destination,
			Title:                          tx.Title,
			Note:                           tx.Note,
			Type:                           transactions.TransactionType(tx.Type),
			Timestamp:                      tx.Timestamp.AsTime(),
			Provider:                       transactions.Provider(tx.Provider),
			State:                          transactions.State(tx.State),
			Amount:                         currency.FromUInt64(tx.Amount, currency.ParseCurrency(tx.AssetCode)),
			LinkedAccountTitle:             tx.LinkedAccountTitle,
			DestinationIdentity:            tx.DestinationIdentity,
			DestinationIdentityType:        tx.DestinationIdentityType,
			Reference:                      tx.Reference,
			RefundState:                    transactions.RefundState(tx.RefundState),
			PaymentProtectionFeePercentage: tx.PaymentProtectionFeePercentage,
		})
	}

	_, err := s.b.AdminTransactions().Seed(ctx, txs)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}
