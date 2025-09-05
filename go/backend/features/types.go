package features

type WalletFeatures struct {
	SendEnabled              bool `db:"send_enabled"`
	ReceiveEnabled           bool `db:"receive_enabled"`
	LinkedAccEnabled         bool `db:"linked_accounts_enabled"`
	CardsEnabled             bool `db:"cards_enabled"`
	BanksEnabled             bool `db:"banks_enabled"`
	IdentitiesEnabled        bool `db:"identities_enabled"`
	TwitterEnabled           bool `db:"twitter_enabled"`
	AddCardsEnabled          bool `db:"add_cards_enabled"`
	InteraccEnabled          bool `db:"interac_enabled"`
	ManageWalletCardsEnabled bool `db:"manage_wallet_cards_enabled"`
	AccountEnabled           bool
}
