package ops

import (
	"gitlab.com/fynbos/backend/user"
)

type Backends interface {
	Users() user.Client
}
