package chimoney

import "github.com/interledger/interledger-app/go/backend/currency"

const (
	ProviderName   = "chimoney"
	AccTypeBalance = "balance"
	AccTypeInterac = "interac"

	LedgerIDCAD   uint32 = 24466639 // Spells chimoney on a Nokia 3320 keyboard
	CADOpsAccount        = "a99a7194-e446-46e0-baad-49eda340b7c9"
)

type Balance struct {
	Total     currency.Amount
	Available currency.Amount
}

type TransferArgs struct {
	SendingWalletID   string
	ReceivingWalletID string
	Amount            currency.Amount
}
