package webhook

import (
	"gitlab.com/fynbos/backend/providers/verygoodsecurity"
)

type Backends interface {
	VGS() verygoodsecurity.Client
}
