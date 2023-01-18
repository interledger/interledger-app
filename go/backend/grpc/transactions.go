package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/transactions"

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

	txs, err := s.b.Transactions().List(ctx, page, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformTransactions(txs, page)
}

func (s *rpcService) ListTransactionsCompleted(ctx context.Context, req *pb.PaginationRequest) (*pb.ListTransactionsResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	page := db.PaginationFromPB(req)

	txs, err := s.b.Transactions().ListCompleted(ctx, page, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformTransactions(txs, page)
}

func (s *rpcService) ListTransactionsWithPending(ctx context.Context, req *pb.PaginationRequest) (*pb.ListTransactionsResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	page := db.PaginationFromPB(req)

	txs, err := s.b.Transactions().ListWithPending(ctx, page, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformTransactions(txs, page)
}

func transformTransactions(txs []transactions.Transaction, page db.Pagination) (*pb.ListTransactionsResponse, error) {
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
				Amount:          tr.Amount.ToPB(),
			}
		}

		res[i] = &pb.Transaction{
			Id:          tx.ID,
			ForeignId:   tx.ForeignID,
			Type:        string(tx.Type),
			Amount:      tx.Amount.ToPB(),
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

func (s *rpcService) LookupTransaction(ctx context.Context, req *pb.LookupTransactionRequest) (*pb.Transaction, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	tx, err := s.b.Transactions().GetTransaction(ctx, wallet.ID, req.GetId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	trs := make([]*pb.Transfer, len(tx.Transfers))
	for y, tr := range tx.Transfers {
		trs[y] = &pb.Transfer{
			ForeignId:       tr.ForeignID,
			Type:            string(tr.Type),
			State:           string(tr.State),
			LinkedAccountId: tr.LinkedAccountID,
			Timestamp:       timestamppb.New(tr.Timestamp),
			Amount:          tr.Amount.ToPB(),
		}
	}

	return &pb.Transaction{
		Id:          tx.ID,
		ForeignId:   tx.ForeignID,
		Type:        string(tx.Type),
		Amount:      tx.Amount.ToPB(),
		Source:      tx.Source,
		Destination: tx.Destination,
		Timestamp:   timestamppb.New(tx.Timestamp),
		State:       string(tx.State),
		Transfers:   trs,
	}, nil
}
