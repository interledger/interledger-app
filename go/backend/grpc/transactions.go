package grpc

import (
	"context"
	"math"
	"strings"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/transactions"
	pb "gitlab.com/fynbos/proto/backend/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *rpcService) ListTransactions(ctx context.Context, req *pb.PaginationRequest) (*pb.ListTransactionsResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	page := db.PaginationFromPB(req)

	txs, err := s.b.Transactions().List(ctx, page, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformTransactions(ctx, s.b, txs, page), nil
}

func (s *rpcService) ListTransactionsCompleted(ctx context.Context, req *pb.PaginationRequest) (*pb.ListTransactionsResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	page := db.PaginationFromPB(req)

	txs, err := s.b.Transactions().ListCompleted(ctx, page, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformTransactions(ctx, s.b, txs, page), nil
}

func (s *rpcService) ListTransactionsWithPending(ctx context.Context, req *pb.PaginationRequest) (*pb.ListTransactionsResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	page := db.PaginationFromPB(req)

	txs, err := s.b.Transactions().ListWithPending(ctx, page, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformTransactions(ctx, s.b, txs, page), nil
}

func transformTransactions(ctx context.Context, b Backends, txs []transactions.Transaction, page db.Pagination) *pb.ListTransactionsResponse {
	var nextPageToken string

	resSize := int(math.Min(float64(len(txs)), float64(page.PageSize)))
	res := make([]*pb.Transaction, resSize)

	for i, tx := range txs {
		// If we have more txs than PageSize, we have a next page.
		if i == page.PageSize {
			// Use the PageSize+1 tx.ID as the start of the next page.
			nextPageToken = tx.ID
			break
		}

		res[i] = transformTransaction(ctx, b, tx)
	}

	return &pb.ListTransactionsResponse{
		Transactions:  res,
		NextPageToken: nextPageToken,
	}
}

func transformTransaction(ctx context.Context, b Backends, tx transactions.Transaction) *pb.Transaction {
	refundState := "NA"
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
		// The transfers are in order.
		// So it will move from state NA -> PENDING -> COMPLETE based on how many transfers there are
		// This is for Card origination transaction so any ACH origin transactions will never get out of NA state
		if refundState == "NA" &&
			tx.State == transactions.StateFailed &&
			tx.Type == transactions.TransactionTypeOpenOutgoingPayment &&
			tr.Type == transactions.TransferTypeDebitCard {
			refundState = "PENDING"
		}
		if refundState == "PENDING" &&
			tx.State == transactions.StateFailed &&
			tx.Type == transactions.TransactionTypeOpenOutgoingPayment &&
			tr.Type == transactions.TransferTypeCreditCard {
			refundState = "COMPLETE"
		}
	}

	amt := tx.Amount.Format()
	title := tx.DestinationIdentity
	if tx.Type == transactions.TransactionTypeOpenPaymentIncoming || tx.Type == transactions.TransactionTypeIncoming {
		title = tx.Source
	}
	w, _ := b.Wallets().GetFromAddress(ctx, title)
	// We don't care about errors, we'll use the source/destination/full wallet address as the fallback
	if w != nil {
		title = w.Name
	}

	// transform identity types to twitter/wallet for the frontend
	destinationIdentityType := tx.DestinationIdentityType
	if destinationIdentityType == payments.IdentityTypeWalletID.String() ||
		destinationIdentityType == payments.IdentityTypeWalletURL.String() ||
		tx.Type == transactions.TransactionTypeOpenPaymentIncoming ||
		tx.Type == transactions.TransactionTypeIncoming {
		destinationIdentityType = "wallet"
	}

	// Remove https if it exists
	title = strings.TrimPrefix(title, "https://")

	fees := currency.FromFloat64(0, tx.Amount.Currency)

	return &pb.Transaction{
		Id:                      tx.ID,
		Type:                    string(tx.Type),
		Amount:                  tx.Amount.ToPB(),
		Source:                  tx.Source,
		Destination:             tx.Destination,
		Timestamp:               timestamppb.New(tx.Timestamp),
		State:                   string(tx.State),
		Transfers:               trs,
		ForeignId:               tx.ForeignID,
		Title:                   title,
		FormattedAmount:         amt,
		FormattedTime:           tx.Timestamp.Format("15:04"),
		FormattedDate:           tx.Timestamp.Format("02 Jan 2006"),
		Subtotal:                amt,
		Fees:                    fees.Format(),
		AccountTitle:            tx.LinkedAccountTitle,
		Reference:               tx.Reference,
		DestinationIdentity:     tx.DestinationIdentity,
		DestinationIdentityType: destinationIdentityType,
		RefundState:             refundState,
	}
}

func (s *rpcService) LookupTransaction(ctx context.Context, req *pb.LookupTransactionRequest) (*pb.Transaction, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	tx, err := s.b.Transactions().GetTransaction(ctx, wallet.ID, req.GetId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformTransaction(ctx, s.b, *tx), nil
}
