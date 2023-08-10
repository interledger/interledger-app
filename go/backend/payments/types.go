package payments

import "gitlab.com/fynbos/backend/currency"

type Payment struct {
	ID              string
	PublicID        string
	State           State
	Sender          Identity
	Receiver        Identity
	SenderAmount    currency.Amount
	ReceiverAmount  currency.Amount
	RequiredActions []RequiredAction
}

//go:generate stringer -type=RequiredAction -trimprefix=RequiredAction
type RequiredAction int

const (
	RequiredActionUnknown     RequiredAction = 0
	RequiredActionThreeDS     RequiredAction = 1
	RequiredActionInformation RequiredAction = 2
	requiredActionSentinel    RequiredAction = 3 // End of range value must be last, no need to public
)

func (i RequiredAction) Valid() bool {
	return i > RequiredActionUnknown && i < requiredActionSentinel
}

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
	stateSentinel   State = 4 // End of range value must be last, no need to public
)

func (s State) Valid() bool {
	return s > StateUnknown && s < stateSentinel
}
