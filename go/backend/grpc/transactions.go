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
	var nextPageToken string
	var res []*pb.Transaction

	for i, tx := range txs {

		// If we have more txs than PageSize, we have a next page.
		if i == page.PageSize {
			// Use the PageSize+1 tx.ID as the start of the next page.
			nextPageToken = tx.ID
			break
		}

		res = append(res, transformTransaction(tx))
	}

	return &pb.ListTransactionsResponse{
		Transactions:  res,
		NextPageToken: nextPageToken,
	}, nil
}

func transformTransaction(tx transactions.Transaction) *pb.Transaction {
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

	amt := tx.Amount.Format()
	title := tx.Source
	if tx.Type == transactions.TransactionTypeOpenOutgoingPayment {
		title = tx.Destination
		amt = "- " + amt
	}

	return &pb.Transaction{
		Id:              tx.ID,
		Type:            string(tx.Type),
		Amount:          tx.Amount.ToPB(),
		Source:          tx.Source,
		Destination:     tx.Destination,
		Timestamp:       timestamppb.New(tx.Timestamp),
		State:           string(tx.State),
		Transfers:       trs,
		ForeignId:       tx.ForeignID,
		Title:           title,
		FormattedAmount: amt,
		FormattedTime:   tx.Timestamp.Format("15:04"),
		FormattedDate:   tx.Timestamp.Format("02 Jan 2006"),
	}
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

	return transformTransaction(*tx), nil
}
