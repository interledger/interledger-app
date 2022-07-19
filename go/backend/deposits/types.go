package deposits

import "fmt"

type Deposit struct {
	ID              string
	AccountID       string `db:"account_id"`
	FundingSourceId string `db:"funding_source_id"`
	Amount          uint64
	State           State
	CreatedAt       string `db:"created_at"`
	UpdatedAt       string `db:"updated_at"`
}

type InitiateDepositArgs struct {
	IdentityID      string `validate:"required,uuid4"`
	AccountID       string `validate:"required,uuid4"`
	FundingSourceID string `validate:"required,uuid4"`
	Amount          uint64 `validate:"required,gt=0"`
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
