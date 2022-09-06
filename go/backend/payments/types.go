package payments

import "fmt"

type OutgoingPayment struct {
	ID          string
	AccountID   string `db:"account_id"`
	Destination string `db:"destination"`
	Amount      uint64
	State       State
	CreatedAt   string `db:"created_at"`
	UpdatedAt   string `db:"updated_at"`
}

const (
	Verified  string = "verified"
	Retry     string = "retry"
	Kba       string = "kba"
	Document  string = "document"
	Suspended string = "suspended"
)

type InitiateOutgoingPaymentArgs struct {
	UserID string `validate:"required"`
	Amount uint64 `validate:"gt=0"`
	To     string `validate:"required"`
	OTP    string `validate:"required"`
}

type State string

const (
	Created    = State("CREATED")
	Processing = State("PROCESSING")
	Complete   = State("COMPLETE")
	Failed     = State("FAILED")
)

func (s State) String() string {
	return string(s)
}

func (s State) IsValid() bool {
	switch s {
	case Created, Processing, Complete, Failed:
		return true
	}
	return false
}

func (s *State) Unmarshall(v string) error {
	state := State(v)
	if !state.IsValid() {
		return fmt.Errorf("%s is not a valid State", v)
	}
	*s = state
	return nil
}
