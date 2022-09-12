package webhooks

import (
	"gitlab.com/fynbos/backend/providers/unit"
	"go.temporal.io/sdk/client"
)

type Backends interface {
	Unit() unit.Client
	Temporal() client.Client
}

var _ Backends = backends{}

type backends struct {
	temporal client.Client
	unit     unit.Client
}

func (b backends) Temporal() client.Client {
	return b.temporal
}

func (b backends) Unit() unit.Client {
	return b.unit
}
