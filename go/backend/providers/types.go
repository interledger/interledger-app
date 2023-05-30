package providers

import (
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/transactions"
)

type TransfersArgs struct {
	FromForeignID       string // ID against which to create transfers
	ToForeignID         string // ID against which to create transfers
	FromPaymentPointer  string // Fully qualified payment pointer
	ToPaymentPointer    string // Fully qualified payment pointer
	FromLinkedAccountID string `validate:"uuid"`
	ToLinkedAccountID   string `validate:"uuid"`
	FromWalletID        string `validate:"uuid"`
	ToWalletID          string `validate:"uuid"`
	Amount              currency.Amount
	FromTransactionID   string
	IPAddress           string
	ThreeDSID           string
	ForceEDD            bool // Used for testing to force the sending of enhanced due diligence info to providers
	ForceNoEDD          bool // Used for testing to force not sending EDD
}

type TransferResponse struct {
	Type                       WorkflowKey
	OutgoingTransferState      transactions.State
	OutgoingTransferExternalID string
	IncomingTransferState      transactions.State
	IncomingTransferExternalID string
}

type WorkflowKey string

const (
	GMTACH2ACH     WorkflowKey = "gmt_ach_2_ach"
	GMTCARD2ACH    WorkflowKey = "gmt_card_2_ach"
	GMTACH2CARD    WorkflowKey = "gmt_ach_2_card"
	GMTUNSUPPORTED WorkflowKey = "gmt_unsupported"
)
