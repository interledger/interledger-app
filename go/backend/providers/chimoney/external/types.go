package external

import (
	"encoding/json"

	"gitlab.com/fynbos/backend/currency"
)

type CreateWalletReq struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
}

type APIResponse struct {
	Status string          `json:"status,omitempty"`
	Error  string          `json:"error,omitempty"`
	Data   json.RawMessage `json:"data,omitempty"`
}

type WalletResp struct {
	ID           string       `json:"id,omitempty"`
	Parent       string       `json:"parent,omitempty"`
	UID          string       `json:"uid,omitempty"`
	Name         string       `json:"name,omitempty"`
	SubAccount   bool         `json:"subAccount,omitempty"`
	Verification Verification `json:"verification,omitempty"`
}

type Verification struct {
	Status string `json:"status,omitempty"`
}

type TransferReq struct {
	SenderSubAccount   string          `json:"subAccount"`
	ReceiverSubAccount string          `json:"receiver"`
	Amount             currency.Amount `json:"amountToSend"`
}

type WithdrawalReq struct {
	Interacs            []Interacs `json:"interacs,omitempty"`
	DebitCurrency       string     `json:"debitCurrency,omitempty"`
	SubAccount          string     `json:"subAccount,omitempty"`
	TurnOffNotification bool       `json:"turnOffNotification,omitempty"`
}

type Interacs struct {
	Name      string  `json:"name,omitempty"`
	Email     string  `json:"email,omitempty"`
	Amount    float64 `json:"amount,omitempty"`
	Narration string  `json:"narration,omitempty"`
}

type DepositReq struct {
	Amount               string `json:"amount,omitempty"`
	Currency             string `json:"currency,omitempty"`
	ChimoneyWallet       string `json:"subAccount,omitempty"`
	Email                string `json:"payerEmail,omitempty"`
	TurnOffNotifications bool   `json:"turnOffNotifications,omitempty"`
}

type DepositResp struct {
	PaymentLink string `json:"paymentLink,omitempty"`
	ValueInUSD  string `json:"valueInUSD,omitempty"`
	Chimoney    string `json:"chimoney,omitempty"`
	IssueID     string `json:"issueID,omitempty"`
	Type        string `json:"type,omitempty"`
	Issuer      string `json:"issuer,omitempty"`
	PayerEmail  string `json:"payerEmail,omitempty"`
	InitiatedBy string `json:"initiatedBy,omitempty"`
	IssueDate   string `json:"issueDate,omitempty"`
	Status      string `json:"status,omitempty"`
	ChiRef      string `json:"chiRef,omitempty"`
	Error       string `json:"error,omitempty"`
}
