package ops

import (
	"github.com/pusher/pusher-http-go/v5"
)

type Backends interface {
	Pusher() *pusher.Client
}
