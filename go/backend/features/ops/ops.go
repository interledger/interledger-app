package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"gitlab.com/fynbos/backend/providers/tabapay"

	"gitlab.com/fynbos/backend/kyc"

	"gitlab.com/fynbos/backend/features"
)

func SetFeatures(ctx context.Context, b Backends, walletID string, feat features.WalletFeatures) (*features.WalletFeatures, error) {

	_, err := b.DB().ExecContext(ctx, "INSERT INTO wallet_features "+
		"(wallet_id, send_enabled, receive_enabled, linked_accounts_enabled, cards_enabled, banks_enabled, identities_enabled, twitter_enabled, add_cards_enabled) "+
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)  ON CONFLICT (wallet_id) DO UPDATE SET "+
		"send_enabled = excluded.send_enabled, "+
		"receive_enabled = excluded.receive_enabled, "+
		"linked_accounts_enabled = excluded.linked_accounts_enabled, "+
		"cards_enabled = excluded.cards_enabled, "+
		"banks_enabled = excluded.banks_enabled, "+
		"identities_enabled = excluded.identities_enabled, "+
		"twitter_enabled = excluded.twitter_enabled, "+
		"add_cards_enabled = excluded.add_cards_enabled, "+
		"updated_at=now()",
		walletID, feat.SendEnabled, feat.ReceiveEnabled, feat.LinkedAccEnabled, feat.CardsEnabled, feat.BanksEnabled, feat.IdentitiesEnabled, feat.TwitterEnabled, feat.AddCardsEnabled)
	if err != nil {
		return nil, fmt.Errorf("%w %s", features.ErrInternal, err)
	}

	return Features(ctx, b, walletID)
}

func Features(ctx context.Context, b Backends, walletID string) (*features.WalletFeatures, error) {
	// Check DB for feature overrides
	var res features.WalletFeatures
	err := b.DB().GetContext(ctx, &res,
		"SELECT send_enabled, receive_enabled, linked_accounts_enabled, cards_enabled, banks_enabled, identities_enabled, twitter_enabled, add_cards_enabled FROM wallet_features WHERE wallet_id=$1",
		walletID)
	if err == nil {
		return &res, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", features.ErrInternal, err)
	}

	kycStatus, err := b.KYC().GetKYCStatus(ctx, walletID)
	if err != nil {
		return nil, err
	}
	// If you are not KYC approved you can do nothing
	if kycStatus != kyc.StatusLevel1 && kycStatus != kyc.StatusLevel2 {
		return &features.WalletFeatures{}, nil
	}

	// Identities are enabled default everywhere
	res = features.WalletFeatures{
		IdentitiesEnabled: true,
		TwitterEnabled:    true,
	}

	id, err := b.KYC().GetIndividualDetails(ctx, walletID)
	if err != nil {
		return nil, err
	}

	canAddCard, err := canAddCards(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	if id.CountryCode == "US" {
		res.ReceiveEnabled = true
		res.LinkedAccEnabled = true
		res.BanksEnabled = false
		res.CardsEnabled = true
		res.AddCardsEnabled = canAddCard
	}
	if id.Address != nil && isGMTSendState(id.Address.State) {
		res.SendEnabled = true
		res.LinkedAccEnabled = true
		res.BanksEnabled = false
		res.CardsEnabled = true
		res.AddCardsEnabled = canAddCard
	}

	return &res, nil
}

func canAddCards(ctx context.Context, b Backends, walletID string) (bool, error) {
	lal, err := b.LinkedAccounts().ListByWalletId(ctx, walletID)
	if err != nil {
		return false, err
	}

	var cnt int
	for _, la := range lal {
		if la.Provider == tabapay.ProviderName &&
			la.Type == tabapay.TypeCard {
			cnt++
		}
	}

	return cnt <= 3, nil
}

var gmtUSStates = []string{
	"US-AL", "US-AK", "US-AZ", "US-AR", "US-CA", "US-CO", "US-CT", "US-DE", "US-DC", "US-FL",
	"US-GA", "US-ID", "US-IL", "US-LA", "US-MD", "US-MA", "US-MI", "US-MN", "US-MO", "US-MT",
	"US-NV", "US-NH", "US-NJ", "US-NM", "US-NY", "US-NC", "US-ND", "US-OR", "US-PA", "US-RI",
	"US-SC", "US-SD", "US-TN", "US-TX", "US-UT", "US-VA", "US-WA",
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
