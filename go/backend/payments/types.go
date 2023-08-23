package payments

import (
	"time"

	"gitlab.com/fynbos/backend/currency"
)

type CreateArgs struct {
	Sender          Identity
	Receiver        Identity
	SenderAmount    currency.Amount
	SenderAccount   string `validate:"omitempty,uuid"`
	ReceiverAmount  currency.Amount
	ReceiverAccount string `validate:"omitempty,uuid"`
	Note            string
	RequiresOTP     bool
	IPAddress       string `validate:"omitempty,ip_addr"`
}

type UpdateArgs struct {
	ID              string
	Receiver        Identity
	ReceiverAmount  currency.Amount
	ReceiverAccount string
	SenderAccount   string
	SenderAmount    currency.Amount
	ThreeDSID       string
	OTP             string
	Note            string
	IPAddress       string
}

type Payment struct {
	ID                   string
	PublicID             string
	State                State
	Sender               Identity
	Receiver             Identity
	SenderAmount         currency.Amount
	ReceiverAmount       currency.Amount
	SenderAccount        string
	ReceiverAccount      string
	SendTransactionID    string
	ReceiveTransactionID string
	RequiredActions      []RequiredActionType
	Note                 string
	IPAddress            string
	UpdatedAt            time.Time
}

//go:generate stringer -type=RequiredActionType -trimprefix=RequiredActionType
type RequiredActionType int

const (
	RequiredActionTypeUnknown            RequiredActionType = 0
	RequiredActionTypeThreeDS            RequiredActionType = 1
	RequiredActionTypeSenderIdentifier   RequiredActionType = 2
	RequiredActionTypeSenderAccount      RequiredActionType = 3
	RequiredActionTypeReceiverIdentifier RequiredActionType = 4
	RequiredActionTypeSenderAmount       RequiredActionType = 5
	RequiredActionTypeReceiverAmount     RequiredActionType = 6
	RequiredActionTypeOTP                RequiredActionType = 7
	RequiredActionTypeIPAddress          RequiredActionType = 8
)

//go:generate stringer -type=IdentityType -trimprefix=IdentityType
type IdentityType int

const (
	IdentityTypeUnknown   IdentityType = 0
	IdentityTypeTwitter   IdentityType = 1
	IdentityTypeWalletID  IdentityType = 2
	IdentityTypeWalletURL IdentityType = 3
	identityTypeSentinel  IdentityType = 4 // End of range value must be last, no need to public
)

func (i IdentityType) Valid() bool {
	return i > IdentityTypeUnknown && i < identityTypeSentinel
}

type Identity struct {
	Type       IdentityType `validate:"omitempty,gt=0,lt=4"`
	Identifier string
	WalletID   string
}

func (i *Identity) IsEmpty() bool {
	return i.Identifier == "" && i.Type == IdentityTypeUnknown
}

func (i *Identity) IsEqual(id Identity) bool {
	return i.Identifier == id.Identifier && i.Type == id.Type
}

//go:generate stringer -type=State -trimprefix=State
type State int

const (
	StateUnknown    State = 0
	StateCreated    State = 1
	StateConfirmed  State = 2
	StateProcessing State = 3
	StateCompleted  State = 4
	StateFailed     State = 5
	stateSentinel   State = 6 // End of range value must be last, no need to public
)

func (s State) Valid() bool {
	return s > StateUnknown && s < stateSentinel
}

func (s State) CanTransitionTo(state State) bool {
	transitions := validTransitions[s]
	for _, s := range transitions {
		if s == state {
			return true
		}
	}

	return false
}

var validTransitions = map[State][]State{
	StateUnknown:    {StateCreated},
	StateCreated:    {StateConfirmed},
	StateConfirmed:  {StateProcessing},
	StateProcessing: {StateFailed, StateCompleted},
	StateCompleted:  {},
	StateFailed:     {StateProcessing, StateCompleted},
}
