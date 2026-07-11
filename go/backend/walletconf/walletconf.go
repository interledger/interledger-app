// Package walletconf declares the "wallet" entity type for entityconf and
// registers a Confs struct duplicating the capabilities of the existing
// wallet_features flags (see go/backend/features). This is intentionally a
// parallel, independent surface: nothing here reads from or writes to
// wallet_features, and nothing outside the admin portal reads from this
// package's Store yet — no application decision-making has been moved onto
// it.
//
// Defaults below mirror wallet_features' schema.hcl column defaults (all
// false except accounts_tab_enabled). They are a placeholder: the real,
// per-flag default review (and the migration off wallet_features) is
// separate, future work.
package walletconf

import "github.com/interledger/interledger-app/go/backend/entityconf"

const EntityWallet entityconf.EntityType = "wallet"

type Confs struct {
	SendEnabled bool `conf:"wallet.send_enabled" default:"false" display:"Send" desc:"Allows a wallet to send payments"`

	ReceiveEnabled bool `conf:"wallet.receive_enabled" default:"false" display:"Receive" desc:"Allows a wallet to receive payments"`

	LinkedAccountsEnabled bool `conf:"wallet.linked_accounts_enabled" default:"false" display:"Linked Accounts" desc:"Allows a wallet to link external accounts"`

	CardsEnabled bool `conf:"wallet.cards_enabled" default:"false" display:"Cards" desc:"Allows a wallet to use cards"`

	BanksEnabled bool `conf:"wallet.banks_enabled" default:"false" display:"Banks" desc:"Allows a wallet to link bank accounts"`

	IdentitiesEnabled bool `conf:"wallet.identities_enabled" default:"false" display:"Identities" desc:"Allows a wallet to manage identities"`

	TwitterEnabled bool `conf:"wallet.twitter_enabled" default:"false" display:"Twitter" desc:"Allows a wallet to link a Twitter account"`

	AddCardsEnabled bool `conf:"wallet.add_cards_enabled" default:"false" display:"Add Cards" desc:"Allows a wallet to add new cards"`

	InteracEnabled bool `conf:"wallet.interac_enabled" default:"false" display:"Interac" desc:"Allows a wallet to use Interac"`

	ManageWalletCardsEnabled bool `conf:"wallet.manage_wallet_cards_enabled" default:"false" display:"Manage Wallet Cards" desc:"Allows a wallet to manage its cards"`

	AccountsTabEnabled bool `conf:"wallet.accounts_tab_enabled" default:"true" display:"Accounts Tab" desc:"Shows the accounts tab for a wallet"`

	DeleteAccountEnabled bool `conf:"wallet.delete_account_enabled" default:"false" display:"Delete Account" desc:"Allows a wallet owner to delete their account"`
}

func init() {
	entityconf.MustRegister(EntityWallet, Confs{})
}
