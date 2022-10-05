package kyc

import "context"

type Client interface {
	GetIndividualDetails(ctx context.Context, walletID string) (*IndividualDetails, error)
	UpdateIndividualDetails(ctx context.Context, args IndividualDetails) (*IndividualDetails, error)
}
