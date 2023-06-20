package workflows

import "gitlab.com/fynbos/backend/linkedaccounts"

type CreateLinkedCardArgs struct {
	ID         string
	WalletID   string
	ProviderID string
	Mask       string
	Name       string
	Nickname   string
	CanSend    bool
	CanReceive bool
	State      linkedaccounts.State
}

type CreateExternalCardArgs struct {
	WalletID            string
	Name                string
	CardNumber          string
	CVV                 string
	ExpirationDate      string
	RejectDuplicateCard bool
}

type QueryCard struct {
	WalletID       string
	CardNumber     string
	ExpirationDate string
	CVV            string
	AVS            bool
}
