package payments

import "gitlab.com/fynbos/backend/currency"

type CreateArgs struct {
	Sender          Identity
	Receiver        Identity
	SenderAmount    currency.Amount
	SenderAccount   string
	ReceiverAmount  currency.Amount
	ReceiverAccount string
}

type Payment struct {
	ID              string
	PublicID        string
	State           State
	Sender          Identity
	Receiver        Identity
	SenderAmount    currency.Amount
	ReceiverAmount  currency.Amount
	SenderAccount   string
	ReceiverAccount string
	RequiredActions []RequiredActionType
}

//go:generate stringer -type=RequiredActionType -trimprefix=RequiredActionType
type RequiredActionType int

const (
	RequiredActionTypeUnknown RequiredActionType = 0
	RequiredActionTypeThreeDS RequiredActionType = 1
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
	Type       IdentityType
	Identifier string
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
