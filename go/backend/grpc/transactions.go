package grpc

import (
	"context"
	"math"

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

		res[i] = transformTransaction(tx, nil)
	}

	return &pb.ListTransactionsResponse{
		Transactions:  res,
		NextPageToken: nextPageToken,
	}
}

func transformTransaction(tx transactions.Transaction, transfers []transactions.Transfer) *pb.Transaction {
	amt := tx.Amount.Format()
	subTotal := amt

	// transform identity types to twitter/wallet for the frontend
	destinationIdentityType := tx.DestinationIdentityType
	if destinationIdentityType == payments.IdentityTypeWalletID.String() ||
		destinationIdentityType == payments.IdentityTypeWalletURL.String() ||
		tx.Type == transactions.TransactionTypeOpenPaymentIncoming ||
		tx.Type == transactions.TransactionTypeReceived {
		destinationIdentityType = "wallet"
	}

	// This should always come from the module directly
	feesPtr := currency.FromUInt64(0, currency.EUR)
	fees := feesPtr.Format()

	if tx.ProviderFee != nil {
		fees = tx.ProviderFee.Format()
	}

	// add the fee to the display amount
	fundsReceived := tx.Amount.Value

	if tx.ProviderFee != nil {
		if tx.ProviderFee.Value >= fundsReceived {
			fundsReceived = 0
		} else {
			fundsReceived -= tx.ProviderFee.Value
		}
	}
	currencyAmt := currency.FromUInt64(fundsReceived, tx.Amount.Currency)

	ret := &pb.Transaction{
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
		FundsReceived:           currencyAmt.Format(),
		Fees:                    fees,
		AccountTitle:            tx.LinkedAccountTitle,
		Reference:               tx.Reference,
		DestinationIdentity:     tx.DestinationIdentity,
		DestinationIdentityType: destinationIdentityType,
		RefundState:             int32(tx.RefundState),
	}

	// only adding the send and receive accounts to withdrawals
	if tx.Type == transactions.TransactionTypeWithdrawal {
		for _, t := range transfers {
			if t.Type == transactions.TransferTypeDebitBalance {
				ret.SenderAccountId = t.LinkedAccountID
			}
			if t.Type == transactions.TransferTypeCreditBankAccount {
				ret.ReceiverAccountId = t.LinkedAccountID
			}
		}
	}

	// only adding the send and receive accounts to withdrawals
	if tx.Type == transactions.TransactionTypeDeposit {
		for _, t := range transfers {
			if t.Type == transactions.TransferTypeCreditBalance {
				ret.ReceiverAccountId = t.LinkedAccountID
			}
		}
	}

	if tx.Type == transactions.TransactionTypeCardTransaction {
		details := pb.CardTransactionDetails{
			CardId:        tx.CardTransactionDetails.CardID,
			CardMaskedPan: tx.CardTransactionDetails.CardMaskedPan,
			Type:          int64(tx.CardTransactionDetails.Type),
		}
		ret.CardTransactionDetails = &details
	}

	return ret
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

	transfers, err := s.b.Transactions().ListTransfers(ctx, tx.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformTransaction(*tx, transfers), nil
}
