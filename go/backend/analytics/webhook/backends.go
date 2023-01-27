package webhook

import (
	"gitlab.com/fynbos/backend/analytics"
)

type Backends interface {
	Analytics() analytics.Client
}
