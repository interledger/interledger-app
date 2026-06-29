package rafiki

import "github.com/interledger/interledger-app/go/backend/currency"

const (
	Provider          = "rafiki"
	ZARBalanceAccount = "905e2c9b-a8b7-4cf2-9449-197e7029052e"
)

type Grant struct {
	Id                 string
	Client             string
	State              string
	FinalizationReason string
	CreatedAt          string
	Access             []Access
}

type Access struct {
	ID         string
	Identifier string
	Type       string
	Actions    []string
	Limits     Limits
}

type Limits struct {
	Receiver      string
	Interval      string
	DebitAmount   currency.Amount
	ReceiveAmount currency.Amount
}

type WalletAddress struct {
	ID         string
	AssetCode  string
	AssetScale uint8
	URL        string
}
type UpdateAddressStatus struct {
	ID   string `db:"payment_pointer_id"`
	Name string `db:"name"`
}

type IncomingPaymentState string

const (
	IncomingPaymentStatePending    IncomingPaymentState = "PENDING"
	IncomingPaymentStateProcessing IncomingPaymentState = "PROCESSING"
	IncomingPaymentStateCompleted  IncomingPaymentState = "COMPLETED"
	IncomingPaymentStateExpired    IncomingPaymentState = "EXPIRED"
)

type IncomingPayment struct {
	ID              string
	WalletAddressID string
	State           IncomingPaymentState
	ExpiresAt       string
	CreatedAt       string
}
