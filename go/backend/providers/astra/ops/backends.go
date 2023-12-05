package ops

import (
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/user"
)

type Backends interface {
	KYC() kyc.Client
	Twilio() twilio.Service
	Users() user.Client
}
