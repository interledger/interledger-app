package dto

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/TransactionTypeEnum.java
type TransactionTypeEnum string

func (tt TransactionTypeEnum) String() string {
	return string(tt)
}

const (
	DEPOSIT    TransactionTypeEnum = "DEPOSIT"
	WITHDRAWAL TransactionTypeEnum = "WITHDRAWAL"

	// not currently supported by us, but may be added in the future
	// PAYMENT  TransactionTypeEnum = "PAYMENT"
	// TRANSFER TransactionTypeEnum = "TRANSFER"
	// SELL     TransactionTypeEnum = "SELL"
	// BUY      TransactionTypeEnum = "BUY"
	// MINT     TransactionTypeEnum = "MINT"
	// TRADE    TransactionTypeEnum = "TRADE"
)
