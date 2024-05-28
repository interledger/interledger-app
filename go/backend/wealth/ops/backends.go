package ops

import (
	"gitlab.com/fynbos/backend/vault"
	temporal "go.temporal.io/sdk/client"
)

type Backends interface {
	Vault() vault.Client
	Temporal() temporal.Client
}
