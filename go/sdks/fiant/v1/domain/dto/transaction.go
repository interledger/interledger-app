package dto

import "encoding/json"

type TransactionOption func(*Transaction)

func WithTransactionID(id string) TransactionOption {
	return func(t *Transaction) {
		t.ID = id
	}
}

func WithTransactionType(transactionType TransactionTypeEnum) TransactionOption {
	return func(t *Transaction) {
		t.TransactionType = transactionType
	}
}

func WithTransactionGroup(transactionGroup string) TransactionOption {
	return func(t *Transaction) {
		t.TransactionGroup = transactionGroup
	}
}

func WithSubClientID(subClientID string) TransactionOption {
	return func(t *Transaction) {
		t.SubClientID = subClientID
	}
}

func WithTransactionTotal(transactionTotal *Total) TransactionOption {
	return func(t *Transaction) {
		t.TransactionTotal = transactionTotal
	}
}

func WithUSDValue(usdValue float64) TransactionOption {
	return func(t *Transaction) {
		t.USDValue = usdValue
	}
}

func WithAmount(amount float64) TransactionOption {
	return func(t *Transaction) {
		t.Amount = amount
	}
}

func WithDate(date string) TransactionOption {
	return func(t *Transaction) {
		t.Date = date
	}
}

func WithInitiator(initiator User) TransactionOption {
	return func(t *Transaction) {
		t.Initiator = initiator
	}
}

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/TransactionType.java
type Transaction struct {
	ID string `json:"id,omitempty"`

	TransactionType TransactionTypeEnum `json:"type,omitempty"`

	TransactionGroup string `json:"transactionGroupId,omitempty"`
	SubClientID      string `json:"subClientId,omitempty"`

	TransactionTotal *Total `json:"total,omitempty"`

	USDValue float64 `json:"usdValue,omitempty"`
	Amount   float64 `json:"amount,omitempty"`
	Date     string  `json:"date,omitempty"`

	Initiator User `json:"initiator,omitempty"`
}

func NewTransaction(opts ...TransactionOption) *Transaction {
	t := &Transaction{}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func (t Transaction) MarshalJSON() ([]byte, error) {
	type alias Transaction
	return json.Marshal(alias(t))
}

func (t *Transaction) UnmarshalJSON(data []byte) error {
	type alias Transaction
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*t = Transaction(a)
	return nil
}

type DepositToWalletOption func(*DepositToWallet)

func WithSourceMethod(sourceMethod PaymentMethod) DepositToWalletOption {
	return func(d *DepositToWallet) {
		d.SourceMethod = sourceMethod
	}
}

func WithDestinationMethod(destinationMethod PaymentMethod) DepositToWalletOption {
	return func(d *DepositToWallet) {
		d.DestinationMethod = destinationMethod
	}
}

type DepositToWallet struct {
	Type TransactionTypeEnum `json:"type"`

	Transaction

	SourceMethod      PaymentMethod `json:"sourceMethod,omitempty"`
	DestinationMethod PaymentMethod `json:"destinationMethod,omitempty"`
}

func NewDepositToWallet(opts ...DepositToWalletOption) *DepositToWallet {
	d := &DepositToWallet{}
	d.Type = DEPOSIT
	for _, opt := range opts {
		opt(d)
	}
	return d
}
