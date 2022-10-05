package client

import (
	"context"

	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/kyc/ops"
)

var _ kyc.Client = client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) kyc.Client {
	return &client{
		b: b,
	}
}

func (c client) GetIndividualDetails(ctx context.Context, walletID string) (*kyc.IndividualDetails, error) {
	return ops.GetIndividualDetails(ctx, c.b, walletID)
}

func (c client) UpdateIndividualDetails(ctx context.Context, args kyc.IndividualDetails) (*kyc.IndividualDetails, error) {
	return ops.UpdateIndividualDetails(ctx, c.b, args)
}
