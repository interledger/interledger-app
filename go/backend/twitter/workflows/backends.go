package workflows

import (
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/twitter"
)

type Backends interface {
	Twitter() twitter.Client
	Identities() identities.Client
	OpenPayments() openpayments.Client
}
