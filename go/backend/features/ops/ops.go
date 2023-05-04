package ops

import (
	"context"
	"strings"

	"gitlab.com/fynbos/backend/kyc"

	"gitlab.com/fynbos/backend/features"
)

func Features(ctx context.Context, b Backends, walletID string) (*features.WalletFeatures, error) {
	kycStatus, err := b.KYC().GetKYCStatus(ctx, walletID)
	if err != nil {
		return nil, err
	}
	// If you are not KYC approved you can do nothing
	if kycStatus != kyc.StatusApproved {
		return &features.WalletFeatures{}, nil
	}

	// Identities are enabled default everywhere
	res := &features.WalletFeatures{
		IdentitiesEnabled: true,
		TwitterEnabled:    true,
	}

	id, err := b.KYC().GetIndividualDetails(ctx, walletID)
	if id.CountryCode == "US" {
		res.RecvEnabled = true
		res.LinkedAccEnabled = true
		res.BanksEnabled = true
		res.CardsEnabled = true
	}
	if id.Address != nil && !isGMTSendState(id.Address.State) {
		res.SendEnabled = true
		res.LinkedAccEnabled = true
		res.BanksEnabled = true
		res.CardsEnabled = true
	}

	return res, nil
}

var gmtUSStates = []string{
	"US-AK", "US-DC", "US-MA", "US-ND", "US-OR", "US-UT",
	"US-AL", "US-DE", "US-MD", "US-NH", "US-RI", "US-VA",
	"US-AR", "US-FL", "US-MN", "US-NJ", "US-SC", "US-WA",
	"US-CA", "US-GA", "US-MO", "US-NM", "US-SD",
	"US-CO", "US-ID", "US-MT", "US-NV", "US-TN",
	"US-CT", "US-IL", "US-NC", "US-NY", "US-TX",
}

func isGMTSendState(state string) bool {
	// Check if the state is in format US-CA
	if len(state) != 5 {
		state = "US-" + state
	}

	for _, s := range gmtUSStates {
		if strings.EqualFold(s, state) {
			return true
		}
	}
	return false
}
