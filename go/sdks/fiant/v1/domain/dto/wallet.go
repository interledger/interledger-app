package dto

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/resources/wallets/requests/WalletCreation.java#L88
type WalletTypeEnum string

const (
	// Fiant has "WALLET" as the only type
	// and is hadrcoded in the Java SDK, so we do the same here for now
	// see the type definition above
	WalletTypeStandard WalletTypeEnum = "WALLET"
)

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/Wallet.java
// the balance values are defined as float64, as the Java SDK defines them as doubles
// it's up to the caller to deal with it
type Wallet struct {
	ID       string `json:"id,omitempty"`
	Label    string `json:"label,omitempty"`
	Currency string `json:"currency,omitempty"`

	AvailableBalance float64 `json:"availableBalance,omitempty"`
	LockedBalance    float64 `json:"lockedBalance,omitempty"`
	PendingBalance   float64 `json:"pendingBalance,omitempty"`
	TotalBalance     float64 `json:"totalBalance,omitempty"`

	Type WalletTypeEnum `json:"type,omitempty"`
}

type WalletOption func(*Wallet)

func WithWalletLabel(label string) WalletOption {
	return func(w *Wallet) {
		w.Label = label
	}
}

func WithWalletCurrency(currency string) WalletOption {
	return func(w *Wallet) {
		w.Currency = currency
	}
}

func WithWalletTypeStandard() WalletOption {
	return func(w *Wallet) {
		w.Type = WalletTypeStandard
	}
}

func NewWallet(opts ...WalletOption) Wallet {
	w := Wallet{}
	for _, opt := range opts {
		opt(&w)
	}

	return w
}
