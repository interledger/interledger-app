package kyc

import "context"

type Client interface {
	GetIndividualDetails(ctx context.Context, userID string) (*IndividualDetails, error)
	UpdateIndividualDetails(ctx context.Context, args IndividualDetails) (*IndividualDetails, error)
}
