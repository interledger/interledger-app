package grpc

import (
	"context"
	"errors"
	"math"
	"sync"

	"gitlab.com/fynbos/backend/openpayments"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/transactions"
	pb "gitlab.com/fynbos/proto/backend/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
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

	return transformTransactions(ctx, s.b, txs, page)
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

	return transformTransactions(ctx, s.b, txs, page)
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

	return transformTransactions(ctx, s.b, txs, page)
}

func transformTransactions(ctx context.Context, b Backends, txs []transactions.Transaction, page db.Pagination) (*pb.ListTransactionsResponse, error) {
	var nextPageToken string

	resSize := int(math.Min(float64(len(txs)), float64(page.PageSize)))
	res := make([]*pb.Transaction, resSize)

	var wg sync.WaitGroup
	var anyErr error

	for i, tx := range txs {
		// If we have more txs than PageSize, we have a next page.
		if i == page.PageSize {
			// Use the PageSize+1 tx.ID as the start of the next page.
			nextPageToken = tx.ID
			break
		}

		wg.Add(1)

		go func(index int, ttx transactions.Transaction) {
			defer wg.Done()
			tt, err := transformTransaction(ctx, b, ttx)
			if err != nil {
				anyErr = err
			}
			// This is thread safe, because we are writing to unique/specific indexes
			res[index] = tt
		}(i, tx)
	}

	wg.Wait()

	if anyErr != nil {
		return nil, toGRPCError(anyErr)
	}

	return &pb.ListTransactionsResponse{
		Transactions:  res,
		NextPageToken: nextPageToken,
	}, nil
}

func transformTransaction(ctx context.Context, b Backends, tx transactions.Transaction) (*pb.Transaction, error) {
	var laid string
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
		if tr.LinkedAccountID != "" {
			laid = tr.LinkedAccountID
		}
	}

	var laTitle string
	if laid != "" {
		la, err := b.LinkedAccounts().Get(ctx, laid)
		if err != nil {
			return nil, err
		}
		laTitle = la.Nickname
		if la.Nickname == "" {
			laTitle = la.Mask
		}
	}

	amt := tx.Amount.Format()
	title := tx.Source
	var reference string
	if tx.Type == transactions.TransactionTypeOpenOutgoingPayment {
		title = tx.Destination

		op, err := b.OpenPayments().GetOutgoingPayment(ctx, tx.ForeignID)
		if err != nil && !errors.Is(err, openpayments.ErrNotFound) {
			return nil, err
		}
		if op != nil {
			reference = op.Description
		}

	} else if tx.Type == transactions.TransactionTypeOpenPaymentIncoming {
		ip, err := b.OpenPayments().GetIncomingPayment(ctx, tx.ForeignID)
		if err != nil && !errors.Is(err, openpayments.ErrNotFound) {
			return nil, err
		}
		if ip != nil {
			reference = ip.ExternalRef
			if reference == "" {
				reference = ip.Description
			}
		}
	}

	fees := currency.FromFloat64(0, tx.Amount.Currency)

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
		AccountTitle:    laTitle,
		Fees:            fees.Format(),
		Reference:       reference,
		Subtotal:        amt,
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

	return transformTransaction(ctx, s.b, *tx)
}
