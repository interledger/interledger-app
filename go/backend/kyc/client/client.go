package client

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/kyc/persona"
	"gitlab.com/fynbos/backend/kyc/smileid"

	"gitlab.com/fynbos/env"

	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/kyc/address"
	"gitlab.com/fynbos/backend/kyc/ops"
)

var _ kyc.Client = client{}

type client struct {
	b   ops.Backends
	val address.Validator
	pc  persona.Client
	sc  smileid.Client
}

func New(b ops.Backends, smartyAuthID, smartyAuthToken string) (kyc.Client, error) {
	if (smartyAuthID == "" || smartyAuthToken == "") &&
		(env.IsSandbox() || env.IsProd()) {
		return nil, errors.New("no auth information for smarty address verification")
	}

	return &client{
		b:   b,
		val: address.New(smartyAuthID, smartyAuthToken),
		pc:  persona.New(),
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

func (c client) GetKYCStatus(ctx context.Context, walletID string) (kyc.Status, error) {
	return ops.GetKYCStatus(ctx, c.b, walletID)
}

func (c client) SetKYCStatus(ctx context.Context, walletID string, status kyc.Status) error {
	return ops.SetKYCStatus(ctx, c.b, walletID, status)
}

func (c client) StartKYC(ctx context.Context, walletID string) error {
	return ops.StartKYC(ctx, c.b, walletID)
}

func (c client) GetPersonaInquiry(ctx context.Context, walletID, idempotencyKey string) (*kyc.PersonaInquiry, error) {
	return ops.GetPersonaInquiry(ctx, c.b, c.pc, walletID, idempotencyKey)
}

func (c client) GetPersonaIDNumbers(ctx context.Context, walletID string) (*kyc.PersonaIDNumbers, error) {
	return ops.GetPersonaIDNumbers(ctx, c.b, c.pc, walletID)
}

func (c client) GetSmileIDToken(ctx context.Context, walletID string) (string, error) {
	return ops.GetSmileIDToken(ctx, c.sc, walletID)
}
