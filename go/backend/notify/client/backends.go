package client

import (
	"github.com/pusher/pusher-http-go/v5"
)

type Backends interface {
}

type opsBackends struct {
	Backends
	pusher *pusher.Client
}

func (o *opsBackends) Pusher() *pusher.Client {
	return o.pusher
}
