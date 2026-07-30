package cppairs

import (
	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/features"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub"
	"github.com/interledger/interledger-app/go/backend/providers/xago"
)

type XagoGatehubPair struct {
}

func (pair *XagoGatehubPair) IsCurrencySupported(senderAcc, receiverAcc *linkedaccounts.LinkedAccount) bool {
	if senderAcc.Provider == xago.ProviderName &&
		senderAcc.SendCurrency == currency.ZAR && receiverAcc.ReceiveCurrency == currency.EUR {
		return true
	}

	if senderAcc.Provider == gatehub.ProviderName &&
		senderAcc.SendCurrency == currency.EUR && receiverAcc.ReceiveCurrency == currency.ZAR {
		return true
	}

	return false
}

func (pair *XagoGatehubPair) IsFeatureFlagEnabled(senderFeats, receiverFeats *features.WalletFeatures) bool {
	if senderFeats.XagoGatehubPaymentsEnabled && receiverFeats.XagoGatehubPaymentsEnabled {
		return true
	}

	return false
}

func (pair *XagoGatehubPair) IsAccTypeSupported(senderAcc, receiverAcc *linkedaccounts.LinkedAccount) bool {
	if pair.isAccTypeSupportedSingleAcct(senderAcc) && pair.isAccTypeSupportedSingleAcct(receiverAcc) {
		return true
	}

	return false
}

// isSupportedAccType returns true if the accounts are of type balance.
func (pair *XagoGatehubPair) isAccTypeSupportedSingleAcct(acc *linkedaccounts.LinkedAccount) bool {
	if acc.Provider == xago.ProviderName && acc.Type == xago.AccTypeBalance {
		return true
	}

	if acc.Provider == gatehub.ProviderName && acc.Type == gatehub.AccTypeBalance {
		return true
	}

	return false
}
