package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/providers/gmt"
	"gitlab.com/fynbos/backend/providers/tabapay"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) Lookup3DS(
	ctx context.Context, req *backendv1.Lookup3DSRequest,
) (*backendv1.Lookup3DSResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	op, err := s.b.OpenPayments().GetOutgoingPayment(ctx, req.GetOutgoingPaymentID())
	if err != nil {
		return nil, toGRPCError(err)
	}

	// make sure payment belongs to wallet
	_, err = s.b.Transactions().GetTransactionByForeignID(ctx, w.ID, op.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	linkedCard, err := s.b.LinkedAccounts().Get(ctx, op.FromLinkedAccount)
	if err != nil {
		return nil, toGRPCError(err)
	}

	lookupResp, err := s.b.Tabapay().Lookup3DS(ctx, tabapay.Lookup3DSArgs{
		ThreeDSID:               req.GetThreeDSID(),
		OutgoingPaymentID:       req.GetOutgoingPaymentID(),
		CardID:                  linkedCard.ProviderID,
		AuthenticationIndicator: tabapay.AuthenticatorIndicatorPayment,
		TransactionMode:         tabapay.TransactionModeComputer,
		ProductCode:             tabapay.ProductCodeQuasiCashTransaction,
		DeviceChannel:           tabapay.DeviceChannelBrowser,
		Amount:                  op.SendAmount,
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

	w, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	op, err := s.b.OpenPayments().GetOutgoingPayment(ctx, req.GetOutgoingPaymentID())
	if err != nil {
		return nil, toGRPCError(err)
	}

	// make sure payment belongs to wallet
	_, err = s.b.Transactions().GetTransactionByForeignID(ctx, w.ID, op.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	authResp, err := s.b.Tabapay().Authenticate3DS(ctx, tabapay.Authenticate3DSArgs{
		OutgoingPaymentID: req.GetOutgoingPaymentID(),
		ThreeDSID:         req.GetThreeDSID(),
		JWT:               req.GetJwt(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	err = s.b.GMT().Authenticate3DS(ctx, gmt.Authenticate3DSArgs{
		OutgoingPaymentID:      req.GetOutgoingPaymentID(),
		JWT:                    req.GetJwt(),
		ThreeDSID:              req.GetThreeDSID(),
		ThreeDSVersion:         authResp.Version3DS,
		ProcessorTransactionID: authResp.ProcessorTransactionID,
		DsTransactionID:        authResp.DsTransactionID,
		Status:                 authResp.Status,
		UCAF:                   authResp.UCAF,
		XID:                    authResp.XID,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.Authenticate3DSResponse{
		Status: authResp.Status,
	}, nil
}
