package grpc

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/providers/rafiki"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

type validateGetQuote struct {
	SendAmount          uint64 `validate:"required_if=ReceiveAmount 0"`
	SendCurrencyCode    string `validate:"required_with=SendAmount,omitempty,iso4217"`
	ReceiveAmount       uint64 `validate:"required_if=SendAmount 0"`
	ReceiveCurrencyCode string `validate:"required_with=ReceiveAmount,omitempty,iso4217"`
	Receiver            string `validate:"required"`
}

func validateGetQuoteDescription(err validator.FieldError) string {
	switch err.Field() {
	case "SendCurrencyCode":
		return "SendCurrencyCode must be an ISO4217 code and specified with SendAmount."
	case "ReceiveCurrencyCode":
		return "ReceiveCurrencyCode must be an ISO4217 code and specified with ReceiveAmount."
	case "Receiver":
		return "Receiver is required."
	case "SendAmount", "ReceiveAmount":
		return "Either SendAmount or ReceiveAmount must be specified."
	default:
		return fmt.Sprintf("%s is invalid.", err.Field())
	}
}

func (s *rpcService) GetQuote(
	ctx context.Context,
	req *backendv1.GetQuoteRequest,
) (*backendv1.Quote, error) {
	if err := s.b.Validator().Struct(&validateGetQuote{
		SendAmount:          req.GetSendAmount(),
		SendCurrencyCode:    req.GetSendCurrencyCode(),
		ReceiveAmount:       req.GetReceiveAmount(),
		ReceiveCurrencyCode: req.GetReceiveCurrencyCode(),
		Receiver:            req.GetReceiverPaymentPointer(),
	}); err != nil {
		return nil, ValidationError(err, validateGetQuoteDescription)
	}

	user, err := s.b.Users().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}
	acc, err := s.b.Accounts().GetByIdentityID(ctx, user.ID)
	if err != nil {
		return nil, InternalError("Account not found.")
	}

	// temporary until `Quoting` service is written.
	identifier, err := s.b.Rafiki().GetIdentifierByAccountAndCurrency(
		ctx,
		acc.ID,
		"USD", // TODO: get currency from somewhere.
	)
	if err != nil {
		return nil, InternalError("Unable to get quote.")
	}

	args := &rafiki.CreateQuoteArgs{
		IdentifierID:           identifier.ID,
		ReceiverPaymentPointer: req.GetReceiverPaymentPointer(),
	}
	if req.GetSendAmount() != 0 {
		args.SendAmount = req.GetSendAmount()
		args.SendAssetCode = req.GetSendCurrencyCode()
		args.SendAssetScale = 2 // assuming for now
	}
	if req.GetReceiveAmount() != 0 {
		args.ReceiveAmount = req.GetReceiveAmount()
		args.ReceiveAssetCode = req.GetReceiveCurrencyCode()
		args.ReceiveAssetScale = 2 // assuming for now
	}
	quote, err := s.b.Rafiki().CreateQuote(ctx, args)
	if err != nil {
		return nil, InternalError("Unable to get quote.")
	}

	return &backendv1.Quote{
		Id:                  quote.ID,
		ReceiveAmount:       quote.ReceiveAmount,
		ReceiveCurrencyCode: quote.ReceiveAssetCode,
		SendAmount:          quote.SendAmount,
		SendCurrencyCode:    quote.SendAssetCode,
		ExpiresAt:           quote.ExpiresAt,
	}, nil
}
