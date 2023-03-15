package workflows

import (
	"context"
	"gitlab.com/fynbos/backend/kyc"
)

type Activity struct {
	b Backends
}

func NewActivity(b Backends) *Activity {
	return &Activity{b: b}
}

func (a *Activity) SetKYCStatus(ctx context.Context, walletID string, state kyc.Status) error {
	err := a.b.KYC().SetKYCStatus(ctx, walletID, state)
	return err
}
