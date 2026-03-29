// note(bradu): this is just a stub
// nothing here is tested or guaranteed to be correct, it's just a starting point for development
// needs more polishing and testing, but this is a starting point for development
package dto

import "encoding/json"

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/PaymentInformationType.java
type PaymentInformationTypeEmum string

const (
	BANK_ACCOUNT_PAYMENT PaymentInformationTypeEmum = "BANK_ACCOUNT"
	WALLET_PAYMENT       PaymentInformationTypeEmum = "WALLET"

	// ENCRYPTED_CREDIT_CARD PaymentInformationTypeEmum = "ENCRYPTED_CREDIT_CARD" // not currently supported by us, but may be added in the future
	// CRYPTO                PaymentInformationTypeEmum = "CRYPTO"
)

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/OneOfExternalPaymentInformation.java
type PaymentInformation struct {
	ID   string                     `json:"id,omitempty"`
	Type PaymentInformationTypeEmum `json:"type,omitempty"`

	*BankAccount
}

// controller expects that all structs that are sent in the body of a request implement MarshalJSON and UnmarshalJSON
func (pi PaymentInformation) MarshalJSON() ([]byte, error) {
	type bankAccountAlias BankAccount
	type alias struct {
		ID   string                     `json:"id,omitempty"`
		Type PaymentInformationTypeEmum `json:"type,omitempty"`
		*bankAccountAlias
	}
	var ba *bankAccountAlias
	if pi.BankAccount != nil {
		a := bankAccountAlias(*pi.BankAccount)
		ba = &a
	}
	return json.Marshal(alias{ID: pi.ID, Type: pi.Type, bankAccountAlias: ba})
}

// controller expects that all structs that are sent in the body of a request implement MarshalJSON and UnmarshalJSON
func (pi *PaymentInformation) UnmarshalJSON(data []byte) error {
	var base struct {
		ID   string                     `json:"id"`
		Type PaymentInformationTypeEmum `json:"type"`
	}
	if err := json.Unmarshal(data, &base); err != nil {
		return err
	}
	pi.ID = base.ID
	pi.Type = base.Type
	pi.BankAccount = &BankAccount{}
	return json.Unmarshal(data, pi.BankAccount)
}

type PaymentInformationOptions func(*PaymentInformation)

func WithPaymentInformationID(id string) PaymentInformationOptions {
	return func(pi *PaymentInformation) {
		pi.ID = id
	}
}

func WithPaymentInformationType(t PaymentInformationTypeEmum) PaymentInformationOptions {
	return func(pi *PaymentInformation) {
		pi.Type = t
	}
}

func WithBankAccount(bankAccount *BankAccount) PaymentInformationOptions {
	return func(pi *PaymentInformation) {
		pi.BankAccount = bankAccount
	}
}

func NewPaymentInformation(opts ...PaymentInformationOptions) *PaymentInformation {
	pi := &PaymentInformation{}
	for _, opt := range opts {
		opt(pi)
	}
	return pi
}

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/BankAccountPaymentInformationBankAccountType.java
type BankAccountTypeEnum string

const (
	CHECKING BankAccountTypeEnum = "CHECKING"
	// SAVINGS  BankAccountTypeEnum = "SAVINGS"
)

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/BankAccount.java
type BankAccount struct {
	BankRoutingNumber     string              `json:"bankRoutingNumber,omitempty"`
	BankAccountType       BankAccountTypeEnum `json:"bankAccountType,omitempty"`
	BankAccountNumner     string              `json:"bankAccountNumber,omitempty"`
	BankSwiftCode         string              `json:"bankSwiftCode,omitempty"`
	BankRoutingCheckDigit string              `json:"bankRoutingCheckDigit,omitempty"`
	AccountBankName       string              `json:"accountBankName,omitempty"`
	AccountHolderName     string              `json:"accountHolderName,omitempty"`

	// PlaidProcessorToken string `json:"plaidProcessorToken,omitempty"` // not currently supported by us, but may be added in the future
}

// controller expects that all structs that are sent in the body of a request implement MarshalJSON and UnmarshalJSON
func (ba BankAccount) MarshalJSON() ([]byte, error) {
	type alias BankAccount
	return json.Marshal(alias(ba))
}

// controller expects that all structs that are sent in the body of a request implement MarshalJSON and UnmarshalJSON
func (ba *BankAccount) UnmarshalJSON(data []byte) error {
	type alias BankAccount
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*ba = BankAccount(a)
	return nil
}

type BankAccountOptions func(*BankAccount)

func WithAccountNumber(number string) BankAccountOptions {
	return func(bapi *BankAccount) {
		bapi.BankAccountNumner = number
	}
}

func WithAccountType(t BankAccountTypeEnum) BankAccountOptions {
	return func(bapi *BankAccount) {
		bapi.BankAccountType = t
	}
}

func WithSwiftCode(code string) BankAccountOptions {
	return func(bapi *BankAccount) {
		bapi.BankSwiftCode = code
	}
}

func WithRoutingNumber(number string) BankAccountOptions {
	return func(bapi *BankAccount) {
		bapi.BankRoutingNumber = number
	}
}

func WithRoutingCheckDigit(digit string) BankAccountOptions {
	return func(bapi *BankAccount) {
		bapi.BankRoutingCheckDigit = digit
	}
}

func WithBankName(name string) BankAccountOptions {
	return func(bapi *BankAccount) {
		bapi.AccountBankName = name
	}
}

func WithAccountHolderName(name string) BankAccountOptions {
	return func(bapi *BankAccount) {
		bapi.AccountHolderName = name
	}
}

func NewBankAccount(opts ...BankAccountOptions) *BankAccount {
	bapi := &BankAccount{}
	for _, opt := range opts {
		opt(bapi)
	}
	return bapi
}
