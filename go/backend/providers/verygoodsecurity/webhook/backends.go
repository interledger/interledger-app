package webhook

import (
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/providers/verygoodsecurity"
)

type Backends interface {
	VGS() verygoodsecurity.Client
	Tabapay() tabapay.Client
}
