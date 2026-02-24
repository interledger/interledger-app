package external

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"gitlab.com/fynbos/backend/currency"
)

// FlexibleFloat handles JSON fields that can be either string or float64
type FlexibleFloat float64

func (f *FlexibleFloat) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case float64:
		*f = FlexibleFloat(value)
	case string:
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float string: %w", err)
		}
		*f = FlexibleFloat(parsed)
	default:
		return fmt.Errorf("amount must be string or number, got %T", v)
	}

	return nil
}

func (f FlexibleFloat) Float64() float64 {
	return float64(f)
}

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
	SenderSubAccount    string          `json:"subAccount"`
	ReceiverSubAccount  string          `json:"receiver"`
	Amount              currency.Amount `json:"amountToSend"`
	TurnOffNotification bool            `json:"turnOffNotification,omitempty"`
}

type WithdrawalReq struct {
	Interacs            []Interacs `json:"interacs,omitempty"`
	DebitCurrency       string     `json:"debitCurrency,omitempty"`
	SubAccount          string     `json:"subAccount,omitempty"`
	TurnOffNotification bool       `json:"turnOffNotification,omitempty"`
}

type WithdrawResponse struct {
	Data        []WithdrawData `json:"data"`
	Error       string         `json:"error"`
	PaymentLink string         `json:"paymentLink"`
}

type WithdrawData struct {
	ID       string  `json:"id"`
	IssueID  string  `json:"issueID"`
	Amount   float64 `json:"amount"`
	Fee      float64 `json:"fee"`
	Currency string  `json:"debitCurrency"`
	Type     string  `json:"type"`
	ChiRef   string  `json:"chiref"`
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
	TurnOffNotifications bool    `json:"turnOffNotification,omitempty"`
	RedirectURL          string  `json:"redirect_url,omitempty"`
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
	Meta                PaymentMeta `json:"meta"`
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
	PaymentType         string      `json:"paymentType,omitempty"`
}

type PaymentMeta struct {
	Amount        FlexibleFloat         `json:"amount"`
	ProcessingFee *PaymentProcessingFee `json:"processingFee,omitempty"`
}

type PaymentProcessingFee struct {
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	GrossAmount float64 `json:"grossAmount"`
	NetAmount   float64 `json:"netAmount"`
	Provider    string  `json:"provider"`
}

type VerifyPaymentReq struct {
	IssueID   string `json:"id,omitempty"`
	ChiWallet string `json:"subAccount,omitempty"`
}

type PayoutStatusRequest struct {
	ChiWallet           string `json:"subAccount,omitempty"`
	TurnOffNotification bool   `json:"turnOffNotification,omitempty"`
	Reference           string `json:"chiRef"`
}

type PayoutStatusResponse struct {
	ID      string  `json:"id"`
	Amount  float64 `json:"amount"`
	Fee     float64 `json:"fee"`
	Type    string  `json:"type"`
	IssueID string  `json:"issueID"`
	Status  string  `json:"status"`
}
