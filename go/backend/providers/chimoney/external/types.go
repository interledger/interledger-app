package external

import (
	"encoding/json"

	"gitlab.com/fynbos/backend/currency"
)

type CreateSubAccountReq struct {
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

type CreateSubAccountResp struct {
	ID         string `json:"id,omitempty"`
	Parent     string `json:"parent,omitempty"`
	UID        string `json:"uid,omitempty"`
	Name       string `json:"name,omitempty"`
	SubAccount bool   `json:"subAccount,omitempty"`
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
	Name      string `json:"name,omitempty"`
	Email     string `json:"email,omitempty"`
	Amount    int    `json:"amount,omitempty"`
	Narration string `json:"narration,omitempty"`
}

type DepositReq struct {
	Amount       currency.Currency
	SubAccountID string
	Email        string
}
