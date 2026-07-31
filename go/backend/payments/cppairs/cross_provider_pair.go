package cppairs

import (
	"context"
	"fmt"

	"github.com/interledger/interledger-app/go/backend/features"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	"github.com/interledger/interledger-app/go/backend/payments"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub"
	"github.com/interledger/interledger-app/go/backend/providers/xago"
)

// crossProviderPairKey key used for lookup into crossProviderPairs map
type crossProviderPairKey struct {
	ProviderA string
	ProviderB string
}

var crossProviderPairs = map[crossProviderPairKey]CrossProviderPair{
	{ProviderA: xago.ProviderName, ProviderB: gatehub.ProviderName}: &XagoGatehubPair{},
}

// Backends is the subset of ops.Backends needed here. Defined locally
// to avoid an import cycle with the ops package.
type Backends interface {
	Features() features.Client
}

// CrossProviderPair is an interface used to enable behaviors specific to particular provider pairs.
type CrossProviderPair interface {
	IsCurrencySupported(senderAcc, receiverAcc *linkedaccounts.LinkedAccount) bool
	IsFeatureFlagEnabled(senderFeats, receiverFeats *features.WalletFeatures) bool
	IsAccTypeSupported(senderAcc, receiverAcc *linkedaccounts.LinkedAccount) bool
}

// LookupCrossProviderPair looks up the CrossProviderPair for the given accounts, matching
// either sender-receiver or receiver-sender provider order.
func LookupCrossProviderPair(senderAcc, receiverAcc *linkedaccounts.LinkedAccount) CrossProviderPair {
	return LookupCrossProviderPairByProviderNames(senderAcc.Provider, receiverAcc.Provider)
}

func LookupCrossProviderPairByProviderNames(senderProviderName, receiverProviderName string) CrossProviderPair {
	// sender-receiver order
	key := crossProviderPairKey{ProviderA: senderProviderName, ProviderB: receiverProviderName}
	if pair, ok := crossProviderPairs[key]; ok {
		return pair
	}

	// receiver-sender order
	key = crossProviderPairKey{ProviderA: receiverProviderName, ProviderB: senderProviderName}
	pair := crossProviderPairs[key]
	return pair
}

// IsCrossProviderPair returns true if accounts have different providers
func IsCrossProviderPair(senderAcc, receiverAcc *linkedaccounts.LinkedAccount) bool {
	return senderAcc.Provider != receiverAcc.Provider
}

// IsSpecificCrossProviderPair returns true if the accounts' providers match providerA and providerB,
// in either order.
func IsSpecificCrossProviderPair(senderAcc, receiverAcc *linkedaccounts.LinkedAccount, providerA, providerB string) bool {
	return (senderAcc.Provider == providerA && receiverAcc.Provider == providerB) ||
		(senderAcc.Provider == providerB && receiverAcc.Provider == providerA)
}

// IsSupportedCrossProviderPair returns true if the accounts have different providers between which we can route payments
func IsSupportedCrossProviderPair(senderAcc, receiverAcc *linkedaccounts.LinkedAccount) bool {
	cppair := LookupCrossProviderPair(senderAcc, receiverAcc)
	return cppair != nil
}

// ValidateAccounts performs cross-provider specific validations.
// Includes compatibility and feature flag checks, so it is somewhat slower than the other validations.
func ValidateAccounts(ctx context.Context, b Backends, typ payments.Type, senderAcc, receiverAcc *linkedaccounts.LinkedAccount) error {
	if !IsCrossProviderPair(senderAcc, receiverAcc) {
		return fmt.Errorf("%w: %s", payments.ErrIncompatibleAccounts, "cp validation: called with same provider accounts")
	}

	if !IsSupportedCrossProviderPair(senderAcc, receiverAcc) {
		return fmt.Errorf("%w: %s", payments.ErrIncompatibleAccounts, "cp validation: provider pair not supported")
	}

	if typ != payments.TypePeer2Peer {
		return fmt.Errorf("%w: %s", payments.ErrIncompatibleAccounts, "cp validation: payment type not supported")
	}

	err := ValidateAccountCompatibility(senderAcc, receiverAcc)
	if err != nil {
		return err
	}

	// Feature flag checks
	err = ValidateWalletFlags(ctx, b, senderAcc, receiverAcc)
	if err != nil {
		return err
	}

	return nil
}

// ValidateAccountCompatibility validates whether the two accounts are compatible for cross provider payments.
// This is the cross-provider equivalent of la.CanPay and it returns the reason for incompatibility.
// It checks the supported currency, state, can send/receive, account type.
// It does not do feature flag checks. Use ValidateAccounts or ValidateWalletFlags for that.
func ValidateAccountCompatibility(senderAcc *linkedaccounts.LinkedAccount, receiverAcc *linkedaccounts.LinkedAccount) error {
	cpPair := LookupCrossProviderPair(senderAcc, receiverAcc)
	if cpPair == nil {
		return fmt.Errorf("%w: %s", payments.ErrIncompatibleAccounts, "cp compat validation: unsupported cross provider pair")
	}

	if !cpPair.IsCurrencySupported(senderAcc, receiverAcc) {
		return fmt.Errorf("%w: %s", payments.ErrIncompatibleAccounts, "cp compat validation: currency pair not supported")
	}

	if !isStateValid(senderAcc, receiverAcc) {
		return fmt.Errorf("%w: %s", payments.ErrIncompatibleAccounts, "cp compat validation: invalid accounts state")
	}

	if !canSendAndReceive(senderAcc, receiverAcc) {
		return fmt.Errorf("%w: %s", payments.ErrIncompatibleAccounts, "cp compat validation: sender can't send or receiver can't receive")
	}

	if !cpPair.IsAccTypeSupported(senderAcc, receiverAcc) {
		return fmt.Errorf("%w: %s", payments.ErrIncompatibleAccounts, "cp compat validation: account type is not supported")
	}

	return nil
}

// ValidateWalletFlags checks for feature flags to be enabled for both accounts.
func ValidateWalletFlags(ctx context.Context, b Backends, senderAcc *linkedaccounts.LinkedAccount, receiverAcc *linkedaccounts.LinkedAccount) error {
	cpPair := LookupCrossProviderPair(senderAcc, receiverAcc)
	if cpPair == nil {
		return fmt.Errorf("%w: %s", payments.ErrIncompatibleAccounts, "cp feature flag validation: unsupported cross provider pair")
	}

	senderFeats, err := b.Features().Features(ctx, senderAcc.WalletID)
	if err != nil {
		return err
	}
	receiverFeats, err := b.Features().Features(ctx, receiverAcc.WalletID)
	if err != nil {
		return err
	}

	if !cpPair.IsFeatureFlagEnabled(senderFeats, receiverFeats) {
		return fmt.Errorf("%w: %s", payments.ErrIncompatibleAccounts, "cp feature flag validation: XagoGatehubPaymentsEnabled is false for the sender or the receiver")
	}

	return nil
}

// isStateValid returns true if the state of the accounts is verified
func isStateValid(senderAcc, receiverAcc *linkedaccounts.LinkedAccount) bool {
	if senderAcc.State == linkedaccounts.Verified && receiverAcc.State == linkedaccounts.Verified {
		return true
	}

	return false
}

// canSendAndReceive returns true if the sender account can send and the receiver account can receive.
func canSendAndReceive(senderAcc, receiverAcc *linkedaccounts.LinkedAccount) bool {
	if senderAcc.CanSend && receiverAcc.CanReceive {
		return true
	}

	return false
}
