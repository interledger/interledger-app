package grpc

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/paymentsv2"
	"gitlab.com/fynbos/geo"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) CreatePaymentV2(ctx context.Context, req *pb.CreatePaymentV2Request) (*pb.CreatePaymentV2Response, error) {
	senderWallet, err := s.b.Wallets().Get(ctx, "ecf1a947-8d96-4768-832e-9ddebec4f082")
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	{ // KYC CHECKS
		// TODO: Validate receiver KYC as well if payment is local
		// TODO: accept a string slice with IDs - if all of them are approved, return true, otherwise false

		approved, err := s.b.KYC().IsKYCApproved(ctx, senderWallet.ID)
		if err != nil {
			return nil, toGRPCError(err)
		}

		if !approved {
			return nil, ForbiddenError("KYC not approved")
		}
	}

	{ // KYC LIMITS CHECKS
		// exceedsLimits, limitType, err := s.b.Limits().ExceedsKYCLimits(ctx, senderWallet.ID, req.SenderCurrency.Amount)
		// if err != nil {
		// 	return nil, toGRPCError(err)
		// }
		// if exceedsLimits {
		// 	var description string
		// 	switch limitType {
		// 	case limits.LimitTypeTransaction:
		// 		description = "Exceeds per transaction limit."
		// 	case limits.LimitTypeDaily:
		// 		description = "Exceeds daily limit."
		// 	case limits.LimitTypeMonthly:
		// 		description = "Exceeds monthly limit."
		// 	case limits.LimitType6Monthly:
		// 		description = "Exceeds 6 monthly limit."
		// 	case limits.LimitTypeYearly:
		// 		description = "Exceeds yearly limit."
		// 	default:
		// 		description = "Exceeds account limit."
		// 	}
		// 	return nil, NewValidationError("amount", description)
		// }
	}

	senderLinkedAccount, err := s.b.LinkedAccounts().Get(ctx, req.SenderAccountId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	// Get receiver wallet id and account id from wallet address
	receiverWallet, err := s.b.Wallets().GetFromAddress(ctx, req.ReceiverWalletAddress)
	if err != nil {
		return nil, toGRPCError(err)
	}

	{ // GET SENDER BALANCE
		// TODO: Validate balance is sufficient
		balance, err := s.b.Xago().GetBalance(ctx, senderLinkedAccount.ID)
		if err != nil {
			return nil, toGRPCError(err)
		}

		senderAmount, err := str2uint64(req.GetSenderCurrency().Amount)
		if err != nil {
			return nil, toGRPCError(err)
		}

		if balance.Available.Value < senderAmount {
			return nil, NewValidationError("amount", fmt.Sprintf("Insufficient balance. Available: %v, required: %v", balance.Available.Format(), senderAmount))
		}
	}

	// HELPER FUNCTION TO GET BALANCE ACCOUNT
	receiverLinkedAccounts, err := s.b.LinkedAccounts().ListByWalletId(ctx, receiverWallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	var receiverLinkedAccount *linkedaccounts.LinkedAccount
	for _, la := range receiverLinkedAccounts {
		if la.Type == "balance" {
			receiverLinkedAccount = &la
			break
		}
	}

	{ // ENSURE BOTH ACCOUNTS ARE ON THE SAME PROVIDER
		// TODO: More validation needed
		if senderLinkedAccount.Provider != receiverLinkedAccount.Provider {
			return nil, NewValidationError("amount", "Sender and receiver accounts must be on the same provider")
		}
	}

	{ // ENSURE BOTH ACCOUNTS ARE ON THE SAME TYPE
		// TODO: More validation needed
		if senderLinkedAccount.Type != receiverLinkedAccount.Type {
			return nil, NewValidationError("amount", "Sender and receiver accounts must be on the same type")
		}
	}

	// TODO: Confirm payment - UPDATE state to CONFIRMED
	// Start Workflow

	senderCurrency, err := geo.CurrencyFromProtoGeoV1(req.SenderCurrency)
	if err != nil {
		return nil, err
	}

	receiverCurrency := senderCurrency.Clone()

	payment := paymentsv2.Payment{
		ID:                uuid.NewString(),
		SenderWalletID:    senderWallet.ID,
		ReceiverWalletID:  receiverWallet.ID,
		SenderAccountID:   senderLinkedAccount.ID,
		ReceiverAccountID: receiverLinkedAccount.ID,
		SenderCurrency:    senderCurrency,
		ReceiverCurrency:  receiverCurrency,
		State:             "INITIATED",
		Transfers:         []string{},
	}

	if err := s.b.PaymentsV2().Store(ctx, &payment); err != nil {
		return nil, toGRPCError(err)
	}

	err = s.b.PaymentsV2().Process(ctx, &payment)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.CreatePaymentV2Response{Id: payment.ID}, nil
}

func str2uint64(str string) (uint64, error) {
	i, err := strconv.ParseInt(str, 10, 64)
	return uint64(i), err
}
