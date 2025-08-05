package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/linkedaccounts"

	"gitlab.com/fynbos/backend/providers/astra"
	"gitlab.com/fynbos/backend/providers/chimoney"
	"gitlab.com/fynbos/backend/providers/xago"

	"gitlab.com/fynbos/backend/kyc"

	"gitlab.com/fynbos/backend/features"
	"gitlab.com/fynbos/env"
)

func SetFeatures(ctx context.Context, b Backends, walletID string, feat features.WalletFeatures) (*features.WalletFeatures, error) {

	_, err := b.DB().ExecContext(ctx, "INSERT INTO wallet_features "+
		"(wallet_id, send_enabled, receive_enabled, linked_accounts_enabled, cards_enabled, banks_enabled, identities_enabled, twitter_enabled, add_cards_enabled, interac_enabled, manage_wallet_cards_enabled, account_enabled) "+
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)  ON CONFLICT (wallet_id) DO UPDATE SET "+
		"send_enabled = excluded.send_enabled, "+
		"receive_enabled = excluded.receive_enabled, "+
		"linked_accounts_enabled = excluded.linked_accounts_enabled, "+
		"cards_enabled = excluded.cards_enabled, "+
		"banks_enabled = excluded.banks_enabled, "+
		"identities_enabled = excluded.identities_enabled, "+
		"twitter_enabled = excluded.twitter_enabled, "+
		"add_cards_enabled = excluded.add_cards_enabled, "+
		"interac_enabled = excluded.interac_enabled, "+
		"manage_wallet_cards_enabled = excluded.manage_wallet_cards_enabled, "+
		"account_enabled = excluded.account_enabled, "+
		"updated_at=now()",
		walletID, feat.SendEnabled, feat.ReceiveEnabled, feat.LinkedAccEnabled, feat.CardsEnabled, feat.BanksEnabled, feat.IdentitiesEnabled, feat.TwitterEnabled, feat.AddCardsEnabled, feat.InteraccEnabled, feat.ManageWalletCardsEnabled, feat.AccountEnabled)
	if err != nil {
		return nil, fmt.Errorf("%w %s", features.ErrInternal, err)
	}

	return Features(ctx, b, walletID)
}

func Features(ctx context.Context, b Backends, walletID string) (*features.WalletFeatures, error) {
	// Check DB for feature overrides
	var res features.WalletFeatures
	err := b.DB().GetContext(ctx, &res,
		"SELECT send_enabled, receive_enabled, linked_accounts_enabled, cards_enabled, banks_enabled, identities_enabled, twitter_enabled, add_cards_enabled, interac_enabled, manage_wallet_cards_enabled, account_enabled FROM wallet_features WHERE wallet_id=$1",
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

	w, err := b.Wallets().Get(ctx, walletID)
	if err != nil {
		return nil, err
	}
	canEnableAccount := isAccountEnabled(walletID, w.Country)
	// If you are not KYC approved you can do nothing
	if kycStatus != kyc.StatusLevel1 && kycStatus != kyc.StatusLevel2 {
		return &features.WalletFeatures{
			SendEnabled:              false,
			ReceiveEnabled:           false,
			LinkedAccEnabled:         false,
			CardsEnabled:             false,
			BanksEnabled:             false,
			IdentitiesEnabled:        false,
			TwitterEnabled:           false,
			AddCardsEnabled:          false,
			InteraccEnabled:          false,
			ManageWalletCardsEnabled: false,
			AccountEnabled:           canEnableAccount,
		}, nil
	}

	// Identities are enabled default everywhere
	res = features.WalletFeatures{
		IdentitiesEnabled: true,
		TwitterEnabled:    true,
	}

	lal, err := b.LinkedAccounts().ListByWalletId(ctx, walletID)
	if err != nil {
		return nil, err
	}

	canAddCard, err := canAddCards(lal)
	if err != nil {
		return nil, err
	}

	canAddBank, err := canAddBanks(lal)
	if err != nil {
		return nil, err
	}

	canAddInterac, err := canAddInterac(lal)
	if err != nil {
		return nil, err
	}

	if w.Country == country.US {
		res.ReceiveEnabled = true
		res.SendEnabled = true
		res.LinkedAccEnabled = true
		res.BanksEnabled = false
		res.CardsEnabled = true
		res.AddCardsEnabled = canAddCard
		res.ManageWalletCardsEnabled = false
		// it enables the feature by default for sandbox / dev
		res.AccountEnabled = canEnableAccount
	}
	if w.Country == country.ZA {
		res.ReceiveEnabled = true
		res.SendEnabled = true
		res.LinkedAccEnabled = true
		res.BanksEnabled = canAddBank
		res.CardsEnabled = false
		res.AddCardsEnabled = false
		res.ManageWalletCardsEnabled = false
		res.AccountEnabled = canEnableAccount
	}
	if country.EUCountries[w.Country] {
		res.ReceiveEnabled = true
		res.SendEnabled = true

		// eu wallets use the Gatehub ramp widget to add cards, do deposits etc.
		res.LinkedAccEnabled = false
		res.BanksEnabled = false
		res.CardsEnabled = false
		res.AddCardsEnabled = false
		res.ManageWalletCardsEnabled = false
		res.AccountEnabled = canEnableAccount
	}
	if w.Country == country.CA {
		res.ReceiveEnabled = true
		res.SendEnabled = true
		res.LinkedAccEnabled = true
		res.BanksEnabled = false
		res.CardsEnabled = false
		res.AddCardsEnabled = false
		res.InteraccEnabled = canAddInterac
		res.ManageWalletCardsEnabled = false
		res.AccountEnabled = canEnableAccount
	}

	return &res, nil
}

func canAddCards(lal []linkedaccounts.LinkedAccount) (bool, error) {
	var cnt int
	for _, la := range lal {
		if la.DeletedAt.Valid {
			continue
		}

		if la.Provider == astra.ProviderName &&
			la.Type == astra.TypeCard {
			cnt++
		}
	}

	return cnt <= 3, nil
}

// This assumes that the wallet is in ZA. The wallet can only add at most 1 bank
func canAddBanks(lal []linkedaccounts.LinkedAccount) (bool, error) {
	for _, la := range lal {
		if la.DeletedAt.Valid {
			continue
		}

		if la.Provider == xago.ProviderName &&
			la.Type == xago.AccTypeBank {
			return false, nil
		}
	}

	return true, nil
}

func canAddInterac(lal []linkedaccounts.LinkedAccount) (bool, error) {
	for _, la := range lal {
		if la.DeletedAt.Valid {
			continue
		}

		if la.Provider == chimoney.ProviderName &&
			la.Type == chimoney.AccTypeInterac {
			return false, nil
		}
	}

	return true, nil
}

func isAccountEnabled(walletID string, walletCountry country.Country) bool {
	if contains(env.GetBlockListExertions(), walletID) {
		return true
	}

	if country.EUCountries[walletCountry] && contains(env.GetBlockedRegions(), "EU") {
		return false
	}

	if contains(env.GetBlockedRegions(), walletCountry.String()) {
		return false
	}

	return true
}

func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, v := range slice {
		set[v] = struct{}{}
	}

	_, exists := set[item]
	return exists
}
