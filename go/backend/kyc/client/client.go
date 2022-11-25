package client

import (
	"context"
	"errors"

	"gitlab.com/fynbos/env"

	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/kyc/address"
	"gitlab.com/fynbos/backend/kyc/ops"
)

var _ kyc.Client = client{}

type client struct {
	b   ops.Backends
	val address.Validator
}

func New(b ops.Backends, smartyAuthID, smartyAuthToken string) (kyc.Client, error) {
	if (smartyAuthID == "" || smartyAuthToken == "") &&
		(env.IsSandbox() || env.IsProd()) {
		return nil, errors.New("no auth information for smarty address verification")
	}

	return &client{
		b:   b,
		val: address.New(smartyAuthID, smartyAuthToken),
	}, nil
}

func (c client) IsUSPSAddress(ctx context.Context, address kyc.Address) (bool, error) {
	return c.val.USPSAddress(ctx, address)
}

func (c client) GetIndividualDetails(ctx context.Context, walletID string) (*kyc.IndividualDetails, error) {
	return ops.GetIndividualDetails(ctx, c.b, walletID)
}

func (c client) UpdateIndividualDetails(ctx context.Context, args kyc.IndividualDetails) (*kyc.IndividualDetails, error) {
	return ops.UpdateIndividualDetails(ctx, c.b, args)
}
