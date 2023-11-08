package grpc

import (
	"context"
	"errors"
	"fmt"
	"math"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/env"
	pb "gitlab.com/fynbos/proto/backend/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *rpcService) CreatePaymentLink(ctx context.Context, req *pb.CreatePaymentLinkRequest) (*pb.PaymentLink, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	t, err := s.b.Transactions().GetTransaction(ctx, w.ID, req.GetTransactionId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	if t.Provider != "payments_engine" {
		return nil, NotFoundError("")
	}

	// insert receive payment link into db
	link, err := s.b.Payments().CreatePaymentLink(ctx, t.ForeignID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.PaymentLink{
		Url: fmt.Sprintf("%s/collect/%s", env.GetUrl(), link.ID),
	}, nil
}

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

	return transformTransactions(txs, page), nil
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

	return transformTransactions(txs, page), nil
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

	return transformTransactions(txs, page), nil
}

func transformTransactions(txs []transactions.Transaction, page db.Pagination) *pb.ListTransactionsResponse {
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

		res[i] = transformTransaction(tx)
	}

	return &pb.ListTransactionsResponse{
		Transactions:  res,
		NextPageToken: nextPageToken,
	}
}

func transformTransaction(tx transactions.Transaction) *pb.Transaction {
	amt := tx.Amount.Format()
	subTotal := amt

	// transform identity types to twitter/wallet for the frontend
	destinationIdentityType := tx.DestinationIdentityType
	if destinationIdentityType == payments.IdentityTypeWalletID.String() ||
		destinationIdentityType == payments.IdentityTypeWalletURL.String() ||
		tx.Type == transactions.TransactionTypeOpenPaymentIncoming ||
		tx.Type == transactions.TransactionTypeIncoming {
		destinationIdentityType = "wallet"
	}

	fees := currency.FromFloat64(0, tx.Amount.Currency)

	paymentProtection := tx.PaymentProtectionAmount()

	// For outgoing transactions we need to minus the payment protection fee from the amount for the subtotal
	if tx.Type == transactions.TransactionTypeOpenOutgoingPayment || tx.Type == transactions.TransactionTypeOutgoing {
		subTotalAmount := currency.FromUInt64(tx.Amount.Value-paymentProtection.Value, tx.Amount.Currency)
		subTotal = subTotalAmount.Format()
	}

	return &pb.Transaction{
		Id:                      tx.ID,
		Type:                    string(tx.Type),
		Amount:                  tx.Amount.ToPB(),
		Source:                  tx.Source,
		Destination:             tx.Destination,
		Timestamp:               timestamppb.New(tx.Timestamp),
		State:                   string(tx.State),
		ForeignId:               tx.ForeignID,
		Title:                   tx.Title,
		FormattedAmount:         amt,
		FormattedTime:           tx.Timestamp.Format("15:04"),
		FormattedDate:           tx.Timestamp.Format("02 Jan 2006"),
		Subtotal:                subTotal,
		Fees:                    fees.Format(),
		AccountTitle:            tx.LinkedAccountTitle,
		Reference:               tx.Reference,
		DestinationIdentity:     tx.DestinationIdentity,
		DestinationIdentityType: destinationIdentityType,
		RefundState:             int32(tx.RefundState),
		HasPaymentProtection:    tx.PaymentProtectionFeePercentage != 0,
		PaymentProtectionAmount: paymentProtection.Format(),
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
	ret := transformTransaction(*tx)

	link, err := s.b.Payments().GetPaymentLinkByPaymentID(ctx, tx.ForeignID)
	if err != nil && !errors.Is(err, payments.ErrNotFound) {
		return nil, toGRPCError(err)
	}
	if link != nil {
		ret.HasPaymentLink = true
		ret.PaymentLinkCompleted = !link.CompletedAt.IsZero()
		ret.PaymentLinkExpired = !link.ExpiresAt.IsZero()
		ret.FormattedPaymentLinkExpiryDate = link.ExpiresAt.Format("02 Jan 2006")
		ret.PaymentLinkId = link.ID
		ret.PaymentLinkUrl = fmt.Sprintf("%s/collect/%s", env.GetUrl(), link.ID)
	}

	return ret, nil
}
