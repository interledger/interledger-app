package jobs

import (
	"context"
	"time"
)

// Job represents delayed asynchronous work.
type Job struct {
	Delay time.Duration
	Run   func(context.Context) error
}
