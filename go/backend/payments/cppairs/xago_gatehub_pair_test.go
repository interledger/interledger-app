package cppairs

import (
	"testing"

	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/features"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub"
	"github.com/interledger/interledger-app/go/backend/providers/xago"
	"github.com/stretchr/testify/assert"
)

func TestXagoGatehubPair_IsCurrencySupported(t *testing.T) {
	pair := &XagoGatehubPair{}

	tests := []struct {
		name     string
		sender   linkedaccounts.LinkedAccount
		receiver linkedaccounts.LinkedAccount
		want     bool
	}{
		{
			name:     "xago sender ZAR to gatehub receiver EUR",
			sender:   linkedaccounts.LinkedAccount{Provider: xago.ProviderName, SendCurrency: currency.ZAR},
			receiver: linkedaccounts.LinkedAccount{Provider: gatehub.ProviderName, ReceiveCurrency: currency.EUR},
			want:     true,
		},
		{
			name:     "gatehub sender EUR to xago receiver ZAR",
			sender:   linkedaccounts.LinkedAccount{Provider: gatehub.ProviderName, SendCurrency: currency.EUR},
			receiver: linkedaccounts.LinkedAccount{Provider: xago.ProviderName, ReceiveCurrency: currency.ZAR},
			want:     true,
		},
		{
			name:     "xago sender wrong send currency",
			sender:   linkedaccounts.LinkedAccount{Provider: xago.ProviderName, SendCurrency: currency.EUR},
			receiver: linkedaccounts.LinkedAccount{Provider: gatehub.ProviderName, ReceiveCurrency: currency.EUR},
			want:     false,
		},
		{
			name:     "xago sender wrong receive currency",
			sender:   linkedaccounts.LinkedAccount{Provider: xago.ProviderName, SendCurrency: currency.ZAR},
			receiver: linkedaccounts.LinkedAccount{Provider: gatehub.ProviderName, ReceiveCurrency: currency.ZAR},
			want:     false,
		},
		{
			name:     "gatehub sender wrong send currency",
			sender:   linkedaccounts.LinkedAccount{Provider: gatehub.ProviderName, SendCurrency: currency.ZAR},
			receiver: linkedaccounts.LinkedAccount{Provider: xago.ProviderName, ReceiveCurrency: currency.ZAR},
			want:     false,
		},
		{
			name:     "gatehub sender wrong receive currency",
			sender:   linkedaccounts.LinkedAccount{Provider: gatehub.ProviderName, SendCurrency: currency.EUR},
			receiver: linkedaccounts.LinkedAccount{Provider: xago.ProviderName, ReceiveCurrency: currency.EUR},
			want:     false,
		},
		{
			name:     "sender is neither xago nor gatehub",
			sender:   linkedaccounts.LinkedAccount{Provider: "other", SendCurrency: currency.ZAR},
			receiver: linkedaccounts.LinkedAccount{Provider: gatehub.ProviderName, ReceiveCurrency: currency.EUR},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pair.IsCurrencySupported(&tt.sender, &tt.receiver)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestXagoGatehubPair_IsFeatureFlagEnabled(t *testing.T) {
	pair := &XagoGatehubPair{}

	tests := []struct {
		name     string
		sender   features.WalletFeatures
		receiver features.WalletFeatures
		want     bool
	}{
		{
			name:     "both enabled",
			sender:   features.WalletFeatures{XagoGatehubPaymentsEnabled: true},
			receiver: features.WalletFeatures{XagoGatehubPaymentsEnabled: true},
			want:     true,
		},
		{
			name:     "sender disabled",
			sender:   features.WalletFeatures{XagoGatehubPaymentsEnabled: false},
			receiver: features.WalletFeatures{XagoGatehubPaymentsEnabled: true},
			want:     false,
		},
		{
			name:     "receiver disabled",
			sender:   features.WalletFeatures{XagoGatehubPaymentsEnabled: true},
			receiver: features.WalletFeatures{XagoGatehubPaymentsEnabled: false},
			want:     false,
		},
		{
			name:     "both disabled",
			sender:   features.WalletFeatures{XagoGatehubPaymentsEnabled: false},
			receiver: features.WalletFeatures{XagoGatehubPaymentsEnabled: false},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pair.IsFeatureFlagEnabled(&tt.sender, &tt.receiver)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestXagoGatehubPair_IsAccTypeSupported(t *testing.T) {
	pair := &XagoGatehubPair{}

	tests := []struct {
		name     string
		sender   linkedaccounts.LinkedAccount
		receiver linkedaccounts.LinkedAccount
		want     bool
	}{
		{
			name:     "xago balance to gatehub balance",
			sender:   linkedaccounts.LinkedAccount{Provider: xago.ProviderName, Type: xago.AccTypeBalance},
			receiver: linkedaccounts.LinkedAccount{Provider: gatehub.ProviderName, Type: gatehub.AccTypeBalance},
			want:     true,
		},
		{
			name:     "gatehub balance to xago balance",
			sender:   linkedaccounts.LinkedAccount{Provider: gatehub.ProviderName, Type: gatehub.AccTypeBalance},
			receiver: linkedaccounts.LinkedAccount{Provider: xago.ProviderName, Type: xago.AccTypeBalance},
			want:     true,
		},
		{
			name:     "sender not balance type",
			sender:   linkedaccounts.LinkedAccount{Provider: xago.ProviderName, Type: xago.AccTypeBank},
			receiver: linkedaccounts.LinkedAccount{Provider: gatehub.ProviderName, Type: gatehub.AccTypeBalance},
			want:     false,
		},
		{
			name:     "receiver not balance type",
			sender:   linkedaccounts.LinkedAccount{Provider: xago.ProviderName, Type: xago.AccTypeBalance},
			receiver: linkedaccounts.LinkedAccount{Provider: gatehub.ProviderName, Type: "other"},
			want:     false,
		},
		{
			name:     "account provider is neither xago nor gatehub",
			sender:   linkedaccounts.LinkedAccount{Provider: "other", Type: xago.AccTypeBalance},
			receiver: linkedaccounts.LinkedAccount{Provider: gatehub.ProviderName, Type: gatehub.AccTypeBalance},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pair.IsAccTypeSupported(&tt.sender, &tt.receiver)
			assert.Equal(t, tt.want, got)
		})
	}
}
