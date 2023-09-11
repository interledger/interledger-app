package grpc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	http_log "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/backend/providers/tabapay"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
	"go.temporal.io/sdk/temporal"
)

func (s *rpcService) CreateCard(
	ctx context.Context, req *backendv1.CreateCardRequest,
) (*backendv1.LinkedAccount, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	feats, err := s.b.Features().Features(ctx, w.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if !feats.AddCardsEnabled {
		return nil, NewValidationError("Form", "You have connected the maximum number of cards to Fynbos.")
	}

	await, err := s.b.Tabapay().CreateCard(ctx, tabapay.CreateCardArgs{
		WalletID:           w.ID,
		BasisTheoryTokenID: req.GetTokenID(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	var la linkedaccounts.LinkedAccount
	err = await(ctx, &la)
	var applicationError *temporal.ApplicationError
	if errors.As(err, &applicationError) {
		switch applicationError.Type() {
		case "ErrDuplicateCard":
			return nil, AlreadyExistsError("ErrDuplicateCard")
		case "ErrUnsupportedCard":
			return nil, NewValidationError("CardNumber", "Your card is unsupported and cannot be connected to Fynbos.")
		case "ErrUnsupportedCountry":
			return nil, NewValidationError("CardNumber", "Your card originates from an unsupported country and cannot be connected to Fynbos.")
		case "ErrMultiStatus":
			return nil, UnavailableError("ErrMultiStatus")
		}
	}
	if err != nil {
		return nil, toGRPCError(err)
	}
	if la.ID == "" {
		return nil, toGRPCError(errors.New("Linked account not returned from create card workflow."))
	}

	return transformLinkedAccount(la), nil
}

func (s *rpcService) InitQuote3DS(
	ctx context.Context, req *backendv1.InitQuote3DSRequest,
) (*backendv1.Init3DSResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	quote, err := s.b.OpenPayments().GetQuote(ctx, req.GetQuoteID())
	if err != nil {
		return nil, InternalError("Quote not found.")
	}
	orderID := quote.ID
	idxSlash := strings.LastIndex(quote.ID, "/")
	if idxSlash > 0 {
		orderID = orderID[idxSlash+1:]
	}

	fromLinkedAcc, err := s.b.LinkedAccounts().Get(ctx, quote.FromLinkedAccount)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if fromLinkedAcc.WalletID != w.ID {
		return nil, NotFoundError("Quote not found.")
	}
	if !linkedaccounts.Requires3DS(fromLinkedAcc) {
		return nil, InternalError("3DS not supported.")
	}

	newCtx := context.WithValue(ctx, http_log.ContextKey, &http_log.Metadata{
		Context: fmt.Sprintf("linkedAccountID=%s", fromLinkedAcc.ID),
	})
	init3DS, err := s.b.Tabapay().Init3DS(newCtx, tabapay.Init3DSArgs{
		Amount:  quote.SendAmount,
		OrderID: orderID,
		CardID:  fromLinkedAcc.ProviderID,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.Init3DSResponse{
		Id:                  init3DS.ID,
		Jwt:                 init3DS.JWT,
		DeviceCollectionURL: init3DS.DeviceCollectionURL,
		SongbirdURL:         tabapay.GetSongbirdURL(),
	}, nil
}

func (s *rpcService) Init3DS(
	ctx context.Context, req *backendv1.Init3DSRequest,
) (*backendv1.Init3DSResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	payment, err := s.b.Payments().Lookup(ctx, req.GetPaymentID())
	if err != nil {
		return nil, toGRPCError(err)
	}

	fromLinkedAcc, err := s.b.LinkedAccounts().Get(ctx, payment.SenderAccount)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if fromLinkedAcc.WalletID != w.ID {
		return nil, NotFoundError("Payment not found.")
	}
	if !linkedaccounts.Requires3DS(fromLinkedAcc) {
		return nil, InternalError("3DS not supported.")
	}

	newCtx := context.WithValue(ctx, http_log.ContextKey, &http_log.Metadata{
		Context: fmt.Sprintf("linkedAccountID=%s", fromLinkedAcc.ID),
	})
	init3DS, err := s.b.Tabapay().Init3DS(newCtx, tabapay.Init3DSArgs{
		Amount:  payment.SenderAmount,
		OrderID: payment.ID, // TODO: should this be the publicID?
		CardID:  fromLinkedAcc.ProviderID,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	_, err = s.b.Payments().Update(ctx, payments.UpdateArgs{
		ID:        payment.ID,
		ThreeDSID: init3DS.ID,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.Init3DSResponse{
		Id:                  init3DS.ID,
		Jwt:                 init3DS.JWT,
		DeviceCollectionURL: init3DS.DeviceCollectionURL,
		SongbirdURL:         tabapay.GetSongbirdURL(),
	}, nil
}

func (s *rpcService) Lookup3DS(
	ctx context.Context, req *backendv1.Lookup3DSRequest,
) (*backendv1.Lookup3DSResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	session, err := s.b.Tabapay().Get3DSSession(ctx, req.GetThreeDSID())
	if err != nil {
		return nil, toGRPCError(err)
	}

	la, err := s.b.LinkedAccounts().GetByProviderID(ctx, linkedaccounts.GetByProviderIDArgs{
		Provider:   tabapay.ProviderName,
		ProviderID: session.CardID,
		WalletID:   w.ID,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	if la.WalletID != w.ID {
		return nil, NotFoundError("")
	}

	newCtx := context.WithValue(ctx, http_log.ContextKey, &http_log.Metadata{
		Context: fmt.Sprintf("linkedAccountID=%s", la.ID),
	})
	lookupResp, err := s.b.Tabapay().Lookup3DS(newCtx, tabapay.Lookup3DSArgs{
		ThreeDSID:               session.ID,
		OrderID:                 session.OrderID,
		CardID:                  session.CardID,
		AuthenticationIndicator: tabapay.AuthenticatorIndicatorPayment,
		TransactionMode:         tabapay.TransactionModeComputer,
		ProductCode:             tabapay.ProductCodeQuasiCashTransaction,
		DeviceChannel:           tabapay.DeviceChannelBrowser,
		Amount: currency.Amount{
			Value:    session.Amount,
			Currency: currency.ParseCurrency(session.Currency),
		},
		BrowserInfo: tabapay.NewBrowserInfo(tabapay.BrowserInfoFields{
			JavascriptEnabled: req.GetJavascriptEnabled(),
			UserAgent:         req.GetUserAgent(),
			Header:            req.GetHeader(),
			JavaEnabled:       req.GetJavaEnabled(),
			Language:          req.GetLanguage(),
			ColorDepth:        tabapay.GetColorDepth(req.GetColorDepth()),
			ScreenHeight:      req.GetScreenHeight(),
			ScreenWidth:       req.GetScreenWidth(),
			TimeZone:          req.GetTimezone(),
			IpAddress:         req.GetIpAddress(),
		}),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.Lookup3DSResponse{
		ProcessorTransactionID: lookupResp.ProcessorTransactionID,
		ChallengeURL:           lookupResp.ChallengeURL,
		Payload:                lookupResp.Payload,
	}, nil
}

func (s *rpcService) Authenticate3DS(
	ctx context.Context, req *backendv1.Authenticate3DSRequest,
) (*backendv1.Authenticate3DSResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	session, err := s.b.Tabapay().Get3DSSession(ctx, req.GetThreeDSID())
	if err != nil {
		return nil, toGRPCError(err)
	}

	la, err := s.b.LinkedAccounts().GetByProviderID(ctx, linkedaccounts.GetByProviderIDArgs{
		Provider:   tabapay.ProviderName,
		ProviderID: session.CardID,
		WalletID:   w.ID,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	if la.WalletID != w.ID {
		return nil, NotFoundError("")
	}

	newCtx := context.WithValue(ctx, http_log.ContextKey, &http_log.Metadata{
		Context: fmt.Sprintf("linkedAccountID=%s", la.ID),
	})
	authResp, err := s.b.Tabapay().Authenticate3DS(newCtx, tabapay.Authenticate3DSArgs{
		ThreeDSID: req.GetThreeDSID(),
		JWT:       req.GetJwt(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.Authenticate3DSResponse{
		Status: authResp.Status,
	}, nil
}
