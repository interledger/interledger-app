package kyc

import "context"

type Client interface {
	GetIndividualDetails(ctx context.Context, walletID string) (*IndividualDetails, error)
	UpdateIndividualDetails(ctx context.Context, args IndividualDetails) (*IndividualDetails, error)
	IsUSPSAddress(ctx context.Context, address Address) (bool, error)
	GetKYCStatus(ctx context.Context, walletID string) (Status, error)
	SetKYCStatus(ctx context.Context, walletID string, status Status) error
	GetPersonaInquiry(ctx context.Context, walletID, idempotencyKey string) (*PersonaInquiry, error)
	GetPersonaIDNumbers(ctx context.Context, walletID string) (*PersonaIDNumbers, error)
	GetPersonaZAIDNumber(ctx context.Context, walletID string) (string, error)
	GetApprovedPersonaInquiryURL(ctx context.Context, walletID string) (string, error)
	IsKYCApproved(ctx context.Context, walletID string) (bool, error)
}
