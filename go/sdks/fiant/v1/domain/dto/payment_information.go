// note(bradu): this is just a stub
// nothing here is tested or guaranteed to be correct, it's just a starting point for development
// needs more polishing and testing, but this is a starting point for development
package dto

import "encoding/json"

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/PaymentInformationType.java
type PaymentInformationTypeEmum string

const (
	BANK_ACCOUNT PaymentInformationTypeEmum = "BANK_ACCOUNT"
	// ENCRYPTED_CREDIT_CARD PaymentInformationTypeEmum = "ENCRYPTED_CREDIT_CARD" // not currently supported by us, but may be added in the future
	// CRYPTO                PaymentInformationTypeEmum = "CRYPTO"
	// WALLET                PaymentInformationTypeEmum = "WALLET"
)

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/OneOfExternalPaymentInformation.java
// note that this is composite and we support bank accounts for now
type PaymentInformation struct {
	ID   string                     `json:"id,omitempty"`
	Type PaymentInformationTypeEmum `json:"type,omitempty"`

	BankAccountPaymentInformation
}

func (pi PaymentInformation) MarshalJSON() ([]byte, error) {
	type alias PaymentInformation
	return json.Marshal(alias(pi))
}

func (pi *PaymentInformation) UnmarshalJSON(data []byte) error {
	type alias PaymentInformation
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*pi = PaymentInformation(a)
	return nil
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

func WithBankAccountPaymentInformation(bankAccount *BankAccountPaymentInformation) PaymentInformationOptions {
	return func(pi *PaymentInformation) {
		pi.BankAccountPaymentInformation = *bankAccount
	}
}

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/BankAccountPaymentInformationBankAccountType.java
type BankAccountTypeEnum string

const (
	CHECKING BankAccountTypeEnum = "CHECKING"
	// SAVINGS  BankAccountTypeEnum = "SAVINGS"
)

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/BankAccountPaymentInformation.java
type BankAccountPaymentInformation struct {
	BankRoutingNumber     string              `json:"bankRoutingNumber,omitempty"`
	BankAccountType       BankAccountTypeEnum `json:"bankAccountType,omitempty"`
	BankAccountNumner     string              `json:"bankAccountNumber,omitempty"`
	BankSwiftCode         string              `json:"bankSwiftCode,omitempty"`
	BankRoutingCheckDigit string              `json:"bankRoutingCheckDigit,omitempty"`
	AccountBankName       string              `json:"accountBankName,omitempty"`
	AccountHolderName     string              `json:"accountHolderName,omitempty"`

	// PlaidProcessorToken string `json:"plaidProcessorToken,omitempty"` // not currently supported by us, but may be added in the future
}

func (bapi BankAccountPaymentInformation) MarshalJSON() ([]byte, error) {
	type alias BankAccountPaymentInformation
	return json.Marshal(alias(bapi))
}

func (bapi *BankAccountPaymentInformation) UnmarshalJSON(data []byte) error {
	type alias BankAccountPaymentInformation
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*bapi = BankAccountPaymentInformation(a)
	return nil
}

type BankAccountPaymentInformationOptions func(*BankAccountPaymentInformation)

func WithBankAccountPaymentInformationBankAccountNumber(number string) BankAccountPaymentInformationOptions {
	return func(bapi *BankAccountPaymentInformation) {
		bapi.BankAccountNumner = number
	}
}

func WithBankAccountPaymentInformationBankAccountType(t BankAccountTypeEnum) BankAccountPaymentInformationOptions {
	return func(bapi *BankAccountPaymentInformation) {
		bapi.BankAccountType = t
	}
}

func WithBankAccountPaymentInformationBankSwiftCode(code string) BankAccountPaymentInformationOptions {
	return func(bapi *BankAccountPaymentInformation) {
		bapi.BankSwiftCode = code
	}
}

func WithBankAccountPaymentInformationBankRoutingNumber(number string) BankAccountPaymentInformationOptions {
	return func(bapi *BankAccountPaymentInformation) {
		bapi.BankRoutingNumber = number
	}
}

func WithBankAccountPaymentInformationBankRoutingCheckDigit(digit string) BankAccountPaymentInformationOptions {
	return func(bapi *BankAccountPaymentInformation) {
		bapi.BankRoutingCheckDigit = digit
	}
}

func WithBankAccountPaymentInformationAccountBankName(name string) BankAccountPaymentInformationOptions {
	return func(bapi *BankAccountPaymentInformation) {
		bapi.AccountBankName = name
	}
}
