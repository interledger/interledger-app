package ops

import (
	segment "github.com/segmentio/analytics-go/v3"
)

type Backends interface {
	Segment() segment.Client
}
