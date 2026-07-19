package grpc

import (
	"context"

	"github.com/interledger/interledger-app/go/backend/kyc"
)

func (s *rpcService) validateKYCTransactionRestrictions(ctx context.Context, walletID string) error {
	status, err := s.b.KYC().GetKYCStatus(ctx, walletID)
	if err != nil {
		return err
	}

	if status == kyc.StatusDocumentsRequired {
		return kyc.ErrKYCResubmissionRequired
	}

	return nil
}
