package client

import (
	segment "github.com/segmentio/analytics-go/v3"
)

type Backends interface {
}

type opsBackends struct {
	Backends
	segment segment.Client
}

func (o *opsBackends) Segment() segment.Client {
	return o.segment
}
