package external

import (
	"encoding/json"
	"time"

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
	ValueInUSD           float64 `json:"valueInUSD,omitempty"`
	Amount               string  `json:"amount,omitempty"`
	Currency             string  `json:"currency,omitempty"`
	ChimoneyWallet       string  `json:"subAccount,omitempty"`
	Email                string  `json:"payerEmail,omitempty"`
	TurnOffNotifications bool    `json:"turnOffNotifications,omitempty"`
}

type DepositResp struct {
	PaymentLink string  `json:"paymentLink,omitempty"`
	ValueInUSD  float64 `json:"valueInUSD,omitempty"`
	Chimoney    float64 `json:"chimoney,omitempty"`
	IssueID     string  `json:"issueID,omitempty"`
	Type        string  `json:"type,omitempty"`
	Issuer      string  `json:"issuer,omitempty"`
	PayerEmail  string  `json:"payerEmail,omitempty"`
	InitiatedBy string  `json:"initiatedBy,omitempty"`
	IssueDate   string  `json:"issueDate,omitempty"`
	Status      string  `json:"status,omitempty"`
	ChiRef      string  `json:"chiRef,omitempty"`
	Error       string  `json:"error,omitempty"`
}

type ConvertToUSDRequest struct {
	Currency string `json:"originCurrency"`
	Amount   int    `json:"amountInOriginCurrency"`
}

type ConvertToUSDResponse struct {
	OriginCurrency     string  `json:"originCurrency"`
	OriginAmount       string  `json:"amountInOriginCurrency"`
	AmountInUSD        float64 `json:"amountInUSD"`
	ValidUntil         string  `json:"validUntil"`
	ExpiresAt          string  `json:"expiresAt"`
	ExpiresAtTimestamp int     `json:"expiresAtTimestamp"`
}

type Integration struct {
	AppID       string `json:"appID"`
	IsFromScrim bool   `json:"isFromScrim"`
}

type RedeemData struct {
	ValueInUSD int     `json:"valueInUSD"`
	WalletID   string  `json:"walletID"`
	Amount     string  `json:"amount"`
	Currency   string  `json:"currency"`
	InteracFee float64 `json:"interacFee"`
	PayerEmail string  `json:"payerEmail"`
	SubAccount string  `json:"subAccount"`
}

type Payment struct {
	ID                  string      `json:"id"`
	ValueInUSD          int         `json:"valueInUSD"`
	Amount              string      `json:"amount"`
	IssueID             string      `json:"issueID"`
	Fee                 int         `json:"fee"`
	Type                string      `json:"type"`
	SubAccount          string      `json:"subAccount"`
	Issuer              string      `json:"issuer"`
	TID                 int         `json:"t_id"`
	ChiRef              string      `json:"chiRef"`
	Meta                interface{} `json:"meta"`
	Integration         Integration `json:"integration"`
	Currency            string      `json:"currency"`
	InteracFee          float64     `json:"interacFee"`
	TurnOffNotification bool        `json:"turnOffNotification"`
	IssueDate           time.Time   `json:"issueDate"`
	PayerEmail          string      `json:"payerEmail"`
	RedeemData          RedeemData  `json:"redeemData"`
	InitiatedBy         string      `json:"initiatedBy"`
	RedirectURL         string      `json:"redirect_url"`
	Status              string      `json:"status"`
}

type VerifyPaymentReq struct {
	IssueID   string `json:"id,omitempty"`
	ChiWallet string `json:"subAccount,omitempty"`
}
