package workflows

import (
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/linkedin"
	"gitlab.com/fynbos/backend/openpayments"
)

type Backends interface {
	Linkedin() linkedin.Client
	Identities() identities.Client
	OpenPayments() openpayments.Client
}
