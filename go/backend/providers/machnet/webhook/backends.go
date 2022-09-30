package webhook

import (
	"gitlab.com/fynbos/backend/providers/machnet"
)

type Backends interface {
	Machnet() machnet.Client
}
