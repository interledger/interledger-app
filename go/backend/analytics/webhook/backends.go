package webhook

import (
	"gitlab.com/fynbos/backend/analytics"
	"gitlab.com/fynbos/backend/user"
)

type Backends interface {
	Users() user.Client
	Analytics() analytics.Client
}
