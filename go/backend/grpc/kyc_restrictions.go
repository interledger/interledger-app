package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/kyc"
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
